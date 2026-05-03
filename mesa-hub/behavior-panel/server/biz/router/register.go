package router

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/logger/accesslog"

	"github.com/charviki/maze-cradle/logutil"
	cradlemw "github.com/charviki/maze-cradle/middleware"
	"github.com/charviki/mesa-hub-behavior-panel/biz/config"
	"github.com/charviki/mesa-hub-behavior-panel/biz/handler"
	"github.com/charviki/mesa-hub-behavior-panel/biz/model"
	"github.com/charviki/mesa-hub-behavior-panel/biz/reconciler"
	"github.com/charviki/mesa-hub-behavior-panel/biz/runtime"
	"github.com/charviki/mesa-hub-behavior-panel/biz/service"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// dataDir 返回数据文件存储目录。
// 优先使用配置中的 workspace.base_dir，
// 未配置时回退到可执行文件所在目录（开发模式兼容）。
func dataDir(cfg *config.Config) string {
	if cfg.Workspace.BaseDir != "" {
		return cfg.Workspace.BaseDir
	}
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	return "."
}

// CleanupResources 持有优雅关闭时需要清理的资源引用
type CleanupResources struct {
	Registry   *model.NodeRegistry
	SpecMgr    *model.HostSpecManager
	AuditLog   *handler.AuditLogger
	Reconciler *reconciler.Reconciler
	HostSvc    *service.HostService
	NodeSvc    *service.NodeService
	AuditSvc   *service.AuditService
}

// Register 初始化依赖、注册 Hertz 中间件和 WebSocket 路由，启动 Reconciler。
// REST API 路由已迁移到 grpc-gateway（由 main.go 通过 NoRoute 转发），此处不再注册。
// gwmux 参数保持接口一致性，本函数不直接使用。
func Register(h *server.Hertz, cfg *config.Config, logger logutil.Logger, gwmux *gwruntime.ServeMux) *CleanupResources {
	dir := dataDir(cfg)

	registry := model.NewNodeRegistry(filepath.Join(dir, "nodes.json"), logger)
	specMgr := model.NewHostSpecManager(filepath.Join(dir, "host_specs.json"), logger)

	auditLog := handler.NewAuditLogger(filepath.Join(dir, "audit.log"), logger)

	var hostRuntime runtime.HostRuntime
	if cfg.Runtime.Type == "kubernetes" {
		hostRuntime = runtime.NewKubernetesRuntime(cfg.Kubernetes, cfg.Workspace, logger)
	} else {
		hostRuntime = runtime.NewDockerRuntime(cfg.Docker, cfg.Workspace, logger)
	}

	logDir := filepath.Join(dir, "host_logs")

	hostSvc := service.NewHostService(registry, specMgr, hostRuntime, auditLog, cfg, logger, logDir)
	nodeSvc := service.NewNodeService(registry, logger)
	auditSvc := service.NewAuditService(auditLog)

	// SessionProxyHandler 仍需保留：WebSocket 终端代理需要访问 registry、auditLog 等依赖
	sessionProxyHandler := handler.NewSessionProxyHandler(registry, auditLog, auditSvc, logger, cfg.Server.AuthToken, cfg.AllowedOrigins(), cfg.Server.AllowPrivateNetworks)

	// 启动恢复：路由注册后、HTTP 服务启动前执行
	rec := reconciler.NewReconciler(specMgr, registry, hostRuntime, cfg, logger, logDir)
	rec.RecoverOnStartup(context.Background())
	// 启动健康巡检后台 goroutine
	rec.StartHealthCheck(context.Background())

	// Access Log
	h.Use(accesslog.New())

	// CORS：配置了 allowed_origins 时使用白名单，否则允许所有来源（开发模式）
	origins := cfg.AllowedOrigins()
	if len(origins) > 0 {
		h.Use(cradlemw.CORSWithOrigins(origins))
	} else {
		h.Use(cradlemw.CORS())
	}

	// WebSocket 终端代理：前端通过 Manager 代理到 Agent 的 WebSocket 连接。
	// 认证由 Hertz Auth 中间件保护，WebSocket 路由不经过 grpc-gateway（grpc-gateway 不支持 HTTP 升级）
	protected := h.Group("/api/v1", cradlemw.Auth(cfg.Server.AuthToken))
	protected.GET("/nodes/:name/sessions/:id/ws", sessionProxyHandler.ProxyWebSocket)

	return &CleanupResources{
		Registry:   registry,
		SpecMgr:    specMgr,
		AuditLog:   auditLog,
		Reconciler: rec,
		HostSvc:    hostSvc,
		NodeSvc:    nodeSvc,
		AuditSvc:   auditSvc,
	}
}
