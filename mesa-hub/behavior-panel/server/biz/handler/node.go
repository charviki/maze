package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/charviki/maze-cradle/protocol"

	"github.com/charviki/mesa-hub-behavior-panel/biz/model"
	"github.com/charviki/maze-cradle/httputil"
)

// NodeHandler 节点管理 handler，处理 Agent 注册、心跳和查询
type NodeHandler struct {
	registry *model.NodeRegistry
}

// NewNodeHandler 创建 NodeHandler
func NewNodeHandler(registry *model.NodeRegistry) *NodeHandler {
	return &NodeHandler{registry: registry}
}

// Register 处理节点注册请求（携带 capabilities、status、metadata）
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

	node := h.registry.Register(req)
	httputil.Success(c, node)
}

// Heartbeat 处理心跳上报请求（携带完整状态快照）
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
