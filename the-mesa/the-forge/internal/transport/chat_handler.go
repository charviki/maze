package transport

import (
	"encoding/json"
	"net/http"

	"github.com/charviki/maze/fabrication/cradle/llmutil"
	"github.com/charviki/maze/fabrication/cradle/logutil"
	"github.com/charviki/maze/the-mesa/the-forge/internal/service"
)

// ChatHandler 处理 Oracle Chat SSE 和历史管理。
type ChatHandler struct {
	chatSvc *service.ChatService
	logger  logutil.Logger
}

// NewChatHandler 创建 ChatHandler。
func NewChatHandler(chatSvc *service.ChatService, logger logutil.Logger) *ChatHandler {
	return &ChatHandler{chatSvc: chatSvc, logger: logger}
}

// RegisterRoutes 注册 Chat 相关路由到 mux。
func (h *ChatHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/oracle/chat", h.handleChat)
	mux.HandleFunc("GET /api/v1/oracle/history", h.handleGetHistory)
	mux.HandleFunc("DELETE /api/v1/oracle/history", h.handleClearHistory)
}

func (h *ChatHandler) handleChat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}

	llmutil.SetSSEHeaders(w)

	events := make(chan llmutil.SSEEvent, 64)
	done := make(chan struct{})

	// 消费 events 并写入 SSE 流
	go func() {
		defer close(done)
		flusher, canFlush := w.(http.Flusher)
		for {
			select {
			case ev, ok := <-events:
				if !ok {
					return
				}
				_ = llmutil.WriteSSEEvent(w, ev)
				if canFlush {
					flusher.Flush()
				}
			case <-r.Context().Done():
				return
			}
		}
	}()

	if err := h.chatSvc.Chat(r.Context(), req.Message, events); err != nil {
		if h.logger != nil {
			h.logger.Infof("[chat] Chat() error: %v", err)
		}
	}
	close(events)
	<-done
}

func (h *ChatHandler) handleGetHistory(w http.ResponseWriter, r *http.Request) {
	history, err := h.chatSvc.GetHistory(r.Context())
	if err != nil {
		http.Error(w, "failed to get history", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(history)
}

func (h *ChatHandler) handleClearHistory(w http.ResponseWriter, r *http.Request) {
	if err := h.chatSvc.ClearHistory(r.Context()); err != nil {
		http.Error(w, "failed to clear history", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
