package main

import (
	"context"
	"errors"
	"io/fs"
	"time"

	"github.com/charviki/maze/fabrication/cradle/auth"
	casbincasbin "github.com/charviki/maze/fabrication/cradle/auth/casbin"
	cradleDB "github.com/charviki/maze/fabrication/cradle/db"
	"github.com/charviki/maze/fabrication/cradle/grpcutil"
	"google.golang.org/grpc"

	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	"github.com/charviki/maze/fabrication/cradle/gatewayutil"
	"github.com/charviki/maze/fabrication/cradle/lifecycle"
	"github.com/charviki/maze/fabrication/cradle/logutil"
	"github.com/charviki/maze/fabrication/cradle/protocol"
	"github.com/charviki/maze/the-mesa/director-core/internal/agentclient"
	"github.com/charviki/maze/the-mesa/director-core/internal/config"
	"github.com/charviki/maze/the-mesa/director-core/internal/repository/postgres"
	appservice "github.com/charviki/maze/the-mesa/director-core/internal/service"
	"github.com/charviki/maze/the-mesa/director-core/internal/transport"
)

type auditLoggerAdapter struct {
	inner appservice.AuditLogWriter
}

func (a *auditLoggerAdapter) Log(entry gatewayutil.AuditEntry) {
	_ = a.inner.Log(context.Background(), protocol.AuditLogEntry{
		Operator:   entry.Operator,
		TargetNode: entry.TargetNode,
		Action:     entry.Action,
		Result:     entry.Result,
		StatusCode: entry.StatusCode,
	})
}

func main() {
	logger := logutil.New("director-core")

	cfg, err := config.LoadFromExe()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	grpcAddr := cfg.Server.GRPCListenAddr
	if grpcAddr == "" {
		grpcAddr = ":9090"
	}

	hostPool, err := cradleDB.NewPoolWithRetry(context.Background(), cradleDB.PoolConfig{
		Host:     cfg.HostDatabase.Host,
		Port:     cfg.HostDatabase.Port,
		Name:     cfg.HostDatabase.Name,
		User:     cfg.HostDatabase.User,
		Password: cfg.HostDatabase.Password,
	}, 30, logger)
	if err != nil {
		logger.Fatalf("connect host database: %v", err)
	}

	gwmux := gatewayutil.NewServeMux()
	httpServer, resources := newHTTPServer(cfg, logger, gwmux, hostPool)
	defer cleanupResources(resources)

	proxySvc := agentclient.NewProxy(resources.Registry, resources.ConnMgr, cfg.Server.JWTSecret)

	// 构建 gRPC interceptor chain
	validationInterceptor, err := gatewayutil.NewValidationInterceptor()
	if err != nil {
		logger.Fatalf("create validation interceptor: %v", err)
	}
	interceptors := []grpc.UnaryServerInterceptor{
		validationInterceptor,
		gatewayutil.UnaryAuthInterceptor(cfg.Server.JWTSecret),
		gatewayutil.UnaryHostTokenInterceptor(cfg.Server.JWTSecret, resources.Registry),
	}

	var permissionHandler *transport.PermissionServiceServer
	var authHandler *transport.AuthHandler
	var janitorCleanup func()
	if cfg.Authz.Enabled {
		iHandler, aHandler, casbinInterceptor, cleanup, credStore, err := initAuthz(cfg, logger)
		if err != nil {
			logger.Fatalf("init authz: %v", err)
		}
		permissionHandler = iHandler
		authHandler = aHandler
		janitorCleanup = cleanup
		resources.HostSvc.SetCredentialStore(credStore)
		interceptors = append(interceptors, casbinInterceptor)
	}

	// 在 initAuthz 成功后立即 defer cleanup，确保正常退出和错误退出都能执行。
	if janitorCleanup != nil {
		defer janitorCleanup()
	}

	interceptors = append(interceptors,
		gatewayutil.UnaryAuditInterceptor(&auditLoggerAdapter{inner: resources.AuditLog}),
	)

	grpcServer := transport.NewServer(
		resources.HostSvc,
		resources.NodeSvc,
		resources.AuditSvc,
		resources.SkillSvc,
		resources.MCPSvc,
		resources.RuleSvc,
		resources.GitKeySvc,
		proxySvc,
		resources.Registry,
		cfg.Server.JWTSecret,
		logger,
	)
	grpcCore := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
	grpcServer.RegisterGRPC(grpcCore)

	if permissionHandler != nil {
		pb.RegisterPermissionServiceServer(grpcCore, permissionHandler)
	}

	if authHandler != nil {
		pb.RegisterAuthServiceServer(grpcCore, authHandler)
	}

	managedGRPC := grpcutil.NewManagedGRPCServer(grpcAddr, grpcCore, logger)

	// gateway 注册下沉到 transport 层
	ctx := context.Background()
	if err := transport.RegisterGatewayHandlers(ctx, transport.GatewayRegistrationParams{
		GWMux:       gwmux,
		GRPCAddr:    grpcAddr,
		GRPCServer:  grpcCore,
		PermHandler: permissionHandler,
		AuthHandler: authHandler,
	}); err != nil {
		logger.Fatalf("register gateway handlers: %v", err)
	}

	manager := &lifecycle.Manager{
		ShutdownTimeout: 5 * time.Second,
		Logger:          logger,
	}
	manager.Add(httpServer)
	manager.Add(managedGRPC)

	logger.Infof("director-core controller started on %s", cfg.Server.ListenAddr)
	if len(cfg.AllowedOrigins()) == 0 {
		logger.Warnf("[security] running in DEV mode: CORS and WebSocket allow all origins")
	}
	logger.Infof("[host] database=%s:%d/%s", cfg.HostDatabase.Host, cfg.HostDatabase.Port, cfg.HostDatabase.Name)
	if cfg.Authz.Enabled {
		logger.Infof("[authz] permission system enabled, database=%s:%d/%s", cfg.AuthDatabase.Host, cfg.AuthDatabase.Port, cfg.AuthDatabase.Name)
	}

	if err := manager.Run(context.Background()); err != nil {
		logger.Fatalf("run server lifecycle: %v", err)
	}
}

