package file

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/internal/service"
)

// newTestRegistry 创建一个使用临时目录的 NodeRegistry，测试结束后自动清理
func newTestRegistry(t *testing.T) *NodeRegistry {
	t.Helper()
	tmpDir := t.TempDir()
	return NewNodeRegistry(filepath.Join(tmpDir, "nodes.json"), logutil.NewNop())
}

func TestNodeRegistry_Register(t *testing.T) {
	reg := newTestRegistry(t)

	node, _ := reg.Register(context.Background(), protocol.RegisterRequest{
		Name:         "agent-1",
		Address:      "http://192.168.1.10:9090",
		ExternalAddr: "http://10.0.0.1:9090",
	})

	if node.Name != "agent-1" {
		t.Errorf("期望 Name=agent-1, 实际=%s", node.Name)
	}
	if node.Address != "http://192.168.1.10:9090" {
		t.Errorf("期望 Address=http://192.168.1.10:9090, 实际=%s", node.Address)
	}
	if node.ExternalAddr != "http://10.0.0.1:9090" {
		t.Errorf("期望 ExternalAddr=http://10.0.0.1:9090, 实际=%s", node.ExternalAddr)
	}
	if node.Status != "online" {
		t.Errorf("期望 Status=online, 实际=%s", node.Status)
	}
	if node.RegisteredAt.IsZero() {
		t.Error("RegisteredAt 不应为零值")
	}
	if node.LastHeartbeat.IsZero() {
		t.Error("LastHeartbeat 不应为零值")
	}

	reg.WaitSave()

	nodes, _ := reg.List(context.Background())
	if len(nodes) != 1 {
		t.Fatalf("期望 List 返回 1 个节点, 实际=%d", len(nodes))
	}
	if nodes[0].Name != "agent-1" {
		t.Errorf("期望 List[0].Name=agent-1, 实际=%s", nodes[0].Name)
	}

	got, _ := reg.Get(context.Background(), "agent-1")
	if got == nil {
		t.Fatal("期望 Get 返回非 nil")
	}
	if got.Name != "agent-1" {
		t.Errorf("期望 Get 返回 Name=agent-1, 实际=%s", got.Name)
	}
}

func TestNodeRegistry_RegisterOverwrite(t *testing.T) {
	reg := newTestRegistry(t)

	reg.Register(context.Background(), protocol.RegisterRequest{
		Name:    "agent-1",
		Address: "http://192.168.1.10:9090",
	})
	reg.Register(context.Background(), protocol.RegisterRequest{
		Name:    "agent-1",
		Address: "http://192.168.1.20:9090",
	})

	reg.WaitSave()

	// 第二次注册应覆盖第一次的地址
	nodes, _ := reg.List(context.Background())
	if len(nodes) != 1 {
		t.Fatalf("期望 1 个节点, 实际=%d", len(nodes))
	}
	if nodes[0].Address != "http://192.168.1.20:9090" {
		t.Errorf("期望 Address=http://192.168.1.20:9090, 实际=%s", nodes[0].Address)
	}
}

