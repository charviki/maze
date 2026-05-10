package service

import (
	"context"
	"testing"
	"time"

	"github.com/charviki/maze/fabrication/cradle/auth"
	"github.com/charviki/maze/the-mesa/director-core/internal/repository"
)

type stubUserStore struct{}

func (s *stubUserStore) CreateUser(_ context.Context, username, passwordHash, subjectKey string) (*repository.User, error) {
	return &repository.User{Username: username, PasswordHash: passwordHash, SubjectKey: subjectKey}, nil
}

func (s *stubUserStore) GetUserByUsername(_ context.Context, _ string) (*repository.User, error) {
	return nil, nil
}

func (s *stubUserStore) GetUserBySubjectKey(_ context.Context, _ string) (*repository.User, error) {
	return nil, nil
}

type stubCredentialStore struct {
	creds       map[string]*repository.Credential
	revokedHash []string
}

func newStubCredentialStore() *stubCredentialStore {
	return &stubCredentialStore{creds: map[string]*repository.Credential{}}
}

func (s *stubCredentialStore) StoreCredential(_ context.Context, cred *repository.Credential) error {
	s.creds[cred.TokenHash] = cred
	return nil
}

func (s *stubCredentialStore) GetCredentialByTokenHash(_ context.Context, tokenHash string) (*repository.Credential, error) {
	cred, ok := s.creds[tokenHash]
	if !ok {
		return nil, nil
	}
	return cred, nil
}

func (s *stubCredentialStore) RevokeCredential(_ context.Context, tokenHash string) error {
	s.revokedHash = append(s.revokedHash, tokenHash)
	if cred, ok := s.creds[tokenHash]; ok {
		now := time.Now().UTC()
		cred.Status = repository.CredentialStatusRevoked
		cred.RevokedAt = &now
	}
	return nil
}

func (s *stubCredentialStore) RevokeAllBySubject(_ context.Context, _ string) error {
	return nil
}

func (s *stubCredentialStore) CleanupExpired(_ context.Context) (int64, error) {
	return 0, nil
}

func (s *stubCredentialStore) ConsumeRefreshCredential(_ context.Context, tokenHash string) (*repository.Credential, error) {
	cred, ok := s.creds[tokenHash]
	if !ok || cred.Status != repository.CredentialStatusActive || cred.Type != repository.CredentialTypeUserRefresh {
		return nil, nil
	}
	now := time.Now().UTC()
	cred.Status = repository.CredentialStatusRevoked
	cred.RevokedAt = &now
	s.revokedHash = append(s.revokedHash, tokenHash)
	return cred, nil
}

type stubTxManager struct{}

func (stubTxManager) WithinTx(_ context.Context, fn func(context.Context) error) error {
	return fn(context.Background())
}

func newTestAuthService(credentialStore repository.CredentialStore) *AuthService {
	return NewAuthService(
		&stubUserStore{},
		credentialStore,
		nil,
		&stubTxManager{},
		"test-secret",
		15*time.Minute,
		24*time.Hour,
		"admin",
		"admin",
		"user:admin",
	)
}

func TestAuthServiceLogoutRevokesOwnActiveRefreshToken(t *testing.T) {
	credentialStore := newStubCredentialStore()
	refreshToken := "refresh-token"
	tokenHash := auth.HashToken(refreshToken)
	credentialStore.creds[tokenHash] = &repository.Credential{
		Type:       repository.CredentialTypeUserRefresh,
		TokenHash:  tokenHash,
		SubjectKey: "user:alice",
		Status:     repository.CredentialStatusActive,
	}

	svc := newTestAuthService(credentialStore)
	if err := svc.Logout(context.Background(), "user:alice", refreshToken); err != nil {
		t.Fatalf("Logout() 返回错误: %v", err)
	}

	if len(credentialStore.revokedHash) != 1 {
		t.Fatalf("撤销次数 = %d, 期望 1", len(credentialStore.revokedHash))
	}
	if credentialStore.revokedHash[0] != tokenHash {
		t.Fatalf("撤销 tokenHash = %q, 期望 %q", credentialStore.revokedHash[0], tokenHash)
	}
}

