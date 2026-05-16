package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/charviki/maze/fabrication/cradle/protocol"
	hostgen "github.com/charviki/maze/the-mesa/director-core/internal/repository/postgres/sqlc/host"
)

// GitKeyRepository implements git key persistence with PostgreSQL.
type GitKeyRepository struct {
	db hostgen.DBTX
}

// NewGitKeyRepository creates a new GitKeyRepository.
func NewGitKeyRepository(db hostgen.DBTX) *GitKeyRepository {
	return &GitKeyRepository{db: db}
}

func (r *GitKeyRepository) queries(ctx context.Context) *hostgen.Queries {
	return hostgen.New(hostExecutorFromContext(ctx, r.db))
}

// Create persists a new git key.
func (r *GitKeyRepository) Create(ctx context.Context, key *protocol.GitKey) (*protocol.GitKey, error) {
	row, err := r.queries(ctx).CreateGitKey(ctx, hostgen.CreateGitKeyParams{
		Name:           key.Name,
		EncryptedToken: key.Token,
		TokenMask:      key.TokenMask,
	})
	if err != nil {
		return nil, err
	}
	result := gitKeyFromRow(row)
	return &result, nil
}

// Get returns a git key by name. Returns nil if not found.
func (r *GitKeyRepository) Get(ctx context.Context, name string) (*protocol.GitKey, error) {
	row, err := r.queries(ctx).GetGitKeyByName(ctx, name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	result := gitKeyFromRow(row)
	return &result, nil
}

// List returns all git keys.
func (r *GitKeyRepository) List(ctx context.Context) ([]*protocol.GitKey, error) {
	rows, err := r.queries(ctx).ListGitKeys(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*protocol.GitKey, 0, len(rows))
	for _, row := range rows {
		s := gitKeyFromRow(row)
		result = append(result, &s)
	}
	return result, nil
}

// Delete deletes a git key by name.
func (r *GitKeyRepository) Delete(ctx context.Context, name string) error {
	return r.queries(ctx).DeleteGitKeyByName(ctx, name)
}

func gitKeyFromRow(row hostgen.GitKey) protocol.GitKey {
	return protocol.GitKey{
		Name:      row.Name,
		TokenMask: row.TokenMask,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
