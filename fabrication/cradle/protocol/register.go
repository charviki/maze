package protocol

import "time"

// AgentCapabilities 声明 Agent 支持的能力，用于 Director Core 调度决策
type AgentCapabilities struct {
	// SupportedTemplates 该 Agent 支持的模板 ID 列表（如 "claude", "bash"）
	SupportedTemplates []string `json:"supported_templates"`
	// MaxSessions 该 Agent 可同时运行的最大 Session 数量
	MaxSessions int `json:"max_sessions"`
	// Tools 该 Agent 可用的工具列表（如 "tmux", "filesystem"）
	Tools []string `json:"tools"`
}

// AgentStatus Agent 当前运行状态快照
type AgentStatus struct {
	// ActiveSessions 当前活跃 Session 数量
	ActiveSessions int `json:"active_sessions"`
	// CPUUsage CPU 使用率百分比（0-100）
	CPUUsage float64 `json:"cpu_usage"`
	// MemoryUsageMB 内存使用量（MB）
	MemoryUsageMB float64 `json:"memory_usage_mb"`
	// WorkspaceRoot 工作区根目录路径
	WorkspaceRoot string `json:"workspace_root"`
	// SessionDetails 各 Session 的详细状态（心跳时上报）
	SessionDetails []SessionDetail `json:"session_details,omitempty"`
	// LocalConfig Agent 本地记忆的只读视图（工作目录和环境变量）
	LocalConfig *LocalAgentConfig `json:"local_config,omitempty"`
}

// SessionDetail 单个 Session 的状态信息
type SessionDetail struct {
	// ID Session 标识
	ID string `json:"id"`
	// Template 创建该 Session 时使用的模板 ID
	Template string `json:"template"`
	// WorkingDir Session 工作目录
	WorkingDir string `json:"working_dir"`
	// UptimeSeconds Session 已运行时长（秒）
	UptimeSeconds int64 `json:"uptime_seconds"`
}

// AgentMetadata Agent 的静态元数据，注册时上报一次
type AgentMetadata struct {
	// Version Agent 程序版本号
	Version string `json:"version"`
	// Hostname Agent 所在容器/主机的主机名
	Hostname string `json:"hostname"`
	// StartedAt Agent 启动时间
	StartedAt time.Time `json:"started_at"`
}

// LocalAgentConfig Agent 本地记忆配置（存储在 /home/agent/.maze/config.json）
type LocalAgentConfig struct {
	// WorkingDir 基础工作目录展示值，由服务端配置决定，只读。
	WorkingDir string `json:"working_dir"`
	// Env 默认环境变量
	Env map[string]string `json:"env"`
}

// RegisterRequest Agent 向 Director Core 发送的增强版注册请求
type RegisterRequest struct {
	// Name Agent 节点唯一标识（全局唯一）
	Name string `json:"name"`
	// Address Agent 内部监听地址（Director Core 用于回调 Agent API）
	Address string `json:"address"`
	// ExternalAddr Agent 的外部可访问地址（前端可能需要）
	ExternalAddr string `json:"external_addr"`
	// GrpcAddress Agent gRPC 监听地址（Director Core 用于 gRPC 回调）
	GrpcAddress string `json:"grpc_address"`
	// Capabilities Agent 能力声明
	Capabilities AgentCapabilities `json:"capabilities"`
	// Status Agent 当前状态快照
	Status AgentStatus `json:"status"`
	// Metadata Agent 静态元数据
	Metadata AgentMetadata `json:"metadata"`
}

// HeartbeatRequest Agent 向 Director Core 发送的增强版心跳请求
type HeartbeatRequest struct {
	// Name Agent 节点唯一标识
	Name string `json:"name"`
	// Status Agent 当前状态快照（含 Session 详情和本地配置）
	Status AgentStatus `json:"status"`
}
