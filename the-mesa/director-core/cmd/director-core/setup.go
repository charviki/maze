package main

import (
	"context"
	"encoding/hex"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	cradleDB "github.com/charviki/maze/fabrication/cradle/db"
	"github.com/charviki/maze/fabrication/cradle/logutil"
	"github.com/charviki/maze/the-mesa/director-core/internal/agentclient"
	"github.com/charviki/maze/the-mesa/director-core/internal/config"
	"github.com/charviki/maze/the-mesa/director-core/internal/reconciler"
	"github.com/charviki/maze/the-mesa/director-core/internal/repository/postgres"
	"github.com/charviki/maze/the-mesa/director-core/internal/runtime"
	"github.com/charviki/maze/the-mesa/director-core/internal/service"
	"github.com/charviki/maze/the-mesa/director-core/internal/transport"
	"github.com/jackc/pgx/v5/pgxpool"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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
	SkillSvc     *service.SkillService
	MCPSvc       *service.MCPServerService
	RuleSvc      *service.RuleService
	GitKeySvc    *service.GitKeyService
	ConnMgr      *agentclient.ConnectionManager
}

// newHTTPServer 负责依赖注入和 HTTP server 构造，但不再直接拼装路由和 middleware。
func newHTTPServer(cfg *config.Config, logger logutil.Logger, gwmux *gwruntime.ServeMux, hostPool *pgxpool.Pool) (*http.Server, *CleanupResources) {
	ctx := context.Background()

	// Host 数据库迁移
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

	hostSvc := service.NewHostService(registry, hostSpecRepo, hostTxManager, hostRuntime, auditLog, nil, cfg, logger, logDir)
	nodeSvc := service.NewNodeService(registry, logger)
	auditSvc := service.NewAuditService(auditLog)

	skillRepo := postgres.NewSkillRepository(hostPool)
	mcpRepo := postgres.NewMCPServerRepository(hostPool)
	ruleRepo := postgres.NewRuleRepository(hostPool)
	gitKeyRepo := postgres.NewGitKeyRepository(hostPool)
	skillSvc := service.NewSkillService(skillRepo, logger)
	mcpSvc := service.NewMCPServerService(mcpRepo, logger)
	ruleSvc := service.NewRuleService(ruleRepo, logger)

	var gitKeyEncryptKey []byte
	if cfg.GitKeyEncryptionKey != "" {
		decoded, err := hex.DecodeString(cfg.GitKeyEncryptionKey)
		if err != nil {
			logger.Fatalf("invalid git_key_encryption_key: not valid hex: %v", err)
		}
		if len(decoded) != 32 {
			logger.Fatalf("invalid git_key_encryption_key: expected 32 bytes (64 hex chars), got %d bytes", len(decoded))
		}
		gitKeyEncryptKey = decoded
	}
	gitKeySvc, err := service.NewGitKeyService(gitKeyRepo, gitKeyEncryptKey, logger)
	if err != nil {
		logger.Fatalf("%v", err)
	}

	sessionProxyHandler := transport.NewSessionProxyHandler(registry, auditLog, logger, cfg.Server.JWTSecret, cfg.AllowedOrigins(), cfg.Server.AllowPrivateNetworks)

	rec := reconciler.NewReconciler(hostSpecRepo, registry, hostRuntime, cfg, logger, logDir)
	rec.RecoverOnStartup(ctx)
	rec.StartHealthCheck(ctx)

	httpServer := transport.NewHTTPServer(transport.HTTPHandlerParams{
		Config:              cfg,
		Logger:              logger,
		GWMux:               gwmux,
		SessionProxyHandler: sessionProxyHandler,
		JWTSecret:           cfg.Server.JWTSecret,
		AllowedOrigins:      cfg.AllowedOrigins(),
	})

	connMgr := agentclient.NewConnectionManager(logger, cfg.Server.JWTSecret, 5*time.Minute)

	resources := &CleanupResources{
		HostPool:     hostPool,
		Registry:     registry,
		HostSpecRepo: hostSpecRepo,
		AuditLog:     auditLog,
		Reconciler:   rec,
		HostSvc:      hostSvc,
		NodeSvc:      nodeSvc,
		AuditSvc:     auditSvc,
		SkillSvc:     skillSvc,
		MCPSvc:       mcpSvc,
		RuleSvc:      ruleSvc,
		GitKeySvc:    gitKeySvc,
		ConnMgr:      connMgr,
	}

	return httpServer, resources
}

func cleanupResources(resources *CleanupResources) {
	if resources == nil {
		return
	}

	resources.Reconciler.Stop()
	resources.HostSvc.Stop()
	resources.ConnMgr.CloseAll()
	if resources.HostPool != nil {
		resources.HostPool.Close()
	}
}
