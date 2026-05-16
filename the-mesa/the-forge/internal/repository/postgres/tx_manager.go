package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	gen "github.com/charviki/maze/the-mesa/the-forge/internal/repository/postgres/sqlc"
)

type docTxDB interface {
	gen.DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

type docTxContextKey struct{}

func docTxFromContext(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(docTxContextKey{}).(pgx.Tx)
	return tx
}

func docExecutorFromContext(ctx context.Context, db gen.DBTX) gen.DBTX {
	if tx := docTxFromContext(ctx); tx != nil {
		return tx
	}
	return db
}

// TxManager 管理 The Forge 数据库事务。
type TxManager struct {
	db docTxDB
}

// NewTxManager 创建事务管理器。
func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{db: pool}
}

// WithinTx 在事务中执行 fn，已处于事务中时直接执行 fn。
func (m *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if existing := docTxFromContext(ctx); existing != nil {
		return fn(ctx)
	}

	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	txCtx := context.WithValue(ctx, docTxContextKey{}, tx)
	if err := fn(txCtx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
