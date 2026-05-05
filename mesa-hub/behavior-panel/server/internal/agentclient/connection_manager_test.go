package agentclient

import (
	"testing"
	"time"

	"github.com/charviki/maze-cradle/logutil"
)

func TestNewConnectionManager(t *testing.T) {
	cm := NewConnectionManager(logutil.NewNop(), "test-token", 30*time.Second)
	if cm == nil {
		t.Fatal("NewConnectionManager returned nil")
	}
	if cm.authToken != "test-token" {
		t.Errorf("authToken = %q, want test-token", cm.authToken)
	}
	if cm.idleTTL != 30*time.Second {
		t.Errorf("idleTTL = %v, want 30s", cm.idleTTL)
	}
	if cm.conns == nil {
		t.Error("conns map should be initialized")
	}
	if cm.stopCh == nil {
		t.Error("stopCh should be initialized")
	}
	cm.CloseAll()
}

func TestConnectionManager_CloseAll_NoConnections(t *testing.T) {
	cm := NewConnectionManager(logutil.NewNop(), "", time.Minute)
	cm.CloseAll()
}

func TestConnectionManager_CloseAll_TwiceSafe(t *testing.T) {
	cm := NewConnectionManager(logutil.NewNop(), "", time.Minute)
	cm.CloseAll()
	cm.CloseAll()
}

func TestConnectionManager_Remove_NonExistent(t *testing.T) {
	cm := NewConnectionManager(logutil.NewNop(), "", time.Minute)
	defer cm.CloseAll()
	cm.Remove("nonexistent:9090")
}

func TestConnectionManager_CleanupIdleRemovesExpiredConnection(t *testing.T) {
	cm := NewConnectionManager(logutil.NewNop(), "tok", time.Minute)
	defer cm.CloseAll()

	conn, err := cm.GetConn(t.Context(), "127.0.0.1:1")
	if err != nil {
		t.Fatalf("GetConn 返回错误: %v", err)
	}

	cm.mu.Lock()
	entry := cm.conns["127.0.0.1:1"]
	entry.lastUsed = time.Now().Add(-2 * time.Minute)
	cm.mu.Unlock()

	cm.cleanupIdle()

	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if _, ok := cm.conns["127.0.0.1:1"]; ok {
		t.Fatal("cleanupIdle 应移除过期连接")
	}
	if conn.GetState().String() == "" {
		t.Fatal("连接对象应已创建，便于验证清理逻辑")
	}
}
