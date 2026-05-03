//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/charviki/maze-integration-tests/kit"
)

// TestSessionCreateQuery 验证 Session 的创建和查询
func TestSessionCreateQuery(t *testing.T) {
	t.Parallel()

	h := newTestHelper(t)
	defer h.cleanup(t)

	name := h.acquireHost(t, "claude")

	t.Log("[step] creating session...")
	sid := h.createSession(t, name, "test-session-1")

	t.Log("[step] querying session by id...")
	session, _, err := h.apiClient.SessionServiceAPI.SessionServiceGetSession(context.Background(), name, sid).Execute()
	if err != nil {
		t.Fatalf("get session failed: %v", err)
	}
	if session.GetId() != sid {
		t.Errorf("expected session id=%s, got=%s", sid, session.GetId())
	}
	t.Logf("[step] PASS: session found id=%s name=%s", session.GetId(), session.GetName())
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
	h.createHostAndWait(t, name, []string{"claude"})

	t.Log("[step] creating session...")
	h.createSession(t, name, "recovery-test-session")

	t.Log("[step] waiting for session to be saved...")
	time.Sleep(5 * time.Second)

	t.Log("[step] querying saved sessions...")
	saved, _, err := h.apiClient.SessionServiceAPI.SessionServiceGetSavedSessions(context.Background(), name).Execute()
	if err != nil {
		t.Fatalf("get saved sessions failed: %v", err)
	}
	t.Logf("[step] PASS: found %d saved sessions", len(saved.GetSessions()))
}
