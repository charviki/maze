package transport

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	cradleDB "github.com/charviki/maze/fabrication/cradle/db"
	"github.com/charviki/maze/fabrication/cradle/gatewayutil"
	"github.com/charviki/maze/fabrication/cradle/grpcutil"
	"github.com/charviki/maze/fabrication/cradle/lifecycle"
	"github.com/charviki/maze/fabrication/cradle/logutil"
	"github.com/charviki/maze/fabrication/cradle/protocol"
	"github.com/charviki/maze/the-mesa/director-core/internal/agentclient"
	"github.com/charviki/maze/the-mesa/director-core/internal/config"
	"github.com/charviki/maze/the-mesa/director-core/internal/reconciler"
	"github.com/charviki/maze/the-mesa/director-core/internal/repository/postgres"
	"github.com/charviki/maze/the-mesa/director-core/internal/runtime"
	"github.com/charviki/maze/the-mesa/director-core/internal/service"
)

// CleanupResources 聚合 director-core 启动时构造的全部资源，供调用方统一释放。
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

// Cleanup 释放所有资源（停止 reconciler、关闭连接池等）。
func (r *CleanupResources) Cleanup() {
	if r == nil {
		return
	}
	r.Reconciler.Stop()
	r.HostSvc.Stop()
	r.ConnMgr.CloseAll()
	if r.HostPool != nil {
		r.HostPool.Close()
	}
}

// AuthzResult 包含权限系统初始化后的全部产物。
type AuthzResult struct {
	PermHandler *PermissionServiceServer
	AuthHandler *AuthHandler
	Interceptor grpc.UnaryServerInterceptor
	Cleanup     func()
	CredStore   *postgres.CredentialRepository
}

// ServerParams 包含创建 director-core 全套服务所需的参数。
type ServerParams struct {
	Config      *config.Config
	Logger      logutil.Logger
	HostPool    *pgxpool.Pool
	AuthzResult *AuthzResult
}

// NewGRPCGatewayServer 创建 gRPC + gateway + HTTP 全套服务。
// 内部完成: interceptor chain → gRPC server → service 注册 → gateway 注册 → HTTP server。
// 返回 (httpServer, managedGRPC, resources, error)。
func NewGRPCGatewayServer(params ServerParams) (*http.Server, lifecycle.Server, *CleanupResources, error) {
	gwmux := gatewayutil.NewServeMux()

	httpServer, resources, err := newHTTPServer(params.Config, params.Logger, gwmux, params.HostPool)
	if err != nil {
		return nil, nil, nil, err
	}

	grpcAddr := params.Config.Server.GRPCListenAddr
	if grpcAddr == "" {
		grpcAddr = ":9090"
	}

	// 构建 gRPC interceptor chain
	validationInterceptor, err := gatewayutil.NewValidationInterceptor()
	if err != nil {
		return nil, nil, nil, err
	}
	interceptors := []grpc.UnaryServerInterceptor{
		validationInterceptor,
		gatewayutil.UnaryAuthInterceptor(params.Config.Server.JWTSecret),
		gatewayutil.UnaryHostTokenInterceptor(params.Config.Server.JWTSecret, resources.Registry),
	}

	if params.AuthzResult != nil {
		resources.HostSvc.SetCredentialStore(params.AuthzResult.CredStore)
		interceptors = append(interceptors, params.AuthzResult.Interceptor)
	}

	interceptors = append(interceptors,
		gatewayutil.UnaryAuditInterceptor(&auditLoggerAdapter{inner: resources.AuditLog}),
	)

	proxySvc := agentclient.NewProxy(resources.Registry, resources.ConnMgr, params.Config.Server.JWTSecret)

	grpcCore := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
	grpcServer := NewServer(
		resources.HostSvc,
		resources.NodeSvc,
		resources.AuditSvc,
		resources.SkillSvc,
		resources.MCPSvc,
		resources.RuleSvc,
		resources.GitKeySvc,
		proxySvc,
		resources.Registry,
		params.Config.Server.JWTSecret,
		params.Logger,
	)
	grpcServer.RegisterGRPC(grpcCore)

	if params.AuthzResult != nil {
		pb.RegisterPermissionServiceServer(grpcCore, params.AuthzResult.PermHandler)
		pb.RegisterAuthServiceServer(grpcCore, params.AuthzResult.AuthHandler)
	}

	managedGRPC := grpcutil.NewManagedGRPCServer(grpcAddr, grpcCore, params.Logger)

	// gateway 注册
	var permHandler *PermissionServiceServer
	var authHandler *AuthHandler
	if params.AuthzResult != nil {
		permHandler = params.AuthzResult.PermHandler
		authHandler = params.AuthzResult.AuthHandler
	}
	if err := RegisterGatewayHandlers(context.Background(), GatewayRegistrationParams{
		GWMux:       gwmux,
		GRPCAddr:    grpcAddr,
		GRPCServer:  grpcCore,
		PermHandler: permHandler,
		AuthHandler: authHandler,
	}); err != nil {
		return nil, nil, nil, err
	}

	return httpServer, managedGRPC, resources, nil
}

// auditLoggerAdapter 将 service.AuditLogWriter 适配为 gatewayutil.AuditLogger。
type auditLoggerAdapter struct {
	inner service.AuditLogWriter
}

// Log 实现 gatewayutil.AuditLogger。
func (a *auditLoggerAdapter) Log(entry gatewayutil.AuditEntry) {
	_ = a.inner.Log(context.Background(), protocol.AuditLogEntry{
		Operator:   entry.Operator,
		TargetNode: entry.TargetNode,
		Action:     entry.Action,
		Result:     entry.Result,
		StatusCode: entry.StatusCode,
	})
}

func dataDir(cfg *config.Config) string {
	if cfg.Workspace.BaseDir != "" {
		return cfg.Workspace.BaseDir
	}
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	return "."
}

// newHTTPServer 负责依赖注入和 HTTP server 构造。
func newHTTPServer(cfg *config.Config, logger logutil.Logger, gwmux *gwruntime.ServeMux, hostPool *pgxpool.Pool) (*http.Server, *CleanupResources, error) {
	// Host 数据库迁移
	hostMigrationsFS, fsErr := fs.Sub(postgres.HostMigrationsFS, "migrations")
	if fsErr != nil {
		return nil, nil, fsErr
	}
	if err := cradleDB.RunMigrations(hostPool, hostMigrationsFS); err != nil {
		return nil, nil, err
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
			return nil, nil, err
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
			return nil, nil, fmt.Errorf("invalid git_key_encryption_key: not valid hex: %w", err)
		}
		if len(decoded) != 32 {
			return nil, nil, fmt.Errorf("invalid git_key_encryption_key: expected 32 bytes (64 hex chars), got %d bytes", len(decoded))
		}
		gitKeyEncryptKey = decoded
	}
	gitKeySvc, err := service.NewGitKeyService(gitKeyRepo, gitKeyEncryptKey, logger)
	if err != nil {
		return nil, nil, err
	}

	hostSvc.SetResourceRepos(skillRepo, mcpRepo, ruleRepo, gitKeySvc)

	sessionProxyHandler := NewSessionProxyHandler(registry, auditLog, logger, cfg.Server.JWTSecret, cfg.AllowedOrigins(), cfg.Server.AllowPrivateNetworks)

	ctx := context.Background()
	rec := reconciler.NewReconciler(hostSpecRepo, registry, hostRuntime, cfg, logger, logDir)
	rec.RecoverOnStartup(ctx)
	rec.StartHealthCheck(ctx)

	httpServer := NewHTTPServer(HTTPHandlerParams{
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

	return httpServer, resources, nil
}
