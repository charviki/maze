//go:build integration

package integration

import (
	"strings"
	"testing"

	"github.com/charviki/maze-integration-tests/kit"
)

// createTestArchive 是文档测试的辅助函数，创建并返回一个 archive。
func createTestArchive(t *testing.T) *kit.ForgeArchive {
	t.Helper()
	archive, err := suite.forgeClient.CreateArchive("Doc Test Archive", "for testing", "library")
	if err != nil {
		t.Fatalf("create test archive: %v", err)
	}
	return archive
}

func TestForgeDocCRUD(t *testing.T) {
	archive := createTestArchive(t)
	defer suite.forgeClient.DeleteArchive(archive.ID)

	// Create
	doc, err := suite.forgeClient.CreateDoc(archive.ID, "Test Doc", "This is the body content.")
	if err != nil {
		t.Fatalf("CreateDoc: %v", err)
	}
	if doc.Title != "Test Doc" {
		t.Errorf("Title = %s, want Test Doc", doc.Title)
	}
	if doc.ArchiveID != archive.ID {
		t.Errorf("ArchiveID = %s, want %s", doc.ArchiveID, archive.ID)
	}

	// Get
	got, err := suite.forgeClient.GetDoc(doc.ID)
	if err != nil {
		t.Fatalf("GetDoc: %v", err)
	}
	if got.Title != "Test Doc" {
		t.Errorf("Title = %s, want Test Doc", got.Title)
	}
	if got.Content != "This is the body content." {
		t.Errorf("Content = %s, want body content", got.Content)
	}

	// List
	result, err := suite.forgeClient.ListDocs("archiveId=" + archive.ID)
	if err != nil {
		t.Fatalf("ListDocs: %v", err)
	}
	if result.Total < 1 {
		t.Errorf("Total = %d, want >= 1", result.Total)
	}

	// Update
	updated, err := suite.forgeClient.UpdateDoc(doc.ID, map[string]any{
		"title": "Updated Title",
	})
	if err != nil {
		t.Fatalf("UpdateDoc: %v", err)
	}
	if updated.Title != "Updated Title" {
		t.Errorf("Updated Title = %s, want Updated Title", updated.Title)
	}

	// Delete
	if err := suite.forgeClient.DeleteDoc(doc.ID); err != nil {
		t.Fatalf("DeleteDoc: %v", err)
	}
}

func TestForgeDocSearch(t *testing.T) {
	archive := createTestArchive(t)
	defer suite.forgeClient.DeleteArchive(archive.ID)

	// Create doc with unique keyword
	uniqueKeyword := "quantum-entanglement-test-" + archive.ID[:8]
	_, err := suite.forgeClient.CreateDoc(archive.ID, "Search Test", "Content about "+uniqueKeyword+" phenomenon.")
	if err != nil {
		t.Fatalf("CreateDoc: %v", err)
	}

	// Search
	results, err := suite.forgeClient.SearchDocs(uniqueKeyword, archive.ID)
	if err != nil {
		t.Fatalf("SearchDocs: %v", err)
	}
	found := false
	for _, d := range results {
		if strings.Contains(d.Content, uniqueKeyword) || strings.Contains(d.Title, uniqueKeyword) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("search for %q returned %d results, none matched", uniqueKeyword, len(results))
	}
}

func TestForgeDocTree(t *testing.T) {
	archive := createTestArchive(t)
	defer suite.forgeClient.DeleteArchive(archive.ID)

	// Create parent
	parent, err := suite.forgeClient.CreateDocFull(archive.ID, "", "Parent Folder", "", nil, nil, "", nil, "public")
	if err != nil {
		t.Fatalf("CreateDoc (parent): %v", err)
	}

	// Create child
	child, err := suite.forgeClient.CreateDocFull(archive.ID, parent.ID, "Child Doc", "child content", nil, nil, "", []string{"tag1"}, "public")
	if err != nil {
		t.Fatalf("CreateDoc (child): %v", err)
	}

	// Get tree
	nodes, err := suite.forgeClient.GetDocTree("archiveId=" + archive.ID)
	if err != nil {
		t.Fatalf("GetDocTree: %v", err)
	}
	if len(nodes) < 1 {
		t.Errorf("tree nodes = %d, want >= 1", len(nodes))
	}

	// Get ancestors of child
	ancestors, err := suite.forgeClient.GetDocAncestors(child.ID)
	if err != nil {
		t.Fatalf("GetDocAncestors: %v", err)
	}
	found := false
	for _, a := range ancestors {
		if a.ID == parent.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("parent not found in ancestors of child; ancestors=%+v", ancestors)
	}
}

func TestForgeDocStatusPriority(t *testing.T) {
	archive := createTestArchive(t)
	defer suite.forgeClient.DeleteArchive(archive.ID)

	// Create doc with status and priority
	statusVal := "in-progress"
	priority := "high"
	doc, err := suite.forgeClient.CreateDocFull(archive.ID, "", "Task Doc", "task content", &statusVal, &priority, "user1", nil, "public")
	if err != nil {
		t.Fatalf("CreateDoc (with status): %v", err)
	}
	if doc.Status != statusVal {
		t.Errorf("Status = %s, want %s", doc.Status, statusVal)
	}
	if doc.Priority != priority {
		t.Errorf("Priority = %s, want %s", doc.Priority, priority)
	}

	// List docs with status filter
	result, err := suite.forgeClient.ListDocs("archiveId=" + archive.ID + "&status=in-progress")
	if err != nil {
		t.Fatalf("ListDocs (status filter): %v", err)
	}
	if result.Total < 1 {
		t.Errorf("Total (status filter) = %d, want >= 1", result.Total)
	}
}
