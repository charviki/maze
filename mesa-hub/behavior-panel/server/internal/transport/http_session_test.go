package transport

import (
	"testing"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/internal/service"
)

type testNodeRegistry struct{}
type testAuditLogWriter struct{}

func (r *testNodeRegistry) StoreHostToken(name, token string)                 {}
func (r *testNodeRegistry) ValidateHostToken(name, token string) (bool, bool) { return true, true }
func (r *testNodeRegistry) RemoveHostToken(name string)                       {}
func (r *testNodeRegistry) Register(req protocol.RegisterRequest) *service.Node {
	return &service.Node{Name: req.Name}
}
func (r *testNodeRegistry) Heartbeat(req protocol.HeartbeatRequest) *service.Node {
	return &service.Node{Name: req.Name}
}
func (r *testNodeRegistry) List() []*service.Node { return nil }
func (r *testNodeRegistry) Get(name string) *service.Node {
	if name == "" {
		return nil
	}
	return &service.Node{Name: name}
}
func (r *testNodeRegistry) Delete(name string) bool { return true }
func (r *testNodeRegistry) GetNodeCount() int       { return 0 }
func (r *testNodeRegistry) GetOnlineCount() int     { return 0 }
func (r *testNodeRegistry) WaitSave()               {}

func (w *testAuditLogWriter) Log(entry protocol.AuditLogEntry) {}

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
