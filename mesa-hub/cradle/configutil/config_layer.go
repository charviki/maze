package configutil

// ConfigLayer 统一配置层结构，Template/Node/Session 都使用同一套
type ConfigLayer struct {
	Env   map[string]string `json:"env"`
	Files []ConfigFile      `json:"files"`
}

// ConfigFile 配置文件项
type ConfigFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// SessionSchema 定义创建 Session 时需要用户填写的字段
type SessionSchema struct {
	EnvDefs  []EnvDef  `json:"env_defs"`
	FileDefs []FileDef `json:"file_defs"`
}

// EnvDef 环境变量定义
type EnvDef struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Required    bool   `json:"required"`
	Placeholder string `json:"placeholder"`
	Sensitive   bool   `json:"sensitive"`
}

// FileDef 配置文件定义
type FileDef struct {
	Path           string `json:"path"`
	Label          string `json:"label"`
	Required       bool   `json:"required"`
	DefaultContent string `json:"default_content"`
}
