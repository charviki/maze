package service

import "time"

// Archive 知识库（Host 记忆的存储容器）。
type Archive struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Author      string    `json:"author"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Memory 知识文档（一条 Host 的记忆片段）。
type Memory struct {
	ID          string       `json:"id"`
	ArchiveID   string       `json:"archiveId"`
	ParentID    *string      `json:"parentId"`
	Kind        string       `json:"kind"`
	Title       string       `json:"title"`
	Content     string       `json:"content"`
	Summary     string       `json:"summary"`
	Type        string       `json:"type"`
	Tags        []string     `json:"tags"`
	Author      string       `json:"author"`
	Visibility  string       `json:"visibility"`
	SharedWith  []string     `json:"sharedWith"`
	Attachments []Attachment `json:"attachments"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

// ParsedMemory 包含解析后的 frontmatter 元数据。
type ParsedMemory struct {
	Meta    MemoryMeta `json:"meta"`
	Summary string     `json:"summary"`
	Content string     `json:"content"`
}

// MemoryMeta 文档元数据（不含 content）。
type MemoryMeta struct {
	ID          string       `json:"id"`
	ArchiveID   string       `json:"archiveId"`
	ParentID    *string      `json:"parentId"`
	Kind        string       `json:"kind"`
	Title       string       `json:"title"`
	Type        string       `json:"type"`
	Summary     string       `json:"summary"`
	Tags        []string     `json:"tags"`
	Author      string       `json:"author"`
	Visibility  string       `json:"visibility"`
	SharedWith  []string     `json:"sharedWith"`
	Attachments []Attachment `json:"attachments"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	Status      string       `json:"status"`
	Priority    string       `json:"priority"`
}

// Attachment 文件附件。
type Attachment struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
}

// NeuralLink 文档关联（记忆之间的关联）。
type NeuralLink struct {
	ID           string    `json:"id"`
	SourceID     string    `json:"sourceId"`
	TargetID     string    `json:"targetId"`
	RelationType string    `json:"relationType"`
	SourceTitle  string    `json:"sourceTitle"`
	TargetTitle  string    `json:"targetTitle"`
	CreatedAt    time.Time `json:"createdAt"`
}

// Directive 任务（给 Host 下达的执行指令）。
type Directive struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	Priority      string    `json:"priority"`
	Assignee      string    `json:"assignee"`
	Author        string    `json:"author"`
	RequireDocIDs []string  `json:"requireDocIds"`
	NarrativeID   string    `json:"narrativeId"`
	ArchiveID     *string   `json:"archiveId"`
	Visibility    string    `json:"visibility"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// ChatMessage 聊天消息。
type ChatMessage struct {
	ID        int64     `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	ToolCalls []ToolCall `json:"toolCalls"`
	CreatedAt time.Time `json:"createdAt"`
}

// ToolCall 工具调用记录。
type ToolCall struct {
	ID     string                 `json:"id"`
	Name   string                 `json:"name"`
	Input  map[string]interface{} `json:"input"`
	Output string                 `json:"output,omitempty"`
}

// Stats 仪表盘统计。
type Stats struct {
	TotalMemories      int                `json:"totalMemories"`
	TotalDirectives    int                `json:"totalDirectives"`
	DirectivesByStatus map[string]int     `json:"directivesByStatus"`
	RecentMemories     []Memory           `json:"recentMemories"`
}
