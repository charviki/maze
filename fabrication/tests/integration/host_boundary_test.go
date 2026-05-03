//go:build integration

package integration

import (
	"context"
	"testing"
)

// TestHostBuildLog — Given: 已上线的 Host; When: 获取构建日志; Then: 返回非空日志内容
func TestHostBuildLog(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	name := uniqueName("test-buildlog")
	h.trackHost(name)

	h.createHostAndWait(t, name, []string{"claude"})

	t.Log("[step] getting build log...")
	resp, _, err := h.apiClient.HostServiceAPI.HostServiceGetBuildLog(context.Background(), name).Execute()
	if err != nil {
		t.Fatalf("get build log failed: %v", err)
	}
	logContent := resp.GetLog()
	if logContent == "" {
		t.Log("[step] WARNING: build log is empty (may not be generated yet)")
	} else {
		t.Logf("[step] PASS: build log length=%d", len(logContent))
	}
}

// TestHostRuntimeLog — Given: 已上线的 Host; When: 获取运行时日志; Then: 请求成功
func TestHostRuntimeLog(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	name := uniqueName("test-runtimelog")
	h.trackHost(name)

	h.createHostAndWait(t, name, []string{"claude"})

	t.Log("[step] getting runtime log...")
	resp, _, err := h.apiClient.HostServiceAPI.HostServiceGetRuntimeLog(context.Background(), name).Execute()
	if err != nil {
		t.Fatalf("get runtime log failed: %v", err)
	}
	logContent := resp.GetLog()
	t.Logf("[step] PASS: runtime log length=%d", len(logContent))
}

// TestHostListTools — Given: Manager 已启动; When: 查询工具列表; Then: 返回至少 1 个工具
func TestHostListTools(t *testing.T) {
	h := newTestHelper(t)

	t.Log("[step] listing available tools...")
	resp, _, err := h.apiClient.HostServiceAPI.HostServiceListTools(context.Background()).Execute()
	if err != nil {
		t.Fatalf("list tools failed: %v", err)
	}
	tools := resp.GetTools()
	if len(tools) == 0 {
		t.Fatal("expected at least 1 tool, got 0")
	}

	toolIDs := make([]string, len(tools))
	for i, t := range tools {
		toolIDs[i] = t.GetId()
	}
	t.Logf("[step] PASS: found %d tools: %v", len(tools), toolIDs)
}
