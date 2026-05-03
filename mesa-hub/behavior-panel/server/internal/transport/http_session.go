package transport

import (
	"net/http"
	"time"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/mesa-hub-behavior-panel/internal/model"
	"github.com/charviki/mesa-hub-behavior-panel/internal/service"
)

const (
	proxyTimeout  = 30 * time.Second
	auditOperator = "frontend"
)

// SessionProxyHandler Session/Template/Config HTTP 代理 handler，转发请求到 Agent
type SessionProxyHandler struct {
	registry             *model.NodeRegistry
	auditLog             *AuditLogger
	auditSvc             *service.AuditService
	client               *http.Client
	logger               logutil.Logger
	authToken            string
	allowedOrigins       []string
	allowPrivateNetworks bool
}

// NewSessionProxyHandler 创建 SessionProxyHandler 实例
func NewSessionProxyHandler(registry *model.NodeRegistry, auditLog *AuditLogger, auditSvc *service.AuditService, logger logutil.Logger, authToken string, allowedOrigins []string, allowPrivateNetworks bool) *SessionProxyHandler {
	return &SessionProxyHandler{
		registry: registry,
		auditLog: auditLog,
		auditSvc: auditSvc,
		client: &http.Client{
			Timeout: proxyTimeout,
		},
		logger:               logger,
		authToken:            authToken,
		allowedOrigins:       allowedOrigins,
		allowPrivateNetworks: allowPrivateNetworks,
	}
}
