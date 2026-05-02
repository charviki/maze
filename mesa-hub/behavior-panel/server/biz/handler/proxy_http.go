package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/charviki/maze-cradle/httputil"
	"github.com/charviki/maze-cradle/protocol"
)

const (
	proxyTimeout  = 30 * time.Second
	auditOperator = "frontend"
)

// ListSessions 代理到 Agent 端列出所有 Session
func (h *SessionProxyHandler) ListSessions(ctx context.Context, c *app.RequestContext) {
	h.proxyToAgent(c, "list_sessions", http.MethodGet, "/api/v1/sessions", nil)
}

// CreateSession 代理到 Agent 端创建 Session
func (h *SessionProxyHandler) CreateSession(ctx context.Context, c *app.RequestContext) {
	body, _ := c.Body()
	h.proxyToAgent(c, "create_session", http.MethodPost, "/api/v1/sessions", body)
}

// GetSession 代理到 Agent 端获取单个 Session
func (h *SessionProxyHandler) GetSession(ctx context.Context, c *app.RequestContext) {
	sessionID := c.Param("id")
	if sessionID == "" {
		httputil.Error(c, http.StatusBadRequest, "session id is required")
		return
	}
	h.proxyToAgent(c, "get_session", http.MethodGet, "/api/v1/sessions/"+sessionID, nil)
}

// DeleteSession 代理到 Agent 端删除 Session
func (h *SessionProxyHandler) DeleteSession(ctx context.Context, c *app.RequestContext) {
	sessionID := c.Param("id")
	if sessionID == "" {
		httputil.Error(c, http.StatusBadRequest, "session id is required")
		return
	}
	h.proxyToAgent(c, "delete_session", http.MethodDelete, "/api/v1/sessions/"+sessionID, nil)
}

// GetSessionConfig 获取 Session 配置
func (h *SessionProxyHandler) GetSessionConfig(ctx context.Context, c *app.RequestContext) {
	sessionID := c.Param("id")
	if sessionID == "" {
		httputil.Error(c, http.StatusBadRequest, "session id is required")
		return
	}
	h.proxyToAgent(c, "get_session_config", http.MethodGet, "/api/v1/sessions/"+sessionID+"/config", nil)
}

// UpdateSessionConfig 更新 Session 配置
func (h *SessionProxyHandler) UpdateSessionConfig(ctx context.Context, c *app.RequestContext) {
	sessionID := c.Param("id")
	if sessionID == "" {
		httputil.Error(c, http.StatusBadRequest, "session id is required")
		return
	}
	body, _ := c.Body()
	h.proxyToAgent(c, "update_session_config", http.MethodPut, "/api/v1/sessions/"+sessionID+"/config", body)
}

// GetSavedSessions 代理到 Agent 端获取已保存 Session 列表
func (h *SessionProxyHandler) GetSavedSessions(ctx context.Context, c *app.RequestContext) {
	h.proxyToAgent(c, "get_saved_sessions", http.MethodGet, "/api/v1/sessions/saved", nil)
}

// RestoreSession 代理到 Agent 端触发单个 Session 恢复
func (h *SessionProxyHandler) RestoreSession(ctx context.Context, c *app.RequestContext) {
	sessionID := c.Param("id")
	if sessionID == "" {
		httputil.Error(c, http.StatusBadRequest, "session id is required")
		return
	}
	body, _ := c.Body()
	h.proxyToAgent(c, "restore_session", http.MethodPost, "/api/v1/sessions/"+sessionID+"/restore", body)
}

// SaveAllSessions 代理到 Agent 端保存所有 Session
func (h *SessionProxyHandler) SaveAllSessions(ctx context.Context, c *app.RequestContext) {
	h.proxyToAgent(c, "save_all_sessions", http.MethodPost, "/api/v1/sessions/save", nil)
}

