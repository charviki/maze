package file

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
)

func newTestHostSpecRepository(t *testing.T) *HostSpecRepository {
	t.Helper()
	tmpDir := t.TempDir()
	return NewHostSpecRepository(filepath.Join(tmpDir, "host_specs.json"), logutil.NewNop())
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

func TestHostSpecRepository_CreateAndGet(t *testing.T) {
	s := newTestHostSpecRepository(t)

	spec := sampleSpec("host-1")
	ok := s.Create(spec)
	if !ok {
		t.Fatal("期望 Create 返回 true")
	}

	got := s.Get("host-1")
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

func TestHostSpecRepository_CreateDuplicate(t *testing.T) {
	s := newTestHostSpecRepository(t)

	s.Create(sampleSpec("host-1"))
	ok := s.Create(sampleSpec("host-1"))
	if ok {
		t.Error("期望重复 Create 返回 false")
	}
}

func TestHostSpecRepository_GetNotFound(t *testing.T) {
	s := newTestHostSpecRepository(t)
	if s.Get("nonexistent") != nil {
		t.Error("期望 Get 不存在的名称返回 nil")
	}
}

func TestHostSpecRepository_List(t *testing.T) {
	s := newTestHostSpecRepository(t)
	s.Create(sampleSpec("host-b"))
	s.Create(sampleSpec("host-a"))
	s.Create(sampleSpec("host-c"))

	list := s.List()
	if len(list) != 3 {
		t.Fatalf("期望 List 返回 3 个, 实际=%d", len(list))
	}
	if list[0].Name != "host-a" || list[1].Name != "host-b" || list[2].Name != "host-c" {
		names := make([]string, len(list))
		for i, spec := range list {
			names[i] = spec.Name
		}
		t.Errorf("期望按名称排序, 实际=%v", names)
	}
}

func TestHostSpecRepository_UpdateStatus(t *testing.T) {
	s := newTestHostSpecRepository(t)
	s.Create(sampleSpec("host-1"))

	ok := s.UpdateStatus("host-1", protocol.HostStatusDeploying, "")
	if !ok {
		t.Fatal("期望 UpdateStatus 返回 true")
	}
	got := s.Get("host-1")
	if got.Status != protocol.HostStatusDeploying {
		t.Errorf("期望 Status=deploying, 实际=%s", got.Status)
	}
}

func TestHostSpecRepository_UpdateStatusWithErrMsg(t *testing.T) {
	s := newTestHostSpecRepository(t)
	s.Create(sampleSpec("host-1"))

	s.UpdateStatus("host-1", protocol.HostStatusFailed, "docker build failed: exit code 1")
	got := s.Get("host-1")
	if got.Status != protocol.HostStatusFailed {
		t.Errorf("期望 Status=failed, 实际=%s", got.Status)
	}
	if got.ErrorMsg != "docker build failed: exit code 1" {
		t.Errorf("期望 ErrorMsg 包含错误信息, 实际=%s", got.ErrorMsg)
	}
}

func TestHostSpecRepository_UpdateStatusNotFound(t *testing.T) {
	s := newTestHostSpecRepository(t)
	ok := s.UpdateStatus("nonexistent", protocol.HostStatusDeploying, "")
	if ok {
		t.Error("期望更新不存在的 HostSpec 返回 false")
	}
}

func TestHostSpecRepository_Delete(t *testing.T) {
	s := newTestHostSpecRepository(t)
	s.Create(sampleSpec("host-1"))

	ok := s.Delete("host-1")
	if !ok {
		t.Error("期望 Delete 返回 true")
	}
	if s.Get("host-1") != nil {
		t.Error("期望删除后 Get 返回 nil")
	}
}

func TestHostSpecRepository_DeleteNotFound(t *testing.T) {
	s := newTestHostSpecRepository(t)
	ok := s.Delete("nonexistent")
	if ok {
		t.Error("期望删除不存在的 HostSpec 返回 false")
	}
}

func TestHostSpecRepository_IncrementRetry(t *testing.T) {
	s := newTestHostSpecRepository(t)
	s.Create(sampleSpec("host-1"))

	s.IncrementRetry("host-1")
	s.IncrementRetry("host-1")
	got := s.Get("host-1")
	if got.RetryCount != 2 {
		t.Errorf("期望 RetryCount=2, 实际=%d", got.RetryCount)
	}
}

func TestHostSpecRepository_IncrementRetryNotFound(t *testing.T) {
	s := newTestHostSpecRepository(t)
	ok := s.IncrementRetry("nonexistent")
	if ok {
		t.Error("期望递增不存在的 HostSpec 返回 false")
	}
}

func TestHostSpecRepository_PersistAndRecover(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "host_specs.json")

	s1 := NewHostSpecRepository(filePath, logutil.NewNop())
	spec := sampleSpec("host-persist")
	spec.DisplayName = "Persist Test"
	s1.Create(spec)
	s1.UpdateStatus("host-persist", protocol.HostStatusDeploying, "")

	s1.WaitSave()

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("期望持久化文件已创建")
	}

	s2 := NewHostSpecRepository(filePath, logutil.NewNop())
	got := s2.Get("host-persist")
	if got == nil {
		t.Fatal("期望恢复 host-persist，但为 nil")
	}
	if got.DisplayName != "Persist Test" {
		t.Errorf("期望 DisplayName=Persist Test, 实际=%s", got.DisplayName)
	}
	if got.Status != protocol.HostStatusDeploying {
		t.Errorf("期望 Status=deploying, 实际=%s", got.Status)
	}
	s2.WaitSave()
}

func TestHostSpecRepository_CreateClonesInput(t *testing.T) {
	s := newTestHostSpecRepository(t)
	spec := sampleSpec("host-clone")

	if !s.Create(spec) {
		t.Fatal("期望 Create 返回 true")
	}

	spec.Tools[0] = "tampered"
	spec.Status = protocol.HostStatusFailed

	got := s.Get("host-clone")
	if got.Tools[0] != "claude" {
		t.Fatalf("期望存储副本保留原始 Tools，实际=%v", got.Tools)
	}
	if got.Status != protocol.HostStatusPending {
		t.Fatalf("期望存储副本保留原始状态，实际=%s", got.Status)
	}
}

func TestHostSpecRepository_GetReturnsDetachedCopy(t *testing.T) {
	s := newTestHostSpecRepository(t)
	s.Create(sampleSpec("host-copy"))

	got := s.Get("host-copy")
	got.Tools[0] = "tampered"
	got.Status = protocol.HostStatusFailed

	again := s.Get("host-copy")
	if again.Tools[0] != "claude" {
		t.Fatalf("期望 Get 返回独立副本，实际=%v", again.Tools)
	}
	if again.Status != protocol.HostStatusPending {
		t.Fatalf("期望 Get 返回独立副本状态未被污染，实际=%s", again.Status)
	}
}

func TestHostSpecRepository_ListReturnsDetachedCopies(t *testing.T) {
	s := newTestHostSpecRepository(t)
	s.Create(sampleSpec("host-list"))

	list := s.List()
	list[0].Tools[0] = "tampered"
	list[0].Status = protocol.HostStatusFailed

	again := s.Get("host-list")
	if again.Tools[0] != "claude" {
		t.Fatalf("期望 List 返回独立副本，实际=%v", again.Tools)
	}
	if again.Status != protocol.HostStatusPending {
		t.Fatalf("期望 List 返回独立副本状态未被污染，实际=%s", again.Status)
	}
}
