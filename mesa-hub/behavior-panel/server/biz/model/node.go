package model

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
)

const (
	nodeOfflineThreshold = 30 * time.Second
	NodeStatusOnline     = "online"
	NodeStatusOffline    = "offline"
)

// Node Agent 节点信息，包含注册信息、能力声明、健康状态和运行时指标
type Node struct {
	Name          string                     `json:"name"`
	Address       string                     `json:"address"`
	ExternalAddr  string                     `json:"external_addr"`
	Status        string                     `json:"status"`
	RegisteredAt  time.Time                  `json:"registered_at"`
	LastHeartbeat time.Time                  `json:"last_heartbeat"`
	Capabilities  protocol.AgentCapabilities `json:"capabilities"`
	AgentStatus   protocol.AgentStatus       `json:"agent_status"`
	Metadata      protocol.AgentMetadata     `json:"metadata"`
}

// NodeRegistry 节点注册表，JSON 文件持久化存储。使用读写锁保护并发访问。
// Manager 重启后可从文件恢复已注册节点信息，Agent 下次心跳时更新状态。
// 持久化采用 dirty flag + 后台定时刷盘策略，避免每次心跳都触发全量写盘。
type NodeRegistry struct {
	mu     sync.RWMutex
	nodes  map[string]*Node
	path   string
	logger logutil.Logger
	dirty  bool
	stopCh chan struct{}
	doneCh chan struct{}
}

const flushInterval = 30 * time.Second

// NewNodeRegistry 创建节点注册表并从 JSON 文件加载已有数据。
// 文件不存在时为首次启动，以空注册表开始；解析失败时记录错误日志但不阻塞启动。
// 启动后台 flush loop 定期检查 dirty 标记并刷盘。
func NewNodeRegistry(filePath string, logger logutil.Logger) *NodeRegistry {
	r := &NodeRegistry{
		nodes:  make(map[string]*Node),
		path:   filePath,
		logger: logger,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
	r.load()
	go r.flushLoop()
	return r
}

// load 从 JSON 文件加载节点数据。文件不存在视为首次启动，不报错。
func (r *NodeRegistry) load() {
	data, err := os.ReadFile(r.path)
	if err != nil {
		r.logger.Infof("[node-registry] file not found, starting fresh: %s", r.path)
		return
	}
	var nodes map[string]*Node
	if err := json.Unmarshal(data, &nodes); err != nil {
		r.logger.Errorf("[node-registry] parse file %s failed: %v", r.path, err)
		return
	}
	r.mu.Lock()
	r.nodes = nodes
	r.mu.Unlock()
}

// save 持久化当前节点数据到 JSON 文件（原子写入，防止写入中断导致文件损坏）。
// 持久化失败时仅记录错误日志，不阻塞业务流程——内存中的数据仍然有效。
func (r *NodeRegistry) save() {
	r.mu.RLock()
	data, err := json.MarshalIndent(r.nodes, "", "  ")
	r.mu.RUnlock()
	if err != nil {
		r.logger.Errorf("[node-registry] marshal nodes failed: %v", err)
		return
	}
	if err := atomicWriteFile(r.path, data, 0644); err != nil {
		r.logger.Errorf("[node-registry] write file %s failed: %v", r.path, err)
	}
}

// Register 注册新节点，携带 capabilities、status、metadata。
// 同名节点被覆盖时记录告警日志，标记旧节点已被替换。
func (r *NodeRegistry) Register(req protocol.RegisterRequest) *Node {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// 同名节点覆盖告警：如果该 name 已存在，说明旧节点可能已失联
	if existing, ok := r.nodes[req.Name]; ok {
		r.logger.Warnf("[node-registry] node %q replaced (old registered at %v, last heartbeat %v)",
			req.Name, existing.RegisteredAt.Format(time.RFC3339), existing.LastHeartbeat.Format(time.RFC3339))
	}

	node := &Node{
		Name:          req.Name,
		Address:       req.Address,
		ExternalAddr:  req.ExternalAddr,
		Status:        NodeStatusOnline,
		RegisteredAt:  now,
		LastHeartbeat: now,
		Capabilities:  req.Capabilities,
		AgentStatus:   req.Status,
		Metadata:      req.Metadata,
	}
	r.nodes[req.Name] = node
	r.dirty = true
	return node
}

// Heartbeat 更新节点心跳时间和完整状态快照（含 CPU、内存、Session 详情）。
// 从 AgentStatus.SessionDetails 同步 session count。节点不存在时返回 nil。
func (r *NodeRegistry) Heartbeat(req protocol.HeartbeatRequest) *Node {
	r.mu.Lock()
	defer r.mu.Unlock()

	node, ok := r.nodes[req.Name]
	if !ok {
		return nil
	}
	node.LastHeartbeat = time.Now()
	node.AgentStatus = req.Status
	// 从 AgentStatus.SessionDetails 同步 session count，避免冗余字段
	node.AgentStatus.ActiveSessions = len(req.Status.SessionDetails)
	node.Status = NodeStatusOnline
	r.dirty = true
	return node
}

func (r *NodeRegistry) List() []*Node {
	r.mu.Lock()
	defer r.mu.Unlock()

	nodes := make([]*Node, 0, len(r.nodes))
	for _, n := range r.nodes {
		// 在写锁内完成离线检测，消除两阶段锁的 TOCTOU 竞争窗口
		if time.Since(n.LastHeartbeat) > nodeOfflineThreshold && n.Status == NodeStatusOnline {
			n.Status = NodeStatusOffline
			r.dirty = true
		}
		nodes = append(nodes, n)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})
	return nodes
}

