package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"

	"github.com/charviki/maze-cradle/protocol"
)

func TestAuditLogRepositoryLogAndListPage(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("new pgxmock pool: %v", err)
	}
	defer mock.Close()

	repo := NewAuditLogRepository(mock)
	createdAt := time.Now().UTC().Round(time.Microsecond)
	mock.ExpectExec("INSERT INTO audit_entries").
		WithArgs(
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			"frontend",
			"host-1",
			"create_host",
			"tools=[claude]",
			"success",
			int32(201),
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM audit_entries").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT .* FROM audit_entries").
		WithArgs(int32(10), int32(0)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "audit_id", "timestamp", "operator", "target_node", "action", "payload_summary", "result", "status_code",
		}).AddRow(
			int64(1),
			"audit_1",
			pgtype.Timestamptz{Time: createdAt, Valid: true},
			"frontend",
			"host-1",
			"create_host",
			"tools=[claude]",
			"success",
			int32(201),
		))

	err = repo.Log(context.Background(), protocol.AuditLogEntry{
		Operator:       "frontend",
		Action:         "create_host",
		TargetNode:     "host-1",
		PayloadSummary: "tools=[claude]",
		Result:         "success",
		StatusCode:     201,
	})
	if err != nil {
		t.Fatalf("Log 返回错误: %v", err)
	}

	entries, total, err := repo.ListPage(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("ListPage 返回错误: %v", err)
	}
	if total != 1 || len(entries) != 1 {
		t.Fatalf("entries=%d total=%d, want 1/1", len(entries), total)
	}
	if entries[0].Action != "create_host" || entries[0].TargetNode != "host-1" {
		t.Fatalf("entry = %+v, want create_host host-1", entries[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("存在未满足的数据库预期: %v", err)
	}
}
