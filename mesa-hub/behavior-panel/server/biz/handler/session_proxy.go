package handler

import (
	"net/http"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/mesa-hub-behavior-panel/biz/model"
)

// SessionProxyHandler 代理到 Agent 端的 Session 相关 API。
// 所有前端请求经过 Manager 代理，保持可观测性：
// - 前端不再直连 Agent API
// - 每次代理操作都记录审计日志
// - 请求超时保护，避免 Agent 无响应时阻塞 Manager
type SessionProxyHandler struct {
	registry             *model.NodeRegistry
	auditLog             *AuditLogger
	client               *http.Client
	logger               logutil.Logger
	authToken            string
	allowedOrigins       []string
	allowPrivateNetworks bool
}

// NewSessionProxyHandler 创建代理 handler，设置 30 秒超时防止 Agent 无响应阻塞。
// allowedOrigins 用于 WebSocket 升级时的 Origin 校验，为空时允许所有来源。
func NewSessionProxyHandler(registry *model.NodeRegistry, auditLog *AuditLogger, logger logutil.Logger, authToken string, allowedOrigins []string, allowPrivateNetworks bool) *SessionProxyHandler {
	return &SessionProxyHandler{
		registry: registry,
		auditLog: auditLog,
		client: &http.Client{
			Timeout: proxyTimeout,
		},
		logger:               logger,
		authToken:            authToken,
		allowedOrigins:       allowedOrigins,
		allowPrivateNetworks: allowPrivateNetworks,
	}
}
