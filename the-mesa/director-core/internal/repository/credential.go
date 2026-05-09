package repository

import (
	"context"
	"time"
)

// CredentialType 区分 refresh token 与 host token 两种凭证类型。
type CredentialType string

const (
	// CredentialTypeUserRefresh 用户刷新令牌。
	CredentialTypeUserRefresh CredentialType = "user_refresh"
	// CredentialTypeHostToken 主机访问令牌。
	CredentialTypeHostToken CredentialType = "host_token"
)

// CredentialStatus 表示凭证的当前生命周期状态。
type CredentialStatus string

const (
	// CredentialStatusActive 凭证处于有效状态。
	CredentialStatusActive CredentialStatus = "active"
	// CredentialStatusRevoked 凭证已被主动撤销。
	CredentialStatusRevoked CredentialStatus = "revoked"
	// CredentialStatusExpired 凭证已过期。
	CredentialStatusExpired CredentialStatus = "expired"
)

// Credential 表示 auth.credentials 表中的一条凭证记录。
type Credential struct {
	ID         int64
	Type       CredentialType
	TokenHash  string
	SubjectKey string
	ExpiresAt  *time.Time
	Status     CredentialStatus
	CreatedAt  time.Time
	RevokedAt  *time.Time
}

// CredentialStore 定义凭证持久化的最小能力集。
type CredentialStore interface {
	StoreCredential(ctx context.Context, cred *Credential) error
	GetCredentialByTokenHash(ctx context.Context, tokenHash string) (*Credential, error)
	RevokeCredential(ctx context.Context, tokenHash string) error
	RevokeAllBySubject(ctx context.Context, subjectKey string) error
	CleanupExpired(ctx context.Context) (int64, error)
	// ConsumeRefreshCredential 原子地消费一条 active 状态的 refresh 凭证。
	// 使用 UPDATE ... WHERE status='active' ... RETURNING 模式，返回被消费的凭证。
	// 若凭证不存在、已撤销或已被并发消费，返回 nil, nil。
	ConsumeRefreshCredential(ctx context.Context, tokenHash string) (*Credential, error)
}
