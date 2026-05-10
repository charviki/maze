package llmutil

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// AnthropicProvider 实现 LLMProvider，使用 Anthropic Messages API。
type AnthropicProvider struct {
	client    *anthropic.Client
	model     string
	maxTokens int
}

// NewAnthropicProvider 创建 Anthropic provider。
func NewAnthropicProvider(cfg ProviderConfig) *AnthropicProvider {
	opts := []option.RequestOption{option.WithAPIKey(cfg.APIKey)}
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL != "" && baseURL != "https://api.anthropic.com" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	client := anthropic.NewClient(opts...)
	return &AnthropicProvider{client: &client, model: cfg.Model, maxTokens: cfg.maxTokens()}
}

// ChatStream 流式调用 Anthropic Messages API。
func (p *AnthropicProvider) ChatStream(ctx context.Context, messages []LLMMessage, tools []LLMTool, events chan<- SSEEvent) (*LLMMessage, error) {
	var systemBlocks []anthropic.TextBlockParam
	var apiMessages []anthropic.MessageParam

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			systemBlocks = append(systemBlocks, anthropic.TextBlockParam{Text: msg.Content})
		case "tool":
			toolResultBlocks := []anthropic.ContentBlockParamUnion{
				anthropic.NewToolResultBlock(msg.ToolCallID, msg.Content, false),
			}
			apiMessages = append(apiMessages, anthropic.NewUserMessage(toolResultBlocks...))
		default:
			apiMessages = append(apiMessages, p.convertMessage(msg))
		}
	}

	params := anthropic.MessageNewParams{
		Model:     p.model,
		MaxTokens: int64(p.maxTokens),
		Messages:  apiMessages,
	}
	if len(systemBlocks) > 0 {
		params.System = systemBlocks
	}
	if len(tools) > 0 {
		params.Tools = p.convertTools(tools)
	}

	log.Printf("[llm:anthropic] streaming request (model=%s, msgs=%d, tools=%d)", p.model, len(apiMessages), len(tools))

	stream := p.client.Messages.NewStreaming(ctx, params)

	var accumulated anthropic.Message

	for stream.Next() {
		event := stream.Current()
		_ = accumulated.Accumulate(event)

		switch evt := event.AsAny().(type) {
		case anthropic.ContentBlockStartEvent:
			if block, ok := evt.ContentBlock.AsAny().(anthropic.ToolUseBlock); ok {
				events <- SSEEvent{
					Type: EventToolUse,
					Data: map[string]interface{}{
						"id":   block.ID,
						"name": block.Name,
					},
				}
			}
		case anthropic.ContentBlockDeltaEvent:
			switch delta := evt.Delta.AsAny().(type) {
			case anthropic.ThinkingDelta:
				events <- SSEEvent{
					Type: EventThinking,
					Data: map[string]string{"content": delta.Thinking},
				}
			case anthropic.TextDelta:
				events <- SSEEvent{
					Type: EventText,
					Data: delta.Text,
				}
			}
		case anthropic.ContentBlockStopEvent:
			idx := int(evt.Index)
			if idx < len(accumulated.Content) {
				if toolBlock, ok := accumulated.Content[idx].AsAny().(anthropic.ToolUseBlock); ok {
					var args map[string]interface{}
					if err := json.Unmarshal([]byte(toolBlock.JSON.Input.Raw()), &args); err != nil {
						log.Printf("[llm:anthropic] unmarshal tool input: %v", err)
					}
					events <- SSEEvent{
						Type: EventToolUse,
						Data: map[string]interface{}{
							"id":    toolBlock.ID,
							"name":  toolBlock.Name,
							"input": args,
						},
					}
				}
			}
		}
	}

	if err := stream.Err(); err != nil {
		log.Printf("[llm:anthropic] stream error: %v", err)
		return nil, fmt.Errorf("anthropic stream error: %w", err)
	}

	log.Printf("[llm:anthropic] stream complete: stop_reason=%s, content_blocks=%d",
		accumulated.StopReason, len(accumulated.Content))

	return p.buildAssistantMessage(&accumulated), nil
}

func (p *AnthropicProvider) buildAssistantMessage(msg *anthropic.Message) *LLMMessage {
	result := &LLMMessage{
		Role: "assistant",
		Raw:  msg,
	}
	for _, block := range msg.Content {
		switch variant := block.AsAny().(type) {
		case anthropic.TextBlock:
			result.Content += variant.Text
		case anthropic.ToolUseBlock:
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &args); err != nil {
				log.Printf("[llm:anthropic] unmarshal tool args: %v", err)
			}
			result.ToolCalls = append(result.ToolCalls, LLMToolCall{
				ID:        variant.ID,
				Name:      variant.Name,
				Arguments: args,
			})
		}
	}
	return result
}

func (p *AnthropicProvider) convertMessage(msg LLMMessage) anthropic.MessageParam {
	if raw, ok := msg.Raw.(*anthropic.Message); ok && raw != nil {
		return raw.ToParam()
	}

	switch msg.Role {
	case "user":
		return anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content))
	case "assistant":
		blocks := make([]anthropic.ContentBlockParamUnion, 0)
		if msg.Content != "" {
			blocks = append(blocks, anthropic.NewTextBlock(msg.Content))
		}
		for _, tc := range msg.ToolCalls {
			inputJSON, _ := json.Marshal(tc.Arguments)
			blocks = append(blocks, anthropic.NewToolUseBlock(tc.ID, json.RawMessage(inputJSON), tc.Name))
		}
		return anthropic.NewAssistantMessage(blocks...)
	default:
		return anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content))
	}
}

func (p *AnthropicProvider) convertTools(tools []LLMTool) []anthropic.ToolUnionParam {
	result := make([]anthropic.ToolUnionParam, len(tools))
	for i, t := range tools {
		schema := anthropic.ToolInputSchemaParam{}
		if t.Parameters != nil {
			if props, ok := t.Parameters["properties"]; ok {
				schema.Properties = props
			}
			if req, ok := t.Parameters["required"].([]interface{}); ok {
				strs := make([]string, len(req))
				for j, r := range req {
					strs[j], _ = r.(string)
				}
				schema.Required = strs
			}
		}
		result[i] = anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        t.Name,
				Description: anthropic.String(t.Description),
				InputSchema: schema,
			},
		}
	}
	return result
}
