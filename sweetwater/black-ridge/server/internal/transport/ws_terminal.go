package transport

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"sync"

	gorillaws "github.com/gorilla/websocket"

	"github.com/charviki/maze-cradle/httputil"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/internal/service"
	"github.com/creack/pty/v2"
)

const (
	defaultTerminalRows = 24
	defaultTerminalCols = 80
	ptyBufferSize       = 4096
)

// TerminalHandler 终端 WebSocket handler，处理实时终端连接
type TerminalHandler struct {
	tmuxService    service.TmuxService
	defaultLines   int
	logger         logutil.Logger
	allowedOrigins []string
}

// NewTerminalHandler 创建 TerminalHandler，allowedOrigins 用于 WebSocket Origin 校验。
func NewTerminalHandler(tmuxService service.TmuxService, defaultLines int, logger logutil.Logger, allowedOrigins []string) *TerminalHandler {
	return &TerminalHandler{tmuxService: tmuxService, defaultLines: defaultLines, logger: logger, allowedOrigins: allowedOrigins}
}

type wsResizeMessage struct {
	Type string `json:"type"`
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

// HandleWs 处理 WebSocket 终端实时连接。
// 流程：WebSocket 升级 → AttachSession → 双向数据转发（PTY↔WS）+ resize 处理。
// 资源清理使用 sync.Once 确保 WebSocket 和 PTY 只关闭一次。
func (h *TerminalHandler) HandleWs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httputil.Error(w, r, http.StatusBadRequest, "id is required")
		return
	}

	conn, err := httputil.NewUpgrader(h.allowedOrigins).Upgrade(w, r, nil)
	if err != nil {
		h.logger.Errorf("[ws] upgrade failed: %v", err)
		return
	}
	defer func() { _ = conn.Close() }()

	ptmx, err := h.tmuxService.AttachSession(id, defaultTerminalRows, defaultTerminalCols)
	if err != nil {
		h.logger.Errorf("[ws] attach session %s failed: %v", id, err)
		return
	}
	defer func() { _ = ptmx.Close() }()

	// attach 后 tmux 进程在后台运行，若 session 不存在进程会立即退出
	// 使用 sync.Once 确保 WebSocket 和 PTY 只关闭一次。
	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			_ = conn.Close()
			_ = ptmx.Close()
		})
	}
	defer cleanup()

	// PTY → WebSocket：读取终端输出并发送到前端。
	go func() {
		buf := make([]byte, ptyBufferSize)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				if err != io.EOF && !errors.Is(err, os.ErrClosed) {
					h.logger.Errorf("[ws] pty read error: %v", err)
				} else {
					h.logger.Infof("[ws] session %s pty closed (session may have exited or does not exist)", id)
				}
				return
			}
			if err := conn.WriteMessage(gorillaws.BinaryMessage, buf[:n]); err != nil {
				h.logger.Errorf("[ws] ws write error: %v", err)
				return
			}
		}
	}()

	// WebSocket → PTY：读取前端输入和 resize 指令。
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if !gorillaws.IsCloseError(err, gorillaws.CloseNormalClosure, gorillaws.CloseGoingAway) {
				h.logger.Errorf("[ws] ws read error: %v", err)
			}
			return
		}

		var resize wsResizeMessage
		if json.Unmarshal(msg, &resize) == nil && resize.Type == "resize" {
			if resize.Cols > 0 && resize.Rows > 0 {
				if err := pty.Setsize(ptmx, &pty.Winsize{
					Rows: resize.Rows,
					Cols: resize.Cols,
				}); err != nil {
					h.logger.Errorf("[ws] pty resize for session %s failed: %v", id, err)
				}
				if err := h.tmuxService.ResizeSession(id, resize.Rows, resize.Cols); err != nil {
					h.logger.Warnf("[ws] tmux resize for session %s failed: %v", id, err)
				}
			}
			continue
		}

		if _, err := ptmx.Write(msg); err != nil {
			h.logger.Errorf("[ws] pty write error: %v", err)
			return
		}
	}
}
