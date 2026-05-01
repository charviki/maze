package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/charviki/maze-cradle/httputil"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"

	"github.com/charviki/mesa-hub-behavior-panel/biz/model"
)

type NodeHandler struct {
	registry        *model.NodeRegistry
	globalAuthToken string
	logger          logutil.Logger
}

func NewNodeHandler(registry *model.NodeRegistry, globalAuthToken string, logger logutil.Logger) *NodeHandler {
	return &NodeHandler{
		registry:        registry,
		globalAuthToken: globalAuthToken,
		logger:          logger,
	}
}

func (h *NodeHandler) validateHostToken(c *app.RequestContext, name string) bool {
	if h.globalAuthToken == "" {
		return true
	}

	auth := string(c.GetHeader("Authorization"))
	token := strings.TrimPrefix(auth, "Bearer ")

	exists, matched := h.registry.ValidateHostToken(name, token)
	if exists {
		if !matched {
			httputil.Error(c, http.StatusUnauthorized, "unauthorized: invalid host token")
			return false
		}
		return true
	}

	if token != h.globalAuthToken {
		httputil.Error(c, http.StatusUnauthorized, "unauthorized: invalid authorization header")
		return false
	}
	return true
}

func (h *NodeHandler) Register(ctx context.Context, c *app.RequestContext) {
	var req protocol.RegisterRequest
	if err := c.Bind(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		httputil.Error(c, http.StatusBadRequest, "name is required")
		return
	}
	if req.Address == "" {
		httputil.Error(c, http.StatusBadRequest, "address is required")
		return
	}

	if !h.validateHostToken(c, req.Name) {
		return
	}

	node := h.registry.Register(req)
	httputil.Success(c, node)

	go h.restoreAgentSessions(req.Name, req.Address)
}

func (h *NodeHandler) Heartbeat(ctx context.Context, c *app.RequestContext) {
	var req protocol.HeartbeatRequest
	if err := c.Bind(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		httputil.Error(c, http.StatusBadRequest, "name is required")
		return
	}

	if !h.validateHostToken(c, req.Name) {
		return
	}

	node := h.registry.Heartbeat(req)
	if node == nil {
		httputil.Error(c, http.StatusNotFound, "node not found")
		return
	}
	httputil.Success(c, node)
}

func (h *NodeHandler) ListNodes(ctx context.Context, c *app.RequestContext) {
	nodes := h.registry.List()
	httputil.Success(c, nodes)
}

func (h *NodeHandler) GetNode(ctx context.Context, c *app.RequestContext) {
	name := c.Param("name")
	if name == "" {
		httputil.Error(c, http.StatusBadRequest, "name is required")
		return
	}

	node := h.registry.Get(name)
	if node == nil {
		httputil.Error(c, http.StatusNotFound, "node not found")
		return
	}
	httputil.Success(c, node)
}

func (h *NodeHandler) DeleteNode(ctx context.Context, c *app.RequestContext) {
	name := c.Param("name")
	if name == "" {
		httputil.Error(c, http.StatusBadRequest, "name is required")
		return
	}

	if !h.registry.Delete(name) {
		httputil.Error(c, http.StatusNotFound, "node not found")
		return
	}
	httputil.Success(c, nil)
}

// restoreAgentSessions 在 Agent 注册后异步恢复已保存的 session。
// Agent 启动时不会自动恢复 session，需要 Manager 在确认 Agent 可达后触发恢复。
func (h *NodeHandler) restoreAgentSessions(name, address string) {
	client := &http.Client{Timeout: 10 * time.Second}

	savedURL := fmt.Sprintf("%s/api/v1/sessions/saved", strings.TrimRight(address, "/"))
	resp, err := client.Get(savedURL)
	if err != nil {
		h.logger.Warnf("[session-restore] query saved sessions from %s failed: %v", name, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Warnf("[session-restore] read response from %s failed: %v", name, err)
		return
	}

	var result struct {
		Status string `json:"status"`
		Data   []struct {
			SessionName     string `json:"session_name"`
			RestoreStrategy string `json:"restore_strategy"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		h.logger.Warnf("[session-restore] parse response from %s failed: %v", name, err)
		return
	}

	if len(result.Data) == 0 {
		h.logger.Infof("[session-restore] no saved sessions for %s", name)
		return
	}

	restored := 0
	for _, s := range result.Data {
		if s.RestoreStrategy == "running" {
			continue
		}

		restoreURL := fmt.Sprintf("%s/api/v1/sessions/%s/restore",
			strings.TrimRight(address, "/"),
			url.PathEscape(s.SessionName),
		)

		resp, err := client.Post(restoreURL, "application/json", nil)
		if err != nil {
			h.logger.Warnf("[session-restore] restore session %s/%s failed: %v", name, s.SessionName, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			restored++
			h.logger.Infof("[session-restore] restored session %s/%s", name, s.SessionName)
		} else {
			h.logger.Warnf("[session-restore] restore session %s/%s returned %d", name, s.SessionName, resp.StatusCode)
		}
	}

	if restored > 0 {
		h.logger.Infof("[session-restore] restored %d sessions for %s", restored, name)
	}
}
