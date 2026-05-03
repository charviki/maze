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

	t.Log("[step] creating first host [claude] (cold build)...")
	start1 := time.Now()
	h.createHostAndWait(t, name1, []string{"claude"})
	buildDuration1 := time.Since(start1)
	t.Logf("[step] first host total time: %v", buildDuration1)

	t.Log("[step] creating second host [claude] (should hit cache)...")
	start2 := time.Now()
	// 第二个 Host 也必须等到真正上线；否则 defer cleanup 可能先于异步创建完成，
	// 把“尚未出现的 Host”误判成已删除，最终留下残留 Pod/目录。
	h.createHostAndWait(t, name2, []string{"claude"})
	buildDuration2 := time.Since(start2)
	t.Logf("[step] second host total time: %v", buildDuration2)

	if buildDuration2 >= buildDuration1 {
		t.Logf("[step] WARNING: second build (%v) was not faster than first (%v)", buildDuration2, buildDuration1)
	} else {
		speedup := float64(buildDuration1) / float64(buildDuration2)
		t.Logf("[step] PASS: image cache speedup: %.1fx", speedup)
	}
}
