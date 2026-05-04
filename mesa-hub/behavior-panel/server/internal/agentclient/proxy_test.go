package agentclient

import (
	"path/filepath"
	"testing"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	filerepo "github.com/charviki/mesa-hub-behavior-panel/internal/repository/file"
)

func TestProxyGetNodeAddrErrors(t *testing.T) {
	registry := filerepo.NewNodeRegistry(filepath.Join(t.TempDir(), "nodes.json"), logutil.NewNop())
	t.Cleanup(registry.WaitSave)

	proxy := NewProxy(registry, nil)

	if _, err := proxy.getNodeAddr("missing"); err == nil {
		t.Fatal("缺失节点应返回错误")
	}

	registry.Register(protocol.RegisterRequest{
		Name:    "node-1",
		Address: "http://node-1:8080",
	})
	if _, err := proxy.getNodeAddr("node-1"); err == nil {
		t.Fatal("缺少 gRPC 地址应返回错误")
	}
}

func TestProxyGetNodeAddrReturnsGrpcAddress(t *testing.T) {
	registry := filerepo.NewNodeRegistry(filepath.Join(t.TempDir(), "nodes.json"), logutil.NewNop())
	t.Cleanup(registry.WaitSave)

	registry.Register(protocol.RegisterRequest{
		Name:        "node-1",
		Address:     "http://node-1:8080",
		GrpcAddress: "node-1:9090",
	})

	proxy := NewProxy(registry, nil)
	addr, err := proxy.getNodeAddr("node-1")
	if err != nil {
		t.Fatalf("getNodeAddr 返回错误: %v", err)
	}
	if addr != "node-1:9090" {
		t.Fatalf("addr = %q, want %q", addr, "node-1:9090")
	}
}
