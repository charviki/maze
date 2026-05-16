package lifecycle

import (
	"context"
	"sync"

	"github.com/charviki/maze/fabrication/cradle/logutil"
)

// BackgroundRunner 将后台任务适配为 lifecycle.Server，
// 使其能被 Manager 统一管理启停。
type BackgroundRunner struct {
	name     string
	logger   logutil.Logger
	run      func(<-chan struct{})
	stopOnce sync.Once
	stopCh   chan struct{}
	doneCh   chan struct{}
}

// NewBackgroundRunner 创建后台任务运行器。
// run 函数接收一个 stopCh，当 stopCh 关闭时应退出。
func NewBackgroundRunner(name string, logger logutil.Logger, run func(<-chan struct{})) *BackgroundRunner {
	return &BackgroundRunner{
		name:   name,
		logger: logger,
		run:    run,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

// ListenAndServe 阻塞运行后台任务，直到 stopCh 关闭。
func (r *BackgroundRunner) ListenAndServe() error {
	defer close(r.doneCh)
	r.run(r.stopCh)
	return nil
}

// Shutdown 通知后台任务停止，并等待其退出或超时。
func (r *BackgroundRunner) Shutdown(ctx context.Context) error {
	r.stopOnce.Do(func() {
		close(r.stopCh)
	})
	select {
	case <-r.doneCh:
		return nil
	case <-ctx.Done():
		if r.logger != nil {
			r.logger.Warnf("[%s] shutdown timed out: %v", r.name, ctx.Err())
		}
		return nil
	}
}
