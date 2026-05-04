package service_test

import (
	"context"
	"sort"
	"sync"
	"testing"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	service "github.com/charviki/mesa-hub-behavior-panel/internal/service"
)

type nodeServiceRegistryStub struct {
	mu    sync.RWMutex
	nodes map[string]*service.Node
}

func newNodeServiceRegistryStub() *nodeServiceRegistryStub {
	return &nodeServiceRegistryStub{nodes: make(map[string]*service.Node)}
}

func (s *nodeServiceRegistryStub) StoreHostToken(name, token string) {}
func (s *nodeServiceRegistryStub) ValidateHostToken(name, token string) (bool, bool) {
	return false, false
}
func (s *nodeServiceRegistryStub) RemoveHostToken(name string) {}

func (s *nodeServiceRegistryStub) Register(req protocol.RegisterRequest) *service.Node {
	s.mu.Lock()
	defer s.mu.Unlock()
	node := &service.Node{Name: req.Name, Address: req.Address, Status: service.NodeStatusOnline}
	s.nodes[req.Name] = node
	return node
}

func (s *nodeServiceRegistryStub) Heartbeat(req protocol.HeartbeatRequest) *service.Node {
	s.mu.Lock()
	defer s.mu.Unlock()
	node := s.nodes[req.Name]
	if node == nil {
		return nil
	}
	node.AgentStatus = req.Status
	node.Status = service.NodeStatusOnline
	return node
}

func (s *nodeServiceRegistryStub) List() []*service.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*service.Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		out = append(out, node)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *nodeServiceRegistryStub) Get(name string) *service.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nodes[name]
}

func (s *nodeServiceRegistryStub) Delete(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.nodes[name]; !ok {
		return false
	}
	delete(s.nodes, name)
	return true
}

func (s *nodeServiceRegistryStub) GetNodeCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.nodes)
}

func (s *nodeServiceRegistryStub) GetOnlineCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, node := range s.nodes {
		if node.Status == service.NodeStatusOnline {
			count++
		}
	}
	return count
}

func (s *nodeServiceRegistryStub) WaitSave() {}

func newNodeServiceTestEnv(t *testing.T) (*service.NodeService, *nodeServiceRegistryStub) {
	t.Helper()

	registry := newNodeServiceRegistryStub()
	svc := service.NewNodeService(registry, logutil.NewNop())
	return svc, registry
}

func TestNodeService_ListNodes(t *testing.T) {
	svc, registry := newNodeServiceTestEnv(t)
	registry.Register(protocol.RegisterRequest{Name: "node-b", Address: "http://node-b"})
	registry.Register(protocol.RegisterRequest{Name: "node-a", Address: "http://node-a"})

	nodes, err := svc.ListNodes(context.Background())
	if err != nil {
		t.Fatalf("ListNodes 返回错误: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("nodes len = %d, want 2", len(nodes))
	}
	if nodes[0].Name != "node-a" || nodes[1].Name != "node-b" {
		t.Fatalf("nodes order = [%s %s], want [node-a node-b]", nodes[0].Name, nodes[1].Name)
	}
}

func TestNodeService_GetNode(t *testing.T) {
	svc, registry := newNodeServiceTestEnv(t)
	registry.Register(protocol.RegisterRequest{Name: "node-1", Address: "http://node-1"})

	node, err := svc.GetNode(context.Background(), "node-1")
	if err != nil {
		t.Fatalf("GetNode 返回错误: %v", err)
	}
	if node == nil || node.Name != "node-1" {
		t.Fatalf("GetNode = %#v, want node-1", node)
	}
}

func TestNodeService_GetNode_ValidatesInputAndMissingNode(t *testing.T) {
	svc, _ := newNodeServiceTestEnv(t)

	if _, err := svc.GetNode(context.Background(), ""); err == nil {
		t.Fatal("空名称应返回错误")
	}
	if _, err := svc.GetNode(context.Background(), "missing"); err == nil {
		t.Fatal("不存在节点应返回错误")
	}
}

func TestNodeService_DeleteNode(t *testing.T) {
	svc, registry := newNodeServiceTestEnv(t)
	registry.Register(protocol.RegisterRequest{Name: "node-delete", Address: "http://node-delete"})

	if err := svc.DeleteNode(context.Background(), "node-delete"); err != nil {
		t.Fatalf("DeleteNode 返回错误: %v", err)
	}
	if registry.Get("node-delete") != nil {
		t.Fatal("DeleteNode 成功后应从注册表移除节点")
	}
}

func TestNodeService_DeleteNode_ValidatesInputAndMissingNode(t *testing.T) {
	svc, _ := newNodeServiceTestEnv(t)

	if err := svc.DeleteNode(context.Background(), ""); err == nil {
		t.Fatal("空名称应返回错误")
	}
	if err := svc.DeleteNode(context.Background(), "missing"); err == nil {
		t.Fatal("不存在节点应返回错误")
	}
}
