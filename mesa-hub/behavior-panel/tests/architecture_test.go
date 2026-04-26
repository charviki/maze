package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// setupMockAgent 创建一个模拟 Agent 的 HTTP Server
func setupMockAgent(t *testing.T) *httptest.Server {
	t.Helper()
	mockAgent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/sessions":
			if r.Method == http.MethodGet {
				w.WriteHeader(200)
				_, _ = w.Write([]byte(`{"status":"ok","data":[]}`))
			} else if r.Method == http.MethodPost {
				w.WriteHeader(200)
				_, _ = w.Write([]byte(`{"status":"ok","data":{"id":"sess-1","name":"test-session"}}`))
			}
		case "/api/v1/sessions/saved":
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"status":"ok","data":[]}`))
		case "/api/v1/sessions/save":
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"status":"ok","data":{"saved_at":"` + time.Now().Format(time.RFC3339) + `"}}`))
		default:
			if r.Method == http.MethodGet {
				w.WriteHeader(200)
				_, _ = w.Write([]byte(`{"status":"ok","data":{"id":"sess-1"}}`))
			} else if r.Method == http.MethodDelete {
				w.WriteHeader(200)
				_, _ = w.Write([]byte(`{"status":"ok","data":null}`))
			} else {
				w.WriteHeader(200)
				_, _ = w.Write([]byte(`{"status":"ok","data":null}`))
			}
		}
	}))
	t.Cleanup(func() { mockAgent.Close() })
	return mockAgent
}

// doRequest 通用 HTTP 请求辅助函数
func doRequest(t *testing.T, method, url string, body interface{}) *APIResponse {
	t.Helper()
	var req *http.Request
	var err error
	if body != nil {
		data, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			t.Fatalf("marshal body failed: %v", marshalErr)
		}
		req, err = http.NewRequest(method, url, bytes.NewReader(data))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s failed: %v", method, url, err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	return &apiResp
}

// TestIntegration_RegisterHeartbeat 验证注册→心跳→状态聚合完整链路
// 需要运行中的 Manager 和 Agent 服务（通过 docker-compose 或手动启动）
func TestIntegration_RegisterHeartbeat(t *testing.T) {
	agent1 := agent1Port()
	manager := managerPort()

	// 1. 注册
	registerBody := map[string]interface{}{
		"name":           "integration-agent-1",
		"address":        fmt.Sprintf("http://host.docker.internal:%s", agent1),
		"external_addr":  fmt.Sprintf("http://localhost:%s", agent1),
		"capabilities":   map[string]interface{}{
			"supported_templates": []string{"claude", "bash"},
			"max_sessions":        10,
			"tools":               []string{"tmux", "filesystem"},
		},
		"status": map[string]interface{}{
			"active_sessions": 0,
			"cpu_usage":       0,
			"memory_usage_mb": 128,
			"workspace_root":  "/home/agent",
		},
		"metadata": map[string]interface{}{
			"version":    "test-0.1.0",
			"hostname":   "integration-test-agent",
			"started_at": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
		},
	}

	resp := apiPost(t, manager, "/api/v1/nodes/register", registerBody)
	assertOK(t, resp, "注册")

	// 2. 查询节点列表，验证注册成功
	listResp := apiGet(t, manager, "/api/v1/nodes")
	assertOK(t, listResp, "列出节点")

	nodes := decodeData[[]Node](t, listResp)
	found := false
	for _, n := range nodes {
		if n.Name == "integration-agent-1" {
			found = true
			if n.Status != "online" {
				t.Errorf("期望 Status=online, 实际=%s", n.Status)
			}
			break
		}
	}
	if !found {
		t.Error("期望在节点列表中找到 integration-agent-1")
	}

	// 3. 心跳
	heartbeatBody := map[string]interface{}{
		"name": "integration-agent-1",
		"status": map[string]interface{}{
			"cpu_usage":      42.5,
			"memory_usage_mb": 512.0,
			"workspace_root": "/home/agent",
			"session_details": []map[string]interface{}{
				{"id": "sess-1", "template": "claude", "working_dir": "/home/agent/proj", "uptime_seconds": 300},
			},
		},
	}

	hbResp := apiPost(t, manager, "/api/v1/nodes/heartbeat", heartbeatBody)
	assertOK(t, hbResp, "心跳")

	// 4. 查询节点详情，验证心跳更新
	detailResp := apiGet(t, manager, "/api/v1/nodes/integration-agent-1")
	assertOK(t, detailResp, "获取节点详情")
}

