package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/charviki/maze-cradle/httputil"
	"github.com/charviki/maze-cradle/logutil"
	cradlemw "github.com/charviki/maze-cradle/middleware"
	"github.com/charviki/mesa-hub-behavior-panel/internal/config"
	"github.com/charviki/mesa-hub-behavior-panel/internal/model"
	"github.com/charviki/mesa-hub-behavior-panel/internal/reconciler"
	"github.com/charviki/mesa-hub-behavior-panel/internal/runtime"
	"github.com/charviki/mesa-hub-behavior-panel/internal/service"
	"github.com/charviki/mesa-hub-behavior-panel/internal/transport"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

const (
	httpReadTimeout  = 10 * time.Second
	httpWriteTimeout = 30 * time.Second
	httpIdleTimeout  = 120 * time.Second
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

// CleanupResources 持有优雅关闭时需要清理的资源引用。
type CleanupResources struct {
	Registry   *model.NodeRegistry
	SpecMgr    *model.HostSpecManager
	AuditLog   *transport.AuditLogger
	Reconciler *reconciler.Reconciler
	HostSvc    *service.HostService
	NodeSvc    *service.NodeService
	AuditSvc   *service.AuditService
	ConnMgr    *service.ConnectionManager
}

// newHTTPServer 负责装配 HTTP Server、顶层路由和后台依赖。
// setup 与 main 放在同目录，是为了把“服务装配”与“进程生命周期入口”放在一起维护。
func newHTTPServer(cfg *config.Config, logger logutil.Logger, gwmux *gwruntime.ServeMux) (*http.Server, *CleanupResources) {
	dir := dataDir(cfg)

	registry := model.NewNodeRegistry(filepath.Join(dir, "nodes.json"), logger)
	specMgr := model.NewHostSpecManager(filepath.Join(dir, "host_specs.json"), logger)

	auditLog := transport.NewAuditLogger(filepath.Join(dir, "audit.log"), logger)

	var hostRuntime runtime.HostRuntime
	if cfg.Runtime.Type == "kubernetes" {
		k8sRT, err := runtime.NewKubernetesRuntime(cfg.Kubernetes, cfg.Workspace, logger)
		if err != nil {
			logger.Fatalf("kubernetes runtime init failed: %v", err)
		}
		hostRuntime = k8sRT
	} else {
		hostRuntime = runtime.NewDockerRuntime(cfg.Docker, cfg.Workspace, logger)
	}

	logDir := filepath.Join(dir, "host_logs")

	hostSvc := service.NewHostService(registry, specMgr, hostRuntime, auditLog, cfg, logger, logDir)
	nodeSvc := service.NewNodeService(registry, logger)
	auditSvc := service.NewAuditService(auditLog)

	// SessionProxyHandler 仍需保留：WebSocket 终端代理需要访问 registry、auditLog 等依赖
	sessionProxyHandler := transport.NewSessionProxyHandler(registry, auditLog, auditSvc, logger, cfg.Server.AuthToken, cfg.AllowedOrigins(), cfg.Server.AllowPrivateNetworks)

	// 启动恢复：路由注册后、HTTP 服务启动前执行
	rec := reconciler.NewReconciler(specMgr, registry, hostRuntime, cfg, logger, logDir)
	rec.RecoverOnStartup(context.Background())
	// 启动健康巡检后台 goroutine
	rec.StartHealthCheck(context.Background())

	apiHandler := chainHTTP(
		gwmux,
		accessLogMiddleware(logger),
		corsMiddleware(cfg.AllowedOrigins()),
		cradlemw.Auth(cfg.Server.AuthToken),
	)
	agentHandler := chainHTTP(
		gwmux,
		accessLogMiddleware(logger),
		corsMiddleware(cfg.AllowedOrigins()),
	)
	wsHandler := chainHTTP(
		http.HandlerFunc(sessionProxyHandler.ProxyWebSocket),
		accessLogMiddleware(logger),
		corsMiddleware(cfg.AllowedOrigins()),
		cradlemw.Auth(cfg.Server.AuthToken),
	)

	mux := http.NewServeMux()
	mux.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	// Agent 注册/心跳必须绕过 HTTP 层全局 auth，交给 gRPC HostTokenInterceptor 执行分层令牌校验；
	// 否则 Host 专属 token 会在进入 grpc-gateway 前就被全局管理员 token 校验拦截。
	mux.Handle("POST /api/v1/nodes/register", agentHandler)
	mux.Handle("POST /api/v1/nodes/heartbeat", agentHandler)
	mux.Handle("GET /api/v1/nodes/{name}/sessions/{id}/ws", wsHandler)
	mux.Handle("/", apiHandler)

	connMgr := service.NewConnectionManager(logger, cfg.Server.AuthToken, 5*time.Minute)

	resources := &CleanupResources{
		Registry:   registry,
		SpecMgr:    specMgr,
		AuditLog:   auditLog,
		Reconciler: rec,
		HostSvc:    hostSvc,
		NodeSvc:    nodeSvc,
		AuditSvc:   auditSvc,
		ConnMgr:    connMgr,
	}

	return &http.Server{
		Addr:         cfg.Server.ListenAddr,
		Handler:      mux,
		ReadTimeout:  httpReadTimeout,
		WriteTimeout: httpWriteTimeout,
		IdleTimeout:  httpIdleTimeout,
	}, resources
}

func corsMiddleware(origins []string) func(http.Handler) http.Handler {
	if len(origins) == 0 {
		return cradlemw.CORS()
	}
	return cradlemw.CORSWithOrigins(origins)
}

func chainHTTP(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	wrapped := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}
	return wrapped
}

func accessLogMiddleware(logger logutil.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 访问日志需要记录状态码，但不能吞掉 Hijacker/Flusher 等可选能力，
			// 否则 WebSocket 升级会在中间件链里被拦断。
			recorder := httputil.NewStatusRecorder(w)
			startedAt := time.Now()
			next.ServeHTTP(recorder, r)
			if logger != nil {
				// 访问日志在标准库下自行实现，保持迁移后仍可观测请求状态与耗时。
				logger.Infof("[http] %s %s status=%d duration=%s", r.Method, r.URL.Path, recorder.Status(), time.Since(startedAt))
			}
		})
	}
}
