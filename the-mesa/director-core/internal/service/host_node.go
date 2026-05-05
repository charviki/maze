package service

import (
	"context"
	"fmt"
	"time"

	"github.com/charviki/maze-cradle/protocol"
)

const (
	// NodeStatusOnline 表示节点最近仍在发送心跳。
	NodeStatusOnline = "online"
	// NodeStatusOffline 表示节点已经超过心跳窗口未上报。
	NodeStatusOffline = "offline"
)

// Node 是 Host/Node 业务侧拥有的运行时对象。
// repository 只负责持久化与恢复，不拥有这个模型。
type Node struct {
	Name          string                     `json:"name"`
	Address       string                     `json:"address"`
	ExternalAddr  string                     `json:"external_addr"`
	GrpcAddress   string                     `json:"grpc_address"`
	AuthToken     string                     `json:"auth_token"`
	Status        string                     `json:"status"`
	RegisteredAt  time.Time                  `json:"registered_at"`
	LastHeartbeat time.Time                  `json:"last_heartbeat"`
	Capabilities  protocol.AgentCapabilities `json:"capabilities"`
	AgentStatus   protocol.AgentStatus       `json:"agent_status"`
	Metadata      protocol.AgentMetadata     `json:"metadata"`
}

// RefreshOfflineStatus 根据心跳时间更新节点在线态。
func (n *Node) RefreshOfflineStatus(now time.Time, offlineThreshold time.Duration) bool {
	if n == nil {
		return false
	}
	if now.Sub(n.LastHeartbeat) <= offlineThreshold || n.Status != NodeStatusOnline {
		return false
	}
	n.Status = NodeStatusOffline
	return true
}

// FormatNodeSummary 返回节点摘要，用于控制面日志。
func FormatNodeSummary(n *Node) string {
	return fmt.Sprintf("%s (%s) sessions=%d cpu=%.1f%% mem=%.0fMB status=%s",
		n.Name, n.Address, n.AgentStatus.ActiveSessions,
		n.AgentStatus.CPUUsage, n.AgentStatus.MemoryUsageMB, n.Status)
}

// BuildHostInfo 将 HostSpec 与运行时 Node 快照投影成 API 视图。
func BuildHostInfo(spec *protocol.HostSpec, node *Node) *protocol.HostInfo {
	if spec == nil {
		return nil
	}

	info := &protocol.HostInfo{HostSpec: *spec}
	if spec.Status != protocol.HostStatusDeploying || node == nil {
		return info
	}

	if node.Status == NodeStatusOnline {
		info.Status = protocol.HostStatusOnline
	} else {
		info.Status = protocol.HostStatusOffline
	}
	info.Address = node.Address
	info.SessionCount = node.AgentStatus.ActiveSessions
	if !node.LastHeartbeat.IsZero() {
		info.LastHeartbeat = node.LastHeartbeat.Format(time.RFC3339)
	}
	return info
}

// NodeRegistry 定义 Host/Node 域对节点注册表的最小依赖。
type NodeRegistry interface {
	StoreHostToken(ctx context.Context, name, token string) error
	ValidateHostToken(ctx context.Context, name, token string) (exists bool, matched bool, err error)
	RemoveHostToken(ctx context.Context, name string) error
	Register(ctx context.Context, req protocol.RegisterRequest) (*Node, error)
	Heartbeat(ctx context.Context, req protocol.HeartbeatRequest) (*Node, error)
	List(ctx context.Context) ([]*Node, error)
	Get(ctx context.Context, name string) (*Node, error)
	Delete(ctx context.Context, name string) (bool, error)
	GetNodeCount(ctx context.Context) (int, error)
	GetOnlineCount(ctx context.Context) (int, error)
}

// HostSpecRepository 定义 Host 规格的声明式持久化边界。
type HostSpecRepository interface {
	Create(ctx context.Context, spec *protocol.HostSpec) (bool, error)
	Get(ctx context.Context, name string) (*protocol.HostSpec, error)
	List(ctx context.Context) ([]*protocol.HostSpec, error)
	UpdateStatus(ctx context.Context, name, status, errMsg string) (bool, error)
	Delete(ctx context.Context, name string) (bool, error)
	IncrementRetry(ctx context.Context, name string) (bool, error)
}
