package config

import (
	"path/filepath"

	"github.com/charviki/maze-cradle/configutil"
)

// Config 全局配置结构体，包含服务端、tmux、终端、控制器、工作区和自动保存配置
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Tmux       TmuxConfig       `yaml:"tmux"`
	Terminal   TerminalConfig   `yaml:"terminal"`
	Controller ControllerConfig `yaml:"controller"`
	Workspace  WorkspaceConfig  `yaml:"workspace"`
	AutoSave   AutoSaveConfig   `yaml:"autosave"`
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	configutil.ServerConfig `yaml:",inline"`
	GRPCAddr                string `yaml:"grpc_addr"`
	Name                    string `yaml:"name"`
	ExternalAddr            string `yaml:"external_addr"`
	AdvertisedAddr          string `yaml:"advertised_addr"`
}

// AllowedOrigins 返回经过去空白处理的来源白名单。
func (c *Config) AllowedOrigins() []string {
	return c.Server.Origins()
}

// IsDevMode 当鉴权令牌为空时视为开发模式。
func (c *Config) IsDevMode() bool {
	return c.Server.IsDevMode()
}

// TmuxConfig tmux 会话管理配置
type TmuxConfig struct {
	SocketPath   string `yaml:"socket_path"`
	DefaultShell string `yaml:"default_shell"`
}

// TerminalConfig 终端输出相关配置
type TerminalConfig struct {
	DefaultLines int `yaml:"default_lines"`
}

// ControllerConfig Agent Manager 控制器连接配置，用于心跳注册和上报
type ControllerConfig struct {
	// Addr Manager HTTP 地址（用于兼容/回退），格式: http://host:port
	Addr string `yaml:"addr"`
	// GRPCAddr Manager gRPC 地址（用于心跳注册/上报），格式: host:port。
	// 若为空，heartbeat 启动时从 Addr 推导（取 host + :9090）。
	GRPCAddr          string `yaml:"grpc_addr"`
	Enabled           bool   `yaml:"enabled"`
	HeartbeatInterval int    `yaml:"heartbeat_interval"`
	AuthToken         string `yaml:"auth_token"`
}

// WorkspaceConfig Agent 工作区根目录配置
type WorkspaceConfig struct {
	RootDir  string `yaml:"root_dir"`
	StateDir string `yaml:"state_dir"`
}

// AutoSaveConfig 自动保存配置
type AutoSaveConfig struct {
	Interval int `yaml:"interval"` // 自动保存间隔（秒），默认 60
}

// LoadFromExe 从可执行文件所在目录及当前工作目录搜索配置文件并加载。支持可选的文件名参数
func LoadFromExe(filename ...string) (*Config, error) {
	var cfg Config
	if _, err := configutil.LoadFromExe(&cfg, filename...); err != nil {
		return nil, err
	}
	applyEnvOverrides(&cfg)
	validate(&cfg)
	return &cfg, nil
}

// Load 从指定路径加载 YAML 配置文件，依次执行解析→环境变量覆盖→校验
func Load(path string) (*Config, error) {
	var cfg Config
	if err := configutil.LoadYAML(path, &cfg); err != nil {
		return nil, err
	}
	applyEnvOverrides(&cfg)
	validate(&cfg)
	return &cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if err := configutil.ApplyEnvOverrides("AGENT", cfg); err != nil {
		panic(err)
	}
	// 以下环境变量是历史兼容别名，不遵循结构路径命名，需显式保留。
	configutil.ApplyStringOverride(&cfg.Server.GRPCAddr, "AGENT_GRPC_ADDR")
	configutil.ApplyStringOverride(&cfg.Server.Name, "AGENT_NAME")
	configutil.ApplyStringOverride(&cfg.Server.ExternalAddr, "AGENT_EXTERNAL_ADDR")
	configutil.ApplyStringOverride(&cfg.Server.AdvertisedAddr, "AGENT_ADVERTISED_ADDR")
	// 历史兼容别名：不遵循结构路径命名，需显式保留。
	configutil.ApplyStringOverride(&cfg.Controller.GRPCAddr, "AGENT_CONTROLLER_GRPC_ADDR")
	if cfg.Controller.Addr != "" {
		// 只要指定 controller 地址，就默认开启注册/心跳，减少部署时的重复开关配置。
		cfg.Controller.Enabled = true
	}
}

// validate 校验配置完整性，对未设置必填字段填充默认值
func validate(cfg *Config) {
	if cfg.Server.ListenAddr == "" {
		cfg.Server.ListenAddr = ":8080"
	}
	if cfg.Tmux.DefaultShell == "" {
		cfg.Tmux.DefaultShell = "/bin/bash"
	}
	if cfg.Terminal.DefaultLines <= 0 {
		cfg.Terminal.DefaultLines = 50
	}
	if cfg.Controller.HeartbeatInterval <= 0 {
		cfg.Controller.HeartbeatInterval = 10
	}
	if cfg.Workspace.RootDir == "" {
		cfg.Workspace.RootDir = "/home/agent"
	}
	cfg.Workspace.RootDir = configutil.ExpandHomePath(cfg.Workspace.RootDir)
	if cfg.Workspace.StateDir == "" {
		cfg.Workspace.StateDir = filepath.Join(cfg.Workspace.RootDir, ".session-state")
	}
	if cfg.AutoSave.Interval <= 0 {
		cfg.AutoSave.Interval = 60
	}
}
