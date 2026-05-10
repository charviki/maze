package service

import "context"

// TaskRepository 定义任务（Directive）的持久化边界。
type TaskRepository interface {
	CreateDirective(ctx context.Context, title, description, status, priority, assignee, author string, requireDocIDs []string, narrativeID string, archiveID *string, visibility string) (*Directive, error)
	GetDirective(ctx context.Context, id string) (*Directive, error)
	ListDirectives(ctx context.Context, status, assignee, priority, visibility string, limit, offset int32) ([]Directive, error)
	CountDirectives(ctx context.Context, status, assignee, priority, visibility string) (int64, error)
	UpdateDirective(ctx context.Context, id string, params UpdateDirectiveParams) (*Directive, error)
	DeleteDirective(ctx context.Context, id string) error
	ListDirectivesByDocID(ctx context.Context, docID string) ([]Directive, error)
	GetDirectivesByStatus(ctx context.Context) (map[string]int, error)
}

// UpdateDirectiveParams 封装 Directive 更新的可选字段。
type UpdateDirectiveParams struct {
	Title         *string
	Description   *string
	Status        *string
	Priority      *string
	Assignee      *string
	Author        *string
	RequireDocIDs []string
	NarrativeID   *string
	ArchiveID     *string
	Visibility    *string
}
