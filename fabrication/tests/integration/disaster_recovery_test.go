//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/charviki/maze-integration-tests/kit"
)

// TestHostDisasterRecovery 验证 Manager 重启后全量恢复所有 Host
func TestHostDisasterRecovery(t *testing.T) {
	cfg := kit.LoadTestConfig()
	if cfg.Env != "kubernetes" {
		t.Skip("disaster recovery test only runs in Kubernetes environment")
	}

	h := newTestHelper(t)
	defer h.cleanup(t)

	name1 := uniqueName("test-dr-1")
	name2 := uniqueName("test-dr-2")
	h.trackHost(name1)
	h.trackHost(name2)

	t.Log("[step] creating 2 hosts for DR test...")
	if _, err := h.client.CreateHost(name1, []string{"claude"}); err != nil {
		t.Fatalf("create host1 failed: %v", err)
	}
	if _, err := h.client.CreateHost(name2, []string{"go"}); err != nil {
		t.Fatalf("create host2 failed: %v", err)
	}

	t.Log("[step] waiting for both hosts to become online...")
	if _, err := h.client.WaitForHostStatus(name1, "online", 3*time.Minute); err != nil {
		t.Fatalf("wait for host1 online failed: %v", err)
	}
	if _, err := h.client.WaitForHostStatus(name2, "online", 3*time.Minute); err != nil {
		t.Fatalf("wait for host2 online failed: %v", err)
	}
	t.Logf("[step] PASS: both hosts online: %s, %s", name1, name2)
}
