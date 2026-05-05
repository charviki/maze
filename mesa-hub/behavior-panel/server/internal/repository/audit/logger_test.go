package audit

import (
	"context"
	"sync"
	"testing"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
)

func TestLogger_ListConcurrentWithLog(t *testing.T) {
	logger := NewLogger("", logutil.NewNop())
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)

		go func(idx int) {
			defer wg.Done()
			_ = logger.Log(ctx, protocol.AuditLogEntry{
				TargetNode: "agent",
				Action:     "heartbeat",
				StatusCode: 200 + idx,
			})
		}(i)

		go func() {
			defer wg.Done()
			_, _ = logger.List(ctx)
			_, _, _ = logger.ListPage(ctx, 1, 10)
			_, _ = logger.Query(ctx, "agent", "heartbeat")
		}()
	}

	wg.Wait()

	logs, _ := logger.List(ctx)
	if got := len(logs); got != 100 {
		t.Fatalf("期望最终有 100 条日志，实际=%d", got)
	}
}
