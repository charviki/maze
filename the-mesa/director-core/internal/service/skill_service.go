package service

import (
	"context"
	"fmt"

	"github.com/charviki/maze/fabrication/cradle/logutil"
	"github.com/charviki/maze/fabrication/cradle/protocol"
)

// SkillRepo defines the interface for skill persistence.
type SkillRepo interface {
	Create(ctx context.Context, skill *protocol.Skill) (*protocol.Skill, error)
	Get(ctx context.Context, name string) (*protocol.Skill, error)
	List(ctx context.Context) ([]*protocol.Skill, error)
	Update(ctx context.Context, skill *protocol.Skill) (*protocol.Skill, error)
	Delete(ctx context.Context, name string) error
}

// SkillService implements business logic for skill operations.
type SkillService struct {
	repo   SkillRepo
	logger logutil.Logger
}

// NewSkillService creates a new SkillService.
func NewSkillService(repo SkillRepo, logger logutil.Logger) *SkillService {
	return &SkillService{repo: repo, logger: logger}
}

// Create creates a new skill.
func (s *SkillService) Create(ctx context.Context, skill *protocol.Skill) (*protocol.Skill, error) {
	existing, err := s.repo.Get(ctx, skill.Name)
	if err != nil {
		return nil, fmt.Errorf("create skill %q: check existing: %w", skill.Name, err)
	}
	if existing != nil {
		return nil, ErrAlreadyExists
	}
	result, err := s.repo.Create(ctx, skill)
	if err != nil {
		return nil, fmt.Errorf("create skill %q: %w", skill.Name, err)
	}
	return result, nil
}

// Get returns a skill by name.
func (s *SkillService) Get(ctx context.Context, name string) (*protocol.Skill, error) {
	skill, err := s.repo.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get skill %q: %w", name, err)
	}
	if skill == nil {
		return nil, ErrNotFound
	}
	return skill, nil
}

// List returns all skills.
func (s *SkillService) List(ctx context.Context) ([]*protocol.Skill, error) {
	skills, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	return skills, nil
}

// Update updates an existing skill.
func (s *SkillService) Update(ctx context.Context, skill *protocol.Skill) (*protocol.Skill, error) {
	result, err := s.repo.Update(ctx, skill)
	if err != nil {
		return nil, fmt.Errorf("update skill %q: %w", skill.Name, err)
	}
	if result == nil {
		return nil, ErrNotFound
	}
	return result, nil
}

// Delete deletes a skill by name.
func (s *SkillService) Delete(ctx context.Context, name string) error {
	if err := s.repo.Delete(ctx, name); err != nil {
		return fmt.Errorf("delete skill %q: %w", name, err)
	}
	return nil
}
