package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route/param"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/biz/model"
)

func newTestRegistryWithNodes(t *testing.T, nodes map[string]string) *model.NodeRegistry {
	t.Helper()
	tmpDir := t.TempDir()
	registry := model.NewNodeRegistry(filepath.Join(tmpDir, "nodes.json"), logutil.NewNop())
	for name, addr := range nodes {
		registry.Register(protocol.RegisterRequest{Name: name, Address: addr})
	}
	time.Sleep(50 * time.Millisecond)
	return registry
}

func newRequestContextWithParams(params ...param.Param) *app.RequestContext {
	c := app.NewContext(0)
	c.Params = params
	return c
}

// ========== AuditLogger 测试 ==========

func TestAuditLogger_LogAndList(t *testing.T) {
	a := NewAuditLogger("", logutil.NewNop())

	for i := 1; i <= 3; i++ {
		a.Log(protocol.AuditLogEntry{
			TargetNode: fmt.Sprintf("agent-%d", i),
			Action:     "list_sessions",
			Result:     "success",
		})
	}

	logs := a.List()
	if len(logs) != 3 {
		t.Fatalf("期望 3 条日志, 实际=%d", len(logs))
	}
	// 最新的在前
	if logs[0].TargetNode != "agent-3" {
		t.Errorf("最新日志应为 agent-3, 实际=%s", logs[0].TargetNode)
	}
	if logs[2].TargetNode != "agent-1" {
		t.Errorf("最早日志应为 agent-1, 实际=%s", logs[2].TargetNode)
	}
}

func TestAuditLogger_AutoFillFields(t *testing.T) {
	a := NewAuditLogger("", logutil.NewNop())

	before := time.Now()
	a.Log(protocol.AuditLogEntry{
		TargetNode: "agent-1",
		Action:     "create_session",
		Result:     "success",
	})
	after := time.Now()

	logs := a.List()
	if len(logs) != 1 {
		t.Fatalf("期望 1 条日志, 实际=%d", len(logs))
	}
	entry := logs[0]
	if entry.ID == "" {
		t.Error("ID 应被自动填充")
	}
	if entry.Timestamp.Before(before) || entry.Timestamp.After(after) {
		t.Errorf("Timestamp 应在 [%v, %v] 范围内, 实际=%v", before, after, entry.Timestamp)
	}
}

func TestAuditLogger_QueryByNode(t *testing.T) {
	a := NewAuditLogger("", logutil.NewNop())
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-1", Action: "list_sessions"})
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-2", Action: "list_sessions"})
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-1", Action: "create_session"})

	results := a.Query("agent-1", "")
	if len(results) != 2 {
		t.Fatalf("期望 2 条 agent-1 记录, 实际=%d", len(results))
	}
	for _, e := range results {
		if e.TargetNode != "agent-1" {
			t.Errorf("期望 TargetNode=agent-1, 实际=%s", e.TargetNode)
		}
	}
}

func TestAuditLogger_QueryByAction(t *testing.T) {
	a := NewAuditLogger("", logutil.NewNop())
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-1", Action: "list_sessions"})
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-2", Action: "create_session"})
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-1", Action: "create_session"})

	results := a.Query("", "create")
	if len(results) != 2 {
		t.Fatalf("期望 2 条 create 操作记录, 实际=%d", len(results))
	}
	for _, e := range results {
		if e.Action != "create_session" {
			t.Errorf("期望 Action=create_session, 实际=%s", e.Action)
		}
	}
}

func TestAuditLogger_QueryByNodeAndAction(t *testing.T) {
	a := NewAuditLogger("", logutil.NewNop())
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-1", Action: "list_sessions"})
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-1", Action: "create_session"})
	a.Log(protocol.AuditLogEntry{TargetNode: "agent-2", Action: "create_session"})

	results := a.Query("agent-1", "create")
	if len(results) != 1 {
		t.Fatalf("期望 1 条 agent-1+create 记录, 实际=%d", len(results))
	}
	if results[0].TargetNode != "agent-1" || results[0].Action != "create_session" {
		t.Errorf("期望 agent-1+create_session, 实际 %s/%s", results[0].TargetNode, results[0].Action)
	}
}

// ========== SessionProxy 测试 ==========

func setupProxyHandler(t *testing.T, agentHandler http.HandlerFunc) *SessionProxyHandler {
	t.Helper()

	mockAgent := httptest.NewServer(agentHandler)
	t.Cleanup(func() { mockAgent.Close() })

	registry := newTestRegistryWithNodes(t, map[string]string{
		"agent-1": "http://" + mockAgent.Listener.Addr().String(),
	})

	auditLog := NewAuditLogger("", logutil.NewNop())
	return NewSessionProxyHandler(registry, auditLog, logutil.NewNop(), "", nil, true)
}

