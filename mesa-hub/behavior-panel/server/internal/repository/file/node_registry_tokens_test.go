package file

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
)

func TestNodeRegistry_HostTokenPersistenceLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	nodesFile := filepath.Join(tmpDir, "nodes.json")

	reg1 := NewNodeRegistry(nodesFile, logutil.NewNop())
	reg1.StoreHostToken(context.Background(), "host-1", "token-1")

	exists, matched, _ := reg1.ValidateHostToken(context.Background(), "host-1", "token-1")
	if !exists || !matched {
		t.Fatalf("ValidateHostToken(token-1) = (%v, %v), want (true, true)", exists, matched)
	}
	exists, matched, _ = reg1.ValidateHostToken(context.Background(), "host-1", "wrong-token")
	if !exists || matched {
		t.Fatalf("ValidateHostToken(wrong-token) = (%v, %v), want (true, false)", exists, matched)
	}

	reg1.WaitSave()

	// 重新创建注册表，验证预存 token 能从磁盘恢复。
	reg2 := NewNodeRegistry(nodesFile, logutil.NewNop())
	exists, matched, _ = reg2.ValidateHostToken(context.Background(), "host-1", "token-1")
	if !exists || !matched {
		t.Fatalf("恢复后 ValidateHostToken = (%v, %v), want (true, true)", exists, matched)
	}

	reg2.RemoveHostToken(context.Background(), "host-1")
	reg2.WaitSave()

	reg3 := NewNodeRegistry(nodesFile, logutil.NewNop())
	exists, matched, _ = reg3.ValidateHostToken(context.Background(), "host-1", "token-1")
	if exists || matched {
		t.Fatalf("删除后 ValidateHostToken = (%v, %v), want (false, false)", exists, matched)
	}
	reg3.WaitSave()
}

func TestNodeRegistry_GetOnlineCount_UsesHeartbeatWindow(t *testing.T) {
	reg := newTestRegistry(t)

	reg.Register(context.Background(), protocol.RegisterRequest{Name: "node-online", Address: "http://node-online"})
	reg.Register(context.Background(), protocol.RegisterRequest{Name: "node-offline", Address: "http://node-offline"})

	reg.mu.Lock()
	online := reg.nodes["node-online"]
	offline := reg.nodes["node-offline"]
	online.LastHeartbeat = time.Now().Add(-5 * time.Second)
	offline.LastHeartbeat = time.Now().Add(-35 * time.Second)
	reg.mu.Unlock()

	if got, _ := reg.GetOnlineCount(context.Background()); got != 1 {
		t.Fatalf("GetOnlineCount = %d, want 1", got)
	}
}
