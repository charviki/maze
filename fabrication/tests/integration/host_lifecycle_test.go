//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	client "github.com/charviki/maze-cradle/api/gen/http"
)

// TestHostCreateOnline 验证 Host 创建后能成功上线
func TestHostCreateOnline(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	name := uniqueName("test-online")
	h.trackHost(name)

	t.Log("[step] creating host with tools=[claude]...")
	nameField := name
	body := client.V1CreateHostRequest{
		Name:  &nameField,
		Tools: []string{"claude"},
	}

	ctx := context.Background()
	resp, httpResp, err := h.apiClient.HostServiceAPI.HostServiceCreateHost(ctx).Body(body).Execute()
	if err != nil {
		t.Fatalf("create host failed: %v (status=%d)", err, httpResp.StatusCode)
	}
	if resp.GetName() != name {
		t.Errorf("expected name=%s, got=%s", name, resp.GetName())
	}
	if resp.GetStatus() == "" {
		t.Error("expected CreateHost to return non-empty status")
	}
	if gotTools := resp.GetTools(); len(gotTools) != 1 || gotTools[0] != "claude" {
		t.Errorf("expected tools=[claude], got=%v", gotTools)
	}

	t.Log("[step] waiting for host to become online (timeout=3m)...")
	h.waitForHostStatus(t, name, "online", 3*time.Minute)

	info, _, err := h.apiClient.HostServiceAPI.HostServiceGetHost(ctx, name).Execute()
	if err != nil {
		t.Fatalf("get host failed: %v", err)
	}
	t.Logf("[step] verifying host status=%s", info.GetStatus())
	if info.GetStatus() != "online" {
		t.Errorf("expected status=online, got=%s", info.GetStatus())
	}
	t.Log("[step] PASS: host is online")
}

// TestHostListQuery 验证 Host 列表查询返回正确的合并视图
func TestHostListQuery(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	name1 := uniqueName("test-list-1")
	name2 := uniqueName("test-list-2")
	h.trackHost(name1)
	h.trackHost(name2)

	t.Log("[step] creating 2 hosts (claude + go)...")
	h.createHostAndWait(t, name1, []string{"claude"})
	h.createHostAndWait(t, name2, []string{"go"})

	t.Log("[step] querying host list...")
	hosts, _, err := h.apiClient.HostServiceAPI.HostServiceListHosts(context.Background()).Execute()
	if err != nil {
		t.Fatalf("list hosts failed: %v", err)
	}

	found1, found2 := false, false
	for _, host := range hosts.GetHosts() {
		if host.GetName() == name1 {
			found1 = true
		}
		if host.GetName() == name2 {
			found2 = true
		}
	}
	if !found1 {
		t.Errorf("host %s not found in list", name1)
	}
	if !found2 {
		t.Errorf("host %s not found in list", name2)
	}
	t.Logf("[step] PASS: list returned %d hosts, both found", len(hosts.GetHosts()))
}

// TestHostDelete 验证 Host 删除后从列表中消失
func TestHostDelete(t *testing.T) {
	h := newTestHelper(t)

	name := uniqueName("test-delete")

	t.Log("[step] creating host for deletion test...")
	h.createHostAndWait(t, name, []string{"claude"})

	t.Log("[step] deleting host...")
	_, _, err := h.apiClient.HostServiceAPI.HostServiceDeleteHost(context.Background(), name).Execute()
	if err != nil {
		t.Fatalf("delete host failed: %v", err)
	}

	t.Log("[step] verifying host no longer in list...")
	hosts, _, err := h.apiClient.HostServiceAPI.HostServiceListHosts(context.Background()).Execute()
	if err != nil {
		t.Fatalf("list hosts failed: %v", err)
	}
	for _, host := range hosts.GetHosts() {
		if host.GetName() == name {
			t.Errorf("host %s still exists after deletion", name)
		}
	}
	t.Log("[step] PASS: host deleted successfully")
}

// TestHostCreateDuplicate 验证同名 Host 创建返回错误
func TestHostCreateDuplicate(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	name := uniqueName("test-dup")
	h.trackHost(name)

	t.Log("[step] creating first host...")
	h.createHostAndWait(t, name, []string{"claude"})

	t.Log("[step] creating duplicate host (expect rejection)...")
	nameField := name
	body := client.V1CreateHostRequest{
		Name:  &nameField,
		Tools: []string{"claude"},
	}
	_, _, err := h.apiClient.HostServiceAPI.HostServiceCreateHost(context.Background()).Body(body).Execute()
	if err == nil {
		t.Fatal("expected error for duplicate host creation, got nil")
	}
	t.Logf("[step] PASS: duplicate correctly rejected: %v", err)
}