// TestIntegration_ProxySessionLifecycle 验证通过 Manager 代理的 Session CRUD
func TestIntegration_ProxySessionLifecycle(t *testing.T) {
	manager := managerPort()
	nodeName := "integration-agent-1"

	// 1. 通过代理列出 Session
	listResp := apiGet(t, manager, fmt.Sprintf("/api/v1/nodes/%s/sessions", nodeName))
	assertOK(t, listResp, "代理列出 Session")

	// 2. 通过代理创建 Session
	sessID := generateID()
	createResp := apiPost(t, manager, fmt.Sprintf("/api/v1/nodes/%s/sessions", nodeName), map[string]interface{}{
		"name":    sessID,
		"command": "echo hello",
	})
	assertOK(t, createResp, "代理创建 Session")

	// 3. 通过代理获取单个 Session
	getResp := apiGet(t, manager, fmt.Sprintf("/api/v1/nodes/%s/sessions/%s", nodeName, sessID))
	assertOK(t, getResp, "代理获取 Session")

	// 4. 通过代理删除 Session
	deleteResp := apiDelete(t, manager, fmt.Sprintf("/api/v1/nodes/%s/sessions/%s", nodeName, sessID))
	assertOK(t, deleteResp, "代理删除 Session")
}

// TestIntegration_AuditLogs 验证审计日志记录和查询
func TestIntegration_AuditLogs(t *testing.T) {
	manager := managerPort()
	nodeName := "integration-agent-1"

	// 先触发一些代理操作（确保有审计日志）
	_ = apiGet(t, manager, fmt.Sprintf("/api/v1/nodes/%s/sessions", nodeName))

	// 查询审计日志
	logResp := apiGet(t, manager, "/api/v1/audit/logs")
	assertOK(t, logResp, "查询审计日志")
}

// TestIntegration_SameNameNodeReplace 验证同名节点替换
func TestIntegration_SameNameNodeReplace(t *testing.T) {
	manager := managerPort()
	agent1 := agent1Port()

	// 第一次注册
	registerBody1 := map[string]interface{}{
		"name":    "replace-test-agent",
		"address": fmt.Sprintf("http://host.docker.internal:%s", agent1),
		"capabilities": map[string]interface{}{
			"max_sessions": 5,
		},
		"metadata": map[string]interface{}{
			"version": "v1",
		},
	}
	resp1 := apiPost(t, manager, "/api/v1/nodes/register", registerBody1)
	assertOK(t, resp1, "第一次注册")

	// 同名注册（模拟新 Agent 接管）
	registerBody2 := map[string]interface{}{
		"name":    "replace-test-agent",
		"address": fmt.Sprintf("http://host.docker.internal:%s", agent1),
		"capabilities": map[string]interface{}{
			"max_sessions": 20,
		},
		"metadata": map[string]interface{}{
			"version": "v2",
		},
	}
	resp2 := apiPost(t, manager, "/api/v1/nodes/register", registerBody2)
	assertOK(t, resp2, "同名注册覆盖")

	// 验证节点列表中只有一个 replace-test-agent
	listResp := apiGet(t, manager, "/api/v1/nodes")
	assertOK(t, listResp, "列出节点")
	nodes := decodeData[[]Node](t, listResp)
	count := 0
	for _, n := range nodes {
		if n.Name == "replace-test-agent" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("期望 1 个 replace-test-agent 节点, 实际=%d", count)
	}
}
