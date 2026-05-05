package transport

import (
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze/the-mesa/director-core/internal/service"
)

const (
	auditOperator = "frontend"
)

// SessionProxyHandler Session/Template/Config HTTP 代理 handler，转发请求到 Agent
type SessionProxyHandler struct {
	registry             service.NodeRegistry
	auditLog             service.AuditLogWriter
	logger               logutil.Logger
	authToken            string
	allowedOrigins       []string
	allowPrivateNetworks bool
}

// NewSessionProxyHandler 创建 SessionProxyHandler 实例。
func NewSessionProxyHandler(registry service.NodeRegistry, auditLog service.AuditLogWriter, logger logutil.Logger, authToken string, allowedOrigins []string, allowPrivateNetworks bool) *SessionProxyHandler {
	return &SessionProxyHandler{
		registry:             registry,
		auditLog:             auditLog,
		logger:               logger,
		authToken:            authToken,
		allowedOrigins:       allowedOrigins,
		allowPrivateNetworks: allowPrivateNetworks,
	}
}
