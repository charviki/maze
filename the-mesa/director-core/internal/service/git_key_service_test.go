package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/charviki/maze/fabrication/cradle/logutil"
	"github.com/charviki/maze/fabrication/cradle/protocol"
	service "github.com/charviki/maze/the-mesa/director-core/internal/service"
)

// stubGitKeyRepo implements service.GitKeyRepo for testing.
type stubGitKeyRepo struct {
	mu   sync.RWMutex
	keys map[string]*protocol.GitKey
}

func newStubGitKeyRepo() *stubGitKeyRepo {
	return &stubGitKeyRepo{keys: make(map[string]*protocol.GitKey)}
}

func (r *stubGitKeyRepo) Create(_ context.Context, key *protocol.GitKey) (*protocol.GitKey, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.keys[key.Name]; ok {
		return nil, nil
	}
	r.keys[key.Name] = key
	return key, nil
}

func (r *stubGitKeyRepo) Get(_ context.Context, name string) (*protocol.GitKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.keys[name], nil
}

func (r *stubGitKeyRepo) GetByNames(_ context.Context, names []string) ([]*protocol.GitKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*protocol.GitKey
	for _, name := range names {
		if k, ok := r.keys[name]; ok {
			result = append(result, k)
		}
	}
	return result, nil
}

func (r *stubGitKeyRepo) List(_ context.Context) ([]*protocol.GitKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*protocol.GitKey, 0, len(r.keys))
	for _, k := range r.keys {
		out = append(out, k)
	}
	return out, nil
}

func (r *stubGitKeyRepo) Update(_ context.Context, key *protocol.GitKey) (*protocol.GitKey, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.keys[key.Name]; !ok {
		return nil, nil
	}
	r.keys[key.Name] = key
	return key, nil
}

func (r *stubGitKeyRepo) Delete(_ context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.keys, name)
	return nil
}

func newTestGitKeyService(t *testing.T) *service.GitKeyService {
	t.Helper()
	encryptKey := make([]byte, 32)
	for i := range encryptKey {
		encryptKey[i] = byte(i)
	}
	svc, err := service.NewGitKeyService(newStubGitKeyRepo(), encryptKey, logutil.NewNop())
	if err != nil {
		t.Fatalf("create git key service: %v", err)
	}
	return svc
}

func newTestGitKeyServiceWithRepo(t *testing.T, repo *stubGitKeyRepo) *service.GitKeyService {
	t.Helper()
	encryptKey := make([]byte, 32)
	for i := range encryptKey {
		encryptKey[i] = byte(i)
	}
	svc, err := service.NewGitKeyService(repo, encryptKey, logutil.NewNop())
	if err != nil {
		t.Fatalf("create git key service: %v", err)
	}
	return svc
}

func TestGitKeyService_Update_OnlyToken(t *testing.T) {
	repo := newStubGitKeyRepo()
	svc := newTestGitKeyServiceWithRepo(t, repo)

	// Create a key with tokenType and host
	created, err := svc.Create(context.Background(), &protocol.GitKey{
		Name:      "key-1",
		Token:     "original-token",
		TokenType: "SSH_KEY",
		Host:      "github.com",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Update only the token
	updated, err := svc.Update(context.Background(), &protocol.GitKey{
		Name:  "key-1",
		Token: "new-token",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	// tokenType and host should be preserved
	if updated.TokenType != "SSH_KEY" {
		t.Errorf("TokenType = %q, want %q", updated.TokenType, "SSH_KEY")
	}
	if updated.Host != "github.com" {
		t.Errorf("Host = %q, want %q", updated.Host, "github.com")
	}

	// Token should have changed (mask different)
	if updated.TokenMask == created.TokenMask {
		t.Error("TokenMask should have changed after token update")
	}

	// Verify decryption works with new token
	decrypted, err := svc.DecryptToken(context.Background(), "key-1")
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if decrypted != "new-token" {
		t.Errorf("decrypted = %q, want %q", decrypted, "new-token")
	}
}

func TestGitKeyService_Update_OnlyHost(t *testing.T) {
	repo := newStubGitKeyRepo()
	svc := newTestGitKeyServiceWithRepo(t, repo)

	_, err := svc.Create(context.Background(), &protocol.GitKey{
		Name:      "key-2",
		Token:     "original-token",
		TokenType: "PERSONAL_ACCESS_TOKEN",
		Host:      "gitlab.com",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	updated, err := svc.Update(context.Background(), &protocol.GitKey{
		Name: "key-2",
		Host: "github.com",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if updated.Host != "github.com" {
		t.Errorf("Host = %q, want %q", updated.Host, "github.com")
	}
	// tokenType should be preserved
	if updated.TokenType != "PERSONAL_ACCESS_TOKEN" {
		t.Errorf("TokenType = %q, want %q", updated.TokenType, "PERSONAL_ACCESS_TOKEN")
	}
	// Token should still decrypt to original
	decrypted, err := svc.DecryptToken(context.Background(), "key-2")
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if decrypted != "original-token" {
		t.Errorf("decrypted = %q, want %q", decrypted, "original-token")
	}
}

func TestGitKeyService_Update_OnlyTokenType(t *testing.T) {
	repo := newStubGitKeyRepo()
	svc := newTestGitKeyServiceWithRepo(t, repo)

	_, err := svc.Create(context.Background(), &protocol.GitKey{
		Name:      "key-3",
		Token:     "original-token",
		TokenType: "PERSONAL_ACCESS_TOKEN",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	updated, err := svc.Update(context.Background(), &protocol.GitKey{
		Name:      "key-3",
		TokenType: "SSH_KEY",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if updated.TokenType != "SSH_KEY" {
		t.Errorf("TokenType = %q, want %q", updated.TokenType, "SSH_KEY")
	}
}

func TestGitKeyService_Update_NotFound(t *testing.T) {
	svc := newTestGitKeyService(t)

	_, err := svc.Update(context.Background(), &protocol.GitKey{
		Name:  "nonexistent",
		Token: "some-token",
	})
	if !errors.Is(err, service.ErrNotFound) {
		t.Errorf("error = %v, want ErrNotFound", err)
	}
}

func TestGitKeyService_DecryptTokensByNames_Batch(t *testing.T) {
	repo := newStubGitKeyRepo()
	svc := newTestGitKeyServiceWithRepo(t, repo)

	// Create 3 keys
	for _, name := range []string{"batch-a", "batch-b", "batch-c"} {
		_, err := svc.Create(context.Background(), &protocol.GitKey{
			Name:      name,
			Token:     name + "-secret",
			TokenType: "PERSONAL_ACCESS_TOKEN",
			Host:      "github.com",
		})
		if err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}

	items, err := svc.DecryptTokensByNames(context.Background(), []string{"batch-a", "batch-b", "batch-c"})
	if err != nil {
		t.Fatalf("DecryptTokensByNames: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}

	got := make(map[string]string)
	for _, item := range items {
		got[item.Name] = item.DecryptedToken
	}
	for _, name := range []string{"batch-a", "batch-b", "batch-c"} {
		if got[name] != name+"-secret" {
			t.Errorf("decrypted[%s] = %q, want %q", name, got[name], name+"-secret")
		}
	}
}

func TestGitKeyService_DecryptTokensByNames_Empty(t *testing.T) {
	svc := newTestGitKeyService(t)

	items, err := svc.DecryptTokensByNames(context.Background(), nil)
	if err != nil {
		t.Fatalf("DecryptTokensByNames: %v", err)
	}
	if items != nil {
		t.Errorf("got %v, want nil", items)
	}
}