// TestNodeRegistry_RegisterWithCapabilities 验证注册正确存储 Capabilities、AgentStatus、Metadata
func TestNodeRegistry_RegisterWithCapabilities(t *testing.T) {
	reg := newTestRegistry(t)

	startedAt := time.Now().Add(-5 * time.Minute)
	node, _ := reg.Register(context.Background(), protocol.RegisterRequest{
		Name:         "agent-full",
		Address:      "http://192.168.1.10:9090",
		ExternalAddr: "http://10.0.0.1:9090",
		Capabilities: protocol.AgentCapabilities{
			SupportedTemplates: []string{"claude", "bash"},
			MaxSessions:        10,
			Tools:              []string{"tmux", "filesystem"},
		},
		Status: protocol.AgentStatus{
			ActiveSessions: 3,
			CPUUsage:       45.5,
			MemoryUsageMB:  1024.0,
			WorkspaceRoot:  "/home/agent/workspace",
		},
		Metadata: protocol.AgentMetadata{
			Version:   "v1.2.3",
			Hostname:  "agent-host-01",
			StartedAt: startedAt,
		},
	})

	if node.Name != "agent-full" {
		t.Errorf("期望 Name=agent-full, 实际=%s", node.Name)
	}
	if node.Address != "http://192.168.1.10:9090" {
		t.Errorf("期望 Address=http://192.168.1.10:9090, 实际=%s", node.Address)
	}
	if node.Status != "online" {
		t.Errorf("期望 Status=online, 实际=%s", node.Status)
	}

	caps := node.Capabilities
	if len(caps.SupportedTemplates) != 2 || caps.SupportedTemplates[0] != "claude" {
		t.Errorf("期望 SupportedTemplates=[claude bash], 实际=%v", caps.SupportedTemplates)
	}
	if caps.MaxSessions != 10 {
		t.Errorf("期望 MaxSessions=10, 实际=%d", caps.MaxSessions)
	}
	if len(caps.Tools) != 2 || caps.Tools[0] != "tmux" {
		t.Errorf("期望 Tools=[tmux filesystem], 实际=%v", caps.Tools)
	}

	st := node.AgentStatus
	if st.ActiveSessions != 3 {
		t.Errorf("期望 ActiveSessions=3, 实际=%d", st.ActiveSessions)
	}
	if st.CPUUsage != 45.5 {
		t.Errorf("期望 CPUUsage=45.5, 实际=%.1f", st.CPUUsage)
	}
	if st.MemoryUsageMB != 1024.0 {
		t.Errorf("期望 MemoryUsageMB=1024.0, 实际=%.0f", st.MemoryUsageMB)
	}

	meta := node.Metadata
	if meta.Version != "v1.2.3" {
		t.Errorf("期望 Version=v1.2.3, 实际=%s", meta.Version)
	}
	if meta.Hostname != "agent-host-01" {
		t.Errorf("期望 Hostname=agent-host-01, 实际=%s", meta.Hostname)
	}
	if meta.StartedAt.IsZero() {
		t.Error("StartedAt 不应为零值")
	}

	reg.WaitSave()
}

// TestNodeRegistry_RegisterOverwriteCapabilities 验证同名注册第二次覆盖第一次（含 capabilities）
func TestNodeRegistry_RegisterOverwriteCapabilities(t *testing.T) {
	reg := newTestRegistry(t)

	reg.Register(context.Background(), protocol.RegisterRequest{
		Name:    "agent-overwrite",
		Address: "http://192.168.1.10:9090",
		Capabilities: protocol.AgentCapabilities{
			MaxSessions:        5,
			SupportedTemplates: []string{"claude"},
		},
		Metadata: protocol.AgentMetadata{Version: "v1.0.0", Hostname: "host-old"},
	})

	node, _ := reg.Register(context.Background(), protocol.RegisterRequest{
		Name:    "agent-overwrite",
		Address: "http://192.168.1.20:9090",
		Capabilities: protocol.AgentCapabilities{
			MaxSessions:        20,
			SupportedTemplates: []string{"bash", "python"},
		},
		Metadata: protocol.AgentMetadata{Version: "v2.0.0", Hostname: "host-new"},
	})

	reg.WaitSave()

	if node.Address != "http://192.168.1.20:9090" {
		t.Errorf("期望 Address=http://192.168.1.20:9090, 实际=%s", node.Address)
	}
	if node.Capabilities.MaxSessions != 20 {
		t.Errorf("期望 MaxSessions=20, 实际=%d", node.Capabilities.MaxSessions)
	}
	if node.Metadata.Version != "v2.0.0" {
		t.Errorf("期望 Version=v2.0.0, 实际=%s", node.Metadata.Version)
	}

	nodes, _ := reg.List(context.Background())
	if len(nodes) != 1 {
		t.Fatalf("期望 1 个节点, 实际=%d", len(nodes))
	}
}

