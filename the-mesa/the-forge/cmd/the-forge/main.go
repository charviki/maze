package main

import (
	"context"
	"io/fs"
	"time"

	cradleDB "github.com/charviki/maze/fabrication/cradle/db"
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

	pool, err := cradleDB.NewPoolWithRetry(context.Background(), cradleDB.PoolConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		Name:     cfg.Database.Name,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
	}, 30, logger)
	if err != nil {
		logger.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	migrationsFS, fsErr := fs.Sub(postgres.MigrationsFS, "migrations")
	if fsErr != nil {
		logger.Fatalf("create migrations sub FS: %v", fsErr)
	}
	if err := cradleDB.RunMigrations(pool, migrationsFS); err != nil {
		logger.Fatalf("run migrations: %v", err)
	}
	logger.Info("database migrations applied")

	txm := postgres.NewTxManager(pool)
	docSvc := service.NewDocService(postgres.NewDocRepository(pool, txm, logger))

	httpSrv, grpcSrv, err := transport.NewGRPCGatewayServer(transport.ServerParams{
		Config: cfg,
		Logger: logger,
		DocSvc: docSvc,
	})
	if err != nil {
		logger.Fatalf("create server: %v", err)
	}

	manager := &lifecycle.Manager{
		ShutdownTimeout: 5 * time.Second,
		Logger:          logger,
	}
	manager.Add(httpSrv)
	manager.Add(grpcSrv)

	logger.Infof("the-forge started on %s (grpc %s)", cfg.Server.ListenAddr, cfg.Server.GRPCAddr)

	if err := manager.Run(context.Background()); err != nil {
		logger.Fatalf("run lifecycle: %v", err)
	}
}