func TestSessionProxyHandler_ProxySuccess(t *testing.T) {
	handler := setupProxyHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/sessions" {
			t.Errorf("代理路径期望 /api/v1/sessions, 实际=%s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("代理方法期望 GET, 实际=%s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","data":[]}`))
	})

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "agent-1"})
	handler.ListSessions(nil, c)

	if c.Response.StatusCode() != http.StatusOK {
		t.Errorf("期望状态码 200, 实际=%d", c.Response.StatusCode())
	}

	logs := handler.auditLog.List()
	if len(logs) != 1 {
		t.Fatalf("期望 1 条审计日志, 实际=%d", len(logs))
	}
	if logs[0].Action != "list_sessions" {
		t.Errorf("审计日志 Action 期望 list_sessions, 实际=%s", logs[0].Action)
	}
	if logs[0].Result != "success" {
		t.Errorf("审计日志 Result 期望 success, 实际=%s", logs[0].Result)
	}
	if logs[0].TargetNode != "agent-1" {
		t.Errorf("审计日志 TargetNode 期望 agent-1, 实际=%s", logs[0].TargetNode)
	}
}

func TestSessionProxyHandler_ProxyAgentError(t *testing.T) {
	handler := setupProxyHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal server error"}`))
	})

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "agent-1"})
	handler.ListSessions(nil, c)

	if c.Response.StatusCode() != http.StatusInternalServerError {
		t.Errorf("期望状态码 500, 实际=%d", c.Response.StatusCode())
	}

	logs := handler.auditLog.List()
	if len(logs) != 1 {
		t.Fatalf("期望 1 条审计日志, 实际=%d", len(logs))
	}
	if logs[0].StatusCode != 500 {
		t.Errorf("审计日志 StatusCode 期望 500, 实际=%d", logs[0].StatusCode)
	}
}

func TestSessionProxyHandler_ProxyAgentUnreachable(t *testing.T) {
	tmpDir := t.TempDir()
	registry := model.NewNodeRegistry(filepath.Join(tmpDir, "nodes.json"), logutil.NewNop())
	registry.Register(protocol.RegisterRequest{Name: "agent-dead", Address: "http://127.0.0.1:1"})
	time.Sleep(50 * time.Millisecond)

	auditLog := NewAuditLogger("", logutil.NewNop())
	handler := NewSessionProxyHandler(registry, auditLog, logutil.NewNop(), "", nil, true)

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "agent-dead"})
	handler.ListSessions(nil, c)

	if c.Response.StatusCode() != http.StatusBadGateway {
		t.Errorf("期望状态码 502, 实际=%d", c.Response.StatusCode())
	}

	logs := auditLog.List()
	if len(logs) != 1 {
		t.Fatalf("期望 1 条审计日志, 实际=%d", len(logs))
	}
	if logs[0].StatusCode != 502 {
		t.Errorf("审计日志 StatusCode 期望 502, 实际=%d", logs[0].StatusCode)
	}
	if logs[0].TargetNode != "agent-dead" {
		t.Errorf("审计日志 TargetNode 期望 agent-dead, 实际=%s", logs[0].TargetNode)
	}
}

func TestSessionProxyHandler_NodeNotFound(t *testing.T) {
	handler := setupProxyHandler(t, nil)

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "nonexistent"})
	handler.ListSessions(nil, c)

	if c.Response.StatusCode() != http.StatusNotFound {
		t.Errorf("期望状态码 404, 实际=%d", c.Response.StatusCode())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(c.Response.Body(), &resp); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}
	if resp["status"] != "error" {
		t.Errorf("期望 status=error, 实际=%v", resp["status"])
	}

	// 节点未找到属于请求校验失败，不应记录审计日志
	logs := handler.auditLog.List()
	if len(logs) != 0 {
		t.Errorf("期望无审计日志, 实际=%d 条", len(logs))
	}
}

// TestSessionProxyHandler_AddressSchemeFormat 回归测试：
// 验证当 address 含 http:// 前缀时，代理层不会拼接出双重 http:// 前缀。
// 这个测试覆盖了 address 字段必须包含 scheme 的协议规范。
func TestSessionProxyHandler_AddressSchemeFormat(t *testing.T) {
	var receivedURL string
	handler := setupProxyHandler(t, func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "agent-1"})
	handler.ListSessions(nil, c)

	// 验证请求成功到达 mock Agent（不会出现 DNS 解析失败）
	if c.Response.StatusCode() != http.StatusOK {
		t.Errorf("期望状态码 200, 实际=%d, 响应体=%s", c.Response.StatusCode(), string(c.Response.Body()))
	}
	// 验证代理路径正确
	if receivedURL != "/api/v1/sessions" {
		t.Errorf("期望代理路径 /api/v1/sessions, 实际=%s", receivedURL)
	}
}
