package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"google.golang.org/grpc"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/gatewayutil"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/biz/config"
	agentgrpc "github.com/charviki/sweetwater-black-ridge/biz/grpc"
	"github.com/charviki/sweetwater-black-ridge/biz/service"
)

func main() {
	logger := logutil.New("agent")

	hlog.SetLogger(logger)

	cfg, err := config.LoadFromExe()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	// 禁用 Hertz trailing slash 自动重定向（301），
	// 避免 grpc-gateway 路径匹配失败。
	h := server.Default(
		server.WithHostPorts(cfg.Server.ListenAddr),
		server.WithRedirectTrailingSlash(false),
	)

	tmuxService := service.NewTmuxService(&cfg.Tmux, cfg.Workspace.StateDir, logger, &service.ClaudeTrustBootstrapper{})

	localConfig := service.NewLocalConfigStore(cfg.Workspace.RootDir, logger)

	// grpc-gateway ServeMux，使用自定义响应格式包装器（与旧 httputil.Success/Error 格式一致）
	gwmux := gatewayutil.NewServeMux()

	templateStore := register(h, cfg, tmuxService, logger, gwmux)

	grpcAddr := cfg.Server.GRPCAddr
	if grpcAddr == "" {
		grpcAddr = ":9090"
	}
	grpcListenAddr := grpcAddr
	if idx := strings.LastIndex(grpcAddr, ":"); idx >= 0 {
		grpcListenAddr = ":" + grpcAddr[idx+1:]
	}

	grpcServer := agentgrpc.NewServer(tmuxService, localConfig, templateStore, cfg.Workspace.RootDir, logger)

	interceptors := []grpc.UnaryServerInterceptor{
		gatewayutil.UnaryAuthInterceptor(cfg.Server.AuthToken),
	}

	if err := grpcServer.Start(grpcListenAddr, grpc.ChainUnaryInterceptor(interceptors...)); err != nil {
		logger.Fatalf("start grpc server: %v", err)
	}

	// 注册 3 个 Service 到 grpc-gateway（进程内直连，不走网络）
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
		grpcServer.Stop()
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
