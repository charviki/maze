package service

import (
	"context"
	"log"

	"github.com/charviki/maze/fabrication/cradle/llmutil"
)

// ChatService 管理 AI 对话（Oracle）。
type ChatService struct {
	chatRepo  ChatRepository
	provider  llmutil.LLMProvider
	toolDefs  []llmutil.LLMTool
	toolExec  func(name string, args map[string]interface{}) (string, error)
}

// NewChatService 创建 ChatService。
func NewChatService(chatRepo ChatRepository, provider llmutil.LLMProvider) *ChatService {
	return &ChatService{
		chatRepo: chatRepo,
		provider: provider,
	}
}

// SetTools 配置 tool 定义和执行函数。
func (s *ChatService) SetTools(defs []llmutil.LLMTool, exec func(name string, args map[string]interface{}) (string, error)) {
	s.toolDefs = defs
	s.toolExec = exec
}

// Chat 执行一次对话，通过 events channel 流式推送 SSE 事件。
// 返回后 events channel 由调用方关闭。
func (s *ChatService) Chat(ctx context.Context, userMessage string, events chan<- llmutil.SSEEvent) error {
	history, err := s.chatRepo.ListHistory(ctx)
	if err != nil {
		return err
	}

	llmMessages := []llmutil.LLMMessage{{Role: "system", Content: systemPrompt()}}
	for _, msg := range history {
		llm := llmutil.LLMMessage{Role: msg.Role, Content: msg.Content}
		for _, tc := range msg.ToolCalls {
			llm.ToolCalls = append(llm.ToolCalls, llmutil.LLMToolCall{
				ID:        tc.ID,
				Name:      tc.Name,
				Arguments: tc.Input,
			})
		}
		llmMessages = append(llmMessages, llm)
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				llmMessages = append(llmMessages, llmutil.LLMMessage{
					Role:       "tool",
					Content:    tc.Output,
					ToolCallID: tc.ID,
				})
			}
		}
	}
	llmMessages = append(llmMessages, llmutil.LLMMessage{Role: "user", Content: userMessage})

	// 保存用户消息
	if _, err := s.chatRepo.CreateMessage(ctx, "user", userMessage, nil); err != nil {
		log.Printf("[chat] save user message: %v", err)
	}

	tools := s.toolDefs
	for {
		assistantMsg, err := s.provider.ChatStream(ctx, llmMessages, tools, events)
		if err != nil {
			events <- llmutil.SSEEvent{Type: llmutil.EventError, Data: err.Error()}
			return err
		}

		llmMessages = append(llmMessages, *assistantMsg)

		if len(assistantMsg.ToolCalls) == 0 {
			// 保存 assistant 消息
			if _, err := s.chatRepo.CreateMessage(ctx, "assistant", assistantMsg.Content, nil); err != nil {
				log.Printf("[chat] save assistant message: %v", err)
			}
			events <- llmutil.SSEEvent{Type: llmutil.EventDone, Data: "complete"}
			return nil
		}

		// 执行 tool calls
		var toolCalls []ToolCall
		for _, tc := range assistantMsg.ToolCalls {
			// 对 create_knowledge 发送 doc_content 事件
			if tc.Name == "create_knowledge" {
				title, _ := tc.Arguments["title"].(string)
				content, _ := tc.Arguments["content"].(string)
				events <- llmutil.SSEEvent{
					Type: llmutil.EventDocContent,
					Data: map[string]string{"title": title, "content": content},
				}
			}

			output := ""
			if s.toolExec != nil {
				var execErr error
				output, execErr = s.toolExec(tc.Name, tc.Arguments)
				if execErr != nil {
					output = "Error: " + execErr.Error()
				}
			}

			events <- llmutil.SSEEvent{
				Type: llmutil.EventToolResult,
				Data: map[string]interface{}{"id": tc.ID, "name": tc.Name, "result": output},
			}

			toolCalls = append(toolCalls, ToolCall{ID: tc.ID, Name: tc.Name, Input: tc.Arguments, Output: output})
			llmMessages = append(llmMessages, llmutil.LLMMessage{
				Role: "tool", Content: output, ToolCallID: tc.ID,
			})
		}

		// 保存 assistant + tool 消息
		if _, err := s.chatRepo.CreateMessage(ctx, "assistant", assistantMsg.Content, toolCalls); err != nil {
			log.Printf("[chat] save assistant tool message: %v", err)
		}
		for _, tc := range toolCalls {
			if _, err := s.chatRepo.CreateMessage(ctx, "tool", tc.Output, nil); err != nil {
				log.Printf("[chat] save tool message: %v", err)
			}
		}
	}
}

// GetHistory 获取聊天历史。
func (s *ChatService) GetHistory(ctx context.Context) ([]ChatMessage, error) {
	return s.chatRepo.ListHistory(ctx)
}

// ClearHistory 清除聊天历史。
func (s *ChatService) ClearHistory(ctx context.Context) error {
	return s.chatRepo.ClearHistory(ctx)
}

func systemPrompt() string {
	return `You are the AI assistant of **The Forge**, a knowledge and task management system.

## Discussion Mode (IMPORTANT)
By default, you MUST NOT call any tools. Your default behavior is to **discuss** requirements with the user.

### When to Create Documents:
ONLY create documents and tasks when the user explicitly confirms.

## Guidelines
- Search existing documents before creating new ones to avoid duplicates
- Use Chinese if the user speaks Chinese
- Be concise but thorough in document content
- Never create documents without explicit user confirmation`
}
