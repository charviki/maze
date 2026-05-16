//go:build integration

package integration

import "testing"

func TestForgeDocLink(t *testing.T) {
	archive := createTestArchive(t)
	defer suite.forgeClient.DeleteArchive(archive.ID)

	// Create two docs
	doc1, err := suite.forgeClient.CreateDoc(archive.ID, "Source Doc", "content 1")
	if err != nil {
		t.Fatalf("CreateDoc 1: %v", err)
	}
	doc2, err := suite.forgeClient.CreateDoc(archive.ID, "Target Doc", "content 2")
	if err != nil {
		t.Fatalf("CreateDoc 2: %v", err)
	}

	// Create link
	link, err := suite.forgeClient.CreateLink(doc1.ID, doc2.ID, "references")
	if err != nil {
		t.Fatalf("CreateLink: %v", err)
	}
	if link.SourceID != doc1.ID {
		t.Errorf("SourceID = %s, want %s", link.SourceID, doc1.ID)
	}
	if link.TargetID != doc2.ID {
		t.Errorf("TargetID = %s, want %s", link.TargetID, doc2.ID)
	}
	if link.RelationType != "references" {
		t.Errorf("RelationType = %s, want references", link.RelationType)
	}

	// Get links
	links, err := suite.forgeClient.GetLinks(doc1.ID)
	if err != nil {
		t.Fatalf("GetLinks: %v", err)
	}
	if len(links) < 1 {
		t.Fatalf("len(links) = %d, want >= 1", len(links))
	}
	found := false
	for _, l := range links {
		if l.ID == link.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("created link not found in GetLinks result")
	}

	// Delete link
	if err := suite.forgeClient.DeleteLink(doc1.ID, link.ID); err != nil {
		t.Fatalf("DeleteLink: %v", err)
	}

	// Verify deleted
	links2, err := suite.forgeClient.GetLinks(doc1.ID)
	if err != nil {
		t.Fatalf("GetLinks after delete: %v", err)
	}
	for _, l := range links2 {
		if l.ID == link.ID {
			t.Error("link should have been deleted")
		}
	}
}
