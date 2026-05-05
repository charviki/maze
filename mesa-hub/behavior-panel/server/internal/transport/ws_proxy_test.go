package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/internal/service"
)

type wsTestNodeRegistry struct {
	nodes map[string]*service.Node
}

func (r *wsTestNodeRegistry) StoreHostToken(_ context.Context, name, token string) error { return nil }
func (r *wsTestNodeRegistry) ValidateHostToken(_ context.Context, name, token string) (bool, bool, error) {
	return true, true, nil
}
func (r *wsTestNodeRegistry) RemoveHostToken(_ context.Context, name string) error { return nil }
func (r *wsTestNodeRegistry) Register(_ context.Context, req protocol.RegisterRequest) (*service.Node, error) {
	return &service.Node{Name: req.Name}, nil
}
func (r *wsTestNodeRegistry) Heartbeat(_ context.Context, req protocol.HeartbeatRequest) (*service.Node, error) {
	return &service.Node{Name: req.Name}, nil
}
func (r *wsTestNodeRegistry) List(_ context.Context) ([]*service.Node, error) {
	result := make([]*service.Node, 0, len(r.nodes))
	for _, n := range r.nodes {
		result = append(result, n)
	}
	return result, nil
}
func (r *wsTestNodeRegistry) Get(_ context.Context, name string) (*service.Node, error) {
	if r.nodes == nil {
		return nil, nil
	}
	return r.nodes[name], nil
}
func (r *wsTestNodeRegistry) Delete(_ context.Context, name string) (bool, error) { return true, nil }
func (r *wsTestNodeRegistry) GetNodeCount(_ context.Context) (int, error) {
	if r.nodes == nil {
		return 0, nil
	}
	return len(r.nodes), nil
}
func (r *wsTestNodeRegistry) GetOnlineCount(_ context.Context) (int, error) { return 0, nil }

type wsTestAuditLog struct {
	entries []protocol.AuditLogEntry
}

func (a *wsTestAuditLog) Log(_ context.Context, entry protocol.AuditLogEntry) error {
	a.entries = append(a.entries, entry)
	return nil
}

func newWSTestHandler(t *testing.T) *SessionProxyHandler {
	t.Helper()
	return NewSessionProxyHandler(
		&wsTestNodeRegistry{nodes: map[string]*service.Node{
			"agent-1": {Name: "agent-1", Address: "http://10.0.0.1:9090", GrpcAddress: "10.0.0.1:9090"},
		}},
		&wsTestAuditLog{},
		logutil.NewNop(),
		"test-token",
		nil,
		false,
	)
}

func TestProxyWebSocket_MissingPathParams(t *testing.T) {
	h := newWSTestHandler(t)

	tests := []struct {
		name string
		path string
	}{
		{name: "no params", path: "/"},
		{name: "only name", path: "/{name}"},
		{name: "only id", path: "/{id}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/v1/sessions/ws", nil)
			h.ProxyWebSocket(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestProxyWebSocket_NodeNotFound(t *testing.T) {
	h := newWSTestHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/sessions/nonexistent/sess-1/ws", nil)
	req.SetPathValue("name", "nonexistent")
	req.SetPathValue("id", "sess-1")

	h.ProxyWebSocket(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestProxyWebSocket_AuditLogWritten(t *testing.T) {
	audit := &wsTestAuditLog{}
	reg := &wsTestNodeRegistry{nodes: map[string]*service.Node{
		"agent-1": {Name: "agent-1", Address: "http://127.0.0.1:1", GrpcAddress: "127.0.0.1:1"},
	}}
	h := NewSessionProxyHandler(reg, audit, logutil.NewNop(), "", nil, false)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/sessions/agent-1/sess-1/ws", nil)
	req.SetPathValue("name", "agent-1")
	req.SetPathValue("id", "sess-1")
	h.ProxyWebSocket(rec, req)

	if len(audit.entries) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(audit.entries))
	}
	if audit.entries[0].Action != "websocket_connect" {
		t.Errorf("action = %q, want websocket_connect", audit.entries[0].Action)
	}
	if audit.entries[0].TargetNode != "agent-1" {
		t.Errorf("target = %q, want agent-1", audit.entries[0].TargetNode)
	}
}
