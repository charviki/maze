package service_test

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	service "github.com/charviki/maze/the-mesa/director-core/internal/service"
)

type nodeServiceRegistryStub struct {
	mu        sync.RWMutex
	nodes     map[string]*service.Node
	getErr    error
	deleteErr error
}

func newNodeServiceRegistryStub() *nodeServiceRegistryStub {
	return &nodeServiceRegistryStub{nodes: make(map[string]*service.Node)}
}

func (s *nodeServiceRegistryStub) StoreHostToken(_ context.Context, name, token string) error {
	return nil
}
func (s *nodeServiceRegistryStub) ValidateHostToken(_ context.Context, name, token string) (bool, bool, error) {
	return false, false, nil
}
func (s *nodeServiceRegistryStub) RemoveHostToken(_ context.Context, name string) error {
	return nil
}

func (s *nodeServiceRegistryStub) Register(_ context.Context, req protocol.RegisterRequest) (*service.Node, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	node := &service.Node{Name: req.Name, Address: req.Address, Status: service.NodeStatusOnline}
	s.nodes[req.Name] = node
	return node, nil
}

func (s *nodeServiceRegistryStub) Heartbeat(_ context.Context, req protocol.HeartbeatRequest) (*service.Node, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	node := s.nodes[req.Name]
	if node == nil {
		return nil, nil
	}
	node.AgentStatus = req.Status
	node.Status = service.NodeStatusOnline
	return node, nil
}

func (s *nodeServiceRegistryStub) List(_ context.Context) ([]*service.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*service.Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		out = append(out, node)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s *nodeServiceRegistryStub) Get(_ context.Context, name string) (*service.Node, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nodes[name], nil
}

func (s *nodeServiceRegistryStub) Delete(_ context.Context, name string) (bool, error) {
	if s.deleteErr != nil {
		return false, s.deleteErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.nodes[name]; !ok {
		return false, nil
	}
	delete(s.nodes, name)
	return true, nil
}

func (s *nodeServiceRegistryStub) GetNodeCount(_ context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.nodes), nil
}

func (s *nodeServiceRegistryStub) GetOnlineCount(_ context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, node := range s.nodes {
		if node.Status == service.NodeStatusOnline {
			count++
		}
	}
	return count, nil
}

func newNodeServiceTestEnv(t *testing.T) (*service.NodeService, *nodeServiceRegistryStub) {
	t.Helper()

	registry := newNodeServiceRegistryStub()
	svc := service.NewNodeService(registry, logutil.NewNop())
	return svc, registry
}

func TestNodeService_ListNodes(t *testing.T) {
	svc, registry := newNodeServiceTestEnv(t)
	_, _ = registry.Register(context.Background(), protocol.RegisterRequest{Name: "node-b", Address: "http://node-b"})
	_, _ = registry.Register(context.Background(), protocol.RegisterRequest{Name: "node-a", Address: "http://node-a"})

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
	_, _ = registry.Register(context.Background(), protocol.RegisterRequest{Name: "node-1", Address: "http://node-1"})

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
	_, _ = registry.Register(context.Background(), protocol.RegisterRequest{Name: "node-delete", Address: "http://node-delete"})

	if err := svc.DeleteNode(context.Background(), "node-delete"); err != nil {
		t.Fatalf("DeleteNode 返回错误: %v", err)
	}
	if got, _ := registry.Get(context.Background(), "node-delete"); got != nil {
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

func TestNodeService_PropagatesRegistryErrors(t *testing.T) {
	svc, registry := newNodeServiceTestEnv(t)
	registry.getErr = errors.New("get failed")
	registry.deleteErr = errors.New("delete failed")

	if _, err := svc.GetNode(context.Background(), "node-1"); err == nil {
		t.Fatal("registry get 失败时应返回错误")
	}
	if err := svc.DeleteNode(context.Background(), "node-1"); err == nil {
		t.Fatal("registry delete 失败时应返回错误")
	}
}
