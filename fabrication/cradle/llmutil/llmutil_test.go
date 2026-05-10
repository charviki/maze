package llmutil

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestSSEEventTypes(t *testing.T) {
	types := []string{EventThinking, EventText, EventToolUse, EventToolResult, EventDocContent, EventDone, EventError}
	for _, typ := range types {
		if typ == "" {
			t.Errorf("event type should not be empty")
		}
	}
}

func TestSSEEventMarshal(t *testing.T) {
	event := SSEEvent{Type: EventText, Data: "hello world"}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal SSEEvent: %v", err)
	}
	want := `{"type":"text","data":"hello world"}`
	if string(data) != want {
		t.Errorf("marshal SSEEvent = %s, want %s", data, want)
	}
}

func TestLLMMessageMarshal(t *testing.T) {
	msg := LLMMessage{
		Role:    "assistant",
		Content: "test",
		ToolCalls: []LLMToolCall{
			{ID: "tc_1", Name: "search", Arguments: map[string]interface{}{"q": "hello"}},
		},
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal LLMMessage: %v", err)
	}
	var parsed LLMMessage
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal LLMMessage: %v", err)
	}
	if parsed.Role != "assistant" || parsed.Content != "test" {
		t.Errorf("parsed LLMMessage = %+v, want role=assistant content=test", parsed)
	}
	if len(parsed.ToolCalls) != 1 || parsed.ToolCalls[0].Name != "search" {
		t.Errorf("parsed ToolCalls = %+v, want 1 call named search", parsed.ToolCalls)
	}
}

func TestWriteSSEEvent(t *testing.T) {
	rec := httptest.NewRecorder()
	SetSSEHeaders(rec)

	event := SSEEvent{Type: EventText, Data: "hello"}
	if err := WriteSSEEvent(rec, event); err != nil {
		t.Fatalf("WriteSSEEvent: %v", err)
	}

	body := rec.Body.String()
	if body == "" {
		t.Fatal("expected non-empty SSE output")
	}
	// SSE format: "event: text\ndata: \"hello\"\n\n"
	expected := "event: text\ndata: \"hello\"\n\n"
	if body != expected {
		t.Errorf("SSE output = %q, want %q", body, expected)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("Content-Type = %s, want text/event-stream", contentType)
	}
}

func TestWriteSSEString(t *testing.T) {
	rec := httptest.NewRecorder()
	SetSSEHeaders(rec)

	if err := WriteSSEString(rec, EventError, "something failed"); err != nil {
		t.Fatalf("WriteSSEString: %v", err)
	}

	body := rec.Body.String()
	expected := "event: error\ndata: \"something failed\"\n\n"
	if body != expected {
		t.Errorf("SSE output = %q, want %q", body, expected)
	}
}

func TestProviderConfig(t *testing.T) {
	cfg := ProviderConfig{
		APIKey:  "test-key",
		BaseURL: "https://api.openai.com",
		Model:   "gpt-4o",
	}
	if cfg.APIKey != "test-key" {
		t.Errorf("ProviderConfig.APIKey = %s, want test-key", cfg.APIKey)
	}
}

func TestLLMTool(t *testing.T) {
	tool := LLMTool{
		Name:        "search",
		Description: "Search documents",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"q": map[string]interface{}{"type": "string"},
			},
		},
	}
	data, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("marshal LLMTool: %v", err)
	}
	var parsed LLMTool
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal LLMTool: %v", err)
	}
	if parsed.Name != "search" {
		t.Errorf("parsed LLMTool.Name = %s, want search", parsed.Name)
	}
}
