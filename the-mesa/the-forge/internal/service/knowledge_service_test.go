package service

import (
	"context"
	"testing"
	"time"
)

// --- Mock Repositories ---

type mockKnowledgeRepo struct {
	archives []Archive
	memories []Memory
	links    []NeuralLink
}

func (m *mockKnowledgeRepo) CreateArchive(_ context.Context, name, description, icon, author string) (*Archive, error) {
	a := &Archive{ID: "archive-1", Name: name, Description: description, Icon: icon, Author: author, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	m.archives = append(m.archives, *a)
	return a, nil
}

func (m *mockKnowledgeRepo) GetArchive(_ context.Context, id string) (*Archive, error) {
	for _, a := range m.archives {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, errNotFound
}

func (m *mockKnowledgeRepo) ListArchives(_ context.Context) ([]Archive, error) {
	return m.archives, nil
}

func (m *mockKnowledgeRepo) UpdateArchive(_ context.Context, id, name, description, icon string) (*Archive, error) {
	for i := range m.archives {
		if m.archives[i].ID == id {
			m.archives[i].Name = name
			m.archives[i].Description = description
			m.archives[i].Icon = icon
			return &m.archives[i], nil
		}
	}
	return nil, errNotFound
}

func (m *mockKnowledgeRepo) DeleteArchive(_ context.Context, id string) error {
	for i, a := range m.archives {
		if a.ID == id {
			m.archives = append(m.archives[:i], m.archives[i+1:]...)
			return nil
		}
	}
	return errNotFound
}

func (m *mockKnowledgeRepo) CreateMemory(_ context.Context, archiveID string, parentID *string, kind, title, content, summary, memType string, tags []string, author, visibility string, sharedWith []string, attachments []Attachment) (*Memory, error) {
	mem := &Memory{
		ID: "mem-1", ArchiveID: archiveID, ParentID: parentID, Kind: kind,
		Title: title, Content: content, Summary: summary, Type: memType,
		Tags: tags, Author: author, Visibility: visibility, SharedWith: sharedWith,
		Attachments: attachments, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	m.memories = append(m.memories, *mem)
	return mem, nil
}

func (m *mockKnowledgeRepo) GetMemory(_ context.Context, id string) (*Memory, error) {
	for _, mem := range m.memories {
		if mem.ID == id {
			return &mem, nil
		}
	}
	return nil, errNotFound
}

func (m *mockKnowledgeRepo) ListMemories(_ context.Context, _ *string, _ *string, _, _, _, _ string, _, _ int32) ([]Memory, error) {
	return m.memories, nil
}

func (m *mockKnowledgeRepo) CountMemories(_ context.Context, _ *string, _ *string, _, _, _, _ string) (int64, error) {
	return int64(len(m.memories)), nil
}

func (m *mockKnowledgeRepo) UpdateMemory(_ context.Context, id string, params UpdateMemoryParams) (*Memory, error) {
	for i := range m.memories {
		if m.memories[i].ID == id {
			if params.Title != nil {
				m.memories[i].Title = *params.Title
			}
			return &m.memories[i], nil
		}
	}
	return nil, errNotFound
}

func (m *mockKnowledgeRepo) DeleteMemory(_ context.Context, id string) error {
	for i, mem := range m.memories {
		if mem.ID == id {
			m.memories = append(m.memories[:i], m.memories[i+1:]...)
			return nil
		}
	}
	return errNotFound
}

func (m *mockKnowledgeRepo) SearchMemories(_ context.Context, _ string, _ *string, _, _ string) ([]Memory, error) {
	return m.memories, nil
}

func (m *mockKnowledgeRepo) GetMemoryChildren(_ context.Context, _ string) ([]Memory, error) {
	return m.memories, nil
}

func (m *mockKnowledgeRepo) GetMemoryRootChildren(_ context.Context, _ string) ([]Memory, error) {
	return m.memories, nil
}

func (m *mockKnowledgeRepo) GetMemoryAncestors(_ context.Context, _ string) ([]Memory, error) {
	return m.memories, nil
}

func (m *mockKnowledgeRepo) CreateLink(_ context.Context, sourceID, targetID, relationType string) (*NeuralLink, error) {
	link := &NeuralLink{ID: "link-1", SourceID: sourceID, TargetID: targetID, RelationType: relationType, CreatedAt: time.Now()}
	m.links = append(m.links, *link)
	return link, nil
}

func (m *mockKnowledgeRepo) GetOutLinks(_ context.Context, _, _ string) ([]NeuralLink, error) {
	return m.links, nil
}

func (m *mockKnowledgeRepo) GetInLinks(_ context.Context, _, _ string) ([]NeuralLink, error) {
	return m.links, nil
}

func (m *mockKnowledgeRepo) DeleteLink(_ context.Context, id string) error {
	for i, l := range m.links {
		if l.ID == id {
			m.links = append(m.links[:i], m.links[i+1:]...)
			return nil
		}
	}
	return errNotFound
}

func (m *mockKnowledgeRepo) GetTotalMemories(_ context.Context) (int64, error) {
	return int64(len(m.memories)), nil
}

func (m *mockKnowledgeRepo) GetRecentMemories(_ context.Context, limit int32) ([]Memory, error) {
	if int(limit) > len(m.memories) {
		return m.memories, nil
	}
	return m.memories[:limit], nil
}

var errNotFound = &notFoundError{}

type notFoundError struct{}

func (e *notFoundError) Error() string { return "not found" }

// --- Tests ---

func TestKnowledgeService_CreateArchive(t *testing.T) {
	repo := &mockKnowledgeRepo{}
	svc := NewKnowledgeService(repo)

	archive, err := svc.CreateArchive(context.Background(), "Test Archive", "desc", "library", "user1")
	if err != nil {
		t.Fatalf("CreateArchive: %v", err)
	}
	if archive.Name != "Test Archive" {
		t.Errorf("Name = %s, want Test Archive", archive.Name)
	}
	if archive.ID == "" {
		t.Error("ID should not be empty")
	}
}

func TestKnowledgeService_ListArchives(t *testing.T) {
	repo := &mockKnowledgeRepo{}
	svc := NewKnowledgeService(repo)

	_, _ = svc.CreateArchive(context.Background(), "A1", "", "", "")
	_, _ = svc.CreateArchive(context.Background(), "A2", "", "", "")

	archives, err := svc.ListArchives(context.Background())
	if err != nil {
		t.Fatalf("ListArchives: %v", err)
	}
	if len(archives) != 2 {
		t.Errorf("len(archives) = %d, want 2", len(archives))
	}
}

func TestKnowledgeService_CreateMemory(t *testing.T) {
	repo := &mockKnowledgeRepo{}
	svc := NewKnowledgeService(repo)

	mem, err := svc.CreateMemory(context.Background(), "archive-1", nil, "doc", "Test Title", "content", "note", nil, "user1", "private", nil, nil)
	if err != nil {
		t.Fatalf("CreateMemory: %v", err)
	}
	if mem.Title != "Test Title" {
		t.Errorf("Title = %s, want Test Title", mem.Title)
	}
	if mem.Tags == nil {
		t.Error("Tags should be initialized to empty slice, not nil")
	}
}

func TestKnowledgeService_GetMemory(t *testing.T) {
	repo := &mockKnowledgeRepo{memories: []Memory{
		{ID: "mem-1", Title: "Test", Content: "body", Summary: "sum", Tags: []string{}, SharedWith: []string{}, Attachments: []Attachment{}},
	}}
	svc := NewKnowledgeService(repo)

	parsed, err := svc.GetMemory(context.Background(), "mem-1")
	if err != nil {
		t.Fatalf("GetMemory: %v", err)
	}
	if parsed.Meta.Title != "Test" {
		t.Errorf("Meta.Title = %s, want Test", parsed.Meta.Title)
	}
	if parsed.Content != "body" {
		t.Errorf("Content = %s, want body", parsed.Content)
	}
}

func TestKnowledgeService_ListMemories(t *testing.T) {
	repo := &mockKnowledgeRepo{memories: []Memory{
		{ID: "mem-1", Title: "A"},
		{ID: "mem-2", Title: "B"},
	}}
	svc := NewKnowledgeService(repo)

	items, total, err := svc.ListMemories(context.Background(), nil, nil, "", "", "", "", 10, 0)
	if err != nil {
		t.Fatalf("ListMemories: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(items) != 2 {
		t.Errorf("len(items) = %d, want 2", len(items))
	}
}

func TestKnowledgeService_DeleteMemory(t *testing.T) {
	repo := &mockKnowledgeRepo{memories: []Memory{{ID: "mem-1"}}}
	svc := NewKnowledgeService(repo)

	if err := svc.DeleteMemory(context.Background(), "mem-1"); err != nil {
		t.Fatalf("DeleteMemory: %v", err)
	}
	if len(repo.memories) != 0 {
		t.Errorf("len(memories) = %d, want 0", len(repo.memories))
	}
}

func TestKnowledgeService_CreateLink(t *testing.T) {
	repo := &mockKnowledgeRepo{}
	svc := NewKnowledgeService(repo)

	link, err := svc.CreateLink(context.Background(), "mem-1", "mem-2", "references")
	if err != nil {
		t.Fatalf("CreateLink: %v", err)
	}
	if link.SourceID != "mem-1" || link.TargetID != "mem-2" {
		t.Errorf("Link = %+v, want source=mem-1 target=mem-2", link)
	}
}

func TestKnowledgeService_GetLinks(t *testing.T) {
	repo := &mockKnowledgeRepo{links: []NeuralLink{
		{ID: "link-1", SourceID: "mem-1", TargetID: "mem-2"},
	}}
	svc := NewKnowledgeService(repo)

	outLinks, err := svc.GetLinks(context.Background(), "mem-1", "out", "")
	if err != nil {
		t.Fatalf("GetLinks(out): %v", err)
	}
	if len(outLinks) != 1 {
		t.Errorf("len(outLinks) = %d, want 1", len(outLinks))
	}

	inLinks, err := svc.GetLinks(context.Background(), "mem-2", "in", "")
	if err != nil {
		t.Fatalf("GetLinks(in): %v", err)
	}
	if len(inLinks) != 1 {
		t.Errorf("len(inLinks) = %d, want 1", len(inLinks))
	}
}

func TestKnowledgeService_EnsureDefaultArchive(t *testing.T) {
	repo := &mockKnowledgeRepo{}
	svc := NewKnowledgeService(repo)

	archive, err := svc.EnsureDefaultArchive(context.Background(), "user1")
	if err != nil {
		t.Fatalf("EnsureDefaultArchive: %v", err)
	}
	if archive.Name != "Default Archive" {
		t.Errorf("Name = %s, want Default Archive", archive.Name)
	}

	// Second call should return existing archive
	archive2, err := svc.EnsureDefaultArchive(context.Background(), "user1")
	if err != nil {
		t.Fatalf("EnsureDefaultArchive(2): %v", err)
	}
	if archive2.ID != archive.ID {
		t.Errorf("Second call should return same archive")
	}
}

func TestExtractAISummary(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{"empty", "", ""},
		{"no summary", "just some text", ""},
		{"with summary", "<!-- ai-summary -->This is AI<!-- /ai-summary -->\nRest", "This is AI"},
		{"unclosed", "<!-- ai-summary -->no closing", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractAISummary(tt.content)
			if got != tt.want {
				t.Errorf("extractAISummary() = %q, want %q", got, tt.want)
			}
		})
	}
}
