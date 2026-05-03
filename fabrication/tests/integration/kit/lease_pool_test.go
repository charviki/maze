package kit

import "testing"

func TestLeasePoolAcquireRelease(t *testing.T) {
	pool := NewLeasePool(map[string][]string{
		"claude": {"host-a", "host-b"},
	})

	first, err := pool.Acquire("claude")
	if err != nil {
		t.Fatalf("acquire first lease: %v", err)
	}
	second, err := pool.Acquire("claude")
	if err != nil {
		t.Fatalf("acquire second lease: %v", err)
	}
	if first == second {
		t.Fatalf("expected unique leases, got %q twice", first)
	}

	if err := pool.Release("claude", first); err != nil {
		t.Fatalf("release first lease: %v", err)
	}
	reused, err := pool.Acquire("claude")
	if err != nil {
		t.Fatalf("re-acquire released lease: %v", err)
	}
	if reused != first {
		t.Fatalf("expected released lease %q to be reusable, got %q", first, reused)
	}
}

func TestLeasePoolReleaseRejectsOverflow(t *testing.T) {
	pool := NewLeasePool(map[string][]string{
		"go": {"host-go-1"},
	})

	if err := pool.Release("go", "host-go-1"); err == nil {
		t.Fatal("expected overflow release to fail")
	}
}
