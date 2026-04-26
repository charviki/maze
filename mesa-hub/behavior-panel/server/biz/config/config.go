package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charviki/maze-cradle/configutil"
)

// 全局配置结构体
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Workspace WorkspaceConfig `yaml:"workspace"`
}

// 工作区配置
type WorkspaceConfig struct {
	BaseDir string `yaml:"base_dir"`
}

// HTTP 服务配置
type ServerConfig struct {
	ListenAddr           string   `yaml:"listen_addr"`
	AuthToken            string   `yaml:"auth_token"`
	AllowedOrigins       []string `yaml:"allowed_origins"`
	AllowPrivateNetworks bool     `yaml:"allow_private_networks"`
}

// AllowedOrigins 返回配置中的允许来源列表
func (c *Config) AllowedOrigins() []string {
	return c.Server.AllowedOrigins
}

// IsDevMode 当 auth_token 为空时视为开发模式
func (c *Config) IsDevMode() bool {
	return c.Server.AuthToken == ""
}

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

// 用环境变量覆盖 YAML 配置值
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("AGENT_MANAGER_SERVER_LISTEN_ADDR"); v != "" {
		cfg.Server.ListenAddr = v
	}
	if v := os.Getenv("AGENT_MANAGER_SERVER_AUTH_TOKEN"); v != "" {
		cfg.Server.AuthToken = v
	}
	if v := os.Getenv("AGENT_MANAGER_WORKSPACE_BASE_DIR"); v != "" {
		cfg.Workspace.BaseDir = v
	}
	if v := os.Getenv("AGENT_MANAGER_SERVER_ALLOWED_ORIGINS"); v != "" {
		cfg.Server.AllowedOrigins = strings.Split(v, ",")
	}
	if v := os.Getenv("AGENT_MANAGER_ALLOW_PRIVATE_NETWORKS"); v != "" {
		cfg.Server.AllowPrivateNetworks = v == "true" || v == "1"
	}
}

// 校验配置完整性并填充默认值
func validate(cfg *Config) error {
	if cfg.Server.ListenAddr == "" {
		cfg.Server.ListenAddr = ":8080"
	}
	if cfg.Workspace.BaseDir == "" {
		defaultBaseDir := "/root"
		home, err := os.UserHomeDir()
		if err != nil {
			home = defaultBaseDir
		}
		cfg.Workspace.BaseDir = filepath.Join(home, "agents")
	}
	return nil
}
