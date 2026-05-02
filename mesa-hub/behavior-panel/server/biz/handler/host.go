package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/charviki/maze-cradle/httputil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/biz/service"
)

type HostHandler struct {
	svc *service.HostService
}

func NewHostHandler(svc *service.HostService) *HostHandler {
	return &HostHandler{svc: svc}
}

// CreateHost 异步创建 Host：校验 → 持久化 HostSpec → 返回 202 → 后台构建部署
func (h *HostHandler) CreateHost(ctx context.Context, c *app.RequestContext) {
	var req protocol.CreateHostRequest
	if err := c.Bind(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	spec, err := h.svc.CreateHost(ctx, &req)
	if err != nil {
		httputil.Error(c, hostErrorCode(err), err.Error())
		return
	}

	httputil.Success(c, spec)
	c.SetStatusCode(http.StatusAccepted)
}

// ListHosts 返回所有 Host（含合并状态）
func (h *HostHandler) ListHosts(ctx context.Context, c *app.RequestContext) {
	hosts, err := h.svc.ListHosts(ctx)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.Success(c, hosts)
}

// GetHost 返回单个 Host 信息
func (h *HostHandler) GetHost(ctx context.Context, c *app.RequestContext) {
	name := c.Param("name")
	info, err := h.svc.GetHost(ctx, name)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err.Error())
		return
	}
	httputil.Success(c, info)
}

// ListTools 返回可用工具列表
func (h *HostHandler) ListTools(ctx context.Context, c *app.RequestContext) {
	tools, err := h.svc.ListTools(ctx)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.Success(c, tools)
}

// DeleteHost 销毁 Host：删除 HostSpec + 清理令牌 + 停止容器 → 审计日志
func (h *HostHandler) DeleteHost(ctx context.Context, c *app.RequestContext) {
	name := c.Param("name")
	if err := h.svc.DeleteHost(ctx, name); err != nil {
		httputil.Error(c, hostErrorCode(err), err.Error())
		return
	}
	httputil.Success(c, nil)
}

// hostErrorCode 根据服务层错误消息推断 HTTP 状态码
func hostErrorCode(err error) int {
	msg := err.Error()
	if strings.Contains(msg, "already exists") {
		return http.StatusConflict
	}
	if strings.Contains(msg, "required") || strings.Contains(msg, "unknown") {
		return http.StatusBadRequest
	}
	if strings.Contains(msg, "not found") {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

// GetBuildLog 返回构建日志内容
func (h *HostHandler) GetBuildLog(ctx context.Context, c *app.RequestContext) {
	name := c.Param("name")
	logContent, err := h.svc.GetBuildLog(ctx, name)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err.Error())
		return
	}
	if logContent == "" {
		logContent = fmt.Sprintf("(no build log for %s)", name)
	}
	httputil.Success(c, logContent)
}

// GetRuntimeLog 返回运行时日志
func (h *HostHandler) GetRuntimeLog(ctx context.Context, c *app.RequestContext) {
	name := c.Param("name")
	logs, err := h.svc.GetRuntimeLog(ctx, name)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, err.Error())
		return
	}
	httputil.Success(c, logs)
}