func TestNodeRegistry_Heartbeat(t *testing.T) {
	reg := newTestRegistry(t)

	reg.Register(context.Background(), protocol.RegisterRequest{
		Name:    "agent-1",
		Address: "http://192.168.1.10:9090",
	})

	node, _ := reg.Heartbeat(context.Background(), protocol.HeartbeatRequest{
		Name: "agent-1",
		Status: protocol.AgentStatus{
			SessionDetails: []protocol.SessionDetail{
				{ID: "s1"}, {ID: "s2"}, {ID: "s3"}, {ID: "s4"}, {ID: "s5"},
			},
		},
	})
	if node == nil {
		t.Fatal("期望 Heartbeat 返回非 nil")
	}
	// Heartbeat 从 SessionDetails 同步 ActiveSessions
	if node.AgentStatus.ActiveSessions != 5 {
		t.Errorf("期望 ActiveSessions=5, 实际=%d", node.AgentStatus.ActiveSessions)
	}
	if node.Status != "online" {
		t.Errorf("期望 Status=online, 实际=%s", node.Status)
	}

	reg.WaitSave()
}

func TestNodeRegistry_HeartbeatNotFound(t *testing.T) {
	reg := newTestRegistry(t)

	node, _ := reg.Heartbeat(context.Background(), protocol.HeartbeatRequest{
		Name:   "non-existent",
		Status: protocol.AgentStatus{},
	})
	if node != nil {
		t.Error("期望对不存在节点 Heartbeat 返回 nil")
	}
}

// TestNodeRegistry_HeartbeatWithStatus 验证心跳更新完整状态快照（含 SessionDetails）
func TestNodeRegistry_HeartbeatWithStatus(t *testing.T) {
	reg := newTestRegistry(t)

	reg.Register(context.Background(), protocol.RegisterRequest{
		Name:    "agent-hb",
		Address: "http://192.168.1.10:9090",
		Capabilities: protocol.AgentCapabilities{
			MaxSessions: 10,
		},
		Status: protocol.AgentStatus{
			ActiveSessions: 1,
			CPUUsage:       10.0,
			MemoryUsageMB:  512.0,
		},
	})

	node, _ := reg.Heartbeat(context.Background(), protocol.HeartbeatRequest{
		Name: "agent-hb",
		Status: protocol.AgentStatus{
			CPUUsage:      78.3,
			MemoryUsageMB: 2048.0,
			WorkspaceRoot: "/home/agent/ws",
			SessionDetails: []protocol.SessionDetail{
				{ID: "sess-1", Template: "claude", WorkingDir: "/home/agent/proj1", UptimeSeconds: 120},
				{ID: "sess-2", Template: "bash", WorkingDir: "/home/agent/proj2", UptimeSeconds: 60},
			},
		},
	})

	if node == nil {
		t.Fatal("期望 Heartbeat 返回非 nil")
	}
	// Heartbeat 从 SessionDetails 同步 ActiveSessions
	if node.AgentStatus.ActiveSessions != 2 {
		t.Errorf("期望 ActiveSessions=2, 实际=%d", node.AgentStatus.ActiveSessions)
	}
	if node.AgentStatus.CPUUsage != 78.3 {
		t.Errorf("期望 CPUUsage=78.3, 实际=%.1f", node.AgentStatus.CPUUsage)
	}
	if node.AgentStatus.MemoryUsageMB != 2048.0 {
		t.Errorf("期望 MemoryUsageMB=2048.0, 实际=%.0f", node.AgentStatus.MemoryUsageMB)
	}
	if len(node.AgentStatus.SessionDetails) != 2 {
		t.Fatalf("期望 2 个 SessionDetail, 实际=%d", len(node.AgentStatus.SessionDetails))
	}
	if node.AgentStatus.SessionDetails[0].ID != "sess-1" {
		t.Errorf("SessionDetail[0].ID 期望 sess-1, 实际=%s", node.AgentStatus.SessionDetails[0].ID)
	}
	if node.Status != "online" {
		t.Errorf("期望 Status=online, 实际=%s", node.Status)
	}

	reg.WaitSave()
}

