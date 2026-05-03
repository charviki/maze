package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"google.golang.org/grpc"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/gatewayutil"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/biz/config"
	managergrpc "github.com/charviki/mesa-hub-behavior-panel/biz/grpc"
	"github.com/charviki/mesa-hub-behavior-panel/biz/handler"
	"github.com/charviki/mesa-hub-behavior-panel/biz/service"
)

// auditLoggerAdapter 将 handler.AuditLogger 适配为 gatewayutil.AuditLogger 接口。
// gatewayutil interceptor 产生 gatewayutil.AuditEntry，需转换为 handler 层使用的 protocol.AuditLogEntry。
type auditLoggerAdapter struct {
	inner *handler.AuditLogger
}

func (a *auditLoggerAdapter) Log(entry gatewayutil.AuditEntry) {
	a.inner.Log(protocol.AuditLogEntry{
		Operator:   entry.Operator,
		TargetNode: entry.TargetNode,
		Action:     entry.Action,
		Result:     entry.Result,
		StatusCode: entry.StatusCode,
	})
}

func main() {
	logger := logutil.New("manager")

	hlog.SetLogger(logger)

	cfg, err := config.LoadFromExe()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	// 禁用 Hertz trailing slash 自动重定向（301），
	// 否则 /api/v1/nodes/x/sessions/saved 会被重定向到 /api/v1/nodes/x/sessions/saved/，
	// 导致 grpc-gateway 路径匹配失败。
	h := server.Default(
		server.WithHostPorts(cfg.Server.ListenAddr),
		server.WithRedirectTrailingSlash(false),
	)

	// 先创建 grpc-gateway ServeMux，后续需传给 register 层
	gwmux := gatewayutil.NewServeMux()

	// register 层初始化依赖、注册 Hertz 中间件和 WebSocket 路由
	resources := register(h, cfg, logger, gwmux)

	proxySvc := service.NewAgentProxyService(resources.Registry, cfg.Server.AuthToken, logger)

	// 构建 gRPC interceptor chain：认证 → 分层令牌 → 审计
	interceptors := []grpc.UnaryServerInterceptor{
		gatewayutil.UnaryAuthInterceptor(cfg.Server.AuthToken),
		gatewayutil.UnaryHostTokenInterceptor(cfg.Server.AuthToken, resources.Registry),
		gatewayutil.UnaryAuditInterceptor(&auditLoggerAdapter{inner: resources.AuditLog}),
	}

	grpcServer := managergrpc.NewServer(
		resources.HostSvc,
		resources.NodeSvc,
		resources.AuditSvc,
		proxySvc,
		resources.Registry,
		logger,
	)
	if err := grpcServer.Start(":9090",
		grpc.ChainUnaryInterceptor(interceptors...),
	); err != nil {
		logger.Fatalf("start grpc server: %v", err)
	}

	// 注册全部 7 个 Service 的 grpc-gateway handler（进程内直连 gRPC Server）
	ctx := context.Background()
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
	// AgentService 现在也注册到 grpc-gateway，节点注册和心跳走 gRPC interceptor 认证
	if err := pb.RegisterAgentServiceHandlerServer(ctx, gwmux, grpcServer); err != nil {
		logger.Fatalf("register AgentService gateway: %v", err)
	}

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
