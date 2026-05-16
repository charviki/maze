//go:build integration

package integration

import "testing"

func TestForgeArchiveCRUD(t *testing.T) {
	// Create
	archive, err := suite.forgeClient.CreateArchive("Test Archive", "integration test", "library")
	if err != nil {
		t.Fatalf("CreateArchive: %v", err)
	}
	if archive.Name != "Test Archive" {
		t.Errorf("Name = %s, want Test Archive", archive.Name)
	}
	if archive.ID == "" {
		t.Fatal("ID should not be empty")
	}
	if archive.Author != "user:admin" {
		t.Errorf("Author = %s, want user:admin (from JWT)", archive.Author)
	}

	// Get
	got, err := suite.forgeClient.GetArchive(archive.ID)
	if err != nil {
		t.Fatalf("GetArchive: %v", err)
	}
	if got.Name != "Test Archive" {
		t.Errorf("Got Name = %s, want Test Archive", got.Name)
	}

	// List
	archives, err := suite.forgeClient.ListArchives()
	if err != nil {
		t.Fatalf("ListArchives: %v", err)
	}
	found := false
	for _, a := range archives {
		if a.ID == archive.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("created archive not found in list")
	}

	// Update
	updated, err := suite.forgeClient.UpdateArchive(archive.ID, "Updated Name", "updated desc", "folder")
	if err != nil {
		t.Fatalf("UpdateArchive: %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("Updated Name = %s, want Updated Name", updated.Name)
	}

	// Delete
	if err := suite.forgeClient.DeleteArchive(archive.ID); err != nil {
		t.Fatalf("DeleteArchive: %v", err)
	}

	// Verify deleted
	_, err = suite.forgeClient.GetArchive(archive.ID)
	if err == nil {
		t.Error("expected error getting deleted archive")
	}
}
