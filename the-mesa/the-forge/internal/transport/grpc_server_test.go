package transport

import (
	"errors"
	"testing"
	"time"

	"github.com/charviki/maze/the-mesa/the-forge/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func TestDocToProto(t *testing.T) {
	now := time.Now()
	parentID := "parent-1"
	statusVal := "active"
	priority := "high"
	doc := &service.Doc{
		ID: "doc-1", ArchiveID: "arch-1", ParentID: &parentID,
		Title: "Title", Content: "body",
		Summary: "sum", Status: &statusVal, Priority: &priority,
		Assignee: "assignee-1",
		Tags:     []string{"tag1"}, Author: "user1",
		Visibility: "private", SharedWith: []string{"user2"},
		Attachments: []service.Attachment{{ID: "att-1", Key: "key-1", Name: "file.txt", ContentType: "text/plain", Size: 100}},
		CreatedAt:   now, UpdatedAt: now,
	}
	pb := docToProto(doc)
	if pb.GetId() != "doc-1" {
		t.Errorf("Id = %s, want doc-1", pb.GetId())
	}
	if pb.GetParentId() != "parent-1" {
		t.Errorf("ParentId = %s, want parent-1", pb.GetParentId())
	}
	if pb.GetStatus() != "active" {
		t.Errorf("Status = %s, want active", pb.GetStatus())
	}
	if pb.GetPriority() != "high" {
		t.Errorf("Priority = %s, want high", pb.GetPriority())
	}
	if pb.GetAssignee() != "assignee-1" {
		t.Errorf("Assignee = %s, want assignee-1", pb.GetAssignee())
	}
	if len(pb.GetTags()) != 1 || pb.GetTags()[0] != "tag1" {
		t.Errorf("Tags = %v, want [tag1]", pb.GetTags())
	}
	if len(pb.GetAttachments()) != 1 {
		t.Errorf("len(Attachments) = %d, want 1", len(pb.GetAttachments()))
	}
}

func TestDocToProto_NilParent(t *testing.T) {
	doc := &service.Doc{ID: "doc-1", Tags: []string{}, SharedWith: []string{}, Attachments: nil}
	pb := docToProto(doc)
	if pb.GetParentId() != "" {
		t.Errorf("ParentId = %s, want empty", pb.GetParentId())
	}
	if pb.GetAttachments() != nil {
		t.Errorf("Attachments should be nil when input is nil")
	}
	if pb.GetStatus() != "" {
		t.Errorf("Status = %s, want empty", pb.GetStatus())
	}
	if pb.GetPriority() != "" {
		t.Errorf("Priority = %s, want empty", pb.GetPriority())
	}
}

func TestDocLinkToProto(t *testing.T) {
	now := time.Now()
	link := &service.DocLink{
		ID: "link-1", SourceID: "s-1", TargetID: "t-1",
		RelationType: "references", SourceTitle: "Source", TargetTitle: "Target",
		CreatedAt: now,
	}
	pb := docLinkToProto(link)
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

func TestStrPtrOrNil(t *testing.T) {
	if got := strPtrOrNil(""); got != nil {
		t.Errorf("strPtrOrNil(\"\") = %v, want nil", got)
	}
	p := strPtrOrNil("hello")
	if p == nil || *p != "hello" {
		t.Errorf("strPtrOrNil(\"hello\") = %v, want *\"hello\"", p)
	}
}

func TestProtoToAttachments(t *testing.T) {
	doc := &service.Doc{
		ID: "doc-1",
		Attachments: []service.Attachment{
			{ID: "a-1", Key: "k-1", Name: "f.txt", ContentType: "text/plain", Size: 42},
		},
		Tags:       []string{},
		SharedWith: []string{},
	}
	pb := docToProto(doc)
	attachments := pb.GetAttachments()
	if len(attachments) != 1 {
		t.Fatalf("len(Attachments) = %d, want 1", len(attachments))
	}
	if attachments[0].GetName() != "f.txt" {
		t.Errorf("Name = %s, want f.txt", attachments[0].GetName())
	}
}

// --- toStatusError Tests ---

func TestToStatusError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode codes.Code
	}{
		{"nil", nil, codes.OK},
		{"archive not found", service.ErrArchiveNotFound, codes.NotFound},
		{"doc not found", service.ErrDocNotFound, codes.NotFound},
		{"link not found", service.ErrLinkNotFound, codes.NotFound},
		{"already exists", service.ErrAlreadyExists, codes.AlreadyExists},
		{"generic error", errors.New("something broke"), codes.Internal},
		{"already gRPC status", status.Error(codes.PermissionDenied, "denied"), codes.PermissionDenied},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toStatusError(tt.err)
			if tt.err == nil {
				if result != nil {
					t.Errorf("toStatusError(nil) = %v, want nil", result)
				}
				return
			}
			st, ok := status.FromError(result)
			if !ok {
				t.Fatalf("expected gRPC status error, got %v", result)
			}
			if st.Code() != tt.wantCode {
				t.Errorf("code = %v, want %v", st.Code(), tt.wantCode)
			}
		})
	}
}

// --- Tree Building Tests ---

func TestBuildTree(t *testing.T) {
	parentID := "parent-1"
	childID := "child-1"
	grandchildID := "gc-1"

	docs := []service.Doc{
		{ID: "root-1", Title: "Root Doc", Tags: []string{}, SharedWith: []string{}, Attachments: []service.Attachment{}},
		{ID: parentID, Title: "Folder", Tags: []string{}, SharedWith: []string{}, Attachments: []service.Attachment{}},
		{ID: childID, ParentID: &parentID, Title: "Child", Tags: []string{}, SharedWith: []string{}, Attachments: []service.Attachment{}},
		{ID: grandchildID, ParentID: &childID, Title: "Grandchild", Tags: []string{}, SharedWith: []string{}, Attachments: []service.Attachment{}},
	}

	nodes := buildTree(docs)

	if len(nodes) != 2 {
		t.Fatalf("len(root nodes) = %d, want 2", len(nodes))
	}

	// root-1 has no children
	if nodes[0].GetDoc().GetId() != "root-1" {
		t.Errorf("nodes[0].Id = %s, want root-1", nodes[0].GetDoc().GetId())
	}
	if len(nodes[0].GetChildren()) != 0 {
		t.Errorf("root-1 should have no children, got %d", len(nodes[0].GetChildren()))
	}

	// parent-1 has child-1
	if nodes[1].GetDoc().GetId() != parentID {
		t.Errorf("nodes[1].Id = %s, want %s", nodes[1].GetDoc().GetId(), parentID)
	}
	if len(nodes[1].GetChildren()) != 1 {
		t.Fatalf("folder children = %d, want 1", len(nodes[1].GetChildren()))
	}

	// child-1 has grandchild
	child := nodes[1].GetChildren()[0]
	if child.GetDoc().GetId() != childID {
		t.Errorf("child.Id = %s, want %s", child.GetDoc().GetId(), childID)
	}
	if len(child.GetChildren()) != 1 {
		t.Fatalf("grandchild count = %d, want 1", len(child.GetChildren()))
	}
	if child.GetChildren()[0].GetDoc().GetId() != grandchildID {
		t.Errorf("grandchild.Id = %s, want %s", child.GetChildren()[0].GetDoc().GetId(), grandchildID)
	}
}

func TestBuildTree_Empty(t *testing.T) {
	nodes := buildTree(nil)
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes for empty input, got %d", len(nodes))
	}
}
