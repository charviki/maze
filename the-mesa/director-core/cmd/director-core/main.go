package main

import (
	"context"
	"time"

	cradleDB "github.com/charviki/maze/fabrication/cradle/db"
	"github.com/charviki/maze/fabrication/cradle/lifecycle"
	"github.com/charviki/maze/fabrication/cradle/logutil"
	"github.com/charviki/maze/the-mesa/director-core/internal/authz"
	"github.com/charviki/maze/the-mesa/director-core/internal/config"
	"github.com/charviki/maze/the-mesa/director-core/internal/transport"
)

func main() {
	logger := logutil.New("director-core")

	cfg, err := config.LoadFromExe()
	if err != nil {
		logger.Fatalf("load config: %v", err)
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

	authzResult, err := authz.Init(cfg, logger)
	if err != nil {
		logger.Fatalf("init authz: %v", err)
	}
	if authzResult != nil {
		defer authzResult.Cleanup()
	}

	httpSrv, grpcSrv, resources, err := transport.NewGRPCGatewayServer(transport.ServerParams{
		Config:      cfg,
		Logger:      logger,
		HostPool:    hostPool,
		AuthzResult: authzResult,
	})
	if err != nil {
		logger.Fatalf("create server: %v", err)
	}
	defer resources.Cleanup()

	manager := &lifecycle.Manager{
		ShutdownTimeout: 5 * time.Second,
		Logger:          logger,
	}
	manager.Add(httpSrv)
	manager.Add(grpcSrv)

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
