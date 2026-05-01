//go:build integration

package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/charviki/maze-integration-tests/kit"
)

type testHelper struct {
	client *kit.APIClient
	cfg    *kit.TestConfig
	clean  []string
}

func newTestHelper(t *testing.T) *testHelper {
	t.Helper()
	cfg := kit.LoadTestConfig()
	client := kit.NewAPIClient(cfg.ManagerURL, cfg.AuthToken)

	if _, err := client.ListHosts(); err != nil {
		t.Skipf("Manager API not available at %s: %v", cfg.ManagerURL, err)
	}

	return &testHelper{client: client, cfg: cfg}
}

func (h *testHelper) cleanup(t *testing.T) {
	t.Helper()
	t.Logf("[step] cleanup: deleting %d hosts", len(h.clean))
	for _, name := range h.clean {
		if err := h.client.DeleteHost(name); err != nil {
			t.Logf("[cleanup] failed to delete host %s: %v", name, err)
		}
	}
}

func (h *testHelper) trackHost(name string) {
	h.clean = append(h.clean, name)
}

func uniqueName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixMilli())
}

// TestHostCreateOnline 验证 Host 创建后能成功上线
func TestHostCreateOnline(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	name := uniqueName("test-online")
	h.trackHost(name)

	t.Log("[step] creating host with tools=[claude]...")
	info, err := h.client.CreateHost(name, []string{"claude"})
	if err != nil {
		t.Fatalf("create host failed: %v", err)
	}
	if info.Name != name {
		t.Errorf("expected name=%s, got=%s", name, info.Name)
	}

	t.Log("[step] waiting for host to become online (timeout=3m)...")
	onlineInfo, err := h.client.WaitForHostStatus(name, "online", 3*time.Minute)
	if err != nil {
		t.Fatalf("wait for host online failed: %v", err)
	}

	t.Logf("[step] verifying heartbeat (address=%s)", onlineInfo.Address)
	if onlineInfo.LastHeartbeat == "" {
		t.Error("expected last_heartbeat to be set for online host")
	}
	t.Log("[step] PASS: host is online with heartbeat")
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
	if _, err := h.client.CreateHost(name1, []string{"claude"}); err != nil {
		t.Fatalf("create host1 failed: %v", err)
	}
	if _, err := h.client.CreateHost(name2, []string{"go"}); err != nil {
		t.Fatalf("create host2 failed: %v", err)
	}

	t.Log("[step] querying host list...")
	hosts, err := h.client.ListHosts()
	if err != nil {
		t.Fatalf("list hosts failed: %v", err)
	}

	found1, found2 := false, false
	for _, host := range hosts {
		if host.Name == name1 {
			found1 = true
		}
		if host.Name == name2 {
			found2 = true
		}
	}
	if !found1 {
		t.Errorf("host %s not found in list", name1)
	}
	if !found2 {
		t.Errorf("host %s not found in list", name2)
	}
	t.Logf("[step] PASS: list returned %d hosts, both found", len(hosts))
}

// TestHostDelete 验证 Host 删除后从列表中消失
func TestHostDelete(t *testing.T) {
	h := newTestHelper(t)

	name := uniqueName("test-delete")

	t.Log("[step] creating host for deletion test...")
	if _, err := h.client.CreateHost(name, []string{"claude"}); err != nil {
		t.Fatalf("create host failed: %v", err)
	}

	t.Log("[step] deleting host...")
	if err := h.client.DeleteHost(name); err != nil {
		t.Fatalf("delete host failed: %v", err)
	}

	t.Log("[step] verifying host no longer in list...")
	hosts, err := h.client.ListHosts()
	if err != nil {
		t.Fatalf("list hosts failed: %v", err)
	}
	for _, host := range hosts {
		if host.Name == name {
			t.Errorf("host %s still exists after deletion", name)
		}
	}
	t.Log("[step] PASS: host deleted successfully")
}

// TestHostCreateDuplicate 验证同名 Host 创建返回 409
func TestHostCreateDuplicate(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	name := uniqueName("test-dup")
	h.trackHost(name)

	t.Log("[step] creating first host...")
	if _, err := h.client.CreateHost(name, []string{"claude"}); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	t.Log("[step] creating duplicate host (expect rejection)...")
	_, err := h.client.CreateHost(name, []string{"claude"})
	if err == nil {
		t.Fatal("expected error for duplicate host creation, got nil")
	}
	t.Logf("[step] PASS: duplicate correctly rejected: %v", err)
}
