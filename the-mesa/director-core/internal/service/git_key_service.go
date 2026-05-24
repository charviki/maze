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
	"github.com/charviki/maze/fabrication/cradle/maskutil"
	"github.com/charviki/maze/fabrication/cradle/protocol"
)

// GitKeyRepo defines the interface for git key persistence.
type GitKeyRepo interface {
	Create(ctx context.Context, key *protocol.GitKey) (*protocol.GitKey, error)
	Get(ctx context.Context, name string) (*protocol.GitKey, error)
	GetByNames(ctx context.Context, names []string) ([]*protocol.GitKey, error)
	List(ctx context.Context) ([]*protocol.GitKey, error)
	Update(ctx context.Context, key *protocol.GitKey) (*protocol.GitKey, error)
	Delete(ctx context.Context, name string) error
}

// GitKeyReadSvc defines the read interface for git key service (used by HostService for GetHostConfig).
type GitKeyReadSvc interface {
	Get(ctx context.Context, name string) (*protocol.GitKey, error)
	DecryptToken(ctx context.Context, name string) (string, error)
	DecryptTokensByNames(ctx context.Context, names []string) ([]protocol.GitKeyItem, error)
}

// GitKeyService implements business logic for git key operations.
type GitKeyService struct {
	repo   GitKeyRepo
	aesGCM cipher.AEAD
	logger logutil.Logger
}

// NewGitKeyService creates a new GitKeyService.
// Returns an error if encryptKey is empty — tokens must never be stored in plaintext.
func NewGitKeyService(repo GitKeyRepo, encryptKey []byte, logger logutil.Logger) (*GitKeyService, error) {
	if len(encryptKey) == 0 {
		return nil, errors.New("git_key_encryption_key is required: tokens cannot be stored in plaintext")
	}
	block, err := aes.NewCipher(encryptKey)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create aes-gcm: %w", err)
	}
	return &GitKeyService{repo: repo, aesGCM: gcm, logger: logger}, nil
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
	mask := maskutil.MaskedValue(key.Token)

	result, err := s.repo.Create(ctx, &protocol.GitKey{
		Name:      key.Name,
		Token:     encrypted,
		TokenMask: mask,
		TokenType: key.TokenType,
		Host:      key.Host,
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

// DecryptToken decrypts and returns the token for a git key by name.
func (s *GitKeyService) DecryptToken(ctx context.Context, name string) (string, error) {
	key, err := s.repo.Get(ctx, name)
	if err != nil {
		return "", fmt.Errorf("get git key %q: %w", name, err)
	}
	if key == nil {
		return "", ErrNotFound
	}
	return s.decrypt(key.Token)
}

// Update updates an existing git key.
func (s *GitKeyService) Update(ctx context.Context, key *protocol.GitKey) (*protocol.GitKey, error) {
	existing, err := s.repo.Get(ctx, key.Name)
	if err != nil {
		return nil, fmt.Errorf("update git key %q: check existing: %w", key.Name, err)
	}
	if existing == nil {
		return nil, ErrNotFound
	}

	updateKey := &protocol.GitKey{
		Name: key.Name,
	}

	if key.Token != "" {
		encrypted, err := s.encrypt(key.Token)
		if err != nil {
			return nil, fmt.Errorf("update git key %q: encrypt token: %w", key.Name, err)
		}
		updateKey.Token = encrypted
		updateKey.TokenMask = maskutil.MaskedValue(key.Token)
	} else {
		updateKey.Token = existing.Token
		updateKey.TokenMask = existing.TokenMask
	}

	if key.TokenType != "" {
		updateKey.TokenType = key.TokenType
	} else {
		updateKey.TokenType = existing.TokenType
	}

	if key.Host != "" {
		updateKey.Host = key.Host
	} else {
		updateKey.Host = existing.Host
	}

	result, err := s.repo.Update(ctx, updateKey)
	if err != nil {
		return nil, fmt.Errorf("update git key %q: %w", key.Name, err)
	}
	return result, nil
}

// DecryptTokensByNames returns decrypted git key items for the given names in a single batch.
func (s *GitKeyService) DecryptTokensByNames(ctx context.Context, names []string) ([]protocol.GitKeyItem, error) {
	if len(names) == 0 {
		return nil, nil
	}
	keys, err := s.repo.GetByNames(ctx, names)
	if err != nil {
		return nil, fmt.Errorf("batch get git keys: %w", err)
	}
	items := make([]protocol.GitKeyItem, 0, len(keys))
	for _, key := range keys {
		decrypted, err := s.decrypt(key.Token)
		if err != nil {
			s.logger.Warnf("[git-key] decrypt token for %q failed: %v", key.Name, err)
			continue
		}
		items = append(items, protocol.GitKeyItem{
			Name:           key.Name,
			TokenType:      key.TokenType,
			Host:           key.Host,
			DecryptedToken: decrypted,
		})
	}
	return items, nil
}

func (s *GitKeyService) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	nonceSize := s.aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := s.aesGCM.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func (s *GitKeyService) encrypt(plaintext string) (string, error) {
	nonce := make([]byte, s.aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := s.aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
