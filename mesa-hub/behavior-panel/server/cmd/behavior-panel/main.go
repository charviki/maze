package main

import (
	"context"
	"time"

	"github.com/charviki/maze-cradle/grpcutil"
	"google.golang.org/grpc"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/gatewayutil"
	"github.com/charviki/maze-cradle/lifecycle"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/internal/config"
	"github.com/charviki/mesa-hub-behavior-panel/internal/service"
	"github.com/charviki/mesa-hub-behavior-panel/internal/transport"
)

const grpcListenAddr = ":9090"

// auditLoggerAdapter 将 transport.AuditLogger 适配为 gatewayutil.AuditLogger 接口。
// gatewayutil interceptor 产生 gatewayutil.AuditEntry，需转换为 transport 层使用的 protocol.AuditLogEntry。
type auditLoggerAdapter struct {
	inner *transport.AuditLogger
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

	cfg, err := config.LoadFromExe()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	gwmux := gatewayutil.NewServeMux()
	httpServer, resources := newHTTPServer(cfg, logger, gwmux)
	defer cleanupResources(resources)

	proxySvc := service.NewAgentProxyService(resources.Registry, cfg.Server.AuthToken, logger)

	// 构建 gRPC interceptor chain：认证 → 分层令牌 → 审计
	interceptors := []grpc.UnaryServerInterceptor{
		gatewayutil.UnaryAuthInterceptor(cfg.Server.AuthToken),
		gatewayutil.UnaryHostTokenInterceptor(cfg.Server.AuthToken, resources.Registry),
		gatewayutil.UnaryAuditInterceptor(&auditLoggerAdapter{inner: resources.AuditLog}),
	}

	grpcServer := transport.NewServer(
		resources.HostSvc,
		resources.NodeSvc,
		resources.AuditSvc,
		proxySvc,
		resources.Registry,
		cfg.Server.AuthToken,
		logger,
	)
	grpcCore := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
	grpcServer.RegisterGRPC(grpcCore)
	managedGRPC := grpcutil.NewManagedGRPCServer(grpcListenAddr, grpcCore, logger)

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

	manager := &lifecycle.Manager{
		ShutdownTimeout: 5 * time.Second,
		Logger:          logger,
	}
	manager.Add(httpServer)
	manager.Add(managedGRPC)

	logger.Infof("agent-manager controller started on %s", cfg.Server.ListenAddr)
	if cfg.IsDevMode() {
		logger.Warnf("[security] running in DEV mode: auth_token is empty, all API endpoints are open")
	}
	if len(cfg.AllowedOrigins()) == 0 {
		logger.Warnf("[security] running in DEV mode: CORS and WebSocket allow all origins")
	}

	if err := manager.Run(context.Background()); err != nil {
		logger.Fatalf("run server lifecycle: %v", err)
	}
}

func cleanupResources(resources *CleanupResources) {
	if resources == nil {
		return
	}

	resources.Reconciler.Stop()
	resources.Registry.WaitSave()
	resources.SpecMgr.WaitSave()
	resources.AuditLog.Close()
}
