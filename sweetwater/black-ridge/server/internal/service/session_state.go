package service

import (
	"encoding/json"

	"github.com/charviki/maze-cradle/pipeline"
)

// SessionState 记录一个会话的管线配置和环境快照，用于跨重启恢复
type SessionState struct {
	SessionName      string            `json:"session_name"`
	Pipeline         pipeline.Pipeline `json:"pipeline"`
	RestoreStrategy  string            `json:"restore_strategy"`
	RestoreCommand   string            `json:"restore_command,omitempty"`
	WorkingDir       string            `json:"working_dir"`
	TemplateID       string            `json:"template_id,omitempty"`
	CLISessionID     string            `json:"cli_session_id,omitempty"`
	EnvSnapshot      map[string]string `json:"env_snapshot"`
	TerminalSnapshot string            `json:"terminal_snapshot"`
	SavedAt          string            `json:"saved_at"`
}

// ToJSON 将 SessionState 序列化为格式化 JSON
func (s *SessionState) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// FromJSON 从 JSON 数据反序列化填充 SessionState
func (s *SessionState) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}