func (r *NodeRegistry) Get(name string) *Node {
	r.mu.RLock()
	node, ok := r.nodes[name]
	r.mu.RUnlock()
	if !ok {
		return nil
	}
	if time.Since(node.LastHeartbeat) > nodeOfflineThreshold && node.Status == NodeStatusOnline {
		r.mu.Lock()
		if node.Status == NodeStatusOnline {
			node.Status = NodeStatusOffline
			r.dirty = true
		}
		r.mu.Unlock()
	}
	return node
}

// Delete 删除指定节点。持久化失败不影响内存删除。
func (r *NodeRegistry) Delete(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.nodes[name]; !ok {
		return false
	}
	delete(r.nodes, name)
	r.dirty = true
	return true
}

// GetNodeCount 返回已注册节点总数（含 offline）
func (r *NodeRegistry) GetNodeCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.nodes)
}

// GetOnlineCount 返回在线节点数量（心跳未超时）
func (r *NodeRegistry) GetOnlineCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, n := range r.nodes {
		if time.Since(n.LastHeartbeat) <= nodeOfflineThreshold {
			count++
		}
	}
	return count
}

// WaitSave 停止后台 flush loop 并执行最终刷盘，确保优雅关闭时数据不丢失
func (r *NodeRegistry) WaitSave() {
	close(r.stopCh)
	<-r.doneCh
}

// flushLoop 后台定期检查 dirty 标记并刷盘，收到停止信号时执行最终刷盘
func (r *NodeRegistry) flushLoop() {
	defer close(r.doneCh)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.mu.Lock()
			if r.dirty {
				r.dirty = false
				r.mu.Unlock()
				r.save()
			} else {
				r.mu.Unlock()
			}
		case <-r.stopCh:
			// 收到停止信号，执行最终刷盘确保数据完整
			r.mu.Lock()
			if r.dirty {
				r.dirty = false
				r.mu.Unlock()
				r.save()
			} else {
				r.mu.Unlock()
			}
			return
		}
	}
}

// FormatNodeSummary 返回节点摘要字符串，用于日志输出
func FormatNodeSummary(n *Node) string {
	return fmt.Sprintf("%s (%s) sessions=%d cpu=%.1f%% mem=%.0fMB status=%s",
		n.Name, n.Address, n.AgentStatus.ActiveSessions,
		n.AgentStatus.CPUUsage, n.AgentStatus.MemoryUsageMB, n.Status)
}
