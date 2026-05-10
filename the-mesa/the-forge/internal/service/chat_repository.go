package service

import "context"

// ChatRepository 定义聊天历史的持久化边界。
type ChatRepository interface {
	CreateMessage(ctx context.Context, role, content string, toolCalls []ToolCall) (*ChatMessage, error)
	ListHistory(ctx context.Context) ([]ChatMessage, error)
	ClearHistory(ctx context.Context) error
}
