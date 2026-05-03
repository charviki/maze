package storeutil

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/charviki/maze-cradle/configutil"
	"github.com/charviki/maze-cradle/logutil"
)

// JSONStore 泛型 JSON 文件持久化存储，通过原子写入保证数据完整性。
// T 为存储的数据类型，Get/Update/View 等方法提供并发安全的数据访问。
type JSONStore[T any] struct {
	mu     sync.RWMutex
	data   T
	path   string
	logger logutil.Logger
}

// NewJSONStore 创建 JSONStore 并从文件加载已有数据。文件不存在时使用零值初始化。
// logger 用于记录持久化相关的警告和错误日志。
func NewJSONStore[T any](path string, data T, logger logutil.Logger) *JSONStore[T] {
	s := &JSONStore[T]{
		data:   data,
		path:   path,
		logger: logger,
	}
	s.load()
	return s
}

func (s *JSONStore[T]) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	// 反序列化失败时记录错误并降级：保留调用方传入的初始零值
	if err := json.Unmarshal(data, &s.data); err != nil {
		if s.logger != nil {
			s.logger.Warnf("[JSONStore] unmarshal failed, path=%s, error=%v; falling back to initial data", s.path, err)
		}
	}
}

// Save 持久化当前数据到 JSON 文件（原子写入）
func (s *JSONStore[T]) Save() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.data, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}
	return configutil.AtomicWriteFile(s.path, data, 0644)
}

// Get 返回数据的值拷贝。对于包含引用类型（指针、切片、map）的 T，
// 调用方应注意浅拷贝风险。
func (s *JSONStore[T]) Get() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

// GetData 返回数据的只读引用（调用方不应修改返回值，用于性能敏感场景）
func (s *JSONStore[T]) GetData() *T {
	return &s.data
}

// Update 在写锁保护下执行更新函数，并可选地持久化
func (s *JSONStore[T]) Update(fn func(data *T), persist bool) error {
	s.mu.Lock()
	fn(&s.data)
	s.mu.Unlock()
	if persist {
		return s.Save()
	}
	return nil
}

// View 在读锁保护下执行只读回调，用于安全读取数据快照
func (s *JSONStore[T]) View(fn func(data *T)) {
	s.mu.RLock()
	fn(&s.data)
	s.mu.RUnlock()
}
