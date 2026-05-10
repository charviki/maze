package service

import (
	"context"
	"strings"
)

// KnowledgeService 管理知识库、文档、关联和统计。
type KnowledgeService struct {
	repo KnowledgeRepository
}

// NewKnowledgeService 创建 KnowledgeService。
func NewKnowledgeService(repo KnowledgeRepository) *KnowledgeService {
	return &KnowledgeService{repo: repo}
}

// --- Archive ---

// CreateArchive creates a new knowledge archive.
func (s *KnowledgeService) CreateArchive(ctx context.Context, name, description, icon, author string) (*Archive, error) {
	return s.repo.CreateArchive(ctx, name, description, icon, author)
}

// GetArchive retrieves an archive by its ID.
func (s *KnowledgeService) GetArchive(ctx context.Context, id string) (*Archive, error) {
	return s.repo.GetArchive(ctx, id)
}

// ListArchives returns all knowledge archives.
func (s *KnowledgeService) ListArchives(ctx context.Context) ([]Archive, error) {
	return s.repo.ListArchives(ctx)
}

// UpdateArchive updates the name, description, and icon of an existing archive.
func (s *KnowledgeService) UpdateArchive(ctx context.Context, id, name, description, icon string) (*Archive, error) {
	return s.repo.UpdateArchive(ctx, id, name, description, icon)
}

// DeleteArchive removes an archive by its ID.
func (s *KnowledgeService) DeleteArchive(ctx context.Context, id string) error {
	return s.repo.DeleteArchive(ctx, id)
}

// --- Memory ---

// CreateMemory creates a new memory document, extracting an AI summary from the content.
func (s *KnowledgeService) CreateMemory(ctx context.Context, archiveID string, parentID *string, kind, title, content, memType string, tags []string, author, visibility string, sharedWith []string, attachments []Attachment) (*Memory, error) {
	if tags == nil {
		tags = []string{}
	}
	if sharedWith == nil {
		sharedWith = []string{}
	}
	if attachments == nil {
		attachments = []Attachment{}
	}
	summary := extractAISummary(content)
	return s.repo.CreateMemory(ctx, archiveID, parentID, kind, title, content, summary, memType, tags, author, visibility, sharedWith, attachments)
}

// GetMemory retrieves a memory by its ID and returns it as a ParsedMemory with separated meta and content.
func (s *KnowledgeService) GetMemory(ctx context.Context, id string) (*ParsedMemory, error) {
	m, err := s.repo.GetMemory(ctx, id)
	if err != nil {
		return nil, err
	}
	return &ParsedMemory{
		Meta: MemoryMeta{
			ID:          m.ID,
			ArchiveID:   m.ArchiveID,
			ParentID:    m.ParentID,
			Kind:        m.Kind,
			Title:       m.Title,
			Type:        m.Type,
			Summary:     m.Summary,
			Tags:        m.Tags,
			Author:      m.Author,
			Visibility:  m.Visibility,
			SharedWith:  m.SharedWith,
			Attachments: m.Attachments,
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
		},
		Summary: m.Summary,
		Content: m.Content,
	}, nil
}

