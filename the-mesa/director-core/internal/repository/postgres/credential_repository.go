package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/charviki/maze/the-mesa/director-core/internal/repository"
)

var _ repository.CredentialStore = (*CredentialRepository)(nil)

// CredentialRepository 是 PostgreSQL 驱动的凭证仓储实现。
type CredentialRepository struct {
	pool *pgxpool.Pool
}

// NewCredentialRepository 创建 PostgreSQL 凭证仓储。
func NewCredentialRepository(pool *pgxpool.Pool) *CredentialRepository {
	return &CredentialRepository{pool: pool}
}

// credentialExecutor 返回当前事务或连接池作为查询执行器，使方法可在事务内外复用。
func credentialExecutor(ctx context.Context, pool *pgxpool.Pool) dbExecutor {
	if tx, _ := ctx.Value(authTxContextKey{}).(pgx.Tx); tx != nil {
		return tx
	}
	return pool
}

// StoreCredential 插入凭证记录并回写数据库生成的 id 和 created_at。
func (r *CredentialRepository) StoreCredential(ctx context.Context, cred *repository.Credential) error {
	return credentialExecutor(ctx, r.pool).QueryRow(ctx, `
		INSERT INTO auth.credentials (type, token_hash, subject_key, expires_at, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, cred.Type, cred.TokenHash, cred.SubjectKey, cred.ExpiresAt, cred.Status).Scan(&cred.ID, &cred.CreatedAt)
}

// GetCredentialByTokenHash 按令牌哈希查询凭证，不存在时返回 nil, nil。
func (r *CredentialRepository) GetCredentialByTokenHash(ctx context.Context, tokenHash string) (*repository.Credential, error) {
	var c repository.Credential
	err := credentialExecutor(ctx, r.pool).QueryRow(ctx, `
		SELECT id, type, token_hash, subject_key, expires_at, status, created_at, revoked_at
		FROM auth.credentials
		WHERE token_hash = $1
	`, tokenHash).Scan(
		&c.ID, &c.Type, &c.TokenHash, &c.SubjectKey,
		&c.ExpiresAt, &c.Status, &c.CreatedAt, &c.RevokedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// RevokeCredential 撤销指定令牌哈希对应的有效凭证。
func (r *CredentialRepository) RevokeCredential(ctx context.Context, tokenHash string) error {
	_, err := credentialExecutor(ctx, r.pool).Exec(ctx, `
		UPDATE auth.credentials
		SET status = 'revoked', revoked_at = NOW()
		WHERE token_hash = $1 AND status = 'active'
	`, tokenHash)
	return err
}

// RevokeAllBySubject 撤销指定主体下的全部有效凭证。
func (r *CredentialRepository) RevokeAllBySubject(ctx context.Context, subjectKey string) error {
	_, err := credentialExecutor(ctx, r.pool).Exec(ctx, `
		UPDATE auth.credentials
		SET status = 'revoked', revoked_at = NOW()
		WHERE subject_key = $1 AND status = 'active'
	`, subjectKey)
	return err
}

// CleanupExpired 将已过期的有效凭证标记为 expired，保留审计追踪。
func (r *CredentialRepository) CleanupExpired(ctx context.Context) (int64, error) {
	tag, err := credentialExecutor(ctx, r.pool).Exec(ctx, `
		UPDATE auth.credentials
		SET status = $1
		WHERE expires_at IS NOT NULL AND expires_at < NOW() AND status = 'active'
	`, repository.CredentialStatusExpired)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// ConsumeRefreshCredential 原子消费一条 active 状态的 refresh 凭证。
// 使用 UPDATE ... RETURNING 确保并发安全：只有一个请求能成功消费。
func (r *CredentialRepository) ConsumeRefreshCredential(ctx context.Context, tokenHash string) (*repository.Credential, error) {
	var c repository.Credential
	err := credentialExecutor(ctx, r.pool).QueryRow(ctx, `
		UPDATE auth.credentials
		SET status = 'revoked', revoked_at = NOW()
		WHERE token_hash = $1 AND status = 'active' AND type = 'user_refresh'
		RETURNING id, type, token_hash, subject_key, expires_at, status, created_at, revoked_at
	`, tokenHash).Scan(
		&c.ID, &c.Type, &c.TokenHash, &c.SubjectKey,
		&c.ExpiresAt, &c.Status, &c.CreatedAt, &c.RevokedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}
