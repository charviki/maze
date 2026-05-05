package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	hostgen "github.com/charviki/maze/the-mesa/director-core/internal/repository/postgres/sqlc/host"
)

type hostTxDB interface {
	hostgen.DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

type hostTxContextKey struct{}

func hostTxFromContext(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(hostTxContextKey{}).(pgx.Tx)
	return tx
}

func hostExecutorFromContext(ctx context.Context, db hostgen.DBTX) hostgen.DBTX {
	if tx := hostTxFromContext(ctx); tx != nil {
		return tx
	}
	return db
}

// HostTxManager 管理 Host 领域的数据库事务，与 Auth 领域的 authTxContextKey 隔离。
type HostTxManager struct {
	db hostTxDB
}

// NewHostTxManager 创建 Host 事务管理器。
func NewHostTxManager(db hostTxDB) *HostTxManager {
	return &HostTxManager{db: db}
}

// WithinTx 在 context 中透传 Host 领域事务，已处于事务中时直接执行 fn。
func (m *HostTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if existing := hostTxFromContext(ctx); existing != nil {
		return fn(ctx)
	}

	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	txCtx := context.WithValue(ctx, hostTxContextKey{}, tx)
	if err := fn(txCtx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Queries 返回当前 context 下的 sqlc Queries 实例（自动选择 tx 或 pool）。
func (m *HostTxManager) Queries(ctx context.Context) *hostgen.Queries {
	return hostgen.New(hostExecutorFromContext(ctx, m.db))
}
