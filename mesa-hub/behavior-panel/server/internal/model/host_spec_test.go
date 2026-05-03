package model

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
)

func newTestHostSpecManager(t *testing.T) *HostSpecManager {
	t.Helper()
	tmpDir := t.TempDir()
	return NewHostSpecManager(filepath.Join(tmpDir, "host_specs.json"), logutil.NewNop())
}

func sampleSpec(name string) *protocol.HostSpec {
	return &protocol.HostSpec{
		Name:      name,
		Tools:     []string{"claude", "go"},
		AuthToken: name + "-token",
		Status:    protocol.HostStatusPending,
		CreatedAt: time.Now().Truncate(time.Millisecond),
		UpdatedAt: time.Now().Truncate(time.Millisecond),
	}
}

func TestHostSpecManager_CreateAndGet(t *testing.T) {
	m := newTestHostSpecManager(t)

	spec := sampleSpec("host-1")
	ok := m.Create(spec)
	if !ok {
		t.Fatal("期望 Create 返回 true")
	}

	got := m.Get("host-1")
	if got == nil {
		t.Fatal("期望 Get 返回非 nil")
	}
	if got.Name != "host-1" {
		t.Errorf("期望 Name=host-1, 实际=%s", got.Name)
	}
	if got.Status != protocol.HostStatusPending {
		t.Errorf("期望 Status=pending, 实际=%s", got.Status)
	}
}

func TestHostSpecManager_CreateDuplicate(t *testing.T) {
	m := newTestHostSpecManager(t)

	m.Create(sampleSpec("host-1"))
	ok := m.Create(sampleSpec("host-1"))
	if ok {
		t.Error("期望重复 Create 返回 false")
	}
}

func TestHostSpecManager_GetNotFound(t *testing.T) {
	m := newTestHostSpecManager(t)
	if m.Get("nonexistent") != nil {
		t.Error("期望 Get 不存在的名称返回 nil")
	}
}

func TestHostSpecManager_List(t *testing.T) {
	m := newTestHostSpecManager(t)
	m.Create(sampleSpec("host-b"))
	m.Create(sampleSpec("host-a"))
	m.Create(sampleSpec("host-c"))

	list := m.List()
	if len(list) != 3 {
		t.Fatalf("期望 List 返回 3 个, 实际=%d", len(list))
	}
	// 验证排序
	if list[0].Name != "host-a" || list[1].Name != "host-b" || list[2].Name != "host-c" {
		names := make([]string, len(list))
		for i, s := range list {
			names[i] = s.Name
		}
		t.Errorf("期望按名称排序, 实际=%v", names)
	}
}

func TestHostSpecManager_UpdateStatus(t *testing.T) {
	m := newTestHostSpecManager(t)
	m.Create(sampleSpec("host-1"))

	ok := m.UpdateStatus("host-1", protocol.HostStatusDeploying, "")
	if !ok {
		t.Fatal("期望 UpdateStatus 返回 true")
	}
	got := m.Get("host-1")
	if got.Status != protocol.HostStatusDeploying {
		t.Errorf("期望 Status=deploying, 实际=%s", got.Status)
	}
}

func TestHostSpecManager_UpdateStatusWithErrMsg(t *testing.T) {
	m := newTestHostSpecManager(t)
	m.Create(sampleSpec("host-1"))

	m.UpdateStatus("host-1", protocol.HostStatusFailed, "docker build failed: exit code 1")
	got := m.Get("host-1")
	if got.Status != protocol.HostStatusFailed {
		t.Errorf("期望 Status=failed, 实际=%s", got.Status)
	}
	if got.ErrorMsg != "docker build failed: exit code 1" {
		t.Errorf("期望 ErrorMsg 包含错误信息, 实际=%s", got.ErrorMsg)
	}
}

func TestHostSpecManager_UpdateStatusNotFound(t *testing.T) {
	m := newTestHostSpecManager(t)
	ok := m.UpdateStatus("nonexistent", protocol.HostStatusDeploying, "")
	if ok {
		t.Error("期望更新不存在的 HostSpec 返回 false")
	}
}

func TestHostSpecManager_Delete(t *testing.T) {
	m := newTestHostSpecManager(t)
	m.Create(sampleSpec("host-1"))

	ok := m.Delete("host-1")
	if !ok {
		t.Error("期望 Delete 返回 true")
	}
	if m.Get("host-1") != nil {
		t.Error("期望删除后 Get 返回 nil")
	}
}

func TestHostSpecManager_DeleteNotFound(t *testing.T) {
	m := newTestHostSpecManager(t)
	ok := m.Delete("nonexistent")
	if ok {
		t.Error("期望删除不存在的 HostSpec 返回 false")
	}
}

func TestHostSpecManager_IncrementRetry(t *testing.T) {
	m := newTestHostSpecManager(t)
	m.Create(sampleSpec("host-1"))

	m.IncrementRetry("host-1")
	m.IncrementRetry("host-1")
	got := m.Get("host-1")
	if got.RetryCount != 2 {
		t.Errorf("期望 RetryCount=2, 实际=%d", got.RetryCount)
	}
}

