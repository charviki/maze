package service

import (
	"context"
	"testing"
	"time"
)

type mockTaskRepo struct {
	directives []Directive
}

func (m *mockTaskRepo) CreateDirective(_ context.Context, title, description, status, priority, assignee, author string, requireDocIDs []string, narrativeID string, archiveID *string, visibility string) (*Directive, error) {
	d := &Directive{
		ID: "dir-1", Title: title, Description: description, Status: status,
		Priority: priority, Assignee: assignee, Author: author,
		RequireDocIDs: requireDocIDs, NarrativeID: narrativeID,
		ArchiveID: archiveID, Visibility: visibility,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	m.directives = append(m.directives, *d)
	return d, nil
}

func (m *mockTaskRepo) GetDirective(_ context.Context, id string) (*Directive, error) {
	for _, d := range m.directives {
		if d.ID == id {
			return &d, nil
		}
	}
	return nil, errNotFound
}

func (m *mockTaskRepo) ListDirectives(_ context.Context, _, _, _, _ string, _, _ int32) ([]Directive, error) {
	return m.directives, nil
}

func (m *mockTaskRepo) CountDirectives(_ context.Context, _, _, _, _ string) (int64, error) {
	return int64(len(m.directives)), nil
}

func (m *mockTaskRepo) UpdateDirective(_ context.Context, id string, params UpdateDirectiveParams) (*Directive, error) {
	for i := range m.directives {
		if m.directives[i].ID == id {
			if params.Title != nil {
				m.directives[i].Title = *params.Title
			}
			if params.Status != nil {
				m.directives[i].Status = *params.Status
			}
			return &m.directives[i], nil
		}
	}
	return nil, errNotFound
}

func (m *mockTaskRepo) DeleteDirective(_ context.Context, id string) error {
	for i, d := range m.directives {
		if d.ID == id {
			m.directives = append(m.directives[:i], m.directives[i+1:]...)
			return nil
		}
	}
	return errNotFound
}

func (m *mockTaskRepo) ListDirectivesByDocID(_ context.Context, _ string) ([]Directive, error) {
	return m.directives, nil
}

func (m *mockTaskRepo) GetDirectivesByStatus(_ context.Context) (map[string]int, error) {
	result := make(map[string]int)
	for _, d := range m.directives {
		result[d.Status]++
	}
	return result, nil
}

func TestTaskService_CreateDirective(t *testing.T) {
	repo := &mockTaskRepo{}
	svc := NewTaskService(repo)

	dir, err := svc.CreateDirective(context.Background(), "Task 1", "desc", "open", "high", "user1", "author1", nil, "narr-1", nil, "private")
	if err != nil {
		t.Fatalf("CreateDirective: %v", err)
	}
	if dir.Title != "Task 1" {
		t.Errorf("Title = %s, want Task 1", dir.Title)
	}
	if dir.RequireDocIDs == nil {
		t.Error("RequireDocIDs should be initialized to empty slice")
	}
}

func TestTaskService_ListDirectives(t *testing.T) {
	repo := &mockTaskRepo{directives: []Directive{
		{ID: "dir-1", Title: "A", Status: "open"},
		{ID: "dir-2", Title: "B", Status: "closed"},
	}}
	svc := NewTaskService(repo)

	items, total, err := svc.ListDirectives(context.Background(), "", "", "", "", 10, 0)
	if err != nil {
		t.Fatalf("ListDirectives: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(items) != 2 {
		t.Errorf("len(items) = %d, want 2", len(items))
	}
}

func TestTaskService_DeleteDirective(t *testing.T) {
	repo := &mockTaskRepo{directives: []Directive{{ID: "dir-1"}}}
	svc := NewTaskService(repo)

	if err := svc.DeleteDirective(context.Background(), "dir-1"); err != nil {
		t.Fatalf("DeleteDirective: %v", err)
	}
	if len(repo.directives) != 0 {
		t.Errorf("len(directives) = %d, want 0", len(repo.directives))
	}
}

func TestTaskService_UpdateDirective(t *testing.T) {
	repo := &mockTaskRepo{directives: []Directive{{ID: "dir-1", Title: "Old", Status: "open"}}}
	svc := NewTaskService(repo)

	dir, err := svc.UpdateDirective(context.Background(), "dir-1", UpdateDirectiveParams{
		Title:  strPtrHelper("New"),
		Status: strPtrHelper("closed"),
	})
	if err != nil {
		t.Fatalf("UpdateDirective: %v", err)
	}
	if dir.Title != "New" {
		t.Errorf("Title = %s, want New", dir.Title)
	}
	if dir.Status != "closed" {
		t.Errorf("Status = %s, want closed", dir.Status)
	}
}

func TestTaskService_Repo(t *testing.T) {
	repo := &mockTaskRepo{}
	svc := NewTaskService(repo)
	if svc.Repo() != repo {
		t.Error("Repo() should return the underlying repository")
	}
}

func strPtrHelper(s string) *string { return &s }
