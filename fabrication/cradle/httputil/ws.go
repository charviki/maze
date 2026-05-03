package httputil

import (
	gorillaws "github.com/gorilla/websocket"
)

// NewUpgrader 基于统一的 Origin 校验规则创建 WebSocket upgrader。
func NewUpgrader(origins []string) *gorillaws.Upgrader {
	return &gorillaws.Upgrader{
		CheckOrigin: CheckOrigin(origins),
	}
}
