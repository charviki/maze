package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	ListenAddr     string   `yaml:"listen_addr"`
	GRPCAddr       string   `yaml:"grpc_addr"`
	AuthToken      string   `yaml:"auth_token"`
	Name           string   `yaml:"name"`
	ExternalAddr   string   `yaml:"external_addr"`
	AdvertisedAddr string   `yaml:"advertised_addr"`
	AllowedOrigins []string `yaml:"allowed_origins"`
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
	Addr              string `yaml:"addr"`
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
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Load 从指定路径加载 YAML 配置文件，依次执行解析→环境变量覆盖→校验
func Load(path string) (*Config, error) {
	var cfg Config
	if err := configutil.LoadYAML(path, &cfg); err != nil {
		return nil, err
	}
	applyEnvOverrides(&cfg)
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("AGENT_SERVER_LISTEN_ADDR"); v != "" {
		cfg.Server.ListenAddr = v
	}
	if v := os.Getenv("AGENT_GRPC_ADDR"); v != "" {
		cfg.Server.GRPCAddr = v
	}
	if v := os.Getenv("AGENT_SERVER_AUTH_TOKEN"); v != "" {
		cfg.Server.AuthToken = v
	}
	if v := os.Getenv("AGENT_NAME"); v != "" {
		cfg.Server.Name = v
	}
	if v := os.Getenv("AGENT_EXTERNAL_ADDR"); v != "" {
		cfg.Server.ExternalAddr = v
	}
	if v := os.Getenv("AGENT_ADVERTISED_ADDR"); v != "" {
		cfg.Server.AdvertisedAddr = v
	}
	if v := os.Getenv("AGENT_TMUX_SOCKET_PATH"); v != "" {
		cfg.Tmux.SocketPath = v
	}
	if v := os.Getenv("AGENT_TMUX_DEFAULT_SHELL"); v != "" {
		cfg.Tmux.DefaultShell = v
	}
	if v := os.Getenv("AGENT_TERMINAL_DEFAULT_LINES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Terminal.DefaultLines = n
		}
	}
	if v := os.Getenv("AGENT_CONTROLLER_ADDR"); v != "" {
		cfg.Controller.Addr = v
		cfg.Controller.Enabled = true
	}
	if v := os.Getenv("AGENT_CONTROLLER_ENABLED"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Controller.Enabled = b
		}
	}
	if v := os.Getenv("AGENT_CONTROLLER_HEARTBEAT_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Controller.HeartbeatInterval = n
		}
	}
	if v := os.Getenv("AGENT_CONTROLLER_AUTH_TOKEN"); v != "" {
		cfg.Controller.AuthToken = v
	}
	if v := os.Getenv("AGENT_WORKSPACE_ROOT_DIR"); v != "" {
		cfg.Workspace.RootDir = v
	}
	if v := os.Getenv("AGENT_AUTOSAVE_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.AutoSave.Interval = n
		}
	}
	if v := os.Getenv("AGENT_SERVER_ALLOWED_ORIGINS"); v != "" {
		cfg.Server.AllowedOrigins = strings.Split(v, ",")
	}
}

// validate 校验配置完整性，对未设置必填字段填充默认值
func validate(cfg *Config) error {
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
		defaultRootDir := "/home/agent"
		cfg.Workspace.RootDir = defaultRootDir
	}
	if cfg.Workspace.StateDir == "" {
		cfg.Workspace.StateDir = filepath.Join(cfg.Workspace.RootDir, ".session-state")
	}
	if cfg.AutoSave.Interval <= 0 {
		cfg.AutoSave.Interval = 60
	}
	return nil
}
