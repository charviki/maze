package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/charviki/maze/fabrication/cradle/protocol"
	hostgen "github.com/charviki/maze/the-mesa/director-core/internal/repository/postgres/sqlc/host"
)

// RuleRepository implements rule persistence with PostgreSQL.
type RuleRepository struct {
	db hostgen.DBTX
}

// NewRuleRepository creates a new RuleRepository.
func NewRuleRepository(db hostgen.DBTX) *RuleRepository {
	return &RuleRepository{db: db}
}

func (r *RuleRepository) queries(ctx context.Context) *hostgen.Queries {
	return hostgen.New(hostExecutorFromContext(ctx, r.db))
}

// Create persists a new rule.
func (r *RuleRepository) Create(ctx context.Context, rule *protocol.Rule) (*protocol.Rule, error) {
	row, err := r.queries(ctx).CreateRule(ctx, hostgen.CreateRuleParams{
		Name:    rule.Name,
		Content: rule.Content,
	})
	if err != nil {
		return nil, err
	}
	result := ruleFromRow(row)
	return &result, nil
}

// Get returns a rule by name. Returns nil if not found.
func (r *RuleRepository) Get(ctx context.Context, name string) (*protocol.Rule, error) {
	row, err := r.queries(ctx).GetRuleByName(ctx, name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	result := ruleFromRow(row)
	return &result, nil
}

// List returns all rules.
func (r *RuleRepository) List(ctx context.Context) ([]*protocol.Rule, error) {
	rows, err := r.queries(ctx).ListRules(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*protocol.Rule, 0, len(rows))
	for _, row := range rows {
		s := ruleFromRow(row)
		result = append(result, &s)
	}
	return result, nil
}

// Update updates an existing rule. Returns nil if not found.
func (r *RuleRepository) Update(ctx context.Context, rule *protocol.Rule) (*protocol.Rule, error) {
	row, err := r.queries(ctx).UpdateRule(ctx, hostgen.UpdateRuleParams{
		Name:    rule.Name,
		Content: rule.Content,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	result := ruleFromRow(row)
	return &result, nil
}

// Delete deletes a rule by name.
func (r *RuleRepository) Delete(ctx context.Context, name string) error {
	return r.queries(ctx).DeleteRuleByName(ctx, name)
}

// GetByNames returns rules matching the given names.
func (r *RuleRepository) GetByNames(ctx context.Context, names []string) ([]*protocol.Rule, error) {
	if len(names) == 0 {
		return nil, nil
	}
	rows, err := r.queries(ctx).GetRulesByNames(ctx, names)
	if err != nil {
		return nil, err
	}
	result := make([]*protocol.Rule, 0, len(rows))
	for _, row := range rows {
		s := ruleFromRow(row)
		result = append(result, &s)
	}
	return result, nil
}

func ruleFromRow(row hostgen.Rule) protocol.Rule {
	return protocol.Rule{
		Name:      row.Name,
		Content:   row.Content,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
