package file

import (
	"context"
	"encoding/json"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/charviki/maze-cradle/configutil"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/maze-cradle/storeutil"
	"github.com/charviki/maze/the-mesa/director-core/internal/service"
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
func (s *HostSpecRepository) Create(_ context.Context, spec *protocol.HostSpec) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.specs[spec.Name]; ok {
		return false, nil
	}
	// repository 内部保存独立副本，避免调用方在锁外继续修改传入对象。
	s.specs[spec.Name] = cloneHostSpec(spec)
	s.flusher.MarkDirty()
	return true, nil
}

// Get 返回指定 HostSpec。
func (s *HostSpecRepository) Get(_ context.Context, name string) (*protocol.HostSpec, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneHostSpec(s.specs[name]), nil
}

// List 返回所有 HostSpec，并保持按名称排序，保证 API 输出稳定。
func (s *HostSpecRepository) List(_ context.Context) ([]*protocol.HostSpec, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*protocol.HostSpec, 0, len(s.specs))
	for _, spec := range s.specs {
		result = append(result, cloneHostSpec(spec))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}

// UpdateStatus 更新 HostSpec 的状态和错误信息。
func (s *HostSpecRepository) UpdateStatus(_ context.Context, name, status, errMsg string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	spec, ok := s.specs[name]
	if !ok {
		return false, nil
	}
	spec.Status = status
	spec.ErrorMsg = errMsg
	spec.UpdatedAt = time.Now()
	s.flusher.MarkDirty()
	return true, nil
}

// Delete 删除指定 HostSpec。
func (s *HostSpecRepository) Delete(_ context.Context, name string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.specs[name]; !ok {
		return false, nil
	}
	delete(s.specs, name)
	s.flusher.MarkDirty()
	return true, nil
}

// IncrementRetry 递增指定 Host 的重试计数。
func (s *HostSpecRepository) IncrementRetry(_ context.Context, name string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	spec, ok := s.specs[name]
	if !ok {
		return false, nil
	}
	spec.RetryCount++
	spec.UpdatedAt = time.Now()
	s.flusher.MarkDirty()
	return true, nil
}

// WaitSave 停止后台刷盘并执行最终持久化。
func (s *HostSpecRepository) WaitSave() {
	s.flusher.WaitSave()
}

var _ service.HostSpecRepository = (*HostSpecRepository)(nil)
