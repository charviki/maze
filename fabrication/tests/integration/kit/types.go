package kit

import "encoding/json"

// HostInfo 对应 Manager API 返回的 Host 信息
type HostInfo struct {
	Name          string   `json:"name"`
	DisplayName   string   `json:"display_name,omitempty"`
	Tools         []string `json:"tools"`
	Status        string   `json:"status"`
	Address       string   `json:"address,omitempty"`
	SessionCount  int      `json:"session_count"`
	LastHeartbeat string   `json:"last_heartbeat,omitempty"`
	ErrorMsg      string   `json:"error_msg,omitempty"`
	RetryCount    int      `json:"retry_count"`
}

// SessionState 对应 Agent 保存的 Session 状态
type SessionState struct {
	SessionName      string            `json:"session_name"`
	Pipeline         json.RawMessage   `json:"pipeline"`
	RestoreStrategy  string            `json:"restore_strategy"`
	RestoreCommand   string            `json:"restore_command,omitempty"`
	WorkingDir       string            `json:"working_dir"`
	TemplateID       string            `json:"template_id,omitempty"`
	CLISessionID     string            `json:"cli_session_id,omitempty"`
	EnvSnapshot      map[string]string `json:"env_snapshot"`
	TerminalSnapshot string            `json:"terminal_snapshot"`
	SavedAt          string            `json:"saved_at"`
}

// APIResponse 是 Manager API 的标准响应格式
type APIResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// CreateHostRequest 创建 Host 的请求体
type CreateHostRequest struct {
	Name        string            `json:"name"`
	Tools       []string          `json:"tools"`
	DisplayName string            `json:"display_name,omitempty"`
	Resources   map[string]string `json:"resources,omitempty"`
}
