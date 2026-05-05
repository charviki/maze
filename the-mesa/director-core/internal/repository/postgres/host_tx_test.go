package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
)

func TestHostTxManagerWithinTx_CommitsOnSuccess(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("new pgxmock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM host_specs").WithArgs("host-1").WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectCommit()

	txm := NewHostTxManager(mock)
	if err := txm.WithinTx(context.Background(), func(ctx context.Context) error {
		_, execErr := hostExecutorFromContext(ctx, mock).Exec(ctx, "DELETE FROM host_specs WHERE name = $1", "host-1")
		return execErr
	}); err != nil {
		t.Fatalf("WithinTx 返回错误: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("存在未满足的数据库预期: %v", err)
	}
}

func TestHostTxManagerWithinTx_RollsBackOnError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("new pgxmock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	txm := NewHostTxManager(mock)
	wantErr := errors.New("boom")
	err = txm.WithinTx(context.Background(), func(ctx context.Context) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("WithinTx error = %v, want %v", err, wantErr)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("存在未满足的数据库预期: %v", err)
	}
}
