package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/charviki/maze-cradle/httputil"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/biz/builder"
	"github.com/charviki/mesa-hub-behavior-panel/biz/config"
	"github.com/charviki/mesa-hub-behavior-panel/biz/model"
	"github.com/charviki/mesa-hub-behavior-panel/biz/runtime"
)

// HostHandler Host 生命周期管理 handler
type HostHandler struct {
	registry *model.NodeRegistry
	runtime  runtime.HostRuntime
	auditLog *AuditLogger
	cfg      *config.Config
	logger   logutil.Logger
}

// NewHostHandler 创建 HostHandler
func NewHostHandler(
	registry *model.NodeRegistry,
	rt runtime.HostRuntime,
	auditLog *AuditLogger,
	cfg *config.Config,
	logger logutil.Logger,
) *HostHandler {
	return &HostHandler{
		registry: registry,
		runtime:  rt,
		auditLog: auditLog,
		cfg:      cfg,
		logger:   logger,
	}
}

// CreateHost 创建 Host：验证 → 生成 Dockerfile → 构建部署规格 → 交给运行时部署 → 审计日志
func (h *HostHandler) CreateHost(ctx context.Context, c *app.RequestContext) {
	var req protocol.CreateHostRequest
	if err := c.Bind(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		httputil.Error(c, http.StatusBadRequest, "name is required")
		return
	}
	if len(req.Tools) == 0 {
		httputil.Error(c, http.StatusBadRequest, "at least one tool is required")
		return
	}

	// 验证名称唯一性
	if existing := h.registry.Get(req.Name); existing != nil {
		httputil.Error(c, http.StatusConflict, fmt.Sprintf("host %q already exists", req.Name))
		return
	}

	// 验证工具列表
	if unknown := builder.ValidateTools(req.Tools); len(unknown) > 0 {
		httputil.Error(c, http.StatusBadRequest,
			fmt.Sprintf("unknown tools: %s. available: %s",
				strings.Join(unknown, ", "),
				strings.Join(func() []string {
					tools := builder.ListAvailableTools()
					ids := make([]string, len(tools))
					for i, t := range tools {
						ids[i] = t.ID
					}
					return ids
				}(), ", ")))
		return
	}

	// 生成 Dockerfile（基于 agent 基础镜像 + 选配工具链）
	dockerfileContent, err := builder.GenerateHostDockerfile(req.Tools, h.cfg.Docker.AgentBaseImage)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, fmt.Sprintf("generate dockerfile: %v", err))
		return
	}

	// 构建运行时无关的部署规格
	spec := &protocol.HostDeploySpec{
		Name:      req.Name,
		Tools:     req.Tools,
		Resources: req.Resources,
		AuthToken: h.cfg.Server.AuthToken,
	}

	// 交给运行时实现部署（构建镜像 + 启动容器）
	resp, err := h.runtime.DeployHost(ctx, spec, dockerfileContent)
	if err != nil {
		h.auditLog.Log(protocol.AuditLogEntry{
			Operator:       "frontend",
			Action:         "create_host",
			TargetNode:     req.Name,
			PayloadSummary: fmt.Sprintf("tools=%v", req.Tools),
			Result:         "failed",
			StatusCode:     http.StatusInternalServerError,
		})
		httputil.Error(c, http.StatusInternalServerError, fmt.Sprintf("deploy host failed: %v", err))
		return
	}

	// 记录审计日志
	h.auditLog.Log(protocol.AuditLogEntry{
		Operator:       "frontend",
		Action:         "create_host",
		TargetNode:     req.Name,
		PayloadSummary: fmt.Sprintf("tools=%v, image=%s, container=%s", req.Tools, resp.ImageTag, resp.ContainerID),
		Result:         "success",
		StatusCode:     http.StatusOK,
	})

	h.logger.Infof("[host] created host %q: image=%s, container=%s", req.Name, resp.ImageTag, resp.ContainerID)

	httputil.Success(c, resp)
}

// ListTools 返回可用工具列表
func (h *HostHandler) ListTools(ctx context.Context, c *app.RequestContext) {
	tools := builder.ListAvailableTools()
	httputil.Success(c, tools)
}

// DeleteHost 销毁 Host：停止容器 → 移除容器 → 删除节点 → 审计日志
func (h *HostHandler) DeleteHost(ctx context.Context, c *app.RequestContext) {
	name := c.Param("name")
	if name == "" {
		httputil.Error(c, http.StatusBadRequest, "name is required")
		return
	}

	// 停止并移除容器（忽略"不存在"的错误）
	if err := h.runtime.RemoveHost(ctx, name); err != nil {
		h.logger.Warnf("[host] remove host %q failed: %v", name, err)
	}

	// 从节点注册表删除
	h.registry.Delete(name)

	h.auditLog.Log(protocol.AuditLogEntry{
		Operator:       "frontend",
		Action:         "delete_host",
		TargetNode:     name,
		PayloadSummary: fmt.Sprintf("container=%s", name),
		Result:         "success",
		StatusCode:     http.StatusOK,
	})

	h.logger.Infof("[host] deleted host %q", name)
	httputil.Success(c, nil)
}
