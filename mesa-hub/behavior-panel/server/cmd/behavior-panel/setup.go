package main

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	cradleDB "github.com/charviki/maze-cradle/db"
	"github.com/charviki/maze-cradle/httputil"
	"github.com/charviki/maze-cradle/logutil"
	cradlemw "github.com/charviki/maze-cradle/middleware"
	"github.com/charviki/mesa-hub-behavior-panel/internal/agentclient"
	"github.com/charviki/mesa-hub-behavior-panel/internal/config"
	"github.com/charviki/mesa-hub-behavior-panel/internal/reconciler"
	"github.com/charviki/mesa-hub-behavior-panel/internal/repository/postgres"
	"github.com/charviki/mesa-hub-behavior-panel/internal/runtime"
	"github.com/charviki/mesa-hub-behavior-panel/internal/service"
	"github.com/charviki/mesa-hub-behavior-panel/internal/transport"
	"github.com/jackc/pgx/v5/pgxpool"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

const (
	httpReadTimeout  = 10 * time.Second
	httpWriteTimeout = 30 * time.Second
	httpIdleTimeout  = 120 * time.Second
)

func dataDir(cfg *config.Config) string {
	if cfg.Workspace.BaseDir != "" {
		return cfg.Workspace.BaseDir
	}
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	return "."
}

type CleanupResources struct {
	HostPool     *pgxpool.Pool
	Registry     service.NodeRegistry
	HostSpecRepo service.HostSpecRepository
	AuditLog     service.AuditLogRepository
	Reconciler   *reconciler.Reconciler
	HostSvc      *service.HostService
	NodeSvc      *service.NodeService
	AuditSvc     *service.AuditService
	ConnMgr      *agentclient.ConnectionManager
}

func newHTTPServer(cfg *config.Config, logger logutil.Logger, gwmux *gwruntime.ServeMux, hostPool *pgxpool.Pool) (*http.Server, *CleanupResources) {
	ctx := context.Background()

	// Host 数据库迁移（00003_init_host_schema.sql）
	hostMigrationsFS, fsErr := fs.Sub(postgres.HostMigrationsFS, "migrations")
	if fsErr != nil {
		logger.Fatalf("[host] sub host migrations: %v", fsErr)
	}
	if err := cradleDB.RunMigrations(hostPool, hostMigrationsFS); err != nil {
		logger.Fatalf("[host] run migrations: %v", err)
	}
	logger.Infof("[host] database migrations completed")

	registry := postgres.NewNodeRegistry(hostPool)
	hostSpecRepo := postgres.NewHostSpecRepository(hostPool)
	auditLog := postgres.NewAuditLogRepository(hostPool)
	hostTxManager := postgres.NewHostTxManager(hostPool)

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

	dir := dataDir(cfg)
	logDir := filepath.Join(dir, "host_logs")

	hostSvc := service.NewHostService(registry, hostSpecRepo, hostTxManager, hostRuntime, auditLog, cfg, logger, logDir)
	nodeSvc := service.NewNodeService(registry, logger)
	auditSvc := service.NewAuditService(auditLog)

	sessionProxyHandler := transport.NewSessionProxyHandler(registry, auditLog, logger, cfg.Server.AuthToken, cfg.AllowedOrigins(), cfg.Server.AllowPrivateNetworks)

	rec := reconciler.NewReconciler(hostSpecRepo, registry, hostRuntime, cfg, logger, logDir)
	rec.RecoverOnStartup(ctx)
	rec.StartHealthCheck(ctx)

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
	mux.Handle("POST /api/v1/nodes/register", agentHandler)
	mux.Handle("POST /api/v1/nodes/heartbeat", agentHandler)
	mux.Handle("GET /api/v1/nodes/{name}/sessions/{id}/ws", wsHandler)
	mux.Handle("/", apiHandler)

	connMgr := agentclient.NewConnectionManager(logger, cfg.Server.AuthToken, 5*time.Minute)

	resources := &CleanupResources{
		HostPool:     hostPool,
		Registry:     registry,
		HostSpecRepo: hostSpecRepo,
		AuditLog:     auditLog,
		Reconciler:   rec,
		HostSvc:      hostSvc,
		NodeSvc:      nodeSvc,
		AuditSvc:     auditSvc,
		ConnMgr:      connMgr,
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
			recorder := httputil.NewStatusRecorder(w)
			startedAt := time.Now()
			next.ServeHTTP(recorder, r)
			if logger != nil {
				logger.Infof("[http] %s %s status=%d duration=%s", r.Method, r.URL.Path, recorder.Status(), time.Since(startedAt))
			}
		})
	}
}
