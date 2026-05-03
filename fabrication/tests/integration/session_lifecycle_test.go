//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	client "github.com/charviki/maze-cradle/api/gen/http"
)

// TestSessionSaveAndRestore — Given: 已上线的 Host 和活跃 Session; When: 保存→查询→恢复; Then: Session 状态完整保留
func TestSessionSaveAndRestore(t *testing.T) {
	t.Parallel()

	h := newTestHelper(t)
	defer h.cleanup(t)

	nodeName := h.acquireHost(t, "claude")

	t.Log("[step] creating session...")
	sid := h.createSession(t, nodeName, "save-restore-session")

	t.Log("[step] saving all sessions...")
	_, _, err := h.apiClient.SessionServiceAPI.SessionServiceSaveSessions(context.Background(), nodeName).Execute()
	if err != nil {
		t.Fatalf("save sessions failed: %v", err)
	}

	t.Log("[step] querying saved sessions...")
	saved, _, err := h.apiClient.SessionServiceAPI.SessionServiceGetSavedSessions(context.Background(), nodeName).Execute()
	if err != nil {
		t.Fatalf("get saved sessions failed: %v", err)
	}

	found := false
	for _, s := range saved.GetSessions() {
		if s.GetSessionName() == "save-restore-session" {
			found = true
			t.Logf("[step] found saved session: name=%s template_id=%s", s.GetSessionName(), s.GetTemplateId())
		}
	}
	if !found {
		t.Error("saved session 'save-restore-session' not found")
	}

	t.Log("[step] restoring session...")
	_, _, err = h.apiClient.SessionServiceAPI.SessionServiceRestoreSession(context.Background(), nodeName, sid).Body(client.SessionServiceRestoreSessionBody{}).Execute()
	if err != nil {
		t.Fatalf("restore session failed: %v", err)
	}

	t.Log("[step] PASS: save and restore succeeded")
}

// TestSessionConfig — Given: 已上线的 Host 和活跃 Session; When: 查询/更新配置; Then: 配置正确返回
func TestSessionConfig(t *testing.T) {
	t.Parallel()

	h := newTestHelper(t)
	defer h.cleanup(t)

	nodeName := h.acquireHost(t, "claude")

	t.Log("[step] creating session...")
	// 只有带模板的会话才有固定文件定义，才能验证配置读写链路。
	sid := h.createSessionWithTemplate(t, nodeName, "config-test-session", "claude")

	t.Log("[step] getting session config...")
	configView, _, err := h.apiClient.SessionServiceAPI.SessionServiceGetSessionConfig(context.Background(), nodeName, sid).Execute()
	if err != nil {
		t.Fatalf("get session config failed: %v", err)
	}
	if configView.GetSessionId() != sid {
		t.Fatalf("expected session id=%s, got=%s", sid, configView.GetSessionId())
	}
	if len(configView.GetFiles()) == 0 {
		t.Fatal("expected session config to expose editable files")
	}
	firstFile := configView.GetFiles()[0]
	updatedContent := firstFile.GetContent() + "\n# integration session config\n"
	t.Logf("[step] session config: id=%s scope=%s files=%d", configView.GetSessionId(), configView.GetScope(), len(configView.GetFiles()))

	t.Log("[step] updating session config...")
	updateBody := client.SessionServiceUpdateSessionConfigBody{
		Files: []client.V1ConfigFileUpdate{
			{
				Path:     ptrStr(firstFile.GetPath()),
				Content:  ptrStr(updatedContent),
				BaseHash: ptrStr(firstFile.GetHash()),
			},
		},
	}
	updatedView, _, err := h.apiClient.SessionServiceAPI.SessionServiceUpdateSessionConfig(context.Background(), nodeName, sid).
		Body(updateBody).Execute()
	if err != nil {
		t.Fatalf("update session config failed: %v", err)
	}
	if len(updatedView.GetFiles()) == 0 || updatedView.GetFiles()[0].GetContent() != updatedContent {
		t.Fatalf("expected updated session config content to be persisted")
	}

	t.Log("[step] re-reading session config...")
	rereadView, _, err := h.apiClient.SessionServiceAPI.SessionServiceGetSessionConfig(context.Background(), nodeName, sid).Execute()
	if err != nil {
		t.Fatalf("re-read session config failed: %v", err)
	}
	if len(rereadView.GetFiles()) == 0 || rereadView.GetFiles()[0].GetContent() != updatedContent {
		t.Fatalf("expected session config re-read to return persisted content")
	}

	t.Log("[step] PASS: session config get/update succeeded")
}

// TestSessionDelete — Given: 已上线的 Host 和活跃 Session; When: 删除 Session; Then: Session 不再出现在列表中
func TestSessionDelete(t *testing.T) {
	t.Parallel()

	h := newTestHelper(t)
	defer h.cleanup(t)

	nodeName := h.acquireHost(t, "claude")

	t.Log("[step] creating session...")
	sid := h.createSession(t, nodeName, "delete-test-session")

	t.Log("[step] deleting session...")
	_, _, err := h.apiClient.SessionServiceAPI.SessionServiceDeleteSession(context.Background(), nodeName, sid).Execute()
	if err != nil {
		t.Fatalf("delete session failed: %v", err)
	}

	t.Log("[step] verifying session no longer in list...")
	time.Sleep(1 * time.Second)
	sessions, _, err := h.apiClient.SessionServiceAPI.SessionServiceListSessions(context.Background(), nodeName).Execute()
	if err != nil {
		t.Fatalf("list sessions failed: %v", err)
	}
	for _, s := range sessions.GetSessions() {
		if s.GetId() == sid {
			t.Errorf("session %s still exists after deletion", sid)
		}
	}
	t.Log("[step] PASS: session deleted successfully")
}
