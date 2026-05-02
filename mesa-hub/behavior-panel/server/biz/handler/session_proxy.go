package handler

import (
	"net/http"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/mesa-hub-behavior-panel/biz/model"
	"github.com/charviki/mesa-hub-behavior-panel/biz/service"
)

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