func TestNodeRegistry_Delete(t *testing.T) {
	reg := newTestRegistry(t)

	reg.Register(context.Background(), protocol.RegisterRequest{
		Name:    "agent-1",
		Address: "http://192.168.1.10:9090",
	})

	ok, _ := reg.Delete(context.Background(), "agent-1")
	if !ok {
		t.Error("期望 Delete 已存在节点返回 true")
	}

	nodes, _ := reg.List(context.Background())
	if len(nodes) != 0 {
		t.Errorf("期望删除后 List 为空, 实际=%d 个节点", len(nodes))
	}
	if got, _ := reg.Get(context.Background(), "agent-1"); got != nil {
		t.Error("期望删除后 Get 返回 nil")
	}

	reg.WaitSave()
}

func TestNodeRegistry_DeleteNotFound(t *testing.T) {
	reg := newTestRegistry(t)

	ok, _ := reg.Delete(context.Background(), "non-existent")
	if ok {
		t.Error("期望删除不存在的节点返回 false")
	}
}

func TestNodeRegistry_List_OfflineStatus(t *testing.T) {
	reg := newTestRegistry(t)

	reg.Register(context.Background(), protocol.RegisterRequest{
		Name:    "agent-1",
		Address: "http://192.168.1.10:9090",
	})

	reg.WaitSave()

	// 手动将 LastHeartbeat 设为 31 秒前，超过 30 秒离线阈值
	reg.mu.Lock()
	node := reg.nodes["agent-1"]
	node.LastHeartbeat = time.Now().Add(-31 * time.Second)
	reg.mu.Unlock()

	nodes, _ := reg.List(context.Background())
	if len(nodes) != 1 {
		t.Fatalf("期望 1 个节点, 实际=%d", len(nodes))
	}
	if nodes[0].Status != "offline" {
		t.Errorf("期望 Status=offline, 实际=%s", nodes[0].Status)
	}
}

func TestNodeRegistry_Get_OfflineStatus(t *testing.T) {
	reg := newTestRegistry(t)

	reg.Register(context.Background(), protocol.RegisterRequest{
		Name:    "agent-1",
		Address: "http://192.168.1.10:9090",
	})

	reg.WaitSave()

	// 手动将 LastHeartbeat 设为 31 秒前，超过 30 秒离线阈值
	reg.mu.Lock()
	node := reg.nodes["agent-1"]
	node.LastHeartbeat = time.Now().Add(-31 * time.Second)
	reg.mu.Unlock()

	got, _ := reg.Get(context.Background(), "agent-1")
	if got == nil {
		t.Fatal("期望 Get 返回非 nil")
	}
	if got.Status != "offline" {
		t.Errorf("期望 Status=offline, 实际=%s", got.Status)
	}
}

// TestNodeRegistry_ConcurrentAccess 验证并发读写不会触发 data race。
// 需使用 go test -race 运行以检测竞争条件
func TestNodeRegistry_ConcurrentAccess(t *testing.T) {
	reg := newTestRegistry(t)
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(4)

		go func() {
			defer wg.Done()
			reg.Register(context.Background(), protocol.RegisterRequest{
				Name:    "agent-concurrent",
				Address: "http://10.0.0.1:9090",
			})
		}()

		go func() {
			defer wg.Done()
			reg.List(context.Background())
		}()

		go func() {
			defer wg.Done()
			reg.Get(context.Background(), "agent-concurrent")
		}()

		go func(idx int) {
			defer wg.Done()
			reg.Heartbeat(context.Background(), protocol.HeartbeatRequest{
				Name: "agent-concurrent",
				Status: protocol.AgentStatus{
					SessionDetails: make([]protocol.SessionDetail, idx),
				},
			})
		}(i)
	}

	wg.Wait()

	reg.WaitSave()

	node, _ := reg.Get(context.Background(), "agent-concurrent")
	if node == nil {
		t.Fatal("期望并发操作后节点仍存在")
	}
	if node.Name != "agent-concurrent" {
		t.Errorf("期望 Name=agent-concurrent, 实际=%s", node.Name)
	}
}

