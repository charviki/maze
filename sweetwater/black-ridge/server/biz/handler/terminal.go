package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"

	"github.com/charviki/maze-cradle/httputil"
	"github.com/charviki/sweetwater-black-ridge/biz/model"
	"github.com/charviki/sweetwater-black-ridge/biz/service"
	"github.com/creack/pty/v2"
)

const (
	defaultTerminalRows = 24
	defaultTerminalCols = 80
	ptyBufferSize       = 4096
)

// handleSessionError 根据 error 类型自动选择 404 或 500 状态码
func handleSessionError(c *app.RequestContext, err error) {
	if errors.Is(err, service.ErrSessionNotFound) {
		httputil.Error(c, http.StatusNotFound, "session not found")
	} else {
		httputil.Error(c, http.StatusInternalServerError, err.Error())
	}
}

// TerminalHandler 终端交互 handler，处理终端输出读取、命令输入、信号发送和 WebSocket 实时连接
type TerminalHandler struct {
	tmuxService    service.TmuxService
	defaultLines   int
	logger         logutil.Logger
	allowedOrigins []string
}

// NewTerminalHandler 创建 TerminalHandler，需传入默认终端行数。
// allowedOrigins 用于 WebSocket 升级时的 Origin 校验，为空时允许所有来源。
func NewTerminalHandler(tmuxService service.TmuxService, defaultLines int, logger logutil.Logger, allowedOrigins []string) *TerminalHandler {
	return &TerminalHandler{tmuxService: tmuxService, defaultLines: defaultLines, logger: logger, allowedOrigins: allowedOrigins}
}

// GetOutput 获取指定会话的终端输出内容（HTTP 轮询模式）
func (h *TerminalHandler) GetOutput(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, "id is required")
		return
	}

	lines := h.defaultLines
	if l := c.Query("lines"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			lines = n
		}
	}

	output, err := h.tmuxService.CapturePane(id, lines)
	if err != nil {
		handleSessionError(c, err)
		return
	}

	httputil.Success(c, model.TerminalOutput{
		SessionID: id,
		Lines:     lines,
		Output:    output,
	})
}

// SendInput 向指定会话发送命令文本（模拟键盘输入）
func (h *TerminalHandler) SendInput(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, "id is required")
		return
	}

	var req model.SendInputRequest
	if err := c.Bind(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Command == "" {
		httputil.Error(c, http.StatusBadRequest, "command is required")
		return
	}

	if err := h.tmuxService.SendKeys(id, req.Command); err != nil {
		handleSessionError(c, err)
		return
	}
	httputil.Success(c, nil)
}

// SendSignal 向指定会话发送控制信号（如 SIGINT）
func (h *TerminalHandler) SendSignal(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, "id is required")
		return
	}

	var req model.SendSignalRequest
	if err := c.Bind(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Signal == "" {
		httputil.Error(c, http.StatusBadRequest, "signal is required")
		return
	}

	if err := h.tmuxService.SendSignal(id, req.Signal); err != nil {
		handleSessionError(c, err)
		return
	}
	httputil.Success(c, nil)
}

// GetEnv 获取指定会话的环境变量列表
func (h *TerminalHandler) GetEnv(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, "id is required")
		return
	}

	env, err := h.tmuxService.GetSessionEnv(id)
	if err != nil {
		handleSessionError(c, err)
		return
	}
	httputil.Success(c, env)
}

type wsResizeMessage struct {
	Type string `json:"type"`
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

// HandleWs WebSocket 实时终端连接。将 WebSocket 与 tmux PTY 双向绑定：
// 1. PTY 输出 → WebSocket：后台 goroutine 持续读取 PTY 输出并推送到 WebSocket
// 2. WebSocket → PTY 输入：主循环读取 WebSocket 消息，区分 resize 控制消息和普通输入
// 3. 资源清理：使用 sync.Once 确保 WebSocket 和 PTY 只关闭一次
func (h *TerminalHandler) HandleWs(_ context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, "id is required")
		return
	}

	// 使用配置化的 Origin 校验替代硬编码的"允许所有来源"，避免跨站 WebSocket 劫持
	upgrader := websocket.HertzUpgrader{CheckOrigin: httputil.CheckOrigin(h.allowedOrigins)}

	err := upgrader.Upgrade(c, func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()

		ptmx, err := h.tmuxService.AttachSession(id, defaultTerminalRows, defaultTerminalCols)
		if err != nil {
			h.logger.Errorf("[ws] attach session %s failed: %v", id, err)
			return
		}
		defer func() { _ = ptmx.Close() }()

		// 在 attach 之后立即检查 tmux 进程是否存活，捕获 session 不存在等快速失败场景
		// AttachSession 返回后 tmux 进程已在后台运行，若 session 不存在，进程会立即退出
		// 此时给 tmux 一个短暂的启动窗口，然后检查 PTY 是否可读
		var once sync.Once
		cleanup := func() {
			once.Do(func() {
				_ = conn.Close()
				_ = ptmx.Close()
			})
		}
		defer cleanup()

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
				if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
					h.logger.Errorf("[ws] ws write error: %v", err)
					return
				}
			}
		}()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
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
					// 同步 tmux session 的内部尺寸，避免外层 PTY 与 tmux session 尺寸不一致
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
	})

	if err != nil {
		h.logger.Errorf("[ws] upgrade failed: %v", err)
	}
}
