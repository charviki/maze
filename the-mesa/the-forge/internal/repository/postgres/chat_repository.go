package postgres

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/charviki/maze/the-mesa/the-forge/internal/repository/postgres/sqlc"
	"github.com/charviki/maze/the-mesa/the-forge/internal/service"
)

// ChatRepository 实现 service.ChatRepository。
type ChatRepository struct {
	pool *pgxpool.Pool
	q    *gen.Queries
}

// NewChatRepository 创建 ChatRepository。
func NewChatRepository(pool *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{pool: pool, q: gen.New(pool)}
}

// CreateMessage persists a new chat message with the given role, content, and optional tool calls.
func (r *ChatRepository) CreateMessage(ctx context.Context, role, content string, toolCalls []service.ToolCall) (*service.ChatMessage, error) {
	tcJSON, _ := json.Marshal(toolCalls)
	if tcJSON == nil {
		tcJSON = []byte("[]")
	}
	row, err := r.q.CreateChatMessage(ctx, gen.CreateChatMessageParams{
		Role: role, Content: content, ToolCalls: tcJSON,
	})
	if err != nil {
		return nil, err
	}
	return &service.ChatMessage{
		ID:        int64(row.ID),
		Role:      row.Role,
		Content:   row.Content,
		ToolCalls: jsonToToolCalls(row.ToolCalls),
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

// ListHistory returns all chat messages in chronological order.
func (r *ChatRepository) ListHistory(ctx context.Context) ([]service.ChatMessage, error) {
	rows, err := r.q.ListChatHistory(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]service.ChatMessage, len(rows))
	for i, row := range rows {
		result[i] = service.ChatMessage{
			ID:        int64(row.ID),
			Role:      row.Role,
			Content:   row.Content,
			ToolCalls: jsonToToolCalls(row.ToolCalls),
			CreatedAt: row.CreatedAt.Time,
		}
	}
	return result, nil
}

// ClearHistory deletes all chat messages from the database.
func (r *ChatRepository) ClearHistory(ctx context.Context) error {
	return r.q.ClearChatHistory(ctx)
}

func jsonToToolCalls(data []byte) []service.ToolCall {
	if data == nil {
		return nil
	}
	var result []service.ToolCall
	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("[chat-repo] unmarshal tool calls: %v", err)
	}
	return result
}

var _ service.ChatRepository = (*ChatRepository)(nil)
