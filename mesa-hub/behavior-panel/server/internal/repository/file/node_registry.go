package file

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/charviki/maze-cradle/configutil"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/maze-cradle/storeutil"
	"github.com/charviki/mesa-hub-behavior-panel/internal/service"
)

const (
	nodeOfflineThreshold = 30 * time.Second
	flushInterval        = 30 * time.Second
)

// NodeRegistry 是 Host/Node 域的 JSON 持久化实现。
// 它只负责并发安全和落盘，不承载对外 API 视图拼装等业务投影逻辑。
type NodeRegistry struct {
	mu             sync.RWMutex
	nodes          map[string]*service.Node
	hostTokens     map[string]string
	path           string
	hostTokensPath string
	logger         logutil.Logger
	flusher        *storeutil.DirtyFlusher
}

// NewNodeRegistry 创建节点注册表并从 JSON 文件加载已有数据。
func NewNodeRegistry(nodesFile string, logger logutil.Logger) *NodeRegistry {
	r := &NodeRegistry{
		nodes:          make(map[string]*service.Node),
		hostTokens:     make(map[string]string),
		path:           nodesFile,
		hostTokensPath: filepath.Join(filepath.Dir(nodesFile), "host_tokens.json"),
		logger:         logger,
	}
	// DirtyFlusher 统一管理脏标记，避免心跳频繁上报时每次都同步刷盘。
	r.flusher = storeutil.NewDirtyFlusher(r.save, flushInterval, logger)
	r.load()
	r.flusher.Start()
	return r
}

func (r *NodeRegistry) load() {
	data, err := os.ReadFile(r.path)
	if err != nil {
		r.logger.Infof("[node-registry] file not found, starting fresh: %s", r.path)
	} else {
		var nodes map[string]*service.Node
		if unmarshalErr := json.Unmarshal(data, &nodes); unmarshalErr != nil {
			r.logger.Errorf("[node-registry] parse file %s failed: %v", r.path, unmarshalErr)
		} else {
			r.mu.Lock()
			r.nodes = nodes
			r.mu.Unlock()
		}
	}

	tokensData, err := os.ReadFile(r.hostTokensPath)
	if err != nil {
		r.logger.Infof("[node-registry] host tokens file not found, starting fresh: %s", r.hostTokensPath)
		return
	}
	var tokens map[string]string
	if err := json.Unmarshal(tokensData, &tokens); err != nil {
		r.logger.Errorf("[node-registry] parse host tokens file %s failed: %v", r.hostTokensPath, err)
		return
	}
	r.mu.Lock()
	r.hostTokens = tokens
	r.mu.Unlock()
}

func (r *NodeRegistry) save() error {
	r.mu.RLock()
	//nolint:gosec // AuthToken is intentionally persisted
	data, err := json.MarshalIndent(r.nodes, "", "  ")
	tokensData, tokensErr := json.MarshalIndent(r.hostTokens, "", "  ")
	r.mu.RUnlock()

	if err != nil {
		r.logger.Errorf("[node-registry] marshal nodes failed: %v", err)
	} else if writeErr := configutil.AtomicWriteFile(r.path, data, 0644); writeErr != nil {
		r.logger.Errorf("[node-registry] write file %s failed: %v", r.path, writeErr)
	}

	if tokensErr != nil {
		r.logger.Errorf("[node-registry] marshal host tokens failed: %v", tokensErr)
	} else if writeErr := configutil.AtomicWriteFile(r.hostTokensPath, tokensData, 0644); writeErr != nil {
		r.logger.Errorf("[node-registry] write host tokens file %s failed: %v", r.hostTokensPath, writeErr)
	}
	return nil
}

// StoreHostToken 预存 Host token，供注册/心跳时做分层令牌校验。
func (r *NodeRegistry) StoreHostToken(_ context.Context, name, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hostTokens[name] = token
	r.flusher.MarkDirty()
	return nil
}

