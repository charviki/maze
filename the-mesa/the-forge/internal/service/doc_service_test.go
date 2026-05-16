package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

// --- Mock Repositories ---

type mockDocRepo struct {
	archives []Archive
	docs     []Doc
	links    []DocLink
}

func (m *mockDocRepo) CreateArchive(_ context.Context, name, description, icon, author string) (*Archive, error) {
	a := &Archive{ID: "archive-1", Name: name, Description: description, Icon: icon, Author: author, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	m.archives = append(m.archives, *a)
	return a, nil
}

func (m *mockDocRepo) GetArchive(_ context.Context, id string) (*Archive, error) {
	for _, a := range m.archives {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, ErrArchiveNotFound
}

func (m *mockDocRepo) ListArchives(_ context.Context) ([]Archive, error) {
	return m.archives, nil
}

func (m *mockDocRepo) UpdateArchive(_ context.Context, id, name, description, icon string) (*Archive, error) {
	for i := range m.archives {
		if m.archives[i].ID == id {
			m.archives[i].Name = name
			m.archives[i].Description = description
			m.archives[i].Icon = icon
			return &m.archives[i], nil
		}
	}
	return nil, ErrArchiveNotFound
}

func (m *mockDocRepo) DeleteArchive(_ context.Context, id string) error {
	for i, a := range m.archives {
		if a.ID == id {
			m.archives = append(m.archives[:i], m.archives[i+1:]...)
			return nil
		}
	}
	return ErrArchiveNotFound
}

func (m *mockDocRepo) CreateDoc(_ context.Context, params CreateDocParams) (*Doc, error) {
	doc := &Doc{
		ID: "doc-1", ArchiveID: params.ArchiveID, ParentID: params.ParentID,
		Title: params.Title, Content: params.Content, Summary: params.Summary,
		Status: params.Status, Priority: params.Priority, Assignee: params.Assignee,
		Tags: params.Tags, Author: params.Author, Visibility: params.Visibility,
		SharedWith: params.SharedWith, Attachments: params.Attachments,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	m.docs = append(m.docs, *doc)
	return doc, nil
}

func (m *mockDocRepo) GetDoc(_ context.Context, id string) (*Doc, error) {
	for _, d := range m.docs {
		if d.ID == id {
			return &d, nil
		}
	}
	return nil, ErrDocNotFound
}

func (m *mockDocRepo) ListDocs(_ context.Context, _ *string, _ *string, _, _, _ string, _, _ int32) ([]Doc, error) {
	return m.docs, nil
}

func (m *mockDocRepo) CountDocs(_ context.Context, _ *string, _ *string, _, _, _ string) (int64, error) {
	return int64(len(m.docs)), nil
}

func (m *mockDocRepo) UpdateDoc(_ context.Context, id string, params UpdateDocParams) (*Doc, error) {
	for i := range m.docs {
		if m.docs[i].ID == id {
			if params.Title != nil {
				m.docs[i].Title = *params.Title
			}
			if params.Content != nil {
				m.docs[i].Content = *params.Content
			}
			if params.Summary != nil {
				m.docs[i].Summary = *params.Summary
			}
			if params.ClearStatus {
				m.docs[i].Status = nil
			}
			if params.Status != nil {
				m.docs[i].Status = params.Status
			}
			if params.ClearPriority {
				m.docs[i].Priority = nil
			}
			if params.Priority != nil {
				m.docs[i].Priority = params.Priority
			}
			if params.Assignee != nil {
				m.docs[i].Assignee = *params.Assignee
			}
			if params.Visibility != nil {
				m.docs[i].Visibility = *params.Visibility
			}
			if params.ParentID != nil {
				m.docs[i].ParentID = params.ParentID
			}
			m.docs[i].UpdatedAt = time.Now()
			return &m.docs[i], nil
		}
	}
	return nil, ErrDocNotFound
}

func (m *mockDocRepo) DeleteDoc(_ context.Context, id string) error {
	for i, d := range m.docs {
		if d.ID == id {
			m.docs = append(m.docs[:i], m.docs[i+1:]...)
			return nil
		}
	}
	return ErrDocNotFound
}

func (m *mockDocRepo) SearchDocs(_ context.Context, _ string, _ *string, _, _ string) ([]Doc, error) {
	return m.docs, nil
}

func (m *mockDocRepo) GetDocChildren(_ context.Context, _ string) ([]Doc, error) {
	return m.docs, nil
}

func (m *mockDocRepo) GetDocRootChildren(_ context.Context, _ string) ([]Doc, error) {
	return m.docs, nil
}

func (m *mockDocRepo) GetAllDocsByArchive(_ context.Context, _ string) ([]Doc, error) {
	return m.docs, nil
}

func (m *mockDocRepo) GetDocAncestors(_ context.Context, _ string) ([]Doc, error) {
	return m.docs, nil
}

func (m *mockDocRepo) CreateLink(_ context.Context, sourceID, targetID, relationType string) (*DocLink, error) {
	link := &DocLink{ID: "link-1", SourceID: sourceID, TargetID: targetID, RelationType: relationType, CreatedAt: time.Now()}
	m.links = append(m.links, *link)
	return link, nil
}

func (m *mockDocRepo) GetOutLinks(_ context.Context, _, _ string) ([]DocLink, error) {
	return m.links, nil
}

func (m *mockDocRepo) GetInLinks(_ context.Context, _, _ string) ([]DocLink, error) {
	return m.links, nil
}

func (m *mockDocRepo) DeleteLink(_ context.Context, id string) error {
	for i, l := range m.links {
		if l.ID == id {
			m.links = append(m.links[:i], m.links[i+1:]...)
			return nil
		}
	}
	return ErrLinkNotFound
}

func (m *mockDocRepo) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (m *mockDocRepo) ListDocsHasStatus(_ context.Context, _ *string, _ *string, _, _, _ string, _, _ int32) ([]Doc, error) {
	var result []Doc
	for _, d := range m.docs {
		if d.Status != nil {
			result = append(result, d)
		}
	}
	return result, nil
}

func (m *mockDocRepo) CountDocsHasStatus(_ context.Context, _ *string, _ *string, _, _, _ string) (int64, error) {
	var count int64
	for _, d := range m.docs {
		if d.Status != nil {
			count++
		}
	}
	return count, nil
}

func (m *mockDocRepo) DeleteDocSubtree(_ context.Context, id string) error {
	found := false
	toDelete := map[string]bool{id: true}
	for _, d := range m.docs {
		if d.ParentID != nil && *d.ParentID == id {
			toDelete[d.ID] = true
		}
	}
	var remaining []Doc
	for _, d := range m.docs {
		if !toDelete[d.ID] {
			remaining = append(remaining, d)
		} else {
			found = true
		}
	}
	m.docs = remaining
	if !found {
		return ErrDocNotFound
	}
	return nil
}

// --- Tests ---

func TestDocService_CreateArchive(t *testing.T) {
	repo := &mockDocRepo{}
	svc := NewDocService(repo)

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

func TestDocService_ListArchives(t *testing.T) {
	repo := &mockDocRepo{}
	svc := NewDocService(repo)

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

func TestDocService_CreateDoc(t *testing.T) {
	repo := &mockDocRepo{}
	svc := NewDocService(repo)

	doc, err := svc.CreateDoc(context.Background(), CreateDocParams{
		ArchiveID:  "archive-1",
		Title:      "Test Title",
		Content:    "content",
		Tags:       nil, // should be defaulted to []
		Visibility: "private",
		Author:     "user1",
	})
	if err != nil {
		t.Fatalf("CreateDoc: %v", err)
	}
	if doc.Title != "Test Title" {
		t.Errorf("Title = %s, want Test Title", doc.Title)
	}
	if doc.Tags == nil {
		t.Error("Tags should be initialized to empty slice, not nil")
	}
}

func TestDocService_GetDoc(t *testing.T) {
	repo := &mockDocRepo{docs: []Doc{
		{ID: "doc-1", Title: "Test", Content: "body", Summary: "sum", Tags: []string{}, SharedWith: []string{}, Attachments: []Attachment{}},
	}}
	svc := NewDocService(repo)

	doc, err := svc.GetDoc(context.Background(), "doc-1")
	if err != nil {
		t.Fatalf("GetDoc: %v", err)
	}
	if doc.Title != "Test" {
		t.Errorf("Title = %s, want Test", doc.Title)
	}
	if doc.Content != "body" {
		t.Errorf("Content = %s, want body", doc.Content)
	}
}

func TestDocService_GetDoc_NotFound(t *testing.T) {
	repo := &mockDocRepo{}
	svc := NewDocService(repo)

	_, err := svc.GetDoc(context.Background(), "nonexistent")
	if !errors.Is(err, ErrDocNotFound) {
		t.Errorf("err = %v, want ErrDocNotFound", err)
	}
}

func TestDocService_ListDocs(t *testing.T) {
	repo := &mockDocRepo{docs: []Doc{
		{ID: "doc-1", Title: "A"},
		{ID: "doc-2", Title: "B"},
	}}
	svc := NewDocService(repo)

	items, total, err := svc.ListDocs(context.Background(), nil, nil, "", false, "", "", 10, 0)
	if err != nil {
		t.Fatalf("ListDocs: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(items) != 2 {
		t.Errorf("len(items) = %d, want 2", len(items))
	}
}

func TestDocService_DeleteDoc(t *testing.T) {
	repo := &mockDocRepo{docs: []Doc{{ID: "doc-1"}}}
	svc := NewDocService(repo)

	if err := svc.DeleteDoc(context.Background(), "doc-1"); err != nil {
		t.Fatalf("DeleteDoc: %v", err)
	}
	if len(repo.docs) != 0 {
		t.Errorf("len(docs) = %d, want 0", len(repo.docs))
	}
}

func TestDocService_CreateLink(t *testing.T) {
	repo := &mockDocRepo{}
	svc := NewDocService(repo)

	link, err := svc.CreateLink(context.Background(), "doc-1", "doc-2", "references")
	if err != nil {
		t.Fatalf("CreateLink: %v", err)
	}
	if link.SourceID != "doc-1" || link.TargetID != "doc-2" {
		t.Errorf("Link = %+v, want source=doc-1 target=doc-2", link)
	}
}

func TestDocService_CreateLink_SameSourceTarget(t *testing.T) {
	repo := &mockDocRepo{}
	svc := NewDocService(repo)

	_, err := svc.CreateLink(context.Background(), "doc-1", "doc-1", "references")
	if err == nil {
		t.Fatal("expected error when source_id == target_id")
	}
}

func TestDocService_GetLinks(t *testing.T) {
	repo := &mockDocRepo{links: []DocLink{
		{ID: "link-1", SourceID: "doc-1", TargetID: "doc-2"},
	}}
	svc := NewDocService(repo)

	outLinks, err := svc.GetLinks(context.Background(), "doc-1", "out", "")
	if err != nil {
		t.Fatalf("GetLinks(out): %v", err)
	}
	if len(outLinks) != 1 {
		t.Errorf("len(outLinks) = %d, want 1", len(outLinks))
	}

	inLinks, err := svc.GetLinks(context.Background(), "doc-2", "in", "")
	if err != nil {
		t.Fatalf("GetLinks(in): %v", err)
	}
	if len(inLinks) != 1 {
		t.Errorf("len(inLinks) = %d, want 1", len(inLinks))
	}
}

func TestDocService_GetLinksBothDirections(t *testing.T) {
	repo := &mockDocRepo{links: []DocLink{
		{ID: "link-1", SourceID: "doc-1", TargetID: "doc-2"},
		{ID: "link-2", SourceID: "doc-3", TargetID: "doc-1"},
	}}
	svc := NewDocService(repo)

	// Empty direction should return both in and out links merged
	allLinks, err := svc.GetLinks(context.Background(), "doc-1", "", "")
	if err != nil {
		t.Fatalf("GetLinks(both): %v", err)
	}
	// mockDocRepo returns full list for both GetOutLinks and GetInLinks,
	// so we get 2 + 2 = 4 links
	if len(allLinks) != 4 {
		t.Errorf("len(allLinks) = %d, want 4 (2 out + 2 in)", len(allLinks))
	}

	// Explicit "both" direction
	bothLinks, err := svc.GetLinks(context.Background(), "doc-1", "both", "")
	if err != nil {
		t.Fatalf("GetLinks(both explicit): %v", err)
	}
	if len(bothLinks) != 4 {
		t.Errorf("len(bothLinks) = %d, want 4", len(bothLinks))
	}
}

func TestDocService_EnsureDefaultArchive(t *testing.T) {
	repo := &mockDocRepo{}
	svc := NewDocService(repo)

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

func TestSentinelErrors(t *testing.T) {
	if ErrArchiveNotFound == nil {
		t.Error("ErrArchiveNotFound should not be nil")
	}
	if ErrDocNotFound == nil {
		t.Error("ErrDocNotFound should not be nil")
	}
	if ErrLinkNotFound == nil {
		t.Error("ErrLinkNotFound should not be nil")
	}
	if ErrAlreadyExists == nil {
		t.Error("ErrAlreadyExists should not be nil")
	}
}

func TestDocService_UpdateDoc(t *testing.T) {
	statusVal := "active"
	repo := &mockDocRepo{docs: []Doc{
		{ID: "doc-1", Title: "Old", Content: "old-content", Status: &statusVal, Tags: []string{}, SharedWith: []string{}, Attachments: []Attachment{}},
	}}
	svc := NewDocService(repo)
	newTitle := "New Title"
	doc, err := svc.UpdateDoc(context.Background(), "doc-1", UpdateDocParams{Title: &newTitle})
	if err != nil {
		t.Fatalf("UpdateDoc: %v", err)
	}
	if doc.Title != "New Title" {
		t.Errorf("Title = %s, want New Title", doc.Title)
	}
}

func TestDocService_UpdateDoc_ClearStatus(t *testing.T) {
	statusVal := "active"
	repo := &mockDocRepo{docs: []Doc{
		{ID: "doc-1", Title: "Doc", Status: &statusVal, Tags: []string{}, SharedWith: []string{}, Attachments: []Attachment{}},
	}}
	svc := NewDocService(repo)
	doc, err := svc.UpdateDoc(context.Background(), "doc-1", UpdateDocParams{ClearStatus: true})
	if err != nil {
		t.Fatalf("UpdateDoc: %v", err)
	}
	if doc.Status != nil {
		t.Errorf("Status should be nil after clear, got %v", doc.Status)
	}
}

func TestDocService_UpdateDoc_NotFound(t *testing.T) {
	repo := &mockDocRepo{}
	svc := NewDocService(repo)
	title := "x"
	_, err := svc.UpdateDoc(context.Background(), "nonexistent", UpdateDocParams{Title: &title})
	if !errors.Is(err, ErrDocNotFound) {
		t.Errorf("err = %v, want ErrDocNotFound", err)
	}
}

func TestDocService_DeleteDoc_CascadeChildren(t *testing.T) {
	parentID := "parent-1"
	repo := &mockDocRepo{docs: []Doc{
		{ID: "parent-1", Title: "Parent", Tags: []string{}, SharedWith: []string{}, Attachments: []Attachment{}},
		{ID: "child-1", ParentID: &parentID, Title: "Child", Tags: []string{}, SharedWith: []string{}, Attachments: []Attachment{}},
	}}
	svc := NewDocService(repo)
	if err := svc.DeleteDoc(context.Background(), "parent-1"); err != nil {
		t.Fatalf("DeleteDoc: %v", err)
	}
	if len(repo.docs) != 0 {
		t.Errorf("len(docs) = %d, want 0 (cascade delete)", len(repo.docs))
	}
}
