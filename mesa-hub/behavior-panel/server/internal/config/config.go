package config

import (
	"os"
	"path/filepath"

	"github.com/charviki/maze-cradle/configutil"
)

// Config 全局配置结构体，含 Server、Workspace、Docker、Runtime、Kubernetes、AuthDatabase、HostDatabase、Casbin 维度
type Config struct {
	Server       ServerConfig     `yaml:"server"`
	Workspace    WorkspaceConfig  `yaml:"workspace"`
	Docker       DockerConfig     `yaml:"docker"`
	Runtime      RuntimeConfig    `yaml:"runtime"`
	Kubernetes   KubernetesConfig `yaml:"kubernetes"`
	AuthDatabase DatabaseConfig   `yaml:"auth_database"`
	HostDatabase DatabaseConfig   `yaml:"host_database"`
	Casbin       CasbinConfig     `yaml:"casbin"`
	Authz        AuthzConfig      `yaml:"authz"`
}

// DatabaseConfig PostgreSQL 数据库连接配置
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// CasbinConfig Casbin RBAC 配置
type CasbinConfig struct {
	ModelPath string `yaml:"model_path"`
}

// AuthzConfig 权限系统启用与 bootstrap 配置。
type AuthzConfig struct {
	Enabled          bool   `yaml:"enabled"`
	AdminSubjectKey  string `yaml:"admin_subject_key"`
	AdminDisplayName string `yaml:"admin_display_name"`
}

// RuntimeConfig 运行时类型选择
type RuntimeConfig struct {
	// Type 运行时类型：docker（默认）或 kubernetes
	Type string `yaml:"type"`
}

// KubernetesConfig Kubernetes 运行时配置
type KubernetesConfig struct {
	// Namespace Manager 创建 Agent 资源的目标 namespace
	Namespace string `yaml:"namespace"`
	// Kubeconfig 外部 kubeconfig 文件路径，in-cluster 模式下留空
	Kubeconfig string `yaml:"kubeconfig"`
	// AgentImagePrefix Agent 镜像前缀（生产环境使用，本地动态构建时忽略）
	AgentImagePrefix string `yaml:"agent_image_prefix"`
	// AgentImageTag Agent 镜像 tag（生产环境使用，本地动态构建时忽略）
	AgentImageTag string `yaml:"agent_image_tag"`
	// ImagePullPolicy 镜像拉取策略：IfNotPresent / Always / Never
	ImagePullPolicy string `yaml:"image_pull_policy"`
	// ImagePullSecret 私有仓库认证 Secret 名称
	ImagePullSecret string `yaml:"image_pull_secret"`
	// ServiceAccount Agent Pod 使用的 ServiceAccount 名称
	ServiceAccount string `yaml:"service_account"`
	// PVCStorageClass Agent 持久卷 StorageClass
	PVCStorageClass string `yaml:"pvc_storage_class"`
	// PVCSize Agent 持久卷默认大小
	PVCSize string `yaml:"pvc_size"`
	// ManagerAddr Manager 在 K8s 集群内的 Service DNS
	ManagerAddr string `yaml:"manager_addr"`
	// VolumeType Agent 持久卷类型：pvc（默认，生产用）或 hostpath（本地开发用）
	VolumeType string `yaml:"volume_type"`
	// HostPathBase hostPath 模式下宿主机上的根目录
	HostPathBase string `yaml:"host_path_base"`
}

// WorkspaceConfig 工作区配置，指定 Manager 元数据和 Agent 工作目录的宿主机/容器路径
type WorkspaceConfig struct {
	// BaseDir Manager 元数据根目录（host_specs.json/nodes.json/audit.log/host_logs）
	BaseDir string `yaml:"base_dir"`
	// MountDir Manager 容器内对应的元数据挂载路径（用于文件操作）
	MountDir string `yaml:"mount_dir"`
}

// DockerConfig Docker 运行时配置，用于动态创建 Host 容器
type DockerConfig struct {
	// SocketPath Docker socket 路径
	SocketPath string `yaml:"socket_path"`
	// Network 默认 Docker 网络名（Host 容器加入此网络以访问 Manager）
	Network string `yaml:"network"`
	// BuildContextDir 构建上下文目录（存放临时 Dockerfile）
	BuildContextDir string `yaml:"build_context_dir"`
	// AgentBaseImage Agent 基础镜像名（含 agent 二进制和 entrypoint）
	AgentBaseImage string `yaml:"agent_base_image"`
	// AgentDataDir Docker 模式下 Agent 宿主机根目录；每个 Host 使用其下的 agents/<name> 子目录。
	AgentDataDir string `yaml:"agent_data_dir"`
	// ManagerAddr Manager 在容器网络中的地址（Agent 通过此地址注册心跳）
	ManagerAddr string `yaml:"manager_addr"`
}

// ServerConfig HTTP 服务配置，含监听地址、鉴权令牌、CORS 白名单
type ServerConfig struct {
	configutil.ServerConfig `yaml:",inline"`
	AllowPrivateNetworks    bool   `yaml:"allow_private_networks"`
	GRPCListenAddr          string `yaml:"grpc_listen_addr"`
}

// AllowedOrigins 返回配置中的允许来源列表
func (c *Config) AllowedOrigins() []string {
	return c.Server.Origins()
}

// IsDevMode 当 auth_token 为空时视为开发模式
func (c *Config) IsDevMode() bool {
	return c.Server.IsDevMode()
}

