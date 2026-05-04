package service

// Session 代表一个 tmux 会话的运行时视图
type Session struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	WindowCount int    `json:"window_count"`
}

// ConfigItem 表示一个环境变量或配置键值对
type ConfigItem struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ConfigScope 区分配置文件的作用域：全局或项目级
type ConfigScope string

const (
	// ConfigScopeGlobal 表示用户主目录下的全局配置
	ConfigScopeGlobal ConfigScope = "global"
	// ConfigScopeProject 表示工作目录下的项目级配置
	ConfigScopeProject ConfigScope = "project"
)

// ConfigFileSnapshot 表示配置文件在某个时刻的内容和指纹
type ConfigFileSnapshot struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Exists  bool   `json:"exists"`
	Hash    string `json:"hash"`
}

// TemplateConfigView 展示某个模板关联的配置文件列表
type TemplateConfigView struct {
	TemplateID string               `json:"template_id"`
	Scope      ConfigScope          `json:"scope"`
	Files      []ConfigFileSnapshot `json:"files"`
}

// SessionConfigView 展示某个会话关联的配置文件列表
type SessionConfigView struct {
	SessionID  string               `json:"session_id"`
	TemplateID string               `json:"template_id"`
	WorkingDir string               `json:"working_dir"`
	Scope      ConfigScope          `json:"scope"`
	Files      []ConfigFileSnapshot `json:"files"`
}

// ConfigFileUpdate 表示前端提交的单个配置文件变更
type ConfigFileUpdate struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	BaseHash string `json:"base_hash"`
}

// SaveConfigRequest 是保存配置文件的请求载荷
type SaveConfigRequest struct {
	Files []ConfigFileUpdate `json:"files"`
}

// ConfigConflict 描述单个文件的乐观并发冲突详情
type ConfigConflict struct {
	Path        string `json:"path"`
	CurrentHash string `json:"current_hash"`
}
