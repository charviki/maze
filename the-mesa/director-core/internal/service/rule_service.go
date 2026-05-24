package service

import (
	"context"
	"fmt"

	"github.com/charviki/maze/fabrication/cradle/logutil"
	"github.com/charviki/maze/fabrication/cradle/protocol"
)

// RuleRepo defines the interface for rule persistence.
type RuleRepo interface {
	Create(ctx context.Context, rule *protocol.Rule) (*protocol.Rule, error)
	Get(ctx context.Context, name string) (*protocol.Rule, error)
	GetByNames(ctx context.Context, names []string) ([]*protocol.Rule, error)
	List(ctx context.Context) ([]*protocol.Rule, error)
	Update(ctx context.Context, rule *protocol.Rule) (*protocol.Rule, error)
	Delete(ctx context.Context, name string) error
}

// RuleService implements business logic for rule operations.
type RuleService struct {
	repo   RuleRepo
	logger logutil.Logger
}

// NewRuleService creates a new RuleService.
func NewRuleService(repo RuleRepo, logger logutil.Logger) *RuleService {
	return &RuleService{repo: repo, logger: logger}
}

// Create creates a new rule.
func (s *RuleService) Create(ctx context.Context, rule *protocol.Rule) (*protocol.Rule, error) {
	existing, err := s.repo.Get(ctx, rule.Name)
	if err != nil {
		return nil, fmt.Errorf("create rule %q: check existing: %w", rule.Name, err)
	}
	if existing != nil {
		return nil, ErrAlreadyExists
	}
	result, err := s.repo.Create(ctx, rule)
	if err != nil {
		return nil, fmt.Errorf("create rule %q: %w", rule.Name, err)
	}
	return result, nil
}

// Get returns a rule by name.
func (s *RuleService) Get(ctx context.Context, name string) (*protocol.Rule, error) {
	rule, err := s.repo.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get rule %q: %w", name, err)
	}
	if rule == nil {
		return nil, ErrNotFound
	}
	return rule, nil
}

// List returns all rules.
func (s *RuleService) List(ctx context.Context) ([]*protocol.Rule, error) {
	rules, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list rules: %w", err)
	}
	return rules, nil
}

// Update updates an existing rule.
func (s *RuleService) Update(ctx context.Context, rule *protocol.Rule) (*protocol.Rule, error) {
	result, err := s.repo.Update(ctx, rule)
	if err != nil {
		return nil, fmt.Errorf("update rule %q: %w", rule.Name, err)
	}
	if result == nil {
		return nil, ErrNotFound
	}
	return result, nil
}

// Delete deletes a rule by name.
func (s *RuleService) Delete(ctx context.Context, name string) error {
	if err := s.repo.Delete(ctx, name); err != nil {
		return fmt.Errorf("delete rule %q: %w", name, err)
	}
	return nil
}
