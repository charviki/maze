package service

import (
	"context"
	"testing"

	"github.com/charviki/maze-cradle/protocol"
)

type mockAuditLogRepo struct {
	entries []protocol.AuditLogEntry
}

func (m *mockAuditLogRepo) Log(_ context.Context, entry protocol.AuditLogEntry) error {
	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockAuditLogRepo) List(_ context.Context) ([]protocol.AuditLogEntry, error) {
	return m.entries, nil
}

func (m *mockAuditLogRepo) ListPage(_ context.Context, page, pageSize int) ([]protocol.AuditLogEntry, int, error) {
	total := len(m.entries)
	start := (page - 1) * pageSize
	if start >= total {
		return nil, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return m.entries[start:end], total, nil
}

func newMockAuditLogRepo() *mockAuditLogRepo {
	return &mockAuditLogRepo{}
}

func TestNewAuditService(t *testing.T) {
	svc := NewAuditService(newMockAuditLogRepo())
	if svc == nil {
		t.Fatal("NewAuditService returned nil")
	}
}

func TestAuditService_GetAuditLogs_All(t *testing.T) {
	repo := newMockAuditLogRepo()
	for i := 0; i < 5; i++ {
		_ = repo.Log(context.Background(), protocol.AuditLogEntry{Action: "test"})
	}
	svc := NewAuditService(repo)

	result, err := svc.GetAuditLogs(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("GetAuditLogs returned error: %v", err)
	}
	if result.Total != 5 {
		t.Errorf("Total = %d, want 5", result.Total)
	}
	if len(result.Logs) != 5 {
		t.Errorf("Logs length = %d, want 5", len(result.Logs))
	}
}

func TestAuditService_GetAuditLogs_NegativePage(t *testing.T) {
	repo := newMockAuditLogRepo()
	_ = repo.Log(context.Background(), protocol.AuditLogEntry{Action: "test"})
	svc := NewAuditService(repo)

	result, err := svc.GetAuditLogs(context.Background(), -1, 10)
	if err != nil {
		t.Fatalf("GetAuditLogs returned error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}
}

func TestAuditService_GetAuditLogs_Pagination(t *testing.T) {
	repo := newMockAuditLogRepo()
	for i := 0; i < 25; i++ {
		_ = repo.Log(context.Background(), protocol.AuditLogEntry{Action: "test"})
	}
	svc := NewAuditService(repo)

	result, err := svc.GetAuditLogs(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("GetAuditLogs returned error: %v", err)
	}
	if result.Total != 25 {
		t.Errorf("Total = %d, want 25", result.Total)
	}
	if len(result.Logs) != 10 {
		t.Errorf("Logs length = %d, want 10", len(result.Logs))
	}
	if result.Page != 1 {
		t.Errorf("Page = %d, want 1", result.Page)
	}
}

func TestAuditService_GetAuditLogs_PageBeyondRange(t *testing.T) {
	repo := newMockAuditLogRepo()
	for i := 0; i < 5; i++ {
		_ = repo.Log(context.Background(), protocol.AuditLogEntry{Action: "test"})
	}
	svc := NewAuditService(repo)

	result, err := svc.GetAuditLogs(context.Background(), 10, 10)
	if err != nil {
		t.Fatalf("GetAuditLogs returned error: %v", err)
	}
	if result.Total != 5 {
		t.Errorf("Total = %d, want 5", result.Total)
	}
	if len(result.Logs) != 0 {
		t.Errorf("Logs length = %d, want 0", len(result.Logs))
	}
}

func TestAuditService_AuditLogWriter_Log(t *testing.T) {
	repo := newMockAuditLogRepo()
	_ = repo.Log(context.Background(), protocol.AuditLogEntry{Action: "create", TargetNode: "node-1"})
	_ = repo.Log(context.Background(), protocol.AuditLogEntry{Action: "delete", TargetNode: "node-2"})

	entries, _ := repo.List(context.Background())
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}
