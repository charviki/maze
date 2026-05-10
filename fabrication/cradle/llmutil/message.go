// Package llmutil 提供 LLM Provider 抽象和 SSE 流式响应工具，
// 供需要 LLM 交互能力的模块复用。
package llmutil

// SSE 事件类型常量。
const (
	EventThinking   = "thinking"
	EventText       = "text"
	EventToolUse    = "tool_use"
	EventToolResult = "tool_result"
	EventDocContent = "doc_content"
	EventDone       = "done"
	EventError      = "error"
)

// SSEEvent 表示一个 Server-Sent Events 事件。
type SSEEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// LLMMessage 表示对话中的一条消息。
type LLMMessage struct {
	Role       string        `json:"role"`
	Content    string        `json:"content"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
	ToolCalls  []LLMToolCall `json:"tool_calls,omitempty"`
	// Raw 存储 SDK 原始消息，用于在 tool calling 循环中回传完整内容（含 thinking block）。
	Raw interface{} `json:"-"`
}

// LLMToolCall 表示 LLM 返回的 tool call。
type LLMToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// LLMTool 定义一个可供 LLM 调用的工具。
type LLMTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}
