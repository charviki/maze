package audit

import (
	"sync"
	"testing"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
)

func TestLogger_ListConcurrentWithLog(t *testing.T) {
	logger := NewLogger("", logutil.NewNop())

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)

		go func(idx int) {
			defer wg.Done()
			logger.Log(protocol.AuditLogEntry{
				TargetNode: "agent",
				Action:     "heartbeat",
				StatusCode: 200 + idx,
			})
		}(i)

		go func() {
			defer wg.Done()
			_ = logger.List()
			_, _ = logger.ListPage(1, 10)
			_ = logger.Query("agent", "heartbeat")
		}()
	}

	wg.Wait()

	if got := len(logger.List()); got != 100 {
		t.Fatalf("期望最终有 100 条日志，实际=%d", got)
	}
}
