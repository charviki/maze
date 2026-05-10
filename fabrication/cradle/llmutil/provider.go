package llmutil

import "context"

// LLMProvider 定义流式 LLM 提供商接口。
//
// ChatStream 执行一次流式 LLM 调用，逐事件推送到 events channel。
// 返回完整的 assistant 消息，用于 tool calling 循环回传。
type LLMProvider interface {
	ChatStream(ctx context.Context, messages []LLMMessage, tools []LLMTool, events chan<- SSEEvent) (*LLMMessage, error)
}

// ProviderConfig 包含创建 LLM Provider 的通用配置。
type ProviderConfig struct {
	APIKey    string
	BaseURL   string
	Model     string
	MaxTokens int
}

const defaultMaxTokens = 4096

func (c ProviderConfig) maxTokens() int {
	if c.MaxTokens <= 0 {
		return defaultMaxTokens
	}
	return c.MaxTokens
}
