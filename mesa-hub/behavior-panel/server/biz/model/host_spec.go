package model

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
)

// HostSpecManager 管理所有 Host 的创建规格，复用 dirty flush 模式持久化到 host_specs.json。
// 作为 Manager 恢复和巡检的 source of truth。
type HostSpecManager struct {
	mu     sync.RWMutex
	specs  map[string]*protocol.HostSpec
	path   string
	logger logutil.Logger
	dirty  bool
	stopCh chan struct{}
	doneCh chan struct{}
}

// NewHostSpecManager 创建 HostSpecManager 并从 JSON 文件加载已有数据。
func NewHostSpecManager(filePath string, logger logutil.Logger) *HostSpecManager {
	m := &HostSpecManager{
		specs:  make(map[string]*protocol.HostSpec),
		path:   filePath,
		logger: logger,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
	m.load()
	go m.flushLoop()
	return m
}

func (m *HostSpecManager) load() {
	data, err := os.ReadFile(m.path)
	if err != nil {
		m.logger.Infof("[host-spec] file not found, starting fresh: %s", m.path)
		return
	}
	var specs map[string]*protocol.HostSpec
	if err := json.Unmarshal(data, &specs); err != nil {
		m.logger.Errorf("[host-spec] parse file %s failed: %v", m.path, err)
		return
	}
	m.mu.Lock()
	m.specs = specs
	m.mu.Unlock()
}

func (m *HostSpecManager) save() {
	m.mu.RLock()
	data, err := json.MarshalIndent(m.specs, "", "  ")
	m.mu.RUnlock()

	if err != nil {
		m.logger.Errorf("[host-spec] marshal specs failed: %v", err)
		return
	}
	if writeErr := atomicWriteFile(m.path, data, 0644); writeErr != nil {
		m.logger.Errorf("[host-spec] write file %s failed: %v", m.path, writeErr)
	}
}

// Create 创建新的 HostSpec 并持久化。同名已存在时返回 false。
func (m *HostSpecManager) Create(spec *protocol.HostSpec) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.specs[spec.Name]; ok {
		return false
	}
	m.specs[spec.Name] = spec
	m.dirty = true
	return true
}

// Get 获取指定名称的 HostSpec。
func (m *HostSpecManager) Get(name string) *protocol.HostSpec {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.specs[name]
}

// List 返回所有 HostSpec，按名称排序。
func (m *HostSpecManager) List() []*protocol.HostSpec {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*protocol.HostSpec, 0, len(m.specs))
	for _, s := range m.specs {
		result = append(result, s)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// UpdateStatus 更新 HostSpec 的状态和错误信息。
func (m *HostSpecManager) UpdateStatus(name, status, errMsg string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	spec, ok := m.specs[name]
	if !ok {
		return false
	}
	spec.Status = status
	spec.ErrorMsg = errMsg
	spec.UpdatedAt = time.Now()
	m.dirty = true
	return true
}

// Delete 删除指定名称的 HostSpec。
func (m *HostSpecManager) Delete(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.specs[name]; !ok {
		return false
	}
	delete(m.specs, name)
	m.dirty = true
	return true
}

// IncrementRetry 递增指定 HostSpec 的重试计数。
func (m *HostSpecManager) IncrementRetry(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	spec, ok := m.specs[name]
	if !ok {
		return false
	}
	spec.RetryCount++
	spec.UpdatedAt = time.Now()
	m.dirty = true
	return true
}

// ListMerged 合并 HostSpec 与 NodeRegistry 心跳状态，返回 API 响应用的 HostInfo 列表。
// 合并逻辑：
//   - HostSpec 状态为 pending/failed → 直接使用
//   - HostSpec 状态为 deploying 且 NodeRegistry 中有对应节点 → 根据 node.Offline() 返回 online/offline
//   - 无对应节点 → 保持原状态
func (m *HostSpecManager) ListMerged(registry *NodeRegistry) []*protocol.HostInfo {
	specs := m.List()
	result := make([]*protocol.HostInfo, 0, len(specs))
	for _, spec := range specs {
		info := &protocol.HostInfo{
			HostSpec: *spec,
		}
		// deploying 状态下检查 NodeRegistry 心跳
		if spec.Status == protocol.HostStatusDeploying {
			node := registry.Get(spec.Name)
			if node != nil {
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
			}
		}
		result = append(result, info)
	}
	return result
}

// WaitSave 停止后台 flush loop 并执行最终刷盘。
func (m *HostSpecManager) WaitSave() {
	close(m.stopCh)
	<-m.doneCh
}

func (m *HostSpecManager) flushLoop() {
	defer close(m.doneCh)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			if m.dirty {
				m.dirty = false
				m.mu.Unlock()
				m.save()
			} else {
				m.mu.Unlock()
			}
		case <-m.stopCh:
			m.mu.Lock()
			if m.dirty {
				m.dirty = false
				m.mu.Unlock()
				m.save()
			} else {
				m.mu.Unlock()
			}
			return
		}
	}
}
