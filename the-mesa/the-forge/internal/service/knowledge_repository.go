package service

import "context"

// KnowledgeRepository 定义知识库和文档的持久化边界。
type KnowledgeRepository interface {
	// Archive
	CreateArchive(ctx context.Context, name, description, icon, author string) (*Archive, error)
	GetArchive(ctx context.Context, id string) (*Archive, error)
	ListArchives(ctx context.Context) ([]Archive, error)
	UpdateArchive(ctx context.Context, id, name, description, icon string) (*Archive, error)
	DeleteArchive(ctx context.Context, id string) error

	// Memory
	CreateMemory(ctx context.Context, archiveID string, parentID *string, kind, title, content, summary, memType string, tags []string, author, visibility string, sharedWith []string, attachments []Attachment) (*Memory, error)
	GetMemory(ctx context.Context, id string) (*Memory, error)
	ListMemories(ctx context.Context, archiveID *string, parentID *string, kind, memType, visibility, author string, limit, offset int32) ([]Memory, error)
	CountMemories(ctx context.Context, archiveID *string, parentID *string, kind, memType, visibility, author string) (int64, error)
	UpdateMemory(ctx context.Context, id string, params UpdateMemoryParams) (*Memory, error)
	DeleteMemory(ctx context.Context, id string) error
	SearchMemories(ctx context.Context, query string, archiveID *string, visibility, author string) ([]Memory, error)
	GetMemoryChildren(ctx context.Context, parentID string) ([]Memory, error)
	GetMemoryRootChildren(ctx context.Context, archiveID string) ([]Memory, error)
	GetMemoryAncestors(ctx context.Context, id string) ([]Memory, error)

	// Neural Link
	CreateLink(ctx context.Context, sourceID, targetID, relationType string) (*NeuralLink, error)
	GetOutLinks(ctx context.Context, sourceID, relationType string) ([]NeuralLink, error)
	GetInLinks(ctx context.Context, targetID, relationType string) ([]NeuralLink, error)
	DeleteLink(ctx context.Context, id string) error

	// Stats
	GetTotalMemories(ctx context.Context) (int64, error)
	GetRecentMemories(ctx context.Context, limit int32) ([]Memory, error)
}

// UpdateMemoryParams 封装 Memory 更新的可选字段。
type UpdateMemoryParams struct {
	Title       *string
	Content     *string
	Summary     *string
	Type        *string
	Tags        []string
	Visibility  *string
	SharedWith  []string
	Attachments []Attachment
	ParentID    *string
	Kind        *string
}
