package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/charviki/maze/fabrication/cradle/auth"
	casbincasbin "github.com/charviki/maze/fabrication/cradle/auth/casbin"
	cradleDB "github.com/charviki/maze/fabrication/cradle/db"
	"github.com/charviki/maze/fabrication/cradle/grpcutil"
	"github.com/jackc/pgx/v5/pgxpool"
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

	hostPool, err := newHostPoolWithRetry(context.Background(), cfg, logger)
	if err != nil {
		logger.Fatalf("connect host database: %v", err)
	}

	gwmux := gatewayutil.NewServeMux()
	httpServer, resources := newHTTPServer(cfg, logger, gwmux, hostPool)
	defer cleanupResources(resources)

	proxySvc := agentclient.NewProxy(resources.Registry, resources.ConnMgr, cfg.Server.JWTSecret)

	// 构建 gRPC interceptor chain
	interceptors := []grpc.UnaryServerInterceptor{
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

	pool, err := newAuthzPoolWithRetry(ctx, cfg, logger)
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

	resourceMap := map[string]casbincasbin.ResourceAction{
		pb.HostService_CreateHost_FullMethodName:                        {Resource: "host/*", Action: "create"},
		pb.HostService_ListHosts_FullMethodName:                         {Resource: "host/*", Action: "read"},
		pb.HostService_GetHost_FullMethodName:                           {Resource: "host/*", Action: "read"},
		pb.HostService_DeleteHost_FullMethodName:                        {Resource: "host/*", Action: "delete"},
		pb.HostService_GetBuildLog_FullMethodName:                       {Resource: "host/*/logs/build", Action: "read"},
		pb.HostService_GetRuntimeLog_FullMethodName:                     {Resource: "host/*/logs/runtime", Action: "read"},
		pb.HostService_ListTools_FullMethodName:                         {Resource: "host/tools", Action: "read"},
		pb.NodeService_ListNodes_FullMethodName:                         {Resource: "node/*", Action: "read"},
		pb.NodeService_GetNode_FullMethodName:                           {Resource: "node/*", Action: "read"},
		pb.NodeService_DeleteNode_FullMethodName:                        {Resource: "node/*", Action: "delete"},
		pb.AuditService_GetAuditLogs_FullMethodName:                     {Resource: "audit/*", Action: "read"},
		pb.SessionService_ListSessions_FullMethodName:                   {Resource: "session/*", Action: "read"},
		pb.SessionService_CreateSession_FullMethodName:                  {Resource: "session/*", Action: "create"},
		pb.SessionService_GetSession_FullMethodName:                     {Resource: "session/*", Action: "read"},
		pb.SessionService_DeleteSession_FullMethodName:                  {Resource: "session/*", Action: "delete"},
		pb.SessionService_GetSessionConfig_FullMethodName:               {Resource: "session/*/config", Action: "read"},
		pb.SessionService_UpdateSessionConfig_FullMethodName:            {Resource: "session/*/config", Action: "update"},
		pb.SessionService_RestoreSession_FullMethodName:                 {Resource: "session/*", Action: "restore"},
		pb.SessionService_SaveSessions_FullMethodName:                   {Resource: "session/*", Action: "save"},
		pb.SessionService_GetSavedSessions_FullMethodName:               {Resource: "session/*", Action: "read"},
		pb.SessionService_GetOutput_FullMethodName:                      {Resource: "session/*/terminal", Action: "read"},
		pb.SessionService_SendInput_FullMethodName:                      {Resource: "session/*/terminal", Action: "write"},
		pb.SessionService_SendSignal_FullMethodName:                     {Resource: "session/*/terminal", Action: "signal"},
		pb.SessionService_GetEnv_FullMethodName:                         {Resource: "session/*/env", Action: "read"},
		pb.TemplateService_ListTemplates_FullMethodName:                 {Resource: "template/*", Action: "read"},
		pb.TemplateService_CreateTemplate_FullMethodName:                {Resource: "template/*", Action: "create"},
		pb.TemplateService_GetTemplate_FullMethodName:                   {Resource: "template/*", Action: "read"},
		pb.TemplateService_UpdateTemplate_FullMethodName:                {Resource: "template/*", Action: "update"},
		pb.TemplateService_DeleteTemplate_FullMethodName:                {Resource: "template/*", Action: "delete"},
		pb.TemplateService_GetTemplateConfig_FullMethodName:             {Resource: "template/*/config", Action: "read"},
		pb.TemplateService_UpdateTemplateConfig_FullMethodName:          {Resource: "template/*/config", Action: "update"},
		pb.ConfigService_GetConfig_FullMethodName:                       {Resource: "config/*", Action: "read"},
		pb.ConfigService_UpdateConfig_FullMethodName:                    {Resource: "config/*", Action: "update"},
		pb.PermissionService_CreatePermissionApplication_FullMethodName: {Resource: "permission-application/*", Action: "create"},
		pb.PermissionService_ListPermissionApplications_FullMethodName:  {Resource: "permission-application/*", Action: "read"},
		pb.PermissionService_GetPermissionApplication_FullMethodName:    {Resource: "permission-application/*", Action: "read"},
		pb.PermissionService_ReviewPermissionApplication_FullMethodName: {Resource: "permission-application/*", Action: "review"},
		pb.PermissionService_RevokePermissionApplication_FullMethodName: {Resource: "permission-application/*", Action: "revoke"},
		pb.PermissionService_ListSubjectPermissions_FullMethodName:      {Resource: "subject/*/permissions", Action: "read"},
	}

	casbinInterceptor := casbincasbin.NewUnaryInterceptor(enforcer, extractUser, resourceMap)

	return permissionHandler, authHandler, casbinInterceptor, cleanup, credentialRepo, nil
}

func newHostPoolWithRetry(ctx context.Context, cfg *config.Config, logger logutil.Logger) (*pgxpool.Pool, error) {
	poolCfg := cradleDB.PoolConfig{
		Host:     cfg.HostDatabase.Host,
		Port:     cfg.HostDatabase.Port,
		Name:     cfg.HostDatabase.Name,
		User:     cfg.HostDatabase.User,
		Password: cfg.HostDatabase.Password,
	}

	const maxAttempts = 30
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pool, err := cradleDB.NewPool(ctx, poolCfg)
		if err == nil {
			if attempt > 1 {
				logger.Infof("[host] database became ready after %d attempts", attempt)
			}
			return pool, nil
		}
		lastErr = err
		if attempt == maxAttempts {
			break
		}

		logger.Warnf("[host] database not ready (attempt %d/%d): %v", attempt, maxAttempts, err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return nil, fmt.Errorf("host database not ready after %d attempts: %w", maxAttempts, lastErr)
}

func newAuthzPoolWithRetry(ctx context.Context, cfg *config.Config, logger logutil.Logger) (*pgxpool.Pool, error) {
	poolCfg := cradleDB.PoolConfig{
		Host:     cfg.AuthDatabase.Host,
		Port:     cfg.AuthDatabase.Port,
		Name:     cfg.AuthDatabase.Name,
		User:     cfg.AuthDatabase.User,
		Password: cfg.AuthDatabase.Password,
	}

	const maxAttempts = 30
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pool, err := cradleDB.NewPool(ctx, poolCfg)
		if err == nil {
			if attempt > 1 {
				logger.Infof("[authz] database became ready after %d attempts", attempt)
			}
			return pool, nil
		}
		lastErr = err
		if attempt == maxAttempts {
			break
		}

		logger.Warnf("[authz] database not ready (attempt %d/%d): %v", attempt, maxAttempts, err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return nil, lastErr
}
