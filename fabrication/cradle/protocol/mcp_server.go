package protocol

import "time"

// MCPServer MCP 服务器配置
type MCPServer struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Command   string            `json:"command,omitempty"`
	URL       string            `json:"url,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}
