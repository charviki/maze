package storeutil

import (
	"sync"
	"time"

	"github.com/charviki/maze-cradle/logutil"
)

// DirtyFlusher 通用脏标记定时刷盘器。
// 后台 goroutine 按固定间隔检查 dirty 标记，有变更时调用 flushFn 持久化。
// 调用 WaitSave 可优雅停止并执行最终刷盘，确保数据不丢失。
type DirtyFlusher struct {
	mu            sync.Mutex
	dirty         bool
	flushFn       func() error
	flushInterval time.Duration
	stopCh        chan struct{}
	doneCh        chan struct{}
	logger        logutil.Logger
}

// NewDirtyFlusher 创建 DirtyFlusher。
// flushFn 为实际持久化函数，interval 为刷盘间隔，logger 用于记录刷盘错误。
// 调用 Start() 启动后台 goroutine。
func NewDirtyFlusher(flushFn func() error, interval time.Duration, logger logutil.Logger) *DirtyFlusher {
	return &DirtyFlusher{
		flushFn:       flushFn,
		flushInterval: interval,
		stopCh:        make(chan struct{}),
		doneCh:        make(chan struct{}),
		logger:        logger,
	}
}

// MarkDirty 标记数据已变更，下次 tick 时会触发刷盘。
// 可在调用方已有的写锁内安全调用（DirtyFlusher 使用独立锁，不会死锁）。
func (f *DirtyFlusher) MarkDirty() {
	f.mu.Lock()
	f.dirty = true
	f.mu.Unlock()
}

// Start 启动后台刷盘 goroutine。
func (f *DirtyFlusher) Start() {
	go f.flushLoop()
}

// WaitSave 停止后台 goroutine 并执行最终刷盘，阻塞直到完成。
func (f *DirtyFlusher) WaitSave() {
	close(f.stopCh)
	<-f.doneCh
}

func (f *DirtyFlusher) flushLoop() {
	defer close(f.doneCh)
	ticker := time.NewTicker(f.flushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-f.stopCh:
			// 收到停止信号，执行最终刷盘确保数据完整
			f.flushIfNeeded()
			return
		case <-ticker.C:
			f.flushIfNeeded()
		}
	}
}

// flushIfNeeded 检查 dirty 标记，有变更时执行刷盘。
// 刷盘失败时重新标记 dirty，下次 tick 重试。
func (f *DirtyFlusher) flushIfNeeded() {
	f.mu.Lock()
	if !f.dirty {
		f.mu.Unlock()
		return
	}
	// 先清除标记，刷盘失败时再重新设置
	f.dirty = false
	f.mu.Unlock()

	if err := f.flushFn(); err != nil {
		if f.logger != nil {
			f.logger.Errorf("[flusher] flush failed: %v", err)
		}
		// 刷盘失败，重新标记 dirty 以便下次重试
		f.mu.Lock()
		f.dirty = true
		f.mu.Unlock()
	}
}
