package model

// Session tmux 会话信息
type Session struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	WindowCount int    `json:"window_count"`
}

// ConfigItem Session 创建时的配置项（环境变量或文件）
type ConfigItem struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CreateSessionRequest 创建 Session 的请求体
type CreateSessionRequest struct {
	Name            string       `json:"name"`
	Command         string       `json:"command"`
	WorkingDir      string       `json:"working_dir,omitempty"`
	SessionConfs    []ConfigItem `json:"session_confs"`
	RestoreStrategy string       `json:"restore_strategy,omitempty"`
	TemplateID      string       `json:"template_id,omitempty"`
}

// SendInputRequest 向 Session 发送命令的请求体
type SendInputRequest struct {
	Command string `json:"command"`
}

type SendSignalRequest struct {
	Signal string `json:"signal"`
}

// TerminalOutput 终端输出内容
type TerminalOutput struct {
	SessionID string `json:"session_id"`
	Lines     int    `json:"lines"`
	Output    string `json:"output"`
}

// ConfigScope 表示配置文件所属作用域。
// global 直接映射到 Agent 节点上的全局文件，project 映射到 session 工作目录内的项目级文件。
type ConfigScope string

const (
	ConfigScopeGlobal  ConfigScope = "global"
	ConfigScopeProject ConfigScope = "project"
)

// ConfigFileSnapshot 表示某个固定路径配置文件在读取时的真实状态。
type ConfigFileSnapshot struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Exists  bool   `json:"exists"`
	Hash    string `json:"hash"`
}

// TemplateConfigView 为模板配置页返回真实全局文件快照。
type TemplateConfigView struct {
	TemplateID string               `json:"template_id"`
	Scope      ConfigScope          `json:"scope"`
	Files      []ConfigFileSnapshot `json:"files"`
}

// SessionConfigView 为 session 配置页返回工作目录内真实项目文件快照。
type SessionConfigView struct {
	SessionID  string               `json:"session_id"`
	TemplateID string               `json:"template_id"`
	WorkingDir string               `json:"working_dir"`
	Scope      ConfigScope          `json:"scope"`
	Files      []ConfigFileSnapshot `json:"files"`
}

// ConfigFileUpdate 表示保存单个配置文件时提交的内容与打开页面时的基线 hash。
type ConfigFileUpdate struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	BaseHash string `json:"base_hash"`
}

// SaveConfigRequest 为模板/session 配置保存的统一请求体。
type SaveConfigRequest struct {
	Files []ConfigFileUpdate `json:"files"`
}

// ConfigConflict 用于返回发生外部变更的文件列表。
type ConfigConflict struct {
	Path        string `json:"path"`
	CurrentHash string `json:"current_hash"`
}