// TestNodeRegistry_PersistAndRecover 验证节点数据持久化后可被新 Registry 实例恢复
func TestNodeRegistry_PersistAndRecover(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nodes.json")

	// 第一个 Registry：注册节点
	reg1 := NewNodeRegistry(filePath, logutil.NewNop())
	reg1.Register(context.Background(), protocol.RegisterRequest{
		Name:    "agent-persist",
		Address: "http://192.168.1.10:9090",
	})
	reg1.Register(context.Background(), protocol.RegisterRequest{
		Name:    "agent-persist-2",
		Address: "http://192.168.1.20:9090",
	})

	// 同步刷盘确保数据持久化
	reg1.WaitSave()

	// 验证文件已创建
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("期望持久化文件已创建，但文件不存在")
	}

	// 第二个 Registry：从同一文件加载，模拟 Manager 重启
	reg2 := NewNodeRegistry(filePath, logutil.NewNop())
	nodes, _ := reg2.List(context.Background())
	if len(nodes) != 2 {
		t.Fatalf("期望恢复 2 个节点, 实际=%d", len(nodes))
	}

	got, _ := reg2.Get(context.Background(), "agent-persist")
	if got == nil {
		t.Fatal("期望恢复 agent-persist 节点，但为 nil")
	}
	if got.Address != "http://192.168.1.10:9090" {
		t.Errorf("期望 Address=http://192.168.1.10:9090, 实际=%s", got.Address)
	}

	got2, _ := reg2.Get(context.Background(), "agent-persist-2")
	if got2 == nil {
		t.Fatal("期望恢复 agent-persist-2 节点，但为 nil")
	}
	if got2.Address != "http://192.168.1.20:9090" {
		t.Errorf("期望 Address=http://192.168.1.20:9090, 实际=%s", got2.Address)
	}
}

// TestNodeRegistry_CorruptedFile 验证文件损坏时不 panic，以空注册表启动
func TestNodeRegistry_CorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nodes.json")

	// 写入损坏的 JSON
	if err := os.WriteFile(filePath, []byte("{invalid json!!!"), 0644); err != nil {
		t.Fatalf("写入损坏文件失败: %v", err)
	}

	// 加载损坏文件不应 panic，以空注册表开始
	reg := NewNodeRegistry(filePath, logutil.NewNop())
	nodes, _ := reg.List(context.Background())
	if len(nodes) != 0 {
		t.Errorf("期望损坏文件后以空注册表开始, 实际=%d 个节点", len(nodes))
	}

	// 应该可以正常注册新节点
	reg.Register(context.Background(), protocol.RegisterRequest{
		Name:    "agent-after-corrupt",
		Address: "http://10.0.0.1:9090",
	})

	reg.WaitSave()

	got, _ := reg.Get(context.Background(), "agent-after-corrupt")
	if got == nil {
		t.Fatal("期望损坏文件恢复后可正常注册新节点")
	}
}

// TestNodeRegistry_DeletePersistence 验证删除操作也会持久化
func TestNodeRegistry_DeletePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nodes.json")

	reg1 := NewNodeRegistry(filePath, logutil.NewNop())
	reg1.Register(context.Background(), protocol.RegisterRequest{Name: "agent-del", Address: "http://10.0.0.1:9090"})
	reg1.Register(context.Background(), protocol.RegisterRequest{Name: "agent-keep", Address: "http://10.0.0.2:9090"})

	reg1.Delete(context.Background(), "agent-del")

	// 同步刷盘确保删除操作持久化
	reg1.WaitSave()

	// 从文件重新加载，验证 agent-del 已被删除
	reg2 := NewNodeRegistry(filePath, logutil.NewNop())
	if got, _ := reg2.Get(context.Background(), "agent-del"); got != nil {
		t.Error("期望 agent-del 已被持久化删除")
	}
	if got, _ := reg2.Get(context.Background(), "agent-keep"); got == nil {
		t.Error("期望 agent-keep 仍存在")
	}
	reg2.WaitSave()
}