// ValidateHostToken 检查 Host 是否存在预存 token，以及 token 是否匹配。
func (r *NodeRegistry) ValidateHostToken(_ context.Context, name, token string) (bool, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	expected, ok := r.hostTokens[name]
	if !ok {
		return false, false, nil
	}
	return true, expected == token, nil
}

// RemoveHostToken 删除 Host 专属 token，避免删除后的旧 token 残留。
func (r *NodeRegistry) RemoveHostToken(_ context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.hostTokens, name)
	r.flusher.MarkDirty()
	return nil
}

// Register 写入最新的节点注册快照。
func (r *NodeRegistry) Register(_ context.Context, req protocol.RegisterRequest) (*service.Node, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if existing, ok := r.nodes[req.Name]; ok {
		r.logger.Warnf("[node-registry] node %q replaced (old registered at %v, last heartbeat %v)",
			req.Name, existing.RegisteredAt.Format(time.RFC3339), existing.LastHeartbeat.Format(time.RFC3339))
	}

	node := &service.Node{
		Name:          req.Name,
		Address:       req.Address,
		ExternalAddr:  req.ExternalAddr,
		GrpcAddress:   req.GrpcAddress,
		Status:        service.NodeStatusOnline,
		RegisteredAt:  now,
		LastHeartbeat: now,
		Capabilities:  cloneAgentCapabilities(req.Capabilities),
		AgentStatus:   cloneAgentStatus(req.Status),
		Metadata:      req.Metadata,
	}
	r.nodes[req.Name] = node
	r.flusher.MarkDirty()
	return cloneNode(node), nil
}

// Heartbeat 更新节点最近一次心跳和完整状态快照。
func (r *NodeRegistry) Heartbeat(_ context.Context, req protocol.HeartbeatRequest) (*service.Node, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	node, ok := r.nodes[req.Name]
	if !ok {
		return nil, nil
	}
	node.LastHeartbeat = time.Now()
	node.AgentStatus = cloneAgentStatus(req.Status)
	// Session 数以 Agent 当前快照为准，避免注册请求与心跳请求各自维护冗余来源。
	node.AgentStatus.ActiveSessions = len(req.Status.SessionDetails)
	node.Status = service.NodeStatusOnline
	r.flusher.MarkDirty()
	return cloneNode(node), nil
}

// List 返回节点快照列表，并在读取路径内顺带完成离线检测。
func (r *NodeRegistry) List(_ context.Context) ([]*service.Node, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	nodes := make([]*service.Node, 0, len(r.nodes))
	for _, n := range r.nodes {
		if n.RefreshOfflineStatus(now, nodeOfflineThreshold) {
			r.flusher.MarkDirty()
		}
		nodes = append(nodes, cloneNode(n))
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})
	return nodes, nil
}

// Get 获取单个节点，并在同一临界区内完成离线检测。
func (r *NodeRegistry) Get(_ context.Context, name string) (*service.Node, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	node, ok := r.nodes[name]
	if !ok {
		return nil, nil
	}
	if node.RefreshOfflineStatus(time.Now(), nodeOfflineThreshold) {
		r.flusher.MarkDirty()
	}
	return cloneNode(node), nil
}

// Delete 删除指定节点。
func (r *NodeRegistry) Delete(_ context.Context, name string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.nodes[name]; !ok {
		return false, nil
	}
	delete(r.nodes, name)
	r.flusher.MarkDirty()
	return true, nil
}

// GetNodeCount 返回注册节点总数。
func (r *NodeRegistry) GetNodeCount(_ context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.nodes), nil
}

// GetOnlineCount 返回最近仍在心跳窗口内的节点数。
func (r *NodeRegistry) GetOnlineCount(_ context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	count := 0
	for _, n := range r.nodes {
		if now.Sub(n.LastHeartbeat) <= nodeOfflineThreshold {
			count++
		}
	}
	return count, nil
}

// WaitSave 停止后台刷盘并执行最终持久化。
func (r *NodeRegistry) WaitSave() {
	r.flusher.WaitSave()
}

var _ service.NodeRegistry = (*NodeRegistry)(nil)