// ListMemories returns a paginated list of memories matching the given filters along with the total count.
func (s *KnowledgeService) ListMemories(ctx context.Context, archiveID *string, parentID *string, kind, memType, visibility, author string, limit, offset int32) ([]Memory, int64, error) {
	items, err := s.repo.ListMemories(ctx, archiveID, parentID, kind, memType, visibility, author, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountMemories(ctx, archiveID, parentID, kind, memType, visibility, author)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// UpdateMemory updates the fields of an existing memory.
func (s *KnowledgeService) UpdateMemory(ctx context.Context, id string, params UpdateMemoryParams) (*Memory, error) {
	return s.repo.UpdateMemory(ctx, id, params)
}

// DeleteMemory removes a memory by its ID.
func (s *KnowledgeService) DeleteMemory(ctx context.Context, id string) error {
	return s.repo.DeleteMemory(ctx, id)
}

// SearchMemories performs a full-text search for memories matching the query.
func (s *KnowledgeService) SearchMemories(ctx context.Context, query string, archiveID *string, visibility, author string) ([]Memory, error) {
	return s.repo.SearchMemories(ctx, query, archiveID, visibility, author)
}

// GetMemoryTree returns child memories under the given parent within an archive.
// If parentID is nil, returns root-level memories (parent_id IS NULL) for the archive.
func (s *KnowledgeService) GetMemoryTree(ctx context.Context, archiveID string, parentID *string) ([]Memory, error) {
	if parentID != nil {
		return s.repo.GetMemoryChildren(ctx, *parentID)
	}
	return s.repo.GetMemoryRootChildren(ctx, archiveID)
}

// GetMemoryAncestors returns the ancestor chain of a memory.
func (s *KnowledgeService) GetMemoryAncestors(ctx context.Context, id string) ([]Memory, error) {
	return s.repo.GetMemoryAncestors(ctx, id)
}

// --- Neural Link ---

// CreateLink creates a neural link between two memories.
func (s *KnowledgeService) CreateLink(ctx context.Context, sourceID, targetID, relationType string) (*NeuralLink, error) {
	return s.repo.CreateLink(ctx, sourceID, targetID, relationType)
}

// GetLinks returns neural links for the given memory, filtered by direction ("in" or "out") and relation type.
func (s *KnowledgeService) GetLinks(ctx context.Context, id, direction, relationType string) ([]NeuralLink, error) {
	if direction == "in" {
		return s.repo.GetInLinks(ctx, id, relationType)
	}
	return s.repo.GetOutLinks(ctx, id, relationType)
}

// DeleteLink removes a neural link by its ID.
func (s *KnowledgeService) DeleteLink(ctx context.Context, id string) error {
	return s.repo.DeleteLink(ctx, id)
}

// --- Stats ---

// GetStats returns aggregated statistics including total memories, directives by status, and recent memories.
func (s *KnowledgeService) GetStats(ctx context.Context, repo TaskRepository) (*Stats, error) {
	total, err := s.repo.GetTotalMemories(ctx)
	if err != nil {
		return nil, err
	}
	recent, err := s.repo.GetRecentMemories(ctx, 10)
	if err != nil {
		return nil, err
	}
	byStatus, err := repo.GetDirectivesByStatus(ctx)
	if err != nil {
		return nil, err
	}
	return &Stats{
		TotalMemories:      int(total),
		TotalDirectives:    sumMapValues(byStatus),
		DirectivesByStatus: byStatus,
		RecentMemories:     recent,
	}, nil
}

// --- Memory Directives ---

// GetMemoryDirectives returns all directives that reference the given memory document.
func (s *KnowledgeService) GetMemoryDirectives(ctx context.Context, id string, taskRepo TaskRepository) ([]Directive, error) {
	return taskRepo.ListDirectivesByDocID(ctx, id)
}

func sumMapValues(m map[string]int) int {
	total := 0
	for _, v := range m {
		total += v
	}
	return total
}

// extractAISummary 从 content 中提取 <!-- ai-summary --> 块。
func extractAISummary(content string) string {
	start := "<!-- ai-summary -->"
	end := "<!-- /ai-summary -->"
	sIdx := strings.Index(content, start)
	if sIdx == -1 {
		return ""
	}
	eIdx := strings.Index(content, end)
	if eIdx == -1 || eIdx <= sIdx {
		return ""
	}
	return strings.TrimSpace(content[sIdx+len(start) : eIdx])
}

// EnsureDefaultArchive 确保至少存在一个默认知识库。
func (s *KnowledgeService) EnsureDefaultArchive(ctx context.Context, author string) (*Archive, error) {
	archives, err := s.repo.ListArchives(ctx)
	if err != nil {
		return nil, err
	}
	if len(archives) > 0 {
		return &archives[0], nil
	}
	return s.repo.CreateArchive(ctx, "Default Archive", "Default knowledge archive", "library", author)
}
