package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/charviki/maze-cradle/httputil"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/biz/builder"
	"github.com/charviki/mesa-hub-behavior-panel/biz/config"
	"github.com/charviki/mesa-hub-behavior-panel/biz/model"
	"github.com/charviki/mesa-hub-behavior-panel/biz/runtime"
)

type HostHandler struct {
	registry *model.NodeRegistry
	specMgr  *model.HostSpecManager
	runtime  runtime.HostRuntime
	auditLog *AuditLogger
	cfg      *config.Config
	logger   logutil.Logger
	logDir   string
}

func NewHostHandler(
	registry *model.NodeRegistry,
	specMgr *model.HostSpecManager,
	rt runtime.HostRuntime,
	auditLog *AuditLogger,
	cfg *config.Config,
	logger logutil.Logger,
	logDir string,
) *HostHandler {
	return &HostHandler{
		registry: registry,
		specMgr:  specMgr,
		runtime:  rt,
		auditLog: auditLog,
		cfg:      cfg,
		logger:   logger,
		logDir:   logDir,
	}
}

// CreateHost 异步创建 Host：验证 → 持久化 HostSpec → 返回 202 → 后台构建部署
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

	// 名称唯一性校验（优先检查 HostSpec，再检查 NodeRegistry）
	if h.specMgr.Get(req.Name) != nil {
		httputil.Error(c, http.StatusConflict, fmt.Sprintf("host %q already exists", req.Name))
		return
	}
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

	now := time.Now()
	hostToken := req.Name

	spec := &protocol.HostSpec{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Tools:       req.Tools,
		Resources:   req.Resources,
		AuthToken:   hostToken,
		CreatedAt:   now,
		UpdatedAt:   now,
		Status:      protocol.HostStatusPending,
	}

	if !h.specMgr.Create(spec) {
		httputil.Error(c, http.StatusConflict, fmt.Sprintf("host %q already exists", req.Name))
		return
	}

	// 预存令牌供注册/心跳时校验
	h.registry.StoreHostToken(req.Name, hostToken)

	// 启动后台 goroutine 异步构建部署
	go h.deployHostAsync(spec)

	httputil.Success(c, spec)
	c.SetStatusCode(http.StatusAccepted)
}

// deployHostAsync 后台异步构建部署 Host
func (h *HostHandler) deployHostAsync(spec *protocol.HostSpec) {
	// 更新状态为 deploying
	h.specMgr.UpdateStatus(spec.Name, protocol.HostStatusDeploying, "")

	// 清理可能残留的旧容器（重试场景下容器名冲突），保留持久化数据
	if err := h.runtime.StopHost(context.Background(), spec.Name); err != nil {
		h.logger.Warnf("[host] stop old container for %s: %v", spec.Name, err)
	}

	// 构建日志文件
	if err := os.MkdirAll(h.logDir, 0755); err != nil {
		h.specMgr.UpdateStatus(spec.Name, protocol.HostStatusFailed, fmt.Sprintf("create log dir: %v", err))
		return
	}
	logPath := filepath.Join(h.logDir, spec.Name+".log")
	logFile, err := os.Create(logPath)
	if err != nil {
		h.specMgr.UpdateStatus(spec.Name, protocol.HostStatusFailed, fmt.Sprintf("create log file: %v", err))
		return
	}
	defer logFile.Close()

	// 生成 Dockerfile
	dockerfileContent, err := builder.GenerateHostDockerfile(spec.Tools, h.cfg.Docker.AgentBaseImage)
	if err != nil {
		errMsg := fmt.Sprintf("generate dockerfile: %v", err)
		h.specMgr.UpdateStatus(spec.Name, protocol.HostStatusFailed, errMsg)
		fmt.Fprintf(logFile, "[ERROR] %s\n", errMsg)
		return
	}

	// 构建部署规格
	deploySpec := &protocol.HostDeploySpec{
		Name:            spec.Name,
		Tools:           spec.Tools,
		Resources:       spec.Resources,
		AuthToken:       spec.AuthToken,
		ServerAuthToken: h.cfg.Server.AuthToken,
	}

	// 部署（构建日志通过 io.MultiWriter 同时写文件和 logger）
	multiWriter := io.MultiWriter(logFile, writerFunc(func(p []byte) (int, error) {
		h.logger.Infof("[host-deploy] %s", string(p))
		return len(p), nil
	}))

	// 通过修改 runtime 的输出来捕获构建日志比较复杂，
	// 这里直接用多 writer 记录状态变化
	fmt.Fprintf(multiWriter, "[INFO] Host %s: starting deployment, tools=%v\n", spec.Name, spec.Tools)

	_, deployErr := h.runtime.DeployHost(context.Background(), deploySpec, dockerfileContent)

	if deployErr != nil {
		errMsg := fmt.Sprintf("deploy failed: %v", deployErr)
		h.specMgr.UpdateStatus(spec.Name, protocol.HostStatusFailed, errMsg)
		fmt.Fprintf(multiWriter, "[ERROR] %s\n", errMsg)

		h.auditLog.Log(protocol.AuditLogEntry{
			Operator:       "system",
			Action:         "create_host",
			TargetNode:     spec.Name,
			PayloadSummary: fmt.Sprintf("tools=%v", spec.Tools),
			Result:         "failed",
			StatusCode:     http.StatusInternalServerError,
		})
		return
	}

	// 部署成功，保持 deploying 状态等待 Agent 注册
	fmt.Fprintf(multiWriter, "[INFO] Host %s: deployment complete, waiting for agent registration\n", spec.Name)

	h.auditLog.Log(protocol.AuditLogEntry{
		Operator:       "system",
		Action:         "create_host",
		TargetNode:     spec.Name,
		PayloadSummary: fmt.Sprintf("tools=%v", spec.Tools),
		Result:         "success",
		StatusCode:     http.StatusAccepted,
	})
}

