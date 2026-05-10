package transport

import (
	"testing"
	"time"

	"github.com/charviki/maze/the-mesa/the-forge/internal/service"
)

func TestArchiveToProto(t *testing.T) {
	now := time.Now()
	archive := &service.Archive{
		ID: "id-1", Name: "Test", Description: "desc", Icon: "library",
		Author: "user1", CreatedAt: now, UpdatedAt: now,
	}
	pb := archiveToProto(archive)
	if pb.GetId() != "id-1" {
		t.Errorf("Id = %s, want id-1", pb.GetId())
	}
	if pb.GetName() != "Test" {
		t.Errorf("Name = %s, want Test", pb.GetName())
	}
	if pb.GetCreatedAt() != now.Format(time.RFC3339) {
		t.Errorf("CreatedAt = %s, want %s", pb.GetCreatedAt(), now.Format(time.RFC3339))
	}
}

func TestMemoryToProto(t *testing.T) {
	now := time.Now()
	parentID := "parent-1"
	mem := &service.Memory{
		ID: "mem-1", ArchiveID: "arch-1", ParentID: &parentID,
		Kind: "doc", Title: "Title", Content: "body", Type: "note",
		Summary: "sum", Tags: []string{"tag1"}, Author: "user1",
		Visibility: "private", SharedWith: []string{"user2"},
		Attachments: []service.Attachment{{ID: "att-1", Key: "key-1", Name: "file.txt", ContentType: "text/plain", Size: 100}},
		CreatedAt: now, UpdatedAt: now,
	}
	pb := memoryToProto(mem)
	if pb.GetId() != "mem-1" {
		t.Errorf("Id = %s, want mem-1", pb.GetId())
	}
	if pb.GetParentId() != "parent-1" {
		t.Errorf("ParentId = %s, want parent-1", pb.GetParentId())
	}
	if len(pb.GetTags()) != 1 || pb.GetTags()[0] != "tag1" {
		t.Errorf("Tags = %v, want [tag1]", pb.GetTags())
	}
	if len(pb.GetAttachments()) != 1 {
		t.Errorf("len(Attachments) = %d, want 1", len(pb.GetAttachments()))
	}
}

func TestMemoryToProto_NilParent(t *testing.T) {
	mem := &service.Memory{ID: "mem-1", Tags: []string{}, SharedWith: []string{}, Attachments: nil}
	pb := memoryToProto(mem)
	if pb.GetParentId() != "" {
		t.Errorf("ParentId = %s, want empty", pb.GetParentId())
	}
	if pb.GetAttachments() != nil {
		t.Errorf("Attachments should be nil when input is nil")
	}
}

func TestDirectiveToProto(t *testing.T) {
	now := time.Now()
	archiveID := "arch-1"
	dir := &service.Directive{
		ID: "dir-1", Title: "Task", Description: "desc", Status: "open",
		Priority: "high", Assignee: "user1", Author: "author1",
		RequireDocIDs: []string{"doc-1"}, NarrativeID: "narr-1",
		ArchiveID: &archiveID, Visibility: "private",
		CreatedAt: now, UpdatedAt: now,
	}
	pb := directiveToProto(dir)
	if pb.GetId() != "dir-1" {
		t.Errorf("Id = %s, want dir-1", pb.GetId())
	}
	if pb.GetArchiveId() != "arch-1" {
		t.Errorf("ArchiveId = %s, want arch-1", pb.GetArchiveId())
	}
	if len(pb.GetRequireDocIds()) != 1 {
		t.Errorf("len(RequireDocIds) = %d, want 1", len(pb.GetRequireDocIds()))
	}
}

func TestDirectiveToProto_NilArchiveID(t *testing.T) {
	dir := &service.Directive{ID: "dir-1"}
	pb := directiveToProto(dir)
	if pb.GetArchiveId() != "" {
		t.Errorf("ArchiveId = %s, want empty", pb.GetArchiveId())
	}
}

func TestNeuralLinkToProto(t *testing.T) {
	now := time.Now()
	link := &service.NeuralLink{
		ID: "link-1", SourceID: "s-1", TargetID: "t-1",
		RelationType: "references", SourceTitle: "Source", TargetTitle: "Target",
		CreatedAt: now,
	}
	pb := neuralLinkToProto(link)
	if pb.GetSourceId() != "s-1" {
		t.Errorf("SourceId = %s, want s-1", pb.GetSourceId())
	}
	if pb.GetRelationType() != "references" {
		t.Errorf("RelationType = %s, want references", pb.GetRelationType())
	}
}

func TestSafeInt32(t *testing.T) {
	tests := []struct {
		input int64
		want  int32
	}{
		{0, 0},
		{100, 100},
		{int64(1) << 40, int32(1<<31 - 1)}, // overflow → MaxInt32
	}
	for _, tt := range tests {
		got := safeInt32(tt.input)
		if got != tt.want {
			t.Errorf("safeInt32(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestSafeInt32FromInt(t *testing.T) {
	got := safeInt32FromInt(100)
	if got != 100 {
		t.Errorf("safeInt32FromInt(100) = %d, want 100", got)
	}
}

func TestStrPtr(t *testing.T) {
	p := strPtr("hello")
	if p == nil || *p != "hello" {
		t.Errorf("strPtr(\"hello\") = %v, want *\"hello\"", p)
	}
}

func TestStrValue(t *testing.T) {
	if got := strValue(nil); got != "" {
		t.Errorf("strValue(nil) = %q, want empty", got)
	}
	s := "hello"
	if got := strValue(&s); got != "hello" {
		t.Errorf("strValue(&\"hello\") = %q, want hello", got)
	}
}

func TestProtoToAttachments(t *testing.T) {
	mem := &service.Memory{
		ID: "mem-1",
		Attachments: []service.Attachment{
			{ID: "a-1", Key: "k-1", Name: "f.txt", ContentType: "text/plain", Size: 42},
		},
		Tags:       []string{},
		SharedWith: []string{},
	}
	pb := memoryToProto(mem)
	attachments := pb.GetAttachments()
	if len(attachments) != 1 {
		t.Fatalf("len(Attachments) = %d, want 1", len(attachments))
	}
	if attachments[0].GetName() != "f.txt" {
		t.Errorf("Name = %s, want f.txt", attachments[0].GetName())
	}
}
