package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/charviki/maze/the-mesa/director-core/internal/repository"
)

var _ repository.UserStore = (*UserRepository)(nil)

// UserRepository 是 PostgreSQL 驱动的用户仓储实现。
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository 创建 PostgreSQL 用户仓储。
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// userExecutor 返回当前事务或连接池作为查询执行器，使方法可在事务内外复用。
func userExecutor(ctx context.Context, pool *pgxpool.Pool) dbExecutor {
	if tx, _ := ctx.Value(authTxContextKey{}).(pgx.Tx); tx != nil {
		return tx
	}
	return pool
}

// CreateUser 创建新用户并回写数据库生成的字段。
func (r *UserRepository) CreateUser(ctx context.Context, username, passwordHash, subjectKey string) (*repository.User, error) {
	var u repository.User
	err := userExecutor(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO auth.users (username, password_hash, subject_key)
		VALUES ($1, $2, $3)
		RETURNING id, username, password_hash, subject_key, is_active, created_at, updated_at
	`, username, passwordHash, subjectKey).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.SubjectKey,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByUsername 按用户名查询，不存在时返回 nil, nil。
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*repository.User, error) {
	return r.getUser(ctx, "SELECT id, username, password_hash, subject_key, is_active, created_at, updated_at FROM auth.users WHERE username = $1", username)
}

// GetUserBySubjectKey 按主体标识查询，不存在时返回 nil, nil。
func (r *UserRepository) GetUserBySubjectKey(ctx context.Context, subjectKey string) (*repository.User, error) {
	return r.getUser(ctx, "SELECT id, username, password_hash, subject_key, is_active, created_at, updated_at FROM auth.users WHERE subject_key = $1", subjectKey)
}

func (r *UserRepository) getUser(ctx context.Context, query string, arg any) (*repository.User, error) {
	var u repository.User
	err := userExecutor(ctx, r.pool).QueryRow(ctx, query, arg).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.SubjectKey,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}
