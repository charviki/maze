package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	gorillaws "github.com/gorilla/websocket"
	"github.com/hertz-contrib/websocket"

	"github.com/charviki/maze-cradle/httputil"
	"github.com/charviki/maze-cradle/protocol"
)

// ProxyWebSocket WebSocket 双向代理：前端 ↔ Manager ↔ Agent。
// Manager 将前端的 WebSocket 连接升级后，建立到 Agent 的 WebSocket 客户端连接，
// 然后在两个连接之间双向转发消息（Binary 和 Text）。
// 任意一端断开时，另一端也自动关闭，确保资源不泄漏。
func (h *SessionProxyHandler) ProxyWebSocket(_ context.Context, c *app.RequestContext) {
	nodeName := c.Param("name")
	sessionID := c.Param("id")
	if nodeName == "" || sessionID == "" {
		httputil.Error(c, http.StatusBadRequest, "node name and session id are required")
		return
	}

	node := h.registry.Get(nodeName)
	if node == nil {
		httputil.Error(c, http.StatusNotFound, "node not found")
		return
	}

	h.auditLog.Log(protocol.AuditLogEntry{
		Operator:       auditOperator,
		TargetNode:     nodeName,
		Action:         "websocket_connect",
		PayloadSummary: fmt.Sprintf("session=%s", sessionID),
		Result:         "connecting",
		StatusCode:     http.StatusSwitchingProtocols,
	})

	// 使用配置化的 Origin 校验替代硬编码的"允许所有来源"，避免跨站 WebSocket 劫持
	upgrader := websocket.HertzUpgrader{CheckOrigin: httputil.CheckOrigin(h.allowedOrigins)}

	err := upgrader.Upgrade(c, func(frontendConn *websocket.Conn) {
		defer frontendConn.Close()

		// 构建 Agent WebSocket URL：scheme 替换（http→ws, https→wss），node.Address 已含 scheme
		agentURL := strings.Replace(node.Address, "http://", "ws://", 1)
		agentURL = strings.Replace(agentURL, "https://", "wss://", 1)
		agentURL += fmt.Sprintf("/api/v1/sessions/%s/ws", sessionID)

		if err := validateAgentURL(agentURL, h.allowPrivateNetworks); err != nil {
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
		defer agentConn.Close()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			for {
				msgType, msg, err := agentConn.ReadMessage()
				if err != nil {
					return
				}
				// gorilla websocket 的 TextMessage=1, BinaryMessage=2
				// hertz-contrib websocket 的 TextMessage=1, BinaryMessage=2
				// 两者值相同，可直接传递
				if err := frontendConn.WriteMessage(int(msgType), msg); err != nil {
					return
				}
			}
		}()

		go func() {
			defer wg.Done()
			for {
				msgType, msg, err := frontendConn.ReadMessage()
				if err != nil {
					return
				}
				if err := agentConn.WriteMessage(int(msgType), msg); err != nil {
					return
				}
			}
		}()

		wg.Wait()
	})

	if err != nil {
		h.logger.Errorf("[ws-proxy] upgrade failed: %v", err)
	}
}
