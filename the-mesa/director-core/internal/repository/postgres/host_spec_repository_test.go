package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"

	"github.com/charviki/maze-cradle/protocol"
)

func TestHostSpecRepositoryCreateAndGet(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("new pgxmock pool: %v", err)
	}
	defer mock.Close()

	repo := NewHostSpecRepository(mock)
	spec := &protocol.HostSpec{
		Name:        "host-1",
		DisplayName: "Host One",
		Tools:       []string{"claude", "go"},
		Resources: protocol.ResourceLimits{
			CPULimit:    "2",
			MemoryLimit: "4Gi",
		},
		AuthToken: "token-1",
		Status:    protocol.HostStatusPending,
	}

	createdAt := time.Now().UTC().Round(time.Microsecond)
	updatedAt := createdAt.Add(time.Minute)
	mock.ExpectExec("INSERT INTO host_specs").
		WithArgs("host-1", "Host One", pgxmock.AnyArg(), pgxmock.AnyArg(), "token-1", protocol.HostStatusPending).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectQuery("SELECT .* FROM host_specs WHERE name = \\$1").
		WithArgs("host-1").
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "display_name", "tools", "resources", "auth_token", "status", "error_msg", "retry_count", "created_at", "updated_at",
		}).AddRow(
			int64(1),
			"host-1",
			"Host One",
			[]byte(`["claude","go"]`),
			[]byte(`{"cpu_limit":"2","memory_limit":"4Gi"}`),
			"token-1",
			protocol.HostStatusPending,
			"",
			int32(0),
			pgtype.Timestamptz{Time: createdAt, Valid: true},
			pgtype.Timestamptz{Time: updatedAt, Valid: true},
		))

	created, err := repo.Create(context.Background(), spec)
	if err != nil {
		t.Fatalf("Create 返回错误: %v", err)
	}
	if !created {
		t.Fatal("Create 应返回 true")
	}

	got, err := repo.Get(context.Background(), "host-1")
	if err != nil {
		t.Fatalf("Get 返回错误: %v", err)
	}
	if got == nil {
		t.Fatal("Get 应返回 HostSpec")
	}
	if got.DisplayName != "Host One" || got.AuthToken != "token-1" {
		t.Fatalf("Get 映射结果错误: %+v", got)
	}
	if len(got.Tools) != 2 || got.Tools[1] != "go" {
		t.Fatalf("tools = %+v, want [claude go]", got.Tools)
	}
	if got.Resources.MemoryLimit != "4Gi" {
		t.Fatalf("memory_limit = %q, want 4Gi", got.Resources.MemoryLimit)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("存在未满足的数据库预期: %v", err)
	}
}

func TestHostSpecRepositoryDeleteUsesTransactionExecutor(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("new pgxmock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM host_specs WHERE name = \\$1").
		WithArgs("host-1").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectCommit()

	repo := NewHostSpecRepository(mock)
	txm := NewHostTxManager(mock)
	err = txm.WithinTx(context.Background(), func(ctx context.Context) error {
		deleted, deleteErr := repo.Delete(ctx, "host-1")
		if deleteErr != nil {
			return deleteErr
		}
		if !deleted {
			t.Fatal("Delete 应返回 true")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WithinTx 返回错误: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("存在未满足的数据库预期: %v", err)
	}
}

func TestHostSpecRepositoryGetNotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("new pgxmock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery("SELECT .* FROM host_specs WHERE name = \\$1").
		WithArgs("missing").
		WillReturnError(pgx.ErrNoRows)

	repo := NewHostSpecRepository(mock)
	got, err := repo.Get(context.Background(), "missing")
	if err != nil {
		t.Fatalf("Get 返回错误: %v", err)
	}
	if got != nil {
		t.Fatalf("Get = %+v, want nil", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("存在未满足的数据库预期: %v", err)
	}
}
