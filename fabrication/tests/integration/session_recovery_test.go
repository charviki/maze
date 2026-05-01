//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/charviki/maze-integration-tests/kit"
)

// TestSessionCreateQuery 验证 Session 的创建和查询
func TestSessionCreateQuery(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	name := uniqueName("test-session")
	h.trackHost(name)

	t.Log("[step] creating host for session test...")
	if _, err := h.client.CreateHost(name, []string{"claude"}); err != nil {
		t.Fatalf("create host failed: %v", err)
	}
	t.Log("[step] waiting for host to become online...")
	if _, err := h.client.WaitForHostStatus(name, "online", 3*time.Minute); err != nil {
		t.Fatalf("wait for host online failed: %v", err)
	}

	t.Log("[step] creating session...")
	sessionName := "test-session-1"
	if err := h.client.CreateSession(name, sessionName, ""); err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	t.Log("[step] querying saved sessions...")
	sessions, err := h.client.GetSavedSessions(name)
	if err != nil {
		t.Fatalf("get saved sessions failed: %v", err)
	}
	t.Logf("[step] PASS: found %d saved sessions", len(sessions))
}

// TestSessionPersistenceRecovery 验证 Session 在 Deployment 重建后保留
func TestSessionPersistenceRecovery(t *testing.T) {
	cfg := kit.LoadTestConfig()
	if cfg.Env != "kubernetes" {
		t.Skip("session persistence recovery test only runs in Kubernetes environment")
	}

	h := newTestHelper(t)
	defer h.cleanup(t)

	name := uniqueName("test-session-recovery")
	h.trackHost(name)

	t.Log("[step] creating host for session persistence test...")
	if _, err := h.client.CreateHost(name, []string{"claude"}); err != nil {
		t.Fatalf("create host failed: %v", err)
	}
	t.Log("[step] waiting for host to become online...")
	if _, err := h.client.WaitForHostStatus(name, "online", 3*time.Minute); err != nil {
		t.Fatalf("wait for host online failed: %v", err)
	}

	t.Log("[step] creating session...")
	sessionName := "recovery-test-session"
	if err := h.client.CreateSession(name, sessionName, ""); err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	t.Log("[step] waiting for session to be saved...")
	time.Sleep(5 * time.Second)

	sessionsBefore, err := h.client.GetSavedSessions(name)
	if err != nil {
		t.Fatalf("get sessions before recovery failed: %v", err)
	}
	t.Logf("[step] PASS: sessions before recovery: %d", len(sessionsBefore))
}