// TestNodeRegistry_FileContents 验证持久化文件内容格式正确
func TestNodeRegistry_FileContents(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nodes.json")

	reg := NewNodeRegistry(filePath, logutil.NewNop())
	reg.Register(context.Background(), protocol.RegisterRequest{
		Name:    "agent-fmt",
		Address: "http://10.0.0.1:9090",
	})

	reg.WaitSave()

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取持久化文件失败: %v", err)
	}

	var nodes map[string]*service.Node
	if err := json.Unmarshal(data, &nodes); err != nil {
		t.Fatalf("解析持久化文件失败: %v, 内容: %s", err, string(data))
	}
	if len(nodes) != 1 {
		t.Fatalf("期望 1 个节点, 实际=%d", len(nodes))
	}
	if nodes["agent-fmt"] == nil {
		t.Fatal("期望 agent-fmt 存在")
	}
	if nodes["agent-fmt"].Address != "http://10.0.0.1:9090" {
		t.Errorf("期望 Address=http://10.0.0.1:9090, 实际=%s", nodes["agent-fmt"].Address)
	}
}

// TestNodeRegistry_RegisterPersistWithCapabilities 验证含 capabilities 的注册数据持久化后可恢复
func TestNodeRegistry_RegisterPersistWithCapabilities(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nodes.json")
	startedAt := time.Now().Add(-10 * time.Minute)

	reg1 := NewNodeRegistry(filePath, logutil.NewNop())
	reg1.Register(context.Background(), protocol.RegisterRequest{
		Name:         "agent-persist-enh",
		Address:      "http://192.168.1.10:9090",
		ExternalAddr: "http://10.0.0.1:9090",
		Capabilities: protocol.AgentCapabilities{
			SupportedTemplates: []string{"claude"},
			MaxSessions:        8,
			Tools:              []string{"tmux"},
		},
		Status: protocol.AgentStatus{
			ActiveSessions: 3,
			CPUUsage:       55.0,
			MemoryUsageMB:  1024.0,
		},
		Metadata: protocol.AgentMetadata{
			Version:   "v1.0.0",
			Hostname:  "persist-host",
			StartedAt: startedAt,
		},
	})

	reg1.WaitSave()

	reg2 := NewNodeRegistry(filePath, logutil.NewNop())
	got, _ := reg2.Get(context.Background(), "agent-persist-enh")
	if got == nil {
		t.Fatal("期望恢复 agent-persist-enh 节点，但为 nil")
	}
	if got.Address != "http://192.168.1.10:9090" {
		t.Errorf("期望 Address=http://192.168.1.10:9090, 实际=%s", got.Address)
	}
	if got.Capabilities.MaxSessions != 8 {
		t.Errorf("期望 MaxSessions=8, 实际=%d", got.Capabilities.MaxSessions)
	}
	if got.Metadata.Version != "v1.0.0" {
		t.Errorf("期望 Version=v1.0.0, 实际=%s", got.Metadata.Version)
	}
	if got.Metadata.Hostname != "persist-host" {
		t.Errorf("期望 Hostname=persist-host, 实际=%s", got.Metadata.Hostname)
	}
	if got.Metadata.StartedAt.IsZero() {
		t.Error("期望 StartedAt 已恢复，但为零值")
	}
}

