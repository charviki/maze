package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"

	"github.com/charviki/maze/fabrication/cradle/protocol"
	hostgen "github.com/charviki/maze/the-mesa/director-core/internal/repository/postgres/sqlc/host"
)

// SkillRepository implements skill persistence with PostgreSQL.
type SkillRepository struct {
	db hostgen.DBTX
}

// NewSkillRepository creates a new SkillRepository.
func NewSkillRepository(db hostgen.DBTX) *SkillRepository {
	return &SkillRepository{db: db}
}

func (r *SkillRepository) queries(ctx context.Context) *hostgen.Queries {
	return hostgen.New(hostExecutorFromContext(ctx, r.db))
}

// Create persists a new skill.
func (r *SkillRepository) Create(ctx context.Context, skill *protocol.Skill) (*protocol.Skill, error) {
	config, err := json.Marshal(skill.Config)
	if err != nil {
		return nil, err
	}
	row, err := r.queries(ctx).CreateSkill(ctx, hostgen.CreateSkillParams{
		Name:        skill.Name,
		Description: skill.Description,
		Config:      config,
	})
	if err != nil {
		return nil, err
	}
	result := skillFromRow(row)
	return &result, nil
}

// Get returns a skill by name. Returns nil if not found.
func (r *SkillRepository) Get(ctx context.Context, name string) (*protocol.Skill, error) {
	row, err := r.queries(ctx).GetSkillByName(ctx, name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	result := skillFromRow(row)
	return &result, nil
}

// List returns all skills.
func (r *SkillRepository) List(ctx context.Context) ([]*protocol.Skill, error) {
	rows, err := r.queries(ctx).ListSkills(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*protocol.Skill, 0, len(rows))
	for _, row := range rows {
		s := skillFromRow(row)
		result = append(result, &s)
	}
	return result, nil
}

// Update updates an existing skill. Returns nil if not found.
func (r *SkillRepository) Update(ctx context.Context, skill *protocol.Skill) (*protocol.Skill, error) {
	config, err := json.Marshal(skill.Config)
	if err != nil {
		return nil, err
	}
	row, err := r.queries(ctx).UpdateSkill(ctx, hostgen.UpdateSkillParams{
		Name:        skill.Name,
		Description: skill.Description,
		Config:      config,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	result := skillFromRow(row)
	return &result, nil
}

// Delete deletes a skill by name.
func (r *SkillRepository) Delete(ctx context.Context, name string) error {
	return r.queries(ctx).DeleteSkillByName(ctx, name)
}

// GetByNames returns skills matching the given names.
func (r *SkillRepository) GetByNames(ctx context.Context, names []string) ([]*protocol.Skill, error) {
	if len(names) == 0 {
		return nil, nil
	}
	rows, err := r.queries(ctx).GetSkillsByNames(ctx, names)
	if err != nil {
		return nil, err
	}
	result := make([]*protocol.Skill, 0, len(rows))
	for _, row := range rows {
		s := skillFromRow(row)
		result = append(result, &s)
	}
	return result, nil
}

func skillFromRow(row hostgen.Skill) protocol.Skill {
	var config map[string]string
	if len(row.Config) > 0 {
		_ = json.Unmarshal(row.Config, &config)
	}
	return protocol.Skill{
		Name:        row.Name,
		Description: row.Description,
		Config:      config,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}
