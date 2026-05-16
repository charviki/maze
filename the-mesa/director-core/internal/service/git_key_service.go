package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/charviki/maze/fabrication/cradle/logutil"
	"github.com/charviki/maze/fabrication/cradle/protocol"
)

// GitKeyRepo defines the interface for git key persistence.
type GitKeyRepo interface {
	Create(ctx context.Context, key *protocol.GitKey) (*protocol.GitKey, error)
	Get(ctx context.Context, name string) (*protocol.GitKey, error)
	List(ctx context.Context) ([]*protocol.GitKey, error)
	Delete(ctx context.Context, name string) error
}

// GitKeyService implements business logic for git key operations.
type GitKeyService struct {
	repo       GitKeyRepo
	encryptKey []byte
	logger     logutil.Logger
}

// NewGitKeyService creates a new GitKeyService.
// Returns an error if encryptKey is empty — tokens must never be stored in plaintext.
func NewGitKeyService(repo GitKeyRepo, encryptKey []byte, logger logutil.Logger) (*GitKeyService, error) {
	if len(encryptKey) == 0 {
		return nil, errors.New("git_key_encryption_key is required: tokens cannot be stored in plaintext")
	}
	return &GitKeyService{repo: repo, encryptKey: encryptKey, logger: logger}, nil
}

// Create creates a new git key with an encrypted token.
func (s *GitKeyService) Create(ctx context.Context, key *protocol.GitKey) (*protocol.GitKey, error) {
	existing, err := s.repo.Get(ctx, key.Name)
	if err != nil {
		return nil, fmt.Errorf("create git key %q: check existing: %w", key.Name, err)
	}
	if existing != nil {
		return nil, ErrAlreadyExists
	}

	encrypted, err := s.encrypt(key.Token)
	if err != nil {
		return nil, fmt.Errorf("create git key %q: encrypt token: %w", key.Name, err)
	}
	mask := generateTokenMask(key.Token)

	result, err := s.repo.Create(ctx, &protocol.GitKey{
		Name:      key.Name,
		Token:     encrypted,
		TokenMask: mask,
	})
	if err != nil {
		return nil, fmt.Errorf("create git key %q: %w", key.Name, err)
	}
	return result, nil
}

// Get returns a git key by name.
func (s *GitKeyService) Get(ctx context.Context, name string) (*protocol.GitKey, error) {
	key, err := s.repo.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get git key %q: %w", name, err)
	}
	if key == nil {
		return nil, ErrNotFound
	}
	return key, nil
}

// List returns all git keys.
func (s *GitKeyService) List(ctx context.Context) ([]*protocol.GitKey, error) {
	keys, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list git keys: %w", err)
	}
	return keys, nil
}

// Delete deletes a git key by name.
func (s *GitKeyService) Delete(ctx context.Context, name string) error {
	if err := s.repo.Delete(ctx, name); err != nil {
		return fmt.Errorf("delete git key %q: %w", name, err)
	}
	return nil
}

func (s *GitKeyService) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func generateTokenMask(token string) string {
	if len(token) == 0 {
		return ""
	}
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}
