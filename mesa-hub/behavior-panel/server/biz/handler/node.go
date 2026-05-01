package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/charviki/maze-cradle/protocol"

	"github.com/charviki/mesa-hub-behavior-panel/biz/model"
	"github.com/charviki/maze-cradle/httputil"
)

// NodeHandler 节点管理 handler，处理 Agent 注册、心跳和查询。
// Host 令牌验证采用分层策略：Manager 创建的 Host 使用预存令牌校验，
// 未知 Host 回退到全局 auth token 校验，开发模式下放行所有请求。
type NodeHandler struct {
	registry        *model.NodeRegistry
	globalAuthToken string
}

// NewNodeHandler 创建 NodeHandler，globalAuthToken 用于未知 Host 的回退校验
func NewNodeHandler(registry *model.NodeRegistry, globalAuthToken string) *NodeHandler {
	return &NodeHandler{
		registry:        registry,
		globalAuthToken: globalAuthToken,
	}
}

// validateHostToken 从请求头提取 Bearer token 并与 Host 预存令牌或全局令牌校验。
// 已知 Host（有预存令牌）必须精确匹配；未知 Host 使用全局 auth token 校验；
// 全局 auth token 为空时视为开发模式，放行所有请求。
func (h *NodeHandler) validateHostToken(c *app.RequestContext, name string) bool {
	// 开发模式：全局 auth token 为空时放行所有请求
	if h.globalAuthToken == "" {
		return true
	}

	// 从 Authorization 头提取 Bearer token
	auth := string(c.GetHeader("Authorization"))
	token := strings.TrimPrefix(auth, "Bearer ")

	// 检查是否为 Manager 创建的 Host（有预存令牌）
	exists, matched := h.registry.ValidateHostToken(name, token)
	if exists {
		// 已知 Host：令牌必须精确匹配
		if !matched {
			httputil.Error(c, http.StatusUnauthorized, "unauthorized: invalid host token")
			return false
		}
		return true
	}

	// 未知 Host：使用全局 auth token 校验，兼容未通过 CreateHost 创建的遗留 Agent
	if token != h.globalAuthToken {
		httputil.Error(c, http.StatusUnauthorized, "unauthorized: invalid authorization header")
		return false
	}
	return true
}

// Register 处理节点注册请求（携带 capabilities、status、metadata）。
// 在执行注册逻辑之前先验证 Host 令牌，确保只有经过 Manager 授权的 Agent 才能注册。
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

	// 验证 Host 令牌：已知 Host 校验预存令牌，未知 Host 校验全局 auth token
	if !h.validateHostToken(c, req.Name) {
		return
	}

	node := h.registry.Register(req)
	httputil.Success(c, node)
}

// Heartbeat 处理心跳上报请求（携带完整状态快照）。
// 在执行心跳逻辑之前先验证 Host 令牌，防止未授权节点伪造心跳。
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

	// 验证 Host 令牌：防止未授权节点伪造心跳
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

// ListNodes 列出所有注册节点
func (h *NodeHandler) ListNodes(ctx context.Context, c *app.RequestContext) {
	nodes := h.registry.List()
	httputil.Success(c, nodes)
}

// GetNode 获取指定节点详情
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
