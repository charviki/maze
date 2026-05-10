package llmutil

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

// OpenAIProvider 实现 LLMProvider，使用 OpenAI 兼容 API（含 DeepSeek、Groq 等）。
type OpenAIProvider struct {
	client    *openai.Client
	model     string
	maxTokens int
}

// NewOpenAIProvider 创建 OpenAI 兼容 provider。
func NewOpenAIProvider(cfg ProviderConfig) *OpenAIProvider {
	opts := []option.RequestOption{option.WithAPIKey(cfg.APIKey)}
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL != "" && baseURL != "https://api.openai.com" && baseURL != "https://api.openai.com/v1" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	client := openai.NewClient(opts...)
	return &OpenAIProvider{client: &client, model: cfg.Model, maxTokens: cfg.maxTokens()}
}

// ChatStream 流式调用 OpenAI Chat Completions API。
func (p *OpenAIProvider) ChatStream(ctx context.Context, messages []LLMMessage, tools []LLMTool, events chan<- SSEEvent) (*LLMMessage, error) {
	apiMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))
	for _, msg := range messages {
		apiMessages = append(apiMessages, p.convertMessage(msg))
	}

	params := openai.ChatCompletionNewParams{
		Model:     p.model,
		MaxTokens: openai.Int(int64(p.maxTokens)),
		Messages:  apiMessages,
	}
	if len(tools) > 0 {
		params.Tools = p.convertTools(tools)
	}

	log.Printf("[llm:openai] streaming request (model=%s, msgs=%d, tools=%d)", p.model, len(apiMessages), len(tools))

	stream := p.client.Chat.Completions.NewStreaming(ctx, params)

	var fullContent string
	var toolCalls []LLMToolCall
	pendingToolArgs := make(map[int]string)

	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta

		if delta.Content != "" {
			events <- SSEEvent{Type: EventText, Data: delta.Content}
			fullContent = fmt.Sprintf("%s%s", fullContent, delta.Content)
		}

		for _, tc := range delta.ToolCalls {
			idx := int(tc.Index)
			for len(toolCalls) <= idx {
				toolCalls = append(toolCalls, LLMToolCall{})
			}
			if tc.ID != "" {
				toolCalls[idx].ID = tc.ID
			}
			if tc.Function.Name != "" {
				toolCalls[idx].Name += tc.Function.Name
				events <- SSEEvent{
					Type: EventToolUse,
					Data: map[string]interface{}{
						"id":   toolCalls[idx].ID,
						"name": toolCalls[idx].Name,
					},
				}
			}
			if tc.Function.Arguments != "" {
				pendingToolArgs[idx] += tc.Function.Arguments
			}
		}
	}

	if err := stream.Err(); err != nil {
		log.Printf("[llm:openai] stream error: %v", err)
		return nil, fmt.Errorf("openai stream error: %w", err)
	}

	for idx, args := range pendingToolArgs {
		if idx < len(toolCalls) {
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(args), &parsed); err != nil {
				log.Printf("[llm:openai] unmarshal tool args: %v", err)
			}
			toolCalls[idx].Arguments = parsed
			events <- SSEEvent{
				Type: EventToolUse,
				Data: map[string]interface{}{
					"id":    toolCalls[idx].ID,
					"name":  toolCalls[idx].Name,
					"input": parsed,
				},
			}
		}
	}

	log.Printf("[llm:openai] stream complete: text_len=%d, tool_calls=%d", len(fullContent), len(toolCalls))

	return &LLMMessage{
		Role:      "assistant",
		Content:   fullContent,
		ToolCalls: toolCalls,
	}, nil
}

func (p *OpenAIProvider) convertMessage(msg LLMMessage) openai.ChatCompletionMessageParamUnion {
	switch msg.Role {
	case "system":
		return openai.SystemMessage(msg.Content)
	case "user":
		return openai.UserMessage(msg.Content)
	case "assistant":
		return openai.AssistantMessage(msg.Content)
	case "tool":
		return openai.ToolMessage(msg.Content, msg.ToolCallID)
	default:
		return openai.UserMessage(msg.Content)
	}
}

func (p *OpenAIProvider) convertTools(tools []LLMTool) []openai.ChatCompletionToolParam {
	result := make([]openai.ChatCompletionToolParam, len(tools))
	for i, t := range tools {
		params := shared.FunctionParameters(t.Parameters)
		result[i] = openai.ChatCompletionToolParam{
			Function: shared.FunctionDefinitionParam{
				Name:        t.Name,
				Description: openai.String(t.Description),
				Parameters:  params,
			},
		}
	}
	return result
}
