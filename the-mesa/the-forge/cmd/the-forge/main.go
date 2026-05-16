package main

import (
	"context"
	"fmt"
	"io/fs"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	"github.com/charviki/maze/fabrication/cradle/db"
	"github.com/charviki/maze/fabrication/cradle/gatewayutil"
	"github.com/charviki/maze/fabrication/cradle/grpcutil"
	"github.com/charviki/maze/fabrication/cradle/lifecycle"
	"github.com/charviki/maze/fabrication/cradle/logutil"

	"github.com/charviki/maze/the-mesa/the-forge/internal/config"
	"github.com/charviki/maze/the-mesa/the-forge/internal/repository/postgres"
	"github.com/charviki/maze/the-mesa/the-forge/internal/service"
	"github.com/charviki/maze/the-mesa/the-forge/internal/transport"
)

func main() {
	logger := logutil.New("the-forge")

	cfg, err := config.LoadFromExe()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	// PostgreSQL 连接池（带重试）
	pool, err := newPoolWithRetry(context.Background(), cfg, logger)
	if err != nil {
		logger.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	// Migration
	migrationsFS, fsErr := fs.Sub(postgres.MigrationsFS, "migrations")
	if fsErr != nil {
		logger.Fatalf("create migrations sub FS: %v", fsErr)
	}
	if err := db.RunMigrations(pool, migrationsFS); err != nil {
		logger.Fatalf("run migrations: %v", err)
	}
	logger.Info("database migrations applied")

	// Repository
	txm := postgres.NewTxManager(pool)
	docRepo := postgres.NewDocRepository(pool, txm, logger)

	// Service
	docSvc := service.NewDocService(docRepo)

	// gRPC server
	validationInterceptor, err := gatewayutil.NewValidationInterceptor()
	if err != nil {
		logger.Fatalf("create validation interceptor: %v", err)
	}
	grpcCore := grpc.NewServer(grpc.ChainUnaryInterceptor(
		validationInterceptor,
		gatewayutil.UnaryAuthInterceptor(cfg.Server.JWTSecret),
	))
	docTransport := transport.NewServer(docSvc)
	docTransport.RegisterGRPC(grpcCore)
	managedGRPC := grpcutil.NewManagedGRPCServer(cfg.Server.GRPCAddr, grpcCore, logger)

	// grpc-gateway
	gwMux := gatewayutil.NewServeMux()
	if err := transport.RegisterGatewayHandlers(context.Background(), transport.GatewayRegistrationParams{
		GWMux:    gwMux,
		GRPCAddr: cfg.Server.GRPCAddr,
	}); err != nil {
		logger.Fatalf("register gateway: %v", err)
	}

	// HTTP server
	httpServer := transport.NewHTTPServer(transport.HTTPHandlerParams{
		Config: cfg,
		GWMux:  gwMux,
		Logger: logger,
	})

	// Lifecycle
	manager := &lifecycle.Manager{
		ShutdownTimeout: 5 * time.Second,
		Logger:          logger,
	}
	manager.Add(httpServer)
	manager.Add(managedGRPC)

	logger.Infof("the-forge started on %s (grpc %s)", cfg.Server.ListenAddr, cfg.Server.GRPCAddr)

	if err := manager.Run(context.Background()); err != nil {
		logger.Fatalf("run lifecycle: %v", err)
	}
}

// newPoolWithRetry 尝试连接数据库，最多重试 30 次（每次间隔 1 秒）。
func newPoolWithRetry(ctx context.Context, cfg *config.Config, logger logutil.Logger) (*pgxpool.Pool, error) {
	poolCfg := db.PoolConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		Name:     cfg.Database.Name,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
	}
	const maxAttempts = 30
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pool, err := db.NewPool(ctx, poolCfg)
		if err == nil {
			if attempt > 1 {
				logger.Infof("database became ready after %d attempts", attempt)
			}
			return pool, nil
		}
		lastErr = err
		if attempt == maxAttempts {
			break
		}
		logger.Warnf("database not ready (attempt %d/%d): %v", attempt, maxAttempts, err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return nil, fmt.Errorf("database not ready after %d attempts: %w", maxAttempts, lastErr)
}
