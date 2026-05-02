//go:build integration

package integration

import (
	"context"
	"testing"
)

// TestNodeListQuery — Given: 已上线的 Host; When: 查询节点列表; Then: 返回包含该 Host 的节点
func TestNodeListQuery(t *testing.T) {
	t.Parallel()
	h := newTestHelper(t)
	defer h.cleanup(t)

	name := uniqueName("test-node-list")
	h.trackHost(name)

	h.createHostAndWait(t, name, []string{"claude"})

	t.Log("[step] listing nodes...")
	resp, _, err := h.apiClient.NodeServiceAPI.NodeServiceListNodes(context.Background()).Execute()
	if err != nil {
		t.Fatalf("list nodes failed: %v", err)
	}

	found := false
	for _, n := range resp.GetNodes() {
		if n.GetName() == name {
			found = true
			t.Logf("[step] found node: name=%s status=%s grpc_address=%s", n.GetName(), n.GetStatus(), n.GetGrpcAddress())
		}
	}
	if !found {
		t.Errorf("node %s not found in list of %d nodes", name, len(resp.GetNodes()))
	}
	t.Log("[step] PASS: node list query succeeded")
}

// TestNodeGetDetail — Given: 已上线的 Host; When: 查询节点详情; Then: 返回完整的节点信息
func TestNodeGetDetail(t *testing.T) {
	t.Parallel()
	h := newTestHelper(t)
	defer h.cleanup(t)

	name := uniqueName("test-node-detail")
	h.trackHost(name)

	h.createHostAndWait(t, name, []string{"claude"})

	t.Log("[step] getting node detail...")
	node, _, err := h.apiClient.NodeServiceAPI.NodeServiceGetNode(context.Background(), name).Execute()
	if err != nil {
		t.Fatalf("get node failed: %v", err)
	}

	if node.GetName() != name {
		t.Errorf("expected name=%s, got=%s", name, node.GetName())
	}
	if node.GetStatus() != "online" {
		t.Errorf("expected status=online, got=%s", node.GetStatus())
	}
	t.Logf("[step] PASS: node detail name=%s status=%s address=%s", node.GetName(), node.GetStatus(), node.GetAddress())
}
