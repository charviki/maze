package llmutil

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// WriteSSEEvent 将一个 SSEEvent 写入 HTTP ResponseWriter，格式为 text/event-stream。
func WriteSSEEvent(w http.ResponseWriter, event SSEEvent) error {
	data, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("marshal SSE event data: %w", err)
	}
	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
	if err != nil {
		return err
	}
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}

// WriteSSEString 将一个简单字符串 SSE 事件写入 HTTP ResponseWriter。
func WriteSSEString(w http.ResponseWriter, eventType string, data string) error {
	return WriteSSEEvent(w, SSEEvent{Type: eventType, Data: data})
}

// SetSSEHeaders 设置 HTTP 响应头为 text/event-stream。
func SetSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}
