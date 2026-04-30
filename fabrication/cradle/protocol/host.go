package protocol

import "time"

// ToolConfig 工具配置，描述一个可选配的供应商工具
type ToolConfig struct {
	// ID 工具唯一标识（如 "claude", "go"）
	ID string `json:"id"`
	// Image 供应商镜像名（如 "maze-deps-claude:latest"）
	Image string `json:"image"`
	// SourcePath 供应商镜像中的源路径（如 "/opt/claude"）
	SourcePath string `json:"source_path"`
	// DestPath 目标镜像中的安装路径（如 "/opt/claude"）
	DestPath string `json:"dest_path"`
	// BinPaths 需要加入 PATH 的 bin 目录列表
	BinPaths []string `json:"bin_paths"`
	// EnvVars 额外环境变量（key=value）
	EnvVars map[string]string `json:"env_vars,omitempty"`
	// Description 工具描述
	Description string `json:"description"`
	// Category 工具分类（如 "cli", "language"）
	Category string `json:"category"`
}

// ResourceLimits 容器资源限制
type ResourceLimits struct {
	// CPULimit CPU 核心数（如 "1", "2", "0.5"），对应 Docker --cpus
	CPULimit string `json:"cpu_limit,omitempty"`
	// MemoryLimit 内存上限（如 "512m", "1g", "4g"），对应 Docker --memory
	MemoryLimit string `json:"memory_limit,omitempty"`
}

// HostDeploySpec 运行时无关的 Host 部署规格
// Handler 构建此规格，Runtime 实现负责翻译为具体运行时操作
type HostDeploySpec struct {
	// Name Host 唯一标识名称
	Name string `json:"name"`
	// Tools 选配的工具 ID 列表
	Tools []string `json:"tools"`
	// Resources 资源限制（可选）
	Resources ResourceLimits `json:"resources,omitempty"`
	// AuthToken Manager 的 auth token，用于 Agent 注册和心跳
	AuthToken string `json:"-"`
}

// CreateHostRequest 创建 Host 请求
type CreateHostRequest struct {
	// Name Host 唯一标识名称
	Name string `json:"name"`
	// Tools 选配的工具 ID 列表
	Tools []string `json:"tools"`
	// DisplayName 显示名称（可选）
	DisplayName string `json:"display_name,omitempty"`
	// Resources 资源限制（可选）
	Resources ResourceLimits `json:"resources,omitempty"`
}

// CreateHostResponse 创建 Host 响应
type CreateHostResponse struct {
	// Name Host 名称
	Name string `json:"name"`
	// Tools 选配的工具列表
	Tools []string `json:"tools"`
	// ImageTag 构建的镜像 tag
	ImageTag string `json:"image_tag"`
	// ContainerID 容器 ID
	ContainerID string `json:"container_id"`
	// Status 创建状态（"building", "running", "failed"）
	Status string `json:"status"`
	// BuildLog 构建日志（可选）
	BuildLog string `json:"build_log,omitempty"`
}

// ContainerInfo 容器信息
type ContainerInfo struct {
	// ID 容器 ID
	ID string `json:"id"`
	// Name 容器名称
	Name string `json:"name"`
	// Status 容器状态（"running", "exited", "created" 等）
	Status string `json:"status"`
	// Image 镜像名
	Image string `json:"image"`
	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`
}
