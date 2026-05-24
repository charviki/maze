package main

import (
	"context"
	"time"

	"github.com/charviki/maze/fabrication/cradle/lifecycle"
	"github.com/charviki/maze/fabrication/cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/internal/config"
	"github.com/charviki/sweetwater-black-ridge/internal/service"
	"github.com/charviki/sweetwater-black-ridge/internal/service/provider"
	"github.com/charviki/sweetwater-black-ridge/internal/transport"
	"github.com/charviki/sweetwater-black-ridge/internal/webstatic"
)

func main() {
	logger := logutil.New("agent")

	cfg, err := config.LoadFromExe()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	registry := provider.NewRegistry()
	registry.Register(&provider.ClaudeProvider{})
	registry.Register(&provider.CodexProvider{})
	registry.Register(&provider.BashProvider{})

	provider.RunEntrypointTasks(logger, registry, provider.ResolveHomeDir())

	tmuxService := service.NewTmuxService(&cfg.Tmux, cfg.Workspace.StateDir, logger, registry)
	localConfig := service.NewLocalConfigStore(cfg.Workspace.RootDir, logger)
	configFileService := service.NewConfigFileService()

	httpSrv, grpcSrv, _, err := transport.NewGRPCGatewayServer(transport.ServerParams{
		Config:            cfg,
		Logger:            logger,
		TmuxService:       tmuxService,
		LocalConfig:       localConfig,
		ConfigFileService: configFileService,
		WorkspaceRootDir:  cfg.Workspace.RootDir,
		StaticFiles:       webstatic.Files,
	})
	if err != nil {
		logger.Fatalf("create server: %v", err)
	}

	heartbeatService, err := service.NewHeartbeatService(cfg, tmuxService, localConfig, registry, logger)
	if err != nil {
		logger.Fatalf("create heartbeat service: %v", err)
	}
	autoSaveService := service.NewAutoSaveService(tmuxService, cfg.AutoSave.Interval, logger)
	heartbeatRunner := lifecycle.NewBackgroundRunner("heartbeat", logger, heartbeatService.Start)
	autoSaveRunner := lifecycle.NewBackgroundRunner("autosave", logger, autoSaveService.Start)

	// 注册成功后拉取 Host 配置
	hostConfigService := service.NewHostConfigService(heartbeatService.GetAgentClient(), cfg.Server.Name, cfg.Controller.AuthToken, logger)
	hostConfigCtx, hostConfigCancel := context.WithTimeout(context.Background(), 30*time.Second)
	go func() {
		defer hostConfigCancel()
		<-heartbeatService.RegisteredCh()
		if err := hostConfigService.FetchAndApply(hostConfigCtx); err != nil {
			logger.Warnf("[host-config] fetch and apply failed: %v", err)
		}
	}()

	manager := &lifecycle.Manager{
		ShutdownTimeout: 5 * time.Second,
		Logger:          logger,
	}
	manager.Add(httpSrv)
	manager.Add(grpcSrv)
	manager.Add(heartbeatRunner)
	manager.Add(autoSaveRunner)

	logger.Infof("agent server started on %s", cfg.Server.ListenAddr)
	if err := manager.Run(context.Background()); err != nil {
		logger.Fatalf("run server lifecycle: %v", err)
	}
}
