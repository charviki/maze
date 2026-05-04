package reconciler

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/charviki/maze-cradle/protocol"
)

func TestReconciler_RecoverOnStartup_RestoresHostToken(t *testing.T) {
	rt := newMockReconcilerRuntime()
	rt.healthy["host-token"] = true

	rec, specMgr := newTestReconciler(t, rt)
	specMgr.Create(&protocol.HostSpec{
		Name:      "host-token",
		Tools:     []string{"claude"},
		Status:    protocol.HostStatusOnline,
		AuthToken: "restored-token",
	})

	rec.RecoverOnStartup(context.Background())

	exists, matched := rec.registry.ValidateHostToken("host-token", "restored-token")
	if !exists || !matched {
		t.Fatalf("ValidateHostToken = (%v, %v), want (true, true)", exists, matched)
	}
}

func TestReconciler_HealthCheck_FailedRetryIncrementsRetryCount(t *testing.T) {
	rt := newMockReconcilerRuntime()

	rec, specMgr := newTestReconciler(t, rt)
	specMgr.Create(&protocol.HostSpec{
		Name:       "host-retry",
		Tools:      []string{"claude"},
		Status:     protocol.HostStatusFailed,
		RetryCount: 0,
		AuthToken:  "retry-token",
	})

	rec.runHealthCheck(context.Background())

	if atomic.LoadInt32(&rt.deployCall) != 1 {
		t.Fatalf("deploy calls = %d, want 1", rt.deployCall)
	}

	got := specMgr.Get("host-retry")
	if got == nil {
		t.Fatal("HostSpec 不应丢失")
	}
	if got.RetryCount != 1 {
		t.Fatalf("RetryCount = %d, want 1", got.RetryCount)
	}
	if got.Status != protocol.HostStatusDeploying {
		t.Fatalf("Status = %q, want %q", got.Status, protocol.HostStatusDeploying)
	}
}
