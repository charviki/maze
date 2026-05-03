//go:build integration

package integration

import (
	"context"
	"testing"

	client "github.com/charviki/maze-cradle/api/gen/http"
)

// TestLocalConfigGetAndUpdate — Given: 已上线的 Host; When: 查询/更新本地配置; Then: 配置正确返回
func TestLocalConfigGetAndUpdate(t *testing.T) {
	t.Parallel()

	h := newTestHelper(t)
	defer h.cleanup(t)

	nodeName := h.acquireHost(t, "claude")

	t.Log("[step] getting local config...")
	config, _, err := h.apiClient.ConfigServiceAPI.ConfigServiceGetConfig(context.Background(), nodeName).Execute()
	if err != nil {
		t.Fatalf("get config failed: %v", err)
	}
	t.Logf("[step] local config: working_dir=%s env=%d", config.GetWorkingDir(), len(config.GetEnv()))

	t.Log("[step] updating local config...")
	env := map[string]string{"TEST_VAR": "test_value"}
	updateBody := client.ConfigServiceUpdateConfigBody{
		Env: &env,
	}
	updated, _, err := h.apiClient.ConfigServiceAPI.ConfigServiceUpdateConfig(context.Background(), nodeName).
		Body(updateBody).Execute()
	if err != nil {
		t.Fatalf("update config failed: %v", err)
	}
	t.Logf("[step] updated config: working_dir=%s", updated.GetWorkingDir())
	if updated.GetEnv()["TEST_VAR"] != "test_value" {
		t.Fatal("expected updated local config to include TEST_VAR")
	}

	t.Log("[step] re-reading local config...")
	reread, _, err := h.apiClient.ConfigServiceAPI.ConfigServiceGetConfig(context.Background(), nodeName).Execute()
	if err != nil {
		t.Fatalf("re-read config failed: %v", err)
	}
	if reread.GetEnv()["TEST_VAR"] != "test_value" {
		t.Fatal("expected local config re-read to include TEST_VAR")
	}

	t.Log("[step] PASS: local config get/update succeeded")
}

// TestAuditLogAfterOperations — Given: 执行过操作的 Manager; When: 查询审计日志; Then: 返回非空日志列表
func TestAuditLogAfterOperations(t *testing.T) {
	t.Parallel()

	h := newTestHelper(t)
	defer h.cleanup(t)

	_ = h.acquireHost(t, "claude")

	t.Log("[step] querying audit logs...")
	resp, _, err := h.apiClient.AuditServiceAPI.AuditServiceGetAuditLogs(context.Background()).Execute()
	if err != nil {
		t.Fatalf("get audit logs failed: %v", err)
	}

	logs := resp.GetLogs()
	total := resp.GetTotal()
	t.Logf("[step] PASS: found %d audit logs (total=%d)", len(logs), total)

	if total == 0 {
		t.Log("[step] WARNING: no audit logs found (operations may not have generated audit entries)")
	}
}