func TestAuthServiceLogoutIsIdempotentForMissingToken(t *testing.T) {
	credentialStore := newStubCredentialStore()
	svc := newTestAuthService(credentialStore)

	if err := svc.Logout(context.Background(), "user:alice", "missing-token"); err != nil {
		t.Fatalf("Logout() 返回错误: %v", err)
	}
	if len(credentialStore.revokedHash) != 0 {
		t.Fatalf("撤销次数 = %d, 期望 0", len(credentialStore.revokedHash))
	}
}

func TestAuthServiceLogoutDoesNotRevokeOtherSubjectToken(t *testing.T) {
	credentialStore := newStubCredentialStore()
	refreshToken := "refresh-token"
	tokenHash := auth.HashToken(refreshToken)
	credentialStore.creds[tokenHash] = &repository.Credential{
		Type:       repository.CredentialTypeUserRefresh,
		TokenHash:  tokenHash,
		SubjectKey: "user:bob",
		Status:     repository.CredentialStatusActive,
	}

	svc := newTestAuthService(credentialStore)
	if err := svc.Logout(context.Background(), "user:alice", refreshToken); err != nil {
		t.Fatalf("Logout() 返回错误: %v", err)
	}

	if len(credentialStore.revokedHash) != 0 {
		t.Fatalf("撤销次数 = %d, 期望 0", len(credentialStore.revokedHash))
	}
	if credentialStore.creds[tokenHash].Status != repository.CredentialStatusActive {
		t.Fatalf("credential status = %s, 期望保持 active", credentialStore.creds[tokenHash].Status)
	}
}

func TestAuthServiceLogoutIgnoresInactiveOrExpiredRefreshToken(t *testing.T) {
	past := time.Now().UTC().Add(-time.Minute)
	tests := []struct {
		name string
		cred repository.Credential
	}{
		{
			name: "已撤销 token 幂等成功",
			cred: repository.Credential{
				Type:       repository.CredentialTypeUserRefresh,
				SubjectKey: "user:alice",
				Status:     repository.CredentialStatusRevoked,
			},
		},
		{
			name: "已过期 token 幂等成功",
			cred: repository.Credential{
				Type:       repository.CredentialTypeUserRefresh,
				SubjectKey: "user:alice",
				Status:     repository.CredentialStatusActive,
				ExpiresAt:  &past,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			credentialStore := newStubCredentialStore()
			refreshToken := tt.name
			tokenHash := auth.HashToken(refreshToken)
			cred := tt.cred
			cred.TokenHash = tokenHash
			credentialStore.creds[tokenHash] = &cred

			svc := newTestAuthService(credentialStore)
			if err := svc.Logout(context.Background(), "user:alice", refreshToken); err != nil {
				t.Fatalf("Logout() 返回错误: %v", err)
			}
			if len(credentialStore.revokedHash) != 0 {
				t.Fatalf("撤销次数 = %d, 期望 0", len(credentialStore.revokedHash))
			}
		})
	}
}

func TestAuthServiceRefreshIsAtomic(t *testing.T) {
	store := newStubCredentialStore()
	refreshToken := "refresh-token-for-atomic"
	tokenHash := auth.HashToken(refreshToken)
	future := time.Now().UTC().Add(time.Hour)
	store.creds[tokenHash] = &repository.Credential{
		Type:       repository.CredentialTypeUserRefresh,
		TokenHash:  tokenHash,
		SubjectKey: "user:alice",
		Status:     repository.CredentialStatusActive,
		ExpiresAt:  &future,
	}
	svc := newTestAuthService(store)

	result, err := svc.Refresh(context.Background(), refreshToken)
	if err != nil {
		t.Fatalf("首次 Refresh 失败: %v", err)
	}
	if result.AccessToken == "" {
		t.Fatal("AccessToken 为空")
	}

	// 第二次用同一个 refresh token 应该失败
	_, err = svc.Refresh(context.Background(), refreshToken)
	if err == nil {
		t.Fatal("重复使用 refresh token 应返回错误")
	}
	if err.Error() != "refresh token not found or already used" {
		t.Fatalf("错误消息 = %q, 期望 %q", err.Error(), "refresh token not found or already used")
	}
}