// ListTemplates 查询模板列表
func (h *SessionProxyHandler) ListTemplates(ctx context.Context, c *app.RequestContext) {
	h.proxyToAgent(c, "list_templates", http.MethodGet, "/api/v1/templates", nil)
}

// GetTemplate 获取模板详情
func (h *SessionProxyHandler) GetTemplate(ctx context.Context, c *app.RequestContext) {
	h.proxyToAgent(c, "get_template", http.MethodGet, "/api/v1/templates/"+c.Param("id"), nil)
}

// GetTemplateConfig 获取模板配置
func (h *SessionProxyHandler) GetTemplateConfig(ctx context.Context, c *app.RequestContext) {
	h.proxyToAgent(c, "get_template_config", http.MethodGet, "/api/v1/templates/"+c.Param("id")+"/config", nil)
}

// CreateTemplate 创建新模板
func (h *SessionProxyHandler) CreateTemplate(ctx context.Context, c *app.RequestContext) {
	body, _ := c.Body()
	h.proxyToAgent(c, "create_template", http.MethodPost, "/api/v1/templates", body)
}

// UpdateTemplate 更新模板
func (h *SessionProxyHandler) UpdateTemplate(ctx context.Context, c *app.RequestContext) {
	body, _ := c.Body()
	h.proxyToAgent(c, "update_template", http.MethodPut, "/api/v1/templates/"+c.Param("id"), body)
}

// UpdateTemplateConfig 更新模板配置
func (h *SessionProxyHandler) UpdateTemplateConfig(ctx context.Context, c *app.RequestContext) {
	body, _ := c.Body()
	h.proxyToAgent(c, "update_template_config", http.MethodPut, "/api/v1/templates/"+c.Param("id")+"/config", body)
}

// DeleteTemplate 删除模板
func (h *SessionProxyHandler) DeleteTemplate(ctx context.Context, c *app.RequestContext) {
	h.proxyToAgent(c, "delete_template", http.MethodDelete, "/api/v1/templates/"+c.Param("id"), nil)
}

// GetLocalConfig 获取 Agent 本地配置
func (h *SessionProxyHandler) GetLocalConfig(ctx context.Context, c *app.RequestContext) {
	h.proxyToAgent(c, "get_local_config", http.MethodGet, "/api/v1/local-config", nil)
}

// UpdateLocalConfig 更新 Agent 本地配置
func (h *SessionProxyHandler) UpdateLocalConfig(ctx context.Context, c *app.RequestContext) {
	body, _ := c.Body()
	h.proxyToAgent(c, "update_local_config", http.MethodPut, "/api/v1/local-config", body)
}

// proxyToAgent 将请求代理到目标 Agent 并记录审计日志。
// 通用代理流程：查找节点 → 构建请求 → 发送 → 回写响应 → 记录审计日志
func (h *SessionProxyHandler) proxyToAgent(c *app.RequestContext, action, method, agentPath string, body []byte) {
	nodeName := c.Param("name")
	if nodeName == "" {
		httputil.Error(c, http.StatusBadRequest, "node name is required")
		return
	}

	node := h.registry.Get(nodeName)
	if node == nil {
		httputil.Error(c, http.StatusNotFound, "node not found")
		return
	}

	// node.Address 已含 scheme 前缀（如 http://host:port），直接拼接路径
	url := node.Address + agentPath
	if err := validateAgentURL(url, h.allowPrivateNetworks); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid agent url: "+err.Error())
		return
	}

	var req *http.Request
	var err error
	if len(body) > 0 {
		req, err = http.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, "create proxy request: "+err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	// 透传 Manager→Agent 鉴权 token，与 ProxyWebSocket 保持一致
	if h.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.authToken)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		// 记录代理失败审计日志
		h.auditLog.Log(protocol.AuditLogEntry{
			Operator:       auditOperator,
			TargetNode:     nodeName,
			Action:         action,
			PayloadSummary: truncateString(string(body), truncateLength),
			Result:         "error: " + err.Error(),
			StatusCode:     http.StatusBadGateway,
		})
		httputil.Error(c, http.StatusBadGateway, "agent unreachable: "+err.Error())
		return
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, "read agent response: "+err.Error())
		return
	}

	// 记录代理成功审计日志
	result := "success"
	if resp.StatusCode >= 400 {
		result = fmt.Sprintf("error: agent returned %d", resp.StatusCode)
	}
	h.auditLog.Log(protocol.AuditLogEntry{
		Operator:       auditOperator,
		TargetNode:     nodeName,
		Action:         action,
		PayloadSummary: truncateString(string(body), truncateLength),
		Result:         result,
		StatusCode:     resp.StatusCode,
	})

	c.SetContentType("application/json")
	c.SetStatusCode(resp.StatusCode)
	_, _ = c.Write(respBody)
}

