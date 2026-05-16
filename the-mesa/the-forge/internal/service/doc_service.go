package service

import (
	"context"
	"errors"
	"strings"
)

// DocService manages archives, docs, and links.
type DocService struct {
	repo DocRepository
}

// NewDocService creates a DocService.
func NewDocService(repo DocRepository) *DocService {
	return &DocService{repo: repo}
}

// --- Archive ---

// CreateArchive 创建知识库。
func (s *DocService) CreateArchive(ctx context.Context, name, description, icon, author string) (*Archive, error) {
	if strings.TrimSpace(name) == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}
	return s.repo.CreateArchive(ctx, name, description, icon, author)
}

// GetArchive 通过 ID 获取知识库。
func (s *DocService) GetArchive(ctx context.Context, id string) (*Archive, error) {
	return s.repo.GetArchive(ctx, id)
}

// ListArchives 返回所有知识库。
func (s *DocService) ListArchives(ctx context.Context) ([]Archive, error) {
	return s.repo.ListArchives(ctx)
}

// UpdateArchive 更新知识库。
func (s *DocService) UpdateArchive(ctx context.Context, id, name, description, icon string) (*Archive, error) {
	return s.repo.UpdateArchive(ctx, id, name, description, icon)
}

// DeleteArchive 删除知识库。
func (s *DocService) DeleteArchive(ctx context.Context, id string) error {
	return s.repo.DeleteArchive(ctx, id)
}

// --- Doc ---

// CreateDoc 创建文档。
func (s *DocService) CreateDoc(ctx context.Context, params CreateDocParams) (*Doc, error) {
	if strings.TrimSpace(params.ArchiveID) == "" {
		return nil, &ValidationError{Field: "archiveId", Message: "archive_id is required"}
	}
	if strings.TrimSpace(params.Title) == "" {
		return nil, &ValidationError{Field: "title", Message: "title is required"}
	}
	if params.Tags == nil {
		params.Tags = []string{}
	}
	if params.SharedWith == nil {
		params.SharedWith = []string{}
	}
	if params.Attachments == nil {
		params.Attachments = []Attachment{}
	}
	return s.repo.CreateDoc(ctx, params)
}

// GetDoc 通过 ID 获取文档。
func (s *DocService) GetDoc(ctx context.Context, id string) (*Doc, error) {
	return s.repo.GetDoc(ctx, id)
}

// ListDocs 返回分页过滤的文档列表。hasStatus 为 true 时仅返回有 status 的文档。
func (s *DocService) ListDocs(ctx context.Context, archiveID *string, parentID *string, status string, hasStatus bool, visibility, author string, limit, offset int32) ([]Doc, int64, error) {
	var items []Doc
	var total int64
	var err error

	if hasStatus {
		items, err = s.repo.ListDocsHasStatus(ctx, archiveID, parentID, status, visibility, author, limit, offset)
	} else {
		items, err = s.repo.ListDocs(ctx, archiveID, parentID, status, visibility, author, limit, offset)
	}
	if err != nil {
		return nil, 0, err
	}

	if hasStatus {
		total, err = s.repo.CountDocsHasStatus(ctx, archiveID, parentID, status, visibility, author)
	} else {
		total, err = s.repo.CountDocs(ctx, archiveID, parentID, status, visibility, author)
	}
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// UpdateDoc 更新文档字段。
func (s *DocService) UpdateDoc(ctx context.Context, id string, params UpdateDocParams) (*Doc, error) {
	return s.repo.UpdateDoc(ctx, id, params)
}

// DeleteDoc 删除文档及其所有子文档（事务中递归删除子树）。
func (s *DocService) DeleteDoc(ctx context.Context, id string) error {
	return s.repo.WithinTx(ctx, func(txCtx context.Context) error {
		return s.repo.DeleteDocSubtree(txCtx, id)
	})
}

// SearchDocs 全文搜索文档。
func (s *DocService) SearchDocs(ctx context.Context, query string, archiveID *string, visibility, author string) ([]Doc, error) {
	if strings.TrimSpace(query) == "" {
		return nil, &ValidationError{Field: "query", Message: "query is required"}
	}
	return s.repo.SearchDocs(ctx, query, archiveID, visibility, author)
}

// GetDocTree 返回文档树。
func (s *DocService) GetDocTree(ctx context.Context, archiveID string, parentID *string) ([]Doc, error) {
	if parentID != nil {
		return s.repo.GetDocChildren(ctx, *parentID)
	}
	return s.repo.GetAllDocsByArchive(ctx, archiveID)
}

// GetDocAncestors 返回文档的祖先链。
func (s *DocService) GetDocAncestors(ctx context.Context, id string) ([]Doc, error) {
	return s.repo.GetDocAncestors(ctx, id)
}

// --- Link ---

// CreateLink 创建两个文档之间的关联。
func (s *DocService) CreateLink(ctx context.Context, sourceID, targetID, relationType string) (*DocLink, error) {
	if strings.TrimSpace(sourceID) == "" {
		return nil, &ValidationError{Field: "sourceId", Message: "source_id is required"}
	}
	if strings.TrimSpace(targetID) == "" {
		return nil, &ValidationError{Field: "targetId", Message: "target_id is required"}
	}
	if strings.TrimSpace(relationType) == "" {
		return nil, &ValidationError{Field: "relationType", Message: "relation_type is required"}
	}
	if sourceID == targetID {
		return nil, &ValidationError{Field: "sourceId", Message: "source_id and target_id must be different"}
	}
	return s.repo.CreateLink(ctx, sourceID, targetID, relationType)
}

// GetLinks 返回文档的关联列表。
func (s *DocService) GetLinks(ctx context.Context, id, direction, relationType string) ([]DocLink, error) {
	switch direction {
	case "in":
		return s.repo.GetInLinks(ctx, id, relationType)
	case "out":
		return s.repo.GetOutLinks(ctx, id, relationType)
	default:
		outLinks, err := s.repo.GetOutLinks(ctx, id, relationType)
		if err != nil {
			return nil, err
		}
		inLinks, err := s.repo.GetInLinks(ctx, id, relationType)
		if err != nil {
			return nil, err
		}
		return append(outLinks, inLinks...), nil
	}
}

// DeleteLink 删除文档关联。
func (s *DocService) DeleteLink(ctx context.Context, id string) error {
	return s.repo.DeleteLink(ctx, id)
}

// EnsureDefaultArchive ensures at least one default archive exists.
func (s *DocService) EnsureDefaultArchive(ctx context.Context, author string) (*Archive, error) {
	var archive *Archive
	err := s.repo.WithinTx(ctx, func(txCtx context.Context) error {
		archives, err := s.repo.ListArchives(txCtx)
		if err != nil {
			return err
		}
		if len(archives) > 0 {
			archive = &archives[0]
			return nil
		}
		archive, err = s.repo.CreateArchive(txCtx, "Default Archive", "Default knowledge archive", "library", author)
		if err != nil {
			if errors.Is(err, ErrAlreadyExists) {
				archives, err2 := s.repo.ListArchives(txCtx)
				if err2 != nil {
					return err2
				}
				if len(archives) > 0 {
					archive = &archives[0]
					return nil
				}
			}
			return err
		}
		return nil
	})
	return archive, err
}