func TestHostSpecManager_IncrementRetryNotFound(t *testing.T) {
	m := newTestHostSpecManager(t)
	ok := m.IncrementRetry("nonexistent")
	if ok {
		t.Error("期望递增不存在的 HostSpec 返回 false")
	}
}

func TestHostSpecManager_PersistAndRecover(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "host_specs.json")

	m1 := NewHostSpecManager(filePath, logutil.NewNop())
	spec := sampleSpec("host-persist")
	spec.DisplayName = "Persist Test"
	m1.Create(spec)
	m1.UpdateStatus("host-persist", protocol.HostStatusDeploying, "")

	m1.WaitSave()

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("期望持久化文件已创建")
	}

	m2 := NewHostSpecManager(filePath, logutil.NewNop())
	got := m2.Get("host-persist")
	if got == nil {
		t.Fatal("期望恢复 host-persist，但为 nil")
	}
	if got.DisplayName != "Persist Test" {
		t.Errorf("期望 DisplayName=Persist Test, 实际=%s", got.DisplayName)
	}
	if got.Status != protocol.HostStatusDeploying {
		t.Errorf("期望 Status=deploying, 实际=%s", got.Status)
	}
	m2.WaitSave()
}

func TestHostSpecManager_ListMerged_PendingNoNode(t *testing.T) {
	m := newTestHostSpecManager(t)
	reg := newTestRegistry(t)

	m.Create(sampleSpec("host-1"))
	m.UpdateStatus("host-1", protocol.HostStatusPending, "")

	merged := m.ListMerged(reg)
	if len(merged) != 1 {
		t.Fatalf("期望 1 个, 实际=%d", len(merged))
	}
	if merged[0].Status != protocol.HostStatusPending {
		t.Errorf("pending + 无 Node → status=pending, 实际=%s", merged[0].Status)
	}
}

func TestHostSpecManager_ListMerged_DeployingWithOnlineNode(t *testing.T) {
	m := newTestHostSpecManager(t)
	reg := newTestRegistry(t)

	m.Create(sampleSpec("host-1"))
	m.UpdateStatus("host-1", protocol.HostStatusDeploying, "")

	// 注册一个在线节点
	reg.Register(protocol.RegisterRequest{
		Name:    "host-1",
		Address: "http://host-1:8080",
		Status: protocol.AgentStatus{
			ActiveSessions: 3,
		},
	})

	merged := m.ListMerged(reg)
	if merged[0].Status != protocol.HostStatusOnline {
		t.Errorf("deploying + Node 在线 → status=online, 实际=%s", merged[0].Status)
	}
	if merged[0].Address != "http://host-1:8080" {
		t.Errorf("期望 Address=http://host-1:8080, 实际=%s", merged[0].Address)
	}
	if merged[0].SessionCount != 3 {
		t.Errorf("期望 SessionCount=3, 实际=%d", merged[0].SessionCount)
	}
}

func TestHostSpecManager_ListMerged_DeployingWithOfflineNode(t *testing.T) {
	m := newTestHostSpecManager(t)
	reg := newTestRegistry(t)

	m.Create(sampleSpec("host-1"))
	m.UpdateStatus("host-1", protocol.HostStatusDeploying, "")

	// 注册节点然后让它离线
	node := reg.Register(protocol.RegisterRequest{
		Name:    "host-1",
		Address: "http://host-1:8080",
	})
	reg.mu.Lock()
	node.LastHeartbeat = time.Now().Add(-31 * time.Second)
	reg.mu.Unlock()

	merged := m.ListMerged(reg)
	if merged[0].Status != protocol.HostStatusOffline {
		t.Errorf("deploying + Node 离线 → status=offline, 实际=%s", merged[0].Status)
	}
}

func TestHostSpecManager_ListMerged_FailedStatus(t *testing.T) {
	m := newTestHostSpecManager(t)
	reg := newTestRegistry(t)

	m.Create(sampleSpec("host-1"))
	m.UpdateStatus("host-1", protocol.HostStatusFailed, "build error")

	// 即使有 Node，failed 状态也应保持
	reg.Register(protocol.RegisterRequest{
		Name:    "host-1",
		Address: "http://host-1:8080",
	})

	merged := m.ListMerged(reg)
	if merged[0].Status != protocol.HostStatusFailed {
		t.Errorf("failed 状态应保持, 实际=%s", merged[0].Status)
	}
}

func TestHostSpecManager_ListMerged_DeployingNoNode(t *testing.T) {
	m := newTestHostSpecManager(t)
	reg := newTestRegistry(t)

	m.Create(sampleSpec("host-1"))
	m.UpdateStatus("host-1", protocol.HostStatusDeploying, "")

	merged := m.ListMerged(reg)
	// deploying 但无 Node → 保持 deploying
	if merged[0].Status != protocol.HostStatusDeploying {
		t.Errorf("deploying + 无 Node → status=deploying, 实际=%s", merged[0].Status)
	}
}
