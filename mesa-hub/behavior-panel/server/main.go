package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/mesa-hub-behavior-panel/biz/config"
	managergrpc "github.com/charviki/mesa-hub-behavior-panel/biz/grpc"
	"github.com/charviki/mesa-hub-behavior-panel/biz/service"
)

func main() {
	logger := logutil.New("manager")

	hlog.SetLogger(logger)

	cfg, err := config.LoadFromExe()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	h := server.Default(server.WithHostPorts(cfg.Server.ListenAddr))

	resources := register(h, cfg, logger)

	proxySvc := service.NewAgentProxyService(resources.Registry, logger)

	grpcServer := managergrpc.NewServer(
		resources.HostSvc,
		resources.NodeSvc,
		resources.AuditSvc,
		proxySvc,
		logger,
	)
	if err := grpcServer.Start(":9090"); err != nil {
		logger.Fatalf("start grpc server: %v", err)
	}

	gwmux := runtime.NewServeMux()
	ctx := context.Background()

	// 注册 6 个 Service gateway handler，进程内直连 gRPC Server
	if err := pb.RegisterHostServiceHandlerServer(ctx, gwmux, grpcServer); err != nil {
		logger.Fatalf("register HostService gateway: %v", err)
	}
	if err := pb.RegisterNodeServiceHandlerServer(ctx, gwmux, grpcServer); err != nil {
		logger.Fatalf("register NodeService gateway: %v", err)
	}
	if err := pb.RegisterAuditServiceHandlerServer(ctx, gwmux, grpcServer); err != nil {
		logger.Fatalf("register AuditService gateway: %v", err)
	}
	if err := pb.RegisterSessionServiceHandlerServer(ctx, gwmux, grpcServer); err != nil {
		logger.Fatalf("register SessionService gateway: %v", err)
	}
	if err := pb.RegisterTemplateServiceHandlerServer(ctx, gwmux, grpcServer); err != nil {
		logger.Fatalf("register TemplateService gateway: %v", err)
	}
	if err := pb.RegisterConfigServiceHandlerServer(ctx, gwmux, grpcServer); err != nil {
		logger.Fatalf("register ConfigService gateway: %v", err)
	}

	// AgentService 不注册（保持 HTTP）

	gatewaySrv := &http.Server{
		Addr:    ":8081",
		Handler: gwmux,
	}
	go func() {
		logger.Infof("[gateway] HTTP gateway started on :8081")
		if err := gatewaySrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("[gateway] error: %v", err)
		}
	}()

	logger.Infof("agent-manager controller started on %s", cfg.Server.ListenAddr)
	if cfg.IsDevMode() {
		logger.Warnf("[security] running in DEV mode: auth_token is empty, all API endpoints are open")
	}
	if len(cfg.AllowedOrigins()) == 0 {
		logger.Warnf("[security] running in DEV mode: CORS and WebSocket allow all origins")
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Infof("shutting down...")

		// 停止 gateway HTTP server
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := gatewaySrv.Shutdown(shutdownCtx); err != nil {
			logger.Errorf("[gateway] shutdown error: %v", err)
		}

		grpcServer.Stop()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.Shutdown(ctx); err != nil {
			logger.Errorf("shutdown error: %v", err)
		}
		resources.Reconciler.Stop()
		resources.Registry.WaitSave()
		resources.SpecMgr.WaitSave()
		resources.AuditLog.Close()
	}()

	h.Spin()
}
