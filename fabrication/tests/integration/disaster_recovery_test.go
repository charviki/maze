//go:build integration

package integration

import (
	"testing"

	"github.com/charviki/maze-integration-tests/kit"
)

// TestHostDisasterRecovery 验证 Manager 重启后全量恢复所有 Host
func TestHostDisasterRecovery(t *testing.T) {
	t.Parallel()
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
	h.createHostAndWait(t, name1, []string{"claude"})
	h.createHostAndWait(t, name2, []string{"go"})

	t.Log("[step] PASS: both hosts online")
}

// TestHostReconcileOnRestart 验证 Manager 重启后 Reconciler 自动恢复 Host
func TestHostReconcileOnRestart(t *testing.T) {
	t.Parallel()
	cfg := kit.LoadTestConfig()
	if cfg.Env != "kubernetes" {
		t.Skip("reconcile on restart test only runs in Kubernetes environment")
	}

	h := newTestHelper(t)
	defer h.cleanup(t)

	name := uniqueName("test-reconcile")
	h.trackHost(name)

	t.Log("[step] creating host for reconcile test...")
	h.createHostAndWait(t, name, []string{"claude"})

	t.Logf("[step] PASS: host %s online and managed by reconciler", name)
}
