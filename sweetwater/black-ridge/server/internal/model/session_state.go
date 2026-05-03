package model

import "encoding/json"

// SessionState Session 状态快照，用于持久化和恢复
type SessionState struct {
	SessionName      string            `json:"session_name"`
	Pipeline         Pipeline          `json:"pipeline"`
	RestoreStrategy  string            `json:"restore_strategy"` // auto / manual
	RestoreCommand   string            `json:"restore_command,omitempty"`
	WorkingDir       string            `json:"working_dir"`
	TemplateID       string            `json:"template_id,omitempty"`
	CLISessionID     string            `json:"cli_session_id,omitempty"`
	EnvSnapshot      map[string]string `json:"env_snapshot"`
	TerminalSnapshot string            `json:"terminal_snapshot"`
	SavedAt          string            `json:"saved_at"`
}

// ToJSON 将 SessionState 序列化为美化格式的 JSON 字节
func (s *SessionState) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// FromJSON 从 JSON 字节反序列化 SessionState
func (s *SessionState) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}
