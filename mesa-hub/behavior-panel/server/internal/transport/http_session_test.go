package transport

import (
	"context"
	"testing"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/internal/service"
)

type testNodeRegistry struct{}
type testAuditLogWriter struct{}

func (r *testNodeRegistry) StoreHostToken(_ context.Context, name, token string) error { return nil }
func (r *testNodeRegistry) ValidateHostToken(_ context.Context, name, token string) (bool, bool, error) {
	return true, true, nil
}
func (r *testNodeRegistry) RemoveHostToken(_ context.Context, name string) error { return nil }
func (r *testNodeRegistry) Register(_ context.Context, req protocol.RegisterRequest) (*service.Node, error) {
	return &service.Node{Name: req.Name}, nil
}
func (r *testNodeRegistry) Heartbeat(_ context.Context, req protocol.HeartbeatRequest) (*service.Node, error) {
	return &service.Node{Name: req.Name}, nil
}
func (r *testNodeRegistry) List(_ context.Context) ([]*service.Node, error) { return nil, nil }
func (r *testNodeRegistry) Get(_ context.Context, name string) (*service.Node, error) {
	if name == "" {
		return nil, nil
	}
	return &service.Node{Name: name}, nil
}
func (r *testNodeRegistry) Delete(_ context.Context, name string) (bool, error) { return true, nil }
func (r *testNodeRegistry) GetNodeCount(_ context.Context) (int, error)         { return 0, nil }
func (r *testNodeRegistry) GetOnlineCount(_ context.Context) (int, error)       { return 0, nil }

func (w *testAuditLogWriter) Log(_ context.Context, entry protocol.AuditLogEntry) error { return nil }

func TestNewSessionProxyHandler(t *testing.T) {
	h := NewSessionProxyHandler(
		&testNodeRegistry{},
		&testAuditLogWriter{},
		logutil.NewNop(),
		"test-token",
		[]string{"http://localhost:3000"},
		false,
	)
	if h == nil {
		t.Fatal("NewSessionProxyHandler returned nil")
	}
	if h.authToken != "test-token" {
		t.Errorf("authToken = %q, want %q", h.authToken, "test-token")
	}
	if len(h.allowedOrigins) != 1 || h.allowedOrigins[0] != "http://localhost:3000" {
		t.Errorf("allowedOrigins = %v, want [http://localhost:3000]", h.allowedOrigins)
	}
	if h.allowPrivateNetworks {
		t.Error("allowPrivateNetworks should be false")
	}
}
