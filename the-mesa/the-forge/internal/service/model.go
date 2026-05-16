package service

import "time"

// Archive 知识库容器，Doc 的顶层归属。
type Archive struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Author      string    `json:"author"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Doc 知识文档，一切皆文档。通过 parent_id 支持树形嵌套。status 为 nil 时是纯知识文档，非 nil 时是待办/任务。
type Doc struct {
	ID          string       `json:"id"`
	ArchiveID   string       `json:"archiveId"`
	ParentID    *string      `json:"parentId"`
	Title       string       `json:"title"`
	Content     string       `json:"content"`
	Summary     string       `json:"summary"`
	Status      *string      `json:"status"`
	Priority    *string      `json:"priority"`
	Assignee    string       `json:"assignee"`
	Tags        []string     `json:"tags"`
	Author      string       `json:"author"`
	Visibility  string       `json:"visibility"`
	SharedWith  []string     `json:"sharedWith"`
	Attachments []Attachment `json:"attachments"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

// Attachment 文件附件。
type Attachment struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
}

// DocLink 文档关联，表示两个 Doc 之间的有向引用。
type DocLink struct {
	ID           string    `json:"id"`
	SourceID     string    `json:"sourceId"`
	TargetID     string    `json:"targetId"`
	RelationType string    `json:"relationType"`
	SourceTitle  string    `json:"sourceTitle"`
	TargetTitle  string    `json:"targetTitle"`
	CreatedAt    time.Time `json:"createdAt"`
}
