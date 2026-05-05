package transport

import (
	"fmt"
	"net/http"
	"strings"

	gorillaws "github.com/gorilla/websocket"

	"github.com/charviki/maze-cradle/httputil"
	"github.com/charviki/maze-cradle/protocol"
)

// ProxyWebSocket WebSocket 双向代理：前端 ↔ Manager ↔ Agent。
// Manager 将前端的 WebSocket 连接升级后，建立到 Agent 的 WebSocket 客户端连接，
// 然后在两个连接之间双向转发消息（Binary 和 Text）。
// 任意一端断开时，另一端也自动关闭，确保资源不泄漏。
func (h *SessionProxyHandler) ProxyWebSocket(w http.ResponseWriter, r *http.Request) {
	nodeName := r.PathValue("name")
	sessionID := r.PathValue("id")
	if nodeName == "" || sessionID == "" {
		httputil.Error(w, r, http.StatusBadRequest, "node name and session id are required")
		return
	}

	node, err := h.registry.Get(r.Context(), nodeName)
	if err != nil {
		h.logger.Errorf("[ws-proxy] get node %s failed: %v", nodeName, err)
		httputil.Error(w, r, http.StatusInternalServerError, "load node failed")
		return
	}
	if node == nil {
		httputil.Error(w, r, http.StatusNotFound, "node not found")
		return
	}

	if err := h.auditLog.Log(r.Context(), protocol.AuditLogEntry{
		Operator:       auditOperator,
		TargetNode:     nodeName,
		Action:         "websocket_connect",
		PayloadSummary: "session=" + sessionID,
		Result:         "connecting",
		StatusCode:     http.StatusSwitchingProtocols,
	}); err != nil {
		h.logger.Errorf("[ws-proxy] write audit log for %s failed: %v", nodeName, err)
	}

	// 使用配置化的 Origin 校验替代硬编码的"允许所有来源"，避免跨站 WebSocket 劫持
	frontendConn, err := httputil.NewUpgrader(h.allowedOrigins).Upgrade(w, r, nil)
	if err != nil {
		h.logger.Errorf("[ws-proxy] upgrade failed: %v", err)
		return
	}
	defer func() { _ = frontendConn.Close() }()

	// 构建 Agent WebSocket URL：scheme 替换（http→ws, https→wss），node.Address 已含 scheme
	agentURL := strings.Replace(node.Address, "http://", "ws://", 1)
	agentURL = strings.Replace(agentURL, "https://", "wss://", 1)
	agentURL += fmt.Sprintf("/api/v1/sessions/%s/ws", sessionID)

	if err := httputil.ValidateTargetURL(agentURL, h.allowPrivateNetworks); err != nil {
		h.logger.Errorf("[ws-proxy] invalid agent url %s: %v", agentURL, err)
		return
	}

	// 使用 gorilla/websocket 作为客户端连接 Agent，携带 Auth token
	dialHeader := http.Header{}
	if h.authToken != "" {
		dialHeader.Set("Authorization", "Bearer "+h.authToken)
	}
	agentConn, _, err := gorillaws.DefaultDialer.Dial(agentURL, dialHeader)
	if err != nil {
		h.logger.Errorf("[ws-proxy] dial agent %s failed: %v", agentURL, err)
		return
	}
	defer func() { _ = agentConn.Close() }()

	if err := httputil.RelayWebSocket(frontendConn, agentConn); err != nil &&
		!gorillaws.IsCloseError(err, gorillaws.CloseNormalClosure, gorillaws.CloseGoingAway) {
		h.logger.Errorf("[ws-proxy] relay failed: %v", err)
	}
}