// LoadFromExe 搜索并加载配置文件（当前目录 → 可执行文件所在目录 → 上级目录），
// 填充到 Config 结构体，可选择指定文件名（默认 config.yaml）。
func LoadFromExe(filename ...string) (*Config, error) {
	var cfg Config
	if _, err := configutil.LoadFromExe(&cfg, filename...); err != nil {
		return nil, err
	}
	applyEnvOverrides(&cfg)
	validate(&cfg)
	return &cfg, nil
}

// Load 从指定路径加载 YAML 配置文件并填充到 Config 结构体
func Load(path string) (*Config, error) {
	var cfg Config
	if err := configutil.LoadYAML(path, &cfg); err != nil {
		return nil, err
	}
	applyEnvOverrides(&cfg)
	validate(&cfg)
	return &cfg, nil
}

// 用环境变量覆盖 YAML 配置值
func applyEnvOverrides(cfg *Config) {
	if err := configutil.ApplyEnvOverrides("AGENT_MANAGER", cfg); err != nil {
		panic(err)
	}
	// 该字段历史上使用顶层环境变量名，不走 server 前缀；显式保留以兼容既有部署脚本。
	configutil.ApplyBoolOverride(&cfg.Server.AllowPrivateNetworks, "AGENT_MANAGER_ALLOW_PRIVATE_NETWORKS")
}

// validate 校验配置完整性并填充默认值
func validate(cfg *Config) {
	if cfg.Server.ListenAddr == "" {
		cfg.Server.ListenAddr = ":8080"
	}
	if cfg.Workspace.BaseDir == "" {
		defaultBaseDir := "/root"
		home, err := os.UserHomeDir()
		if err != nil {
			home = defaultBaseDir
		}
		cfg.Workspace.BaseDir = filepath.Join(home, ".maze", "docker")
	}
	// 展开 ~/ 路径前缀，Docker volume 挂载要求绝对路径
	cfg.Workspace.BaseDir = configutil.ExpandHomePath(cfg.Workspace.BaseDir)
	if cfg.Workspace.MountDir == "" {
		cfg.Workspace.MountDir = cfg.Workspace.BaseDir
	}
	if cfg.Docker.SocketPath == "" {
		cfg.Docker.SocketPath = "/var/run/docker.sock"
	}
	if cfg.Docker.AgentDataDir == "" {
		// Docker 默认沿用统一目录模型：workspace.base_dir 负责 Manager 元数据，agents/ 子目录负责 Agent 工作目录。
		cfg.Docker.AgentDataDir = filepath.Join(cfg.Workspace.BaseDir, "agents")
	}
	cfg.Docker.AgentDataDir = configutil.ExpandHomePath(cfg.Docker.AgentDataDir)
	if cfg.Docker.ManagerAddr == "" {
		cfg.Docker.ManagerAddr = "http://agent-manager:8080"
	}

	// Kubernetes 运行时默认值：未显式配置时提供安全兜底
	if cfg.Runtime.Type == "" {
		cfg.Runtime.Type = "docker"
	}
	if cfg.Kubernetes.Namespace == "" {
		cfg.Kubernetes.Namespace = "default"
	}
	if cfg.Kubernetes.ImagePullPolicy == "" {
		cfg.Kubernetes.ImagePullPolicy = "IfNotPresent"
	}
	if cfg.Kubernetes.PVCSize == "" {
		cfg.Kubernetes.PVCSize = "10Gi"
	}
	if cfg.Kubernetes.AgentImageTag == "" {
		cfg.Kubernetes.AgentImageTag = "latest"
	}
	if cfg.Kubernetes.VolumeType == "" {
		cfg.Kubernetes.VolumeType = "pvc"
	}
	if cfg.Kubernetes.HostPathBase == "" {
		defaultHome := "/root"
		home, err := os.UserHomeDir()
		if err != nil {
			home = defaultHome
		}
		cfg.Kubernetes.HostPathBase = filepath.Join(home, ".maze", "kubernetes", "agents")
	}
	cfg.Kubernetes.HostPathBase = configutil.ExpandHomePath(cfg.Kubernetes.HostPathBase)

	// AuthDatabase 默认值
	if cfg.AuthDatabase.Port == 0 {
		cfg.AuthDatabase.Port = 5432
	}
	if cfg.AuthDatabase.Name == "" {
		cfg.AuthDatabase.Name = "maze_auth"
	}

	// HostDatabase 默认值：继承 AuthDatabase 的 host/port/user/password
	if cfg.HostDatabase.Host == "" {
		cfg.HostDatabase.Host = cfg.AuthDatabase.Host
	}
	if cfg.HostDatabase.Port == 0 {
		cfg.HostDatabase.Port = cfg.AuthDatabase.Port
	}
	if cfg.HostDatabase.Name == "" {
		cfg.HostDatabase.Name = "maze_host"
	}
	if cfg.HostDatabase.User == "" {
		cfg.HostDatabase.User = cfg.AuthDatabase.User
	}
	if cfg.HostDatabase.Password == "" {
		cfg.HostDatabase.Password = cfg.AuthDatabase.Password
	}

	// Authz 默认值
	if cfg.Authz.AdminSubjectKey == "" {
		cfg.Authz.AdminSubjectKey = "user:admin"
	}
	if cfg.Authz.AdminDisplayName == "" {
		cfg.Authz.AdminDisplayName = "admin"
	}
}