// TestNodeRegistry_FormatNodeSummary 验证 FormatNodeSummary 输出包含关键字段且格式正确
func TestNodeRegistry_FormatNodeSummary(t *testing.T) {
	reg := newTestRegistry(t)

	node, _ := reg.Register(context.Background(), protocol.RegisterRequest{
		Name:    "agent-summary",
		Address: "http://192.168.1.10:9090",
		Status: protocol.AgentStatus{
			ActiveSessions: 5,
			CPUUsage:       67.8,
			MemoryUsageMB:  2048,
		},
	})

	reg.WaitSave()

	summary := service.FormatNodeSummary(node)
	expected := "agent-summary (http://192.168.1.10:9090) sessions=5 cpu=67.8% mem=2048MB status=online"
	if summary != expected {
		t.Errorf("期望摘要:\n  %s\n实际:\n  %s", expected, summary)
	}
}

func TestNodeRegistry_RegisterClonesInput(t *testing.T) {
	reg := newTestRegistry(t)
	req := protocol.RegisterRequest{
		Name:    "agent-clone",
		Address: "http://127.0.0.1:9090",
		Capabilities: protocol.AgentCapabilities{
			SupportedTemplates: []string{"claude"},
			Tools:              []string{"tmux"},
		},
		Status: protocol.AgentStatus{
			SessionDetails: []protocol.SessionDetail{{ID: "sess-1"}},
			LocalConfig: &protocol.LocalAgentConfig{
				WorkingDir: "/workspace",
				Env:        map[string]string{"FOO": "bar"},
			},
		},
	}

	_, _ = reg.Register(context.Background(), req)
	req.Capabilities.SupportedTemplates[0] = "tampered"
	req.Status.SessionDetails[0].ID = "tampered"
	req.Status.LocalConfig.Env["FOO"] = "mutated"

	got, _ := reg.Get(context.Background(), "agent-clone")
	if got.Capabilities.SupportedTemplates[0] != "claude" {
		t.Fatalf("期望注册时克隆输入切片，实际=%v", got.Capabilities.SupportedTemplates)
	}
	if got.AgentStatus.SessionDetails[0].ID != "sess-1" {
		t.Fatalf("期望注册时克隆输入会话快照，实际=%v", got.AgentStatus.SessionDetails)
	}
	if got.AgentStatus.LocalConfig.Env["FOO"] != "bar" {
		t.Fatalf("期望注册时克隆输入配置，实际=%v", got.AgentStatus.LocalConfig.Env)
	}
}

func TestNodeRegistry_GetAndListReturnDetachedCopies(t *testing.T) {
	reg := newTestRegistry(t)
	reg.Register(context.Background(), protocol.RegisterRequest{
		Name:    "agent-copy",
		Address: "http://127.0.0.1:9090",
		Capabilities: protocol.AgentCapabilities{
			SupportedTemplates: []string{"claude"},
		},
		Status: protocol.AgentStatus{
			SessionDetails: []protocol.SessionDetail{{ID: "sess-1"}},
			LocalConfig: &protocol.LocalAgentConfig{
				Env: map[string]string{"FOO": "bar"},
			},
		},
	})

	got, _ := reg.Get(context.Background(), "agent-copy")
	got.Capabilities.SupportedTemplates[0] = "tampered"
	got.AgentStatus.SessionDetails[0].ID = "tampered"
	got.AgentStatus.LocalConfig.Env["FOO"] = "mutated"

	list, _ := reg.List(context.Background())
	list[0].Status = service.NodeStatusOffline

	again, _ := reg.Get(context.Background(), "agent-copy")
	if again.Capabilities.SupportedTemplates[0] != "claude" {
		t.Fatalf("期望 Get 返回独立副本，实际=%v", again.Capabilities.SupportedTemplates)
	}
	if again.AgentStatus.SessionDetails[0].ID != "sess-1" {
		t.Fatalf("期望 Get 返回独立副本的会话未被污染，实际=%v", again.AgentStatus.SessionDetails)
	}
	if again.AgentStatus.LocalConfig.Env["FOO"] != "bar" {
		t.Fatalf("期望 Get 返回独立副本的配置未被污染，实际=%v", again.AgentStatus.LocalConfig.Env)
	}
	if again.Status != service.NodeStatusOnline {
		t.Fatalf("期望 List 返回副本不污染内部状态，实际=%s", again.Status)
	}
}
