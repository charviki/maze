package transport

import (
	"time"

	"github.com/charviki/maze-cradle/auth"
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
	jwtSecret            string
	allowedOrigins       []string
	allowPrivateNetworks bool
}

// NewSessionProxyHandler 创建 SessionProxyHandler 实例。
func NewSessionProxyHandler(registry service.NodeRegistry, auditLog service.AuditLogWriter, logger logutil.Logger, jwtSecret string, allowedOrigins []string, allowPrivateNetworks bool) *SessionProxyHandler {
	return &SessionProxyHandler{
		registry:             registry,
		auditLog:             auditLog,
		logger:               logger,
		jwtSecret:            jwtSecret,
		allowedOrigins:       allowedOrigins,
		allowPrivateNetworks: allowPrivateNetworks,
	}
}

// generateServiceToken 生成短期 JWT，用于 Director Core 主动回调 Agent 时的服务间认证。
func (h *SessionProxyHandler) generateServiceToken() string {
	if h.jwtSecret == "" {
		return ""
	}
	token, err := auth.GenerateAccessToken(h.jwtSecret, auth.DefaultIssuer, "service:director-core", 5*time.Minute)
	if err != nil {
		return ""
	}
	return token
}
