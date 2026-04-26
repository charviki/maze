package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/biz/config"
	"github.com/charviki/sweetwater-black-ridge/biz/service"
)

func main() {
	logger := logutil.New("agent")

	// 用 slog 替换 Hertz 框架默认 logger，统一 JSON 输出
	hlog.SetLogger(logger)

	cfg, err := config.LoadFromExe()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	h := server.Default(server.WithHostPorts(cfg.Server.ListenAddr))

	tmuxService := service.NewTmuxService(&cfg.Tmux, cfg.Workspace.StateDir, logger, &service.ClaudeTrustBootstrapper{})

	localConfig := service.NewLocalConfigStore(cfg.Workspace.RootDir, logger)

	register(h, cfg, tmuxService, logger)

	stopCh := make(chan struct{})
	heartbeatService := service.NewHeartbeatService(cfg, tmuxService, localConfig, logger)
	go heartbeatService.Start(stopCh)

	autoSaveService := service.NewAutoSaveService(tmuxService, cfg.AutoSave.Interval, logger)
	go autoSaveService.Start(stopCh)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Infof("shutting down...")
		close(stopCh)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.Shutdown(ctx); err != nil {
			logger.Errorf("shutdown error: %v", err)
		}
	}()

	logger.Infof("agent server started on %s", cfg.Server.ListenAddr)
	if cfg.Server.AuthToken == "" {
		logger.Warnf("[security] running in DEV mode: server.auth_token is empty, all API endpoints are open")
	}
	if cfg.Controller.Enabled && cfg.Controller.AuthToken == "" {
		logger.Warnf("[security] running in DEV mode: controller.auth_token is empty, heartbeat/register requests have no auth")
	}
	h.Spin()
}