// GetAuditLogs 返回审计日志列表（委托 AuditService）。
func (h *SessionProxyHandler) GetAuditLogs(ctx context.Context, c *app.RequestContext) {
	pageStr := c.Query("page")
	pageSizeStr := c.Query("page_size")

	page := parsePageParam(pageStr, pageSizeStr)
	pageSize := parsePageSizeParam(pageStr, pageSizeStr)

	result, err := h.auditSvc.GetAuditLogs(ctx, page, pageSize)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	if page <= 0 {
		httputil.Success(c, result.Logs)
		return
	}

	httputil.Success(c, map[string]interface{}{
		"logs":      result.Logs,
		"total":     result.Total,
		"page":      result.Page,
		"page_size": result.PageSize,
	})
}

// validateAgentURL 校验代理目标 URL 安全性，防止 SSRF 攻击：
// - 必须是 http:// 或 https:// 协议
// - 主机名不能解析为内网 IP（127.0.0.0/8, 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 169.254.0.0/16）
// - allowPrivate 为 true 时跳过内网 IP 检查（适用于 Docker/Kubernetes 等容器网络环境）
func validateAgentURL(rawURL string, allowPrivate bool) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "ws" && u.Scheme != "wss" {
		return fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return errors.New("empty host")
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("resolve host %s: %w", host, err)
	}
	if !allowPrivate {
		for _, ip := range ips {
			if isPrivateIP(ip) {
				return fmt.Errorf("host %s resolves to private IP %s", host, ip)
			}
		}
	}
	return nil
}

// isPrivateIP 判断 IP 是否为内网地址
func isPrivateIP(ip net.IP) bool {
	privateRanges := []struct {
		network *net.IPNet
	}{
		{mustParseCIDR("127.0.0.0/8")},
		{mustParseCIDR("10.0.0.0/8")},
		{mustParseCIDR("172.16.0.0/12")},
		{mustParseCIDR("192.168.0.0/16")},
		{mustParseCIDR("169.254.0.0/16")},
		{mustParseCIDR("::1/128")},
		{mustParseCIDR("fc00::/7")},
	}
	for _, r := range privateRanges {
		if r.network.Contains(ip) {
			return true
		}
	}
	return false
}

func mustParseCIDR(s string) *net.IPNet {
	_, network, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return network
}

// parsePageParam 解析 page 参数，页号从 1 开始。
// 无分页参数时返回 0，表示获取全部日志。
func parsePageParam(pageStr, pageSizeStr string) int {
	if pageStr == "" && pageSizeStr == "" {
		return 0
	}
	page := defaultAuditPage
	if n, err := strconv.Atoi(pageStr); err == nil && n > 0 {
		page = n
	}
	return page
}

// parsePageSizeParam 解析 page_size 参数。
func parsePageSizeParam(pageStr, pageSizeStr string) int {
	if pageStr == "" && pageSizeStr == "" {
		return 0
	}
	pageSize := defaultAuditPageSize
	if n, err := strconv.Atoi(pageSizeStr); err == nil && n > 0 {
		pageSize = n
	}
	return pageSize
}
