package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charviki/maze-cradle/configutil"
)

// 全局配置结构体
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Workspace  WorkspaceConfig  `yaml:"workspace"`
	Docker     DockerConfig     `yaml:"docker"`
	Runtime    RuntimeConfig    `yaml:"runtime"`
	Kubernetes KubernetesConfig `yaml:"kubernetes"`
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

	// Runtime / Kubernetes 环境变量覆盖
	if v := os.Getenv("AGENT_MANAGER_RUNTIME_TYPE"); v != "" {
		cfg.Runtime.Type = v
	}
	if v := os.Getenv("AGENT_MANAGER_KUBERNETES_NAMESPACE"); v != "" {
		cfg.Kubernetes.Namespace = v
	}
	if v := os.Getenv("AGENT_MANAGER_KUBERNETES_KUBECONFIG"); v != "" {
		cfg.Kubernetes.Kubeconfig = v
	}
	if v := os.Getenv("AGENT_MANAGER_KUBERNETES_AGENT_IMAGE_PREFIX"); v != "" {
		cfg.Kubernetes.AgentImagePrefix = v
	}
	if v := os.Getenv("AGENT_MANAGER_KUBERNETES_AGENT_IMAGE_TAG"); v != "" {
		cfg.Kubernetes.AgentImageTag = v
	}
	if v := os.Getenv("AGENT_MANAGER_KUBERNETES_IMAGE_PULL_POLICY"); v != "" {
		cfg.Kubernetes.ImagePullPolicy = v
	}
	if v := os.Getenv("AGENT_MANAGER_KUBERNETES_IMAGE_PULL_SECRET"); v != "" {
		cfg.Kubernetes.ImagePullSecret = v
	}
	if v := os.Getenv("AGENT_MANAGER_KUBERNETES_SERVICE_ACCOUNT"); v != "" {
		cfg.Kubernetes.ServiceAccount = v
	}
	if v := os.Getenv("AGENT_MANAGER_KUBERNETES_PVC_STORAGE_CLASS"); v != "" {
		cfg.Kubernetes.PVCStorageClass = v
	}
	if v := os.Getenv("AGENT_MANAGER_KUBERNETES_PVC_SIZE"); v != "" {
		cfg.Kubernetes.PVCSize = v
	}
	if v := os.Getenv("AGENT_MANAGER_KUBERNETES_MANAGER_ADDR"); v != "" {
		cfg.Kubernetes.ManagerAddr = v
	}
	if v := os.Getenv("AGENT_MANAGER_KUBERNETES_VOLUME_TYPE"); v != "" {
		cfg.Kubernetes.VolumeType = v
	}
	if v := os.Getenv("AGENT_MANAGER_KUBERNETES_HOST_PATH_BASE"); v != "" {
		cfg.Kubernetes.HostPathBase = v
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
		cfg.Workspace.BaseDir = filepath.Join(home, ".maze", "docker", "agents")
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
	if strings.HasPrefix(cfg.Kubernetes.HostPathBase, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			cfg.Kubernetes.HostPathBase = filepath.Join(home, cfg.Kubernetes.HostPathBase[2:])
		}
	}

	return nil
}
