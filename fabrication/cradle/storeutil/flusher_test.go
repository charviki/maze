package storeutil

import (
	"sync"
	"testing"
	"time"

	"github.com/charviki/maze-cradle/logutil"
)

func TestNewDirtyFlusher(t *testing.T) {
	flushCount := 0
	flusher := NewDirtyFlusher(func() error {
		flushCount++
		return nil
	}, 50*time.Millisecond, logutil.NewNop())

	if flusher == nil {
		t.Fatal("NewDirtyFlusher returned nil")
	}
	if flusher.flushInterval != 50*time.Millisecond {
		t.Errorf("flushInterval = %v, want 50ms", flusher.flushInterval)
	}
}

func TestDirtyFlusher_StartAndWaitSave(t *testing.T) {
	flushCount := 0
	var mu sync.Mutex
	flusher := NewDirtyFlusher(func() error {
		mu.Lock()
		flushCount++
		mu.Unlock()
		return nil
	}, 50*time.Millisecond, logutil.NewNop())

	flusher.Start()
	flusher.MarkDirty()

	flusher.WaitSave()

	mu.Lock()
	if flushCount == 0 {
		t.Error("flush should have been called at least once")
	}
	mu.Unlock()
}

func TestDirtyFlusher_MarkDirty(t *testing.T) {
	flusher := NewDirtyFlusher(func() error { return nil }, 1*time.Hour, logutil.NewNop())

	if flusher.dirty {
		t.Error("flusher should not be dirty initially")
	}
	flusher.MarkDirty()
	if !flusher.dirty {
		t.Error("flusher should be dirty after MarkDirty")
	}
}

func TestDirtyFlusher_NoFlushWhenNotDirty(t *testing.T) {
	flushCount := 0
	flusher := NewDirtyFlusher(func() error {
		flushCount++
		return nil
	}, 50*time.Millisecond, logutil.NewNop())

	flusher.Start()

	flusher.WaitSave()

	if flushCount > 0 {
		t.Error("flush should not be called when not dirty")
	}
}

func TestDirtyFlusher_MultipleWaitSavePanics(t *testing.T) {
	flusher := NewDirtyFlusher(func() error { return nil }, 50*time.Millisecond, logutil.NewNop())
	flusher.Start()
	flusher.WaitSave()

	defer func() {
		if r := recover(); r == nil {
			t.Error("second WaitSave should panic")
		}
	}()
	flusher.WaitSave()
}

func TestDirtyFlusher_FlushErrorRetry(t *testing.T) {
	flushCount := 0
	flusher := NewDirtyFlusher(func() error {
		flushCount++
		if flushCount == 1 {
			return nil
		}
		return nil
	}, 50*time.Millisecond, logutil.NewNop())

	flusher.Start()
	flusher.MarkDirty()
	flusher.WaitSave()

	if flushCount < 1 {
		t.Error("flush should have been called")
	}
}

func TestDirtyFlusher_ConcurrentMarkDirty(t *testing.T) {
	flusher := NewDirtyFlusher(func() error { return nil }, 1*time.Hour, logutil.NewNop())
	flusher.Start()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			flusher.MarkDirty()
		}()
	}
	wg.Wait()

	if !flusher.dirty {
		t.Error("flusher should be dirty after concurrent MarkDirty")
	}

	flusher.WaitSave()
}

func TestDirtyFlusher_FlushLoopStops(t *testing.T) {
	flushCount := 0
	var mu sync.Mutex
	flusher := NewDirtyFlusher(func() error {
		mu.Lock()
		flushCount++
		mu.Unlock()
		return nil
	}, 50*time.Millisecond, logutil.NewNop())

	flusher.Start()
	flusher.MarkDirty()
	flusher.WaitSave()

	mu.Lock()
	countAfterWaitSave := flushCount
	mu.Unlock()
	if countAfterWaitSave == 0 {
		t.Fatal("WaitSave 前应至少发生一次 flush")
	}

	select {
	case <-flusher.stopCh:
	default:
		t.Fatal("WaitSave 返回后应关闭 stopCh，确保后台循环已停止")
	}
}