// ListHosts 返回所有 Host（含合并状态）
func (h *HostHandler) ListHosts(ctx context.Context, c *app.RequestContext) {
	hosts := h.specMgr.ListMerged(h.registry)
	httputil.Success(c, hosts)
}

// GetHost 返回单个 Host 信息
func (h *HostHandler) GetHost(ctx context.Context, c *app.RequestContext) {
	name := c.Param("name")
	if name == "" {
		httputil.Error(c, http.StatusBadRequest, "name is required")
		return
	}

	spec := h.specMgr.Get(name)
	if spec == nil {
		httputil.Error(c, http.StatusNotFound, fmt.Sprintf("host %q not found", name))
		return
	}

	info := &protocol.HostInfo{HostSpec: *spec}
	// 如果 deploying 状态，合并 NodeRegistry 心跳
	if spec.Status == protocol.HostStatusDeploying {
		node := h.registry.Get(name)
		if node != nil {
			if node.Status == model.NodeStatusOnline {
				info.Status = protocol.HostStatusOnline
			} else {
				info.Status = protocol.HostStatusOffline
			}
			info.Address = node.Address
			info.SessionCount = node.AgentStatus.ActiveSessions
			if !node.LastHeartbeat.IsZero() {
				info.LastHeartbeat = node.LastHeartbeat.Format(time.RFC3339)
			}
		}
	}
	httputil.Success(c, info)
}

// ListTools 返回可用工具列表
func (h *HostHandler) ListTools(ctx context.Context, c *app.RequestContext) {
	tools := builder.ListAvailableTools()
	httputil.Success(c, tools)
}

// DeleteHost 销毁 Host：删除 HostSpec + 清理令牌 + 停止容器 → 审计日志
func (h *HostHandler) DeleteHost(ctx context.Context, c *app.RequestContext) {
	name := c.Param("name")
	if name == "" {
		httputil.Error(c, http.StatusBadRequest, "name is required")
		return
	}

	// 停止并移除容器
	if err := h.runtime.RemoveHost(ctx, name); err != nil {
		h.logger.Warnf("[host] remove host %q failed: %v", name, err)
	}

	// 从节点注册表删除
	h.registry.Delete(name)

	// 清除 Host 预存令牌
	h.registry.RemoveHostToken(name)

	// 删除 HostSpec
	h.specMgr.Delete(name)

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

// GetBuildLog 返回构建日志内容
func (h *HostHandler) GetBuildLog(ctx context.Context, c *app.RequestContext) {
	name := c.Param("name")
	if name == "" {
		httputil.Error(c, http.StatusBadRequest, "name is required")
		return
	}

	if h.specMgr.Get(name) == nil {
		httputil.Error(c, http.StatusNotFound, fmt.Sprintf("host %q not found", name))
		return
	}

	logPath := filepath.Join(h.logDir, name+".log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		// 日志文件不存在说明还没开始构建或构建日志丢失
		httputil.Success(c, "")
		return
	}
	httputil.Success(c, string(data))
}

// GetRuntimeLog 返回运行时日志
func (h *HostHandler) GetRuntimeLog(ctx context.Context, c *app.RequestContext) {
	name := c.Param("name")
	if name == "" {
		httputil.Error(c, http.StatusBadRequest, "name is required")
		return
	}

	if h.specMgr.Get(name) == nil {
		httputil.Error(c, http.StatusNotFound, fmt.Sprintf("host %q not found", name))
		return
	}

	logs, err := h.runtime.GetRuntimeLogs(ctx, name, 500)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, fmt.Sprintf("get runtime logs: %v", err))
		return
	}
	httputil.Success(c, logs)
}

// writerFunc 将函数适配为 io.Writer
type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) {
	return f(p)
}
