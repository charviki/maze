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
	Docker    DockerConfig    `yaml:"docker"`
}

// 工作区配置
type WorkspaceConfig struct {
	// BaseDir 宿主机上的持久化根目录（用于 docker -v 挂载路径）
	BaseDir string `yaml:"base_dir"`
	// MountDir Manager 容器内的挂载路径（用于 os.MkdirAll / os.RemoveAll）
	MountDir string `yaml:"mount_dir"`
}

// Docker 配置（用于动态创建 Host 容器）
type DockerConfig struct {
	// SocketPath Docker socket 路径
	SocketPath string `yaml:"socket_path"`
	// Network 默认 Docker 网络名（Host 容器加入此网络以访问 Manager）
	Network string `yaml:"network"`
	// BuildContextDir 构建上下文目录（存放临时 Dockerfile）
	BuildContextDir string `yaml:"build_context_dir"`
	// AgentBaseImage Agent 基础镜像名（含 agent 二进制和 entrypoint）
	AgentBaseImage string `yaml:"agent_base_image"`
	// ManagerAddr Manager 在容器网络中的地址（Agent 通过此地址注册心跳）
	ManagerAddr string `yaml:"manager_addr"`
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
	if v := os.Getenv("AGENT_MANAGER_WORKSPACE_MOUNT_DIR"); v != "" {
		cfg.Workspace.MountDir = v
	}
	if v := os.Getenv("AGENT_MANAGER_SERVER_ALLOWED_ORIGINS"); v != "" {
		cfg.Server.AllowedOrigins = strings.Split(v, ",")
	}
	if v := os.Getenv("AGENT_MANAGER_ALLOW_PRIVATE_NETWORKS"); v != "" {
		cfg.Server.AllowPrivateNetworks = v == "true" || v == "1"
	}
	if v := os.Getenv("AGENT_MANAGER_DOCKER_SOCKET_PATH"); v != "" {
		cfg.Docker.SocketPath = v
	}
	if v := os.Getenv("AGENT_MANAGER_DOCKER_NETWORK"); v != "" {
		cfg.Docker.Network = v
	}
	if v := os.Getenv("AGENT_MANAGER_DOCKER_BUILD_CONTEXT_DIR"); v != "" {
		cfg.Docker.BuildContextDir = v
	}
	if v := os.Getenv("AGENT_MANAGER_DOCKER_AGENT_BASE_IMAGE"); v != "" {
		cfg.Docker.AgentBaseImage = v
	}
	if v := os.Getenv("AGENT_MANAGER_DOCKER_MANAGER_ADDR"); v != "" {
		cfg.Docker.ManagerAddr = v
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
	// 展开 ~/ 路径前缀，Docker volume 挂载要求绝对路径
	if strings.HasPrefix(cfg.Workspace.BaseDir, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			cfg.Workspace.BaseDir = filepath.Join(home, cfg.Workspace.BaseDir[2:])
		}
	}
	if cfg.Workspace.MountDir == "" {
		cfg.Workspace.MountDir = cfg.Workspace.BaseDir
	}
	if cfg.Docker.SocketPath == "" {
		cfg.Docker.SocketPath = "/var/run/docker.sock"
	}
	if cfg.Docker.ManagerAddr == "" {
		cfg.Docker.ManagerAddr = "http://agent-manager:8080"
	}
	return nil
}
