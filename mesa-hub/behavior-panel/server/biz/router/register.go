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
}

// 注册所有 API 路由并初始化各 Store 和 Handler。
// 通过构造函数注入 Store 依赖到 Handler，实现手动依赖注入。
// 返回 CleanupResources 供调用方在优雅关闭时执行资源清理。
func Register(h *server.Hertz, cfg *config.Config, logger logutil.Logger) *CleanupResources {
	dir := dataDir(cfg)

	registry := model.NewNodeRegistry(filepath.Join(dir, "nodes.json"), logger)
	specMgr := model.NewHostSpecManager(filepath.Join(dir, "host_specs.json"), logger)
	nodeHandler := handler.NewNodeHandler(registry, cfg.Server.AuthToken, logger)

	auditLog := handler.NewAuditLogger(filepath.Join(dir, "audit.log"), logger)
	sessionProxyHandler := handler.NewSessionProxyHandler(registry, auditLog, logger, cfg.Server.AuthToken, cfg.AllowedOrigins(), cfg.Server.AllowPrivateNetworks)

	// 根据运行时类型选择对应的 HostRuntime 实现
	var hostRuntime runtime.HostRuntime
	if cfg.Runtime.Type == "kubernetes" {
		hostRuntime = runtime.NewKubernetesRuntime(cfg.Kubernetes, cfg.Workspace, logger)
	} else {
		hostRuntime = runtime.NewDockerRuntime(cfg.Docker, cfg.Workspace)
	}

	// host_logs 目录用于存储构建日志
	logDir := filepath.Join(dir, "host_logs")
	hostHandler := handler.NewHostHandler(registry, specMgr, hostRuntime, auditLog, cfg, logger, logDir)

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

	api := h.Group("/api/v1")

	// 节点注册和心跳端点：认证在 handler 内部完成（支持 Host 令牌 + 全局令牌分层校验）
	api.POST("/nodes/register", nodeHandler.Register)
	api.POST("/nodes/heartbeat", nodeHandler.Heartbeat)

	// 其余 API 需要 Auth 保护（当 config.auth_token 非空时生效）
	protected := api.Group("", cradlemw.Auth(cfg.Server.AuthToken))

	// 节点查询和删除
	protected.GET("/nodes", nodeHandler.ListNodes)
	protected.GET("/nodes/:name", nodeHandler.GetNode)
	protected.DELETE("/nodes/:name", nodeHandler.DeleteNode)

	// Session/Template/LocalConfig 管理代理路由（前端通过 Manager 代理到 Agent）
	protected.GET("/nodes/:name/sessions", sessionProxyHandler.ListSessions)
	protected.POST("/nodes/:name/sessions", sessionProxyHandler.CreateSession)
	protected.GET("/nodes/:name/sessions/saved", sessionProxyHandler.GetSavedSessions)
	protected.GET("/nodes/:name/sessions/:id", sessionProxyHandler.GetSession)
	protected.DELETE("/nodes/:name/sessions/:id", sessionProxyHandler.DeleteSession)
	protected.GET("/nodes/:name/sessions/:id/config", sessionProxyHandler.GetSessionConfig)
	protected.PUT("/nodes/:name/sessions/:id/config", sessionProxyHandler.UpdateSessionConfig)
	protected.POST("/nodes/:name/sessions/:id/restore", sessionProxyHandler.RestoreSession)
	protected.POST("/nodes/:name/sessions/save", sessionProxyHandler.SaveAllSessions)

	// Template 代理
	protected.GET("/nodes/:name/templates", sessionProxyHandler.ListTemplates)
	protected.POST("/nodes/:name/templates", sessionProxyHandler.CreateTemplate)
	protected.GET("/nodes/:name/templates/:id", sessionProxyHandler.GetTemplate)
	protected.PUT("/nodes/:name/templates/:id", sessionProxyHandler.UpdateTemplate)
	protected.DELETE("/nodes/:name/templates/:id", sessionProxyHandler.DeleteTemplate)
	protected.GET("/nodes/:name/templates/:id/config", sessionProxyHandler.GetTemplateConfig)
	protected.PUT("/nodes/:name/templates/:id/config", sessionProxyHandler.UpdateTemplateConfig)

	// Local Config 代理
	protected.GET("/nodes/:name/local-config", sessionProxyHandler.GetLocalConfig)
	protected.PUT("/nodes/:name/local-config", sessionProxyHandler.UpdateLocalConfig)

	// WebSocket 终端代理（前端通过 Manager 代理到 Agent 的 WebSocket 连接）
	protected.GET("/nodes/:name/sessions/:id/ws", sessionProxyHandler.ProxyWebSocket)

	// Host 生命周期管理（异步创建 + 全生命周期状态）
	protected.POST("/hosts", hostHandler.CreateHost)
	protected.GET("/hosts", hostHandler.ListHosts)
	protected.GET("/hosts/:name", hostHandler.GetHost)
	protected.GET("/host/tools", hostHandler.ListTools)
	protected.DELETE("/hosts/:name", hostHandler.DeleteHost)
	protected.GET("/hosts/:name/logs/build", hostHandler.GetBuildLog)
	protected.GET("/hosts/:name/logs/runtime", hostHandler.GetRuntimeLog)

	// 审计日志路由
	protected.GET("/audit/logs", sessionProxyHandler.GetAuditLogs)

	return &CleanupResources{
		Registry:   registry,
		SpecMgr:    specMgr,
		AuditLog:   auditLog,
		Reconciler: rec,
	}
}
