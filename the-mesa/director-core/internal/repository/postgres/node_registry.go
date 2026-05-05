package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/charviki/maze-cradle/protocol"
	hostgen "github.com/charviki/maze/the-mesa/director-core/internal/repository/postgres/sqlc/host"
	"github.com/charviki/maze/the-mesa/director-core/internal/service"
)

const nodeOfflineThreshold = 30 * time.Second

// NodeRegistry 是 PostgreSQL/sqlc 驱动的节点注册中心实现。
type NodeRegistry struct {
	db hostgen.DBTX
}

// NewNodeRegistry 创建 PG 节点注册中心。
func NewNodeRegistry(db hostgen.DBTX) *NodeRegistry {
	return &NodeRegistry{db: db}
}

func (r *NodeRegistry) queries(ctx context.Context) *hostgen.Queries {
	return hostgen.New(hostExecutorFromContext(ctx, r.db))
}

// StoreHostToken 持久化 Host 令牌（upsert）。
func (r *NodeRegistry) StoreHostToken(ctx context.Context, name, token string) error {
	return r.queries(ctx).UpsertHostToken(ctx, hostgen.UpsertHostTokenParams{
		Name:  name,
		Token: token,
	})
}

// ValidateHostToken 校验 Host 令牌是否存在且匹配。
func (r *NodeRegistry) ValidateHostToken(ctx context.Context, name, token string) (bool, bool, error) {
	expected, err := r.queries(ctx).GetHostTokenByName(ctx, name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, false, nil
		}
		return false, false, err
	}
	return true, expected == token, nil
}

// RemoveHostToken 删除指定 Host 令牌。
func (r *NodeRegistry) RemoveHostToken(ctx context.Context, name string) error {
	return r.queries(ctx).DeleteHostTokenByName(ctx, name)
}

// Register 注册或更新节点信息（upsert），返回最新节点状态。
func (r *NodeRegistry) Register(ctx context.Context, req protocol.RegisterRequest) (*service.Node, error) {
	caps, _ := json.Marshal(req.Capabilities)
	status, _ := json.Marshal(req.Status)
	meta, _ := json.Marshal(req.Metadata)

	row, err := r.queries(ctx).UpsertNode(ctx, hostgen.UpsertNodeParams{
		Name:         req.Name,
		Address:      req.Address,
		ExternalAddr: req.ExternalAddr,
		GrpcAddress:  req.GrpcAddress,
		Capabilities: caps,
		AgentStatus:  status,
		Metadata:     meta,
	})
	if err != nil {
		return nil, err
	}
	n := nodeFromRow(row)
	return &n, nil
}

// Heartbeat 更新节点心跳和 agent_status，节点不存在时返回 nil, nil。
func (r *NodeRegistry) Heartbeat(ctx context.Context, req protocol.HeartbeatRequest) (*service.Node, error) {
	status, _ := json.Marshal(req.Status)
	row, err := r.queries(ctx).UpdateNodeHeartbeat(ctx, hostgen.UpdateNodeHeartbeatParams{
		Name:        req.Name,
		AgentStatus: status,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	n := nodeFromRow(row)
	return &n, nil
}

// List 返回所有节点，自动检测离线状态。
func (r *NodeRegistry) List(ctx context.Context) ([]*service.Node, error) {
	rows, err := r.queries(ctx).ListNodes(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*service.Node, 0, len(rows))
	for _, row := range rows {
		n := nodeFromRow(row)
		if n.RefreshOfflineStatus(time.Now(), nodeOfflineThreshold) {
			_ = r.queries(ctx).UpdateNodeStatusOffline(ctx, n.Name)
		}
		result = append(result, &n)
	}
	return result, nil
}

// Get 按名称查询节点，自动检测离线状态。
func (r *NodeRegistry) Get(ctx context.Context, name string) (*service.Node, error) {
	row, err := r.queries(ctx).GetNodeByName(ctx, name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	n := nodeFromRow(row)
	if n.RefreshOfflineStatus(time.Now(), nodeOfflineThreshold) {
		_ = r.queries(ctx).UpdateNodeStatusOffline(ctx, n.Name)
	}
	return &n, nil
}

// Delete 按名称删除节点，返回是否真正删除。
func (r *NodeRegistry) Delete(ctx context.Context, name string) (bool, error) {
	tag, err := hostExecutorFromContext(ctx, r.db).Exec(ctx, "DELETE FROM nodes WHERE name = $1", name)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// GetNodeCount 返回节点总数。
func (r *NodeRegistry) GetNodeCount(ctx context.Context) (int, error) {
	count, err := r.queries(ctx).CountNodes(ctx)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetOnlineCount 返回在线节点数（30s 内有心跳）。
func (r *NodeRegistry) GetOnlineCount(ctx context.Context) (int, error) {
	count, err := r.queries(ctx).CountOnlineNodes(ctx)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func nodeFromRow(row hostgen.Node) service.Node {
	var caps protocol.AgentCapabilities
	if len(row.Capabilities) > 0 {
		_ = json.Unmarshal(row.Capabilities, &caps)
	}
	var agentStatus protocol.AgentStatus
	if len(row.AgentStatus) > 0 {
		_ = json.Unmarshal(row.AgentStatus, &agentStatus)
	}
	var meta protocol.AgentMetadata
	if len(row.Metadata) > 0 {
		_ = json.Unmarshal(row.Metadata, &meta)
	}
	return service.Node{
		Name:          row.Name,
		Address:       row.Address,
		ExternalAddr:  row.ExternalAddr,
		GrpcAddress:   row.GrpcAddress,
		Status:        row.Status,
		RegisteredAt:  row.RegisteredAt.Time,
		LastHeartbeat: row.LastHeartbeat.Time,
		Capabilities:  caps,
		AgentStatus:   agentStatus,
		Metadata:      meta,
	}
}
