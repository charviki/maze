package file

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/charviki/maze-cradle/configutil"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/maze-cradle/storeutil"
	"github.com/charviki/mesa-hub-behavior-panel/internal/service"
)

// HostSpecRepository 是 HostSpec 的 JSON 持久化实现。
// 它只维护声明式 HostSpec 本身，不在 store 层拼装 HostInfo 业务视图。
type HostSpecRepository struct {
	mu      sync.RWMutex
	specs   map[string]*protocol.HostSpec
	path    string
	logger  logutil.Logger
	flusher *storeutil.DirtyFlusher
}

// NewHostSpecRepository 创建 HostSpecRepository 并从 JSON 文件恢复已有数据。
func NewHostSpecRepository(filePath string, logger logutil.Logger) *HostSpecRepository {
	s := &HostSpecRepository{
		specs:  make(map[string]*protocol.HostSpec),
		path:   filePath,
		logger: logger,
	}
	s.flusher = storeutil.NewDirtyFlusher(s.save, flushInterval, logger)
	s.load()
	s.flusher.Start()
	return s
}

func (s *HostSpecRepository) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		s.logger.Infof("[host-spec] file not found, starting fresh: %s", s.path)
		return
	}
	var specs map[string]*protocol.HostSpec
	if err := json.Unmarshal(data, &specs); err != nil {
		s.logger.Errorf("[host-spec] parse file %s failed: %v", s.path, err)
		return
	}
	s.mu.Lock()
	s.specs = specs
	s.mu.Unlock()
}

func (s *HostSpecRepository) save() error {
	s.mu.RLock()
	//nolint:gosec // AuthToken is intentionally persisted
	data, err := json.MarshalIndent(s.specs, "", "  ")
	s.mu.RUnlock()

	if err != nil {
		s.logger.Errorf("[host-spec] marshal specs failed: %v", err)
		return nil
	}
	if writeErr := configutil.AtomicWriteFile(s.path, data, 0644); writeErr != nil {
		s.logger.Errorf("[host-spec] write file %s failed: %v", s.path, writeErr)
	}
	return nil
}

// Create 创建新的 HostSpec。
func (s *HostSpecRepository) Create(spec *protocol.HostSpec) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.specs[spec.Name]; ok {
		return false
	}
	// repository 内部保存独立副本，避免调用方在锁外继续修改传入对象。
	s.specs[spec.Name] = cloneHostSpec(spec)
	s.flusher.MarkDirty()
	return true
}

// Get 返回指定 HostSpec。
func (s *HostSpecRepository) Get(name string) *protocol.HostSpec {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneHostSpec(s.specs[name])
}

// List 返回所有 HostSpec，并保持按名称排序，保证 API 输出稳定。
func (s *HostSpecRepository) List() []*protocol.HostSpec {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*protocol.HostSpec, 0, len(s.specs))
	for _, spec := range s.specs {
		result = append(result, cloneHostSpec(spec))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// UpdateStatus 更新 HostSpec 的状态和错误信息。
func (s *HostSpecRepository) UpdateStatus(name, status, errMsg string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	spec, ok := s.specs[name]
	if !ok {
		return false
	}
	spec.Status = status
	spec.ErrorMsg = errMsg
	spec.UpdatedAt = time.Now()
	s.flusher.MarkDirty()
	return true
}

// Delete 删除指定 HostSpec。
func (s *HostSpecRepository) Delete(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.specs[name]; !ok {
		return false
	}
	delete(s.specs, name)
	s.flusher.MarkDirty()
	return true
}

// IncrementRetry 递增指定 Host 的重试计数。
func (s *HostSpecRepository) IncrementRetry(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	spec, ok := s.specs[name]
	if !ok {
		return false
	}
	spec.RetryCount++
	spec.UpdatedAt = time.Now()
	s.flusher.MarkDirty()
	return true
}

// WaitSave 停止后台刷盘并执行最终持久化。
func (s *HostSpecRepository) WaitSave() {
	s.flusher.WaitSave()
}

var _ service.HostSpecRepository = (*HostSpecRepository)(nil)
