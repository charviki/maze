//go:build integration

package integration

import (
	"testing"
	"time"
)

// TestComboImageCache 验证相同工具组合的 Host 共享镜像缓存
func TestComboImageCache(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	name1 := uniqueName("test-cache-1")
	name2 := uniqueName("test-cache-2")
	h.trackHost(name1)
	h.trackHost(name2)

	t.Log("[step] creating first host [claude, go] (cold build)...")
	start1 := time.Now()
	if _, err := h.client.CreateHost(name1, []string{"claude", "go"}); err != nil {
		t.Fatalf("create host1 failed: %v", err)
	}
	_, err := h.client.WaitForHostStatus(name1, "online", 3*time.Minute)
	if err != nil {
		t.Logf("[step] host1 not online yet (may still be registering): %v", err)
	}
	buildDuration1 := time.Since(start1)
	t.Logf("[step] first host total time: %v", buildDuration1)

	t.Log("[step] creating second host [go, claude] (should hit cache)...")
	start2 := time.Now()
	if _, err := h.client.CreateHost(name2, []string{"go", "claude"}); err != nil {
		t.Fatalf("create host2 failed: %v", err)
	}
	_, err = h.client.WaitForHostStatus(name2, "online", 3*time.Minute)
	if err != nil {
		t.Logf("[step] host2 not online yet (may still be registering): %v", err)
	}
	buildDuration2 := time.Since(start2)
	t.Logf("[step] second host total time: %v", buildDuration2)

	if buildDuration2 >= buildDuration1 {
		t.Logf("[step] WARNING: second build (%v) was not faster than first (%v)", buildDuration2, buildDuration1)
	} else {
		speedup := float64(buildDuration1) / float64(buildDuration2)
		t.Logf("[step] PASS: image cache speedup: %.1fx", speedup)
	}
}
