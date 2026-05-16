package service

import "context"

// DocRepository 定义 Doc 和 Archive 的持久化边界。
type DocRepository interface {
	// Archive
	CreateArchive(ctx context.Context, name, description, icon, author string) (*Archive, error)
	GetArchive(ctx context.Context, id string) (*Archive, error)
	ListArchives(ctx context.Context) ([]Archive, error)
	UpdateArchive(ctx context.Context, id, name, description, icon string) (*Archive, error)
	DeleteArchive(ctx context.Context, id string) error

	// Doc
	CreateDoc(ctx context.Context, params CreateDocParams) (*Doc, error)
	GetDoc(ctx context.Context, id string) (*Doc, error)
	ListDocs(ctx context.Context, archiveID *string, parentID *string, status, visibility, author string, limit, offset int32) ([]Doc, error)
	ListDocsHasStatus(ctx context.Context, archiveID *string, parentID *string, status, visibility, author string, limit, offset int32) ([]Doc, error)
	CountDocs(ctx context.Context, archiveID *string, parentID *string, status, visibility, author string) (int64, error)
	CountDocsHasStatus(ctx context.Context, archiveID *string, parentID *string, status, visibility, author string) (int64, error)
	UpdateDoc(ctx context.Context, id string, params UpdateDocParams) (*Doc, error)
	DeleteDoc(ctx context.Context, id string) error
	SearchDocs(ctx context.Context, query string, archiveID *string, visibility, author string) ([]Doc, error)
	GetDocChildren(ctx context.Context, parentID string) ([]Doc, error)
	GetDocRootChildren(ctx context.Context, archiveID string) ([]Doc, error)
	GetAllDocsByArchive(ctx context.Context, archiveID string) ([]Doc, error)
	GetDocAncestors(ctx context.Context, id string) ([]Doc, error)
	DeleteDocSubtree(ctx context.Context, id string) error

	// Link
	CreateLink(ctx context.Context, sourceID, targetID, relationType string) (*DocLink, error)
	GetOutLinks(ctx context.Context, sourceID, relationType string) ([]DocLink, error)
	GetInLinks(ctx context.Context, targetID, relationType string) ([]DocLink, error)
	DeleteLink(ctx context.Context, id string) error

	// WithinTx 在事务中执行 fn（支持嵌套事务）。
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// CreateDocParams 封装创建 Doc 所需的参数。
type CreateDocParams struct {
	ArchiveID   string
	ParentID    *string
	Title       string
	Content     string
	Summary     string
	Status      *string
	Priority    *string
	Assignee    string
	Tags        []string
	Author      string
	Visibility  string
	SharedWith  []string
	Attachments []Attachment
}

// UpdateDocParams 封装更新 Doc 的可选字段。
type UpdateDocParams struct {
	Title         *string
	Content       *string
	Summary       *string
	Status        *string
	Priority      *string
	Assignee      *string
	Tags          []byte
	Visibility    *string
	SharedWith    []byte
	Attachments   []byte
	ParentID      *string
	ClearStatus   bool
	ClearPriority bool
}
