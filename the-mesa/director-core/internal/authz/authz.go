package authz

import (
	"context"
	"errors"
	"io/fs"

	"github.com/charviki/maze/fabrication/cradle/auth"
	casbincasbin "github.com/charviki/maze/fabrication/cradle/auth/casbin"
	cradleDB "github.com/charviki/maze/fabrication/cradle/db"
	"github.com/charviki/maze/fabrication/cradle/logutil"
	"github.com/charviki/maze/the-mesa/director-core/internal/config"
	"github.com/charviki/maze/the-mesa/director-core/internal/repository/postgres"
	appservice "github.com/charviki/maze/the-mesa/director-core/internal/service"
	"github.com/charviki/maze/the-mesa/director-core/internal/transport"
)

// Init 初始化 Casbin + 权限服务 + JWT 认证 + janitor。
// cfg.Authz.Enabled 为 false 时返回 nil, nil。
func Init(cfg *config.Config, logger logutil.Logger) (*transport.AuthzResult, error) {
	if !cfg.Authz.Enabled {
		return nil, nil
	}

	ctx := context.Background()
	if cfg.AuthDatabase.Host == "" {
		return nil, errors.New("authz.enabled requires database.host")
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
		return nil, err
	}

	migrationsFS, fsErr := fs.Sub(postgres.AuthMigrationsFS, "migrations")
	if fsErr != nil {
		pool.Close()
		return nil, fsErr
	}
	if err := cradleDB.RunMigrations(pool, migrationsFS); err != nil {
		pool.Close()
		return nil, err
	}
	logger.Infof("[authz] database migrations completed")

	permRepo := postgres.NewPermissionRepository(pool)
	if err := permRepo.EnsureAdminBootstrap(ctx, cfg.Authz.AdminSubjectKey, cfg.Authz.AdminDisplayName); err != nil {
		pool.Close()
		return nil, err
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
		return nil, err
	}
	if err := enforcer.LoadPolicy(); err != nil {
		pool.Close()
		return nil, err
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

	return &transport.AuthzResult{
		PermHandler: permissionHandler,
		AuthHandler: authHandler,
		Interceptor: casbinInterceptor,
		Cleanup:     cleanup,
		CredStore:   credentialRepo,
	}, nil
}
