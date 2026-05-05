package main

import (
	"context"
	"strings"
	"time"

	"github.com/charviki/maze-cradle/grpcutil"
	"google.golang.org/grpc"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/gatewayutil"
	"github.com/charviki/maze-cradle/lifecycle"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/internal/config"
	"github.com/charviki/sweetwater-black-ridge/internal/service"
	"github.com/charviki/sweetwater-black-ridge/internal/transport"
)

const defaultGRPCListenAddr = ":9090"

func grpcListenAddrFor(grpcAddr string) string {
	if grpcAddr == "" {
		return defaultGRPCListenAddr
	}
	if idx := strings.LastIndex(grpcAddr, ":"); idx >= 0 {
		return ":" + grpcAddr[idx+1:]
	}
	return grpcAddr
}

func main() {
	logger := logutil.New("agent")

	cfg, err := config.LoadFromExe()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	tmuxService := service.NewTmuxService(&cfg.Tmux, cfg.Workspace.StateDir, logger, &service.ClaudeTrustBootstrapper{})
	localConfig := service.NewLocalConfigStore(cfg.Workspace.RootDir, logger)
	gwmux := gatewayutil.NewServeMux()
	httpServer, templateStore := newHTTPServer(cfg, tmuxService, logger, gwmux)

	grpcAddr := cfg.Server.GRPCAddr
	grpcListenAddr := grpcListenAddrFor(grpcAddr)

	configFileService := service.NewConfigFileService()
	grpcServer := transport.NewServer(tmuxService, localConfig, templateStore, configFileService, cfg.Workspace.RootDir, logger)
	interceptors := []grpc.UnaryServerInterceptor{
		gatewayutil.UnaryAuthInterceptor(cfg.Server.AuthToken),
	}
	grpcCore := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
	grpcServer.RegisterGRPC(grpcCore)
	managedGRPC := grpcutil.NewManagedGRPCServer(grpcListenAddr, grpcCore, logger)

	ctx := context.Background()
	if err := pb.RegisterSessionServiceHandlerServer(ctx, gwmux, grpcServer); err != nil {
		logger.Fatalf("register session service to gateway: %v", err)
	}
	if err := pb.RegisterTemplateServiceHandlerServer(ctx, gwmux, grpcServer); err != nil {
		logger.Fatalf("register template service to gateway: %v", err)
	}
	if err := pb.RegisterConfigServiceHandlerServer(ctx, gwmux, grpcServer); err != nil {
		logger.Fatalf("register config service to gateway: %v", err)
	}

	heartbeatService, err := service.NewHeartbeatService(cfg, tmuxService, localConfig, logger)
	if err != nil {
		logger.Fatalf("create heartbeat service: %v", err)
	}
	autoSaveService := service.NewAutoSaveService(tmuxService, cfg.AutoSave.Interval, logger)
	heartbeatRunner := newBackgroundRunner("heartbeat", logger, heartbeatService.Start)
	autoSaveRunner := newBackgroundRunner("autosave", logger, autoSaveService.Start)

	manager := &lifecycle.Manager{
		ShutdownTimeout: 5 * time.Second,
		Logger:          logger,
	}
	manager.Add(httpServer)
	manager.Add(managedGRPC)
	manager.Add(heartbeatRunner)
	manager.Add(autoSaveRunner)

	logger.Infof("agent server started on %s", cfg.Server.ListenAddr)
	if cfg.Server.AuthToken == "" {
		logger.Warnf("[security] running in DEV mode: server.auth_token is empty, all API endpoints are open")
	}
	if cfg.Controller.Enabled && cfg.Controller.AuthToken == "" {
		logger.Warnf("[security] running in DEV mode: controller.auth_token is empty, heartbeat/register requests have no auth")
	}
	if err := manager.Run(context.Background()); err != nil {
		logger.Fatalf("run server lifecycle: %v", err)
	}
}
