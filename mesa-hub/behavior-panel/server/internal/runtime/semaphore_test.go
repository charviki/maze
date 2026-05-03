package runtime

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBuildSemaphore_ConcurrencyLimit(t *testing.T) {
	// 使用独立的信号量验证并发限流逻辑，不污染全局 buildSemaphore
	sem := make(chan struct{}, 2)
	const totalTasks = 10
	var concurrentCount atomic.Int32
	var maxConcurrent atomic.Int32
	var wg sync.WaitGroup

	for i := 0; i < totalTasks; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			current := concurrentCount.Add(1)
			for {
				old := maxConcurrent.Load()
				if current <= old || maxConcurrent.CompareAndSwap(old, current) {
					break
				}
			}
			time.Sleep(20 * time.Millisecond)
			concurrentCount.Add(-1)
			<-sem
		}()
	}

	wg.Wait()

	max := maxConcurrent.Load()
	if max > 2 {
		t.Errorf("最大并发构建数不应超过 2, 实际=%d", max)
	}
	if max < 1 {
		t.Errorf("最大并发构建数不应为 0")
	}
}

func TestBuildSemaphore_BlockingRelease(t *testing.T) {
	sem := make(chan struct{}, 2)

	// 填满信号量
	sem <- struct{}{}
	sem <- struct{}{}

	gotThrough := make(chan struct{}, 1)

	// 第三个应该阻塞
	go func() {
		sem <- struct{}{}
		gotThrough <- struct{}{}
		<-sem
	}()

	// 确认第三个被阻塞
	select {
	case <-gotThrough:
		t.Fatal("第三个任务不应在槽位释放前完成")
	case <-time.After(50 * time.Millisecond):
	}

	// 释放一个槽位
	<-sem

	// 第三个应该能继续
	select {
	case <-gotThrough:
	case <-time.After(1 * time.Second):
		t.Fatal("释放槽位后第三个任务应能继续")
	}

	// 清理
	<-sem
}

func TestBuildSemaphore_CapacityIsTwo(t *testing.T) {
	// 验证全局 buildSemaphore 的容量
	if cap(buildSemaphore) != 2 {
		t.Errorf("buildSemaphore 容量应为 2, 实际=%d", cap(buildSemaphore))
	}
}