// initAuthz 初始化权限系统和 JWT 认证服务。
func initAuthz(cfg *config.Config, logger logutil.Logger) (*transport.PermissionServiceServer, *transport.AuthHandler, grpc.UnaryServerInterceptor, func(), *postgres.CredentialRepository, error) {
	ctx := context.Background()
	if cfg.AuthDatabase.Host == "" {
		return nil, nil, nil, nil, nil, errors.New("authz.enabled requires database.host")
	}
	logger.Infof("[authz] connecting to database %s:%d/%s", cfg.AuthDatabase.Host, cfg.AuthDatabase.Port, cfg.AuthDatabase.Name)

	pool, err := cradleDB.NewPoolWithRetry(ctx, cradleDB.PoolConfig{
		Host:     cfg.AuthDatabase.Host,
		Port:     cfg.AuthDatabase.Port,
		Name:     cfg.AuthDatabase.Name,
		User:     cfg.AuthDatabase.User,
		Password: cfg.AuthDatabase.Password,
	}, 30, logger)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	migrationsFS, fsErr := fs.Sub(postgres.AuthMigrationsFS, "migrations")
	if fsErr != nil {
		pool.Close()
		return nil, nil, nil, nil, nil, fsErr
	}
	if err := cradleDB.RunMigrations(pool, migrationsFS); err != nil {
		pool.Close()
		return nil, nil, nil, nil, nil, err
	}
	logger.Infof("[authz] database migrations completed")

	permRepo := postgres.NewPermissionRepository(pool)
	if err := permRepo.EnsureAdminBootstrap(ctx, cfg.Authz.AdminSubjectKey, cfg.Authz.AdminDisplayName); err != nil {
		pool.Close()
		return nil, nil, nil, nil, nil, err
	}

	loadFn := func() ([][]string, error) {
		rules, err := permRepo.ListAllCasbinRules(ctx)
		if err != nil {
			return nil, err
		}
		var result [][]string
		for _, r := range rules {
			row := make([]string, 1, 4)
			row[0] = r.Ptype
			row = append(row, r.V0, r.V1, r.V2)
			result = append(result, row)
		}
		return result, nil
	}

	adapter := casbincasbin.NewDBAdapter(loadFn, nil)
	enforcer, err := casbincasbin.NewEnforcer(adapter)
	if err != nil {
		pool.Close()
		return nil, nil, nil, nil, nil, err
	}
	if err := enforcer.LoadPolicy(); err != nil {
		pool.Close()
		return nil, nil, nil, nil, nil, err
	}

	permSvc := appservice.NewPermissionService(permRepo, permRepo, enforcer.LoadPolicy)

	ctx2, cancel := context.WithCancel(ctx)
	janitor := appservice.NewPermissionJanitor(permSvc)
	go janitor.Run(ctx2)
	cleanup := func() {
		cancel()
		pool.Close()
	}

	permissionHandler := transport.NewPermissionServiceServer(permSvc)

	userRepo := postgres.NewUserRepository(pool)
	credentialRepo := postgres.NewCredentialRepository(pool)
	authSvc := appservice.NewAuthService(
		userRepo,
		credentialRepo,
		permRepo,
		permRepo,
		cfg.JWT.Secret,
		cfg.JWT.AccessDuration(),
		cfg.JWT.RefreshDuration(),
		cfg.JWT.DefaultAdminUsername,
		cfg.JWT.DefaultAdminPassword,
		cfg.Authz.AdminSubjectKey,
	)

	// 在 migration 完成后、服务启动前执行 admin bootstrap，确保 JWT 登录可用。
	if created, err := authSvc.BootstrapAdmin(ctx); err != nil {
		logger.Warnf("[auth] bootstrap admin failed (non-fatal, login may not work): %v", err)
	} else if created {
		logger.Warnf("[auth] default admin created — please change the password immediately via API or environment variable MAZE_JWT_DEFAULT_ADMIN_PASSWORD")
	}

	authHandler := transport.NewAuthHandler(authSvc)

	extractUser := func(ctx context.Context) (*auth.UserInfo, error) {
		if user := auth.GetUserInfo(ctx); user != nil {
			return user, nil
		}
		if cfg.Server.JWTSecret == "" {
			return &auth.UserInfo{SubjectKey: cfg.Authz.AdminSubjectKey, DisplayName: cfg.Authz.AdminDisplayName}, nil
		}
		return nil, errors.New("unauthorized: no user info in context")
	}

		resourceMap := casbincasbin.DirectorCoreResourceMap()

	casbinInterceptor := casbincasbin.NewUnaryInterceptor(enforcer, extractUser, resourceMap)

	return permissionHandler, authHandler, casbinInterceptor, cleanup, credentialRepo, nil
}
