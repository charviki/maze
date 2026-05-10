package service

import "context"

// TaskService 管理任务（Directive）。
type TaskService struct {
	repo TaskRepository
}

// NewTaskService 创建 TaskService。
func NewTaskService(repo TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

// CreateDirective creates a new task directive, defaulting requireDocIDs to an empty slice if nil.
func (s *TaskService) CreateDirective(ctx context.Context, title, description, status, priority, assignee, author string, requireDocIDs []string, narrativeID string, archiveID *string, visibility string) (*Directive, error) {
	if requireDocIDs == nil {
		requireDocIDs = []string{}
	}
	return s.repo.CreateDirective(ctx, title, description, status, priority, assignee, author, requireDocIDs, narrativeID, archiveID, visibility)
}

// GetDirective retrieves a directive by its ID.
func (s *TaskService) GetDirective(ctx context.Context, id string) (*Directive, error) {
	return s.repo.GetDirective(ctx, id)
}

// ListDirectives returns a paginated list of directives matching the given filters along with the total count.
func (s *TaskService) ListDirectives(ctx context.Context, status, assignee, priority, visibility string, limit, offset int32) ([]Directive, int64, error) {
	items, err := s.repo.ListDirectives(ctx, status, assignee, priority, visibility, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountDirectives(ctx, status, assignee, priority, visibility)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// UpdateDirective updates the fields of an existing directive.
func (s *TaskService) UpdateDirective(ctx context.Context, id string, params UpdateDirectiveParams) (*Directive, error) {
	return s.repo.UpdateDirective(ctx, id, params)
}

// DeleteDirective removes a directive by its ID.
func (s *TaskService) DeleteDirective(ctx context.Context, id string) error {
	return s.repo.DeleteDirective(ctx, id)
}

// Repo 返回底层 TaskRepository（用于跨 service 调用）。
func (s *TaskService) Repo() TaskRepository {
	return s.repo
}
