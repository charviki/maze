package handler

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route/param"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/biz/model"
	"github.com/charviki/mesa-hub-behavior-panel/biz/service"
)

// newTestNodeHandler 创建一个使用临时目录的 NodeHandler，测试结束后自动清理
func newTestNodeHandler(t *testing.T) *NodeHandler {
	t.Helper()
	registry := model.NewNodeRegistry(
		filepath.Join(t.TempDir(), "nodes.json"),
		logutil.NewNop(),
	)
	nodeSvc := service.NewNodeService(registry, logutil.NewNop())
	return NewNodeHandler(nodeSvc, registry, "", logutil.NewNop())
}

// newTestNodeHandlerWithNodes 创建一个预注册了指定节点的 NodeHandler
func newTestNodeHandlerWithNodes(t *testing.T, nodes map[string]string) *NodeHandler {
	t.Helper()
	registry := model.NewNodeRegistry(
		filepath.Join(t.TempDir(), "nodes.json"),
		logutil.NewNop(),
	)
	for name, addr := range nodes {
		registry.Register(protocol.RegisterRequest{Name: name, Address: addr})
	}
	nodeSvc := service.NewNodeService(registry, logutil.NewNop())
	return NewNodeHandler(nodeSvc, registry, "", logutil.NewNop())
}

// newPostContext 创建一个携带 JSON body 的 POST 请求上下文
// 必须同时设置 Content-Length，否则 Hertz 的 Bind 会因 ContentLength<=0 跳过 body 解析
func newPostContext(body string) *app.RequestContext {
	c := app.NewContext(0)
	c.Request.SetMethod("POST")
	c.Request.SetRequestURI("/test")
	bodyBytes := []byte(body)
	c.Request.SetBody(bodyBytes)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.SetContentLength(len(bodyBytes))
	return c
}

// parseResponse 解析响应体为通用 map，供断言使用
func parseResponse(t *testing.T, body []byte) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}
	return resp
}

// ========== Register 测试 ==========

// TestNodeHandler_Register 验证合法注册请求返回 200 和完整节点数据
func TestNodeHandler_Register(t *testing.T) {
	h := newTestNodeHandler(t)

	c := newPostContext(`{"name":"agent-1","address":"http://192.168.1.10:9090"}`)
	h.Register(nil, c)

	if c.Response.StatusCode() != http.StatusOK {
		t.Fatalf("期望状态码 200, 实际=%d", c.Response.StatusCode())
	}

	resp := parseResponse(t, c.Response.Body())
	if resp["status"] != "ok" {
		t.Errorf("期望 status=ok, 实际=%v", resp["status"])
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("data 字段类型不是 map[string]interface{}")
	}
	if data["name"] != "agent-1" {
		t.Errorf("期望 data.name=agent-1, 实际=%v", data["name"])
	}
	if data["address"] != "http://192.168.1.10:9090" {
		t.Errorf("期望 data.address=http://192.168.1.10:9090, 实际=%v", data["address"])
	}
	if data["status"] != "online" {
		t.Errorf("期望 data.status=online, 实际=%v", data["status"])
	}
}

// TestNodeHandler_Register_MissingName 验证缺少 name 字段返回 400
func TestNodeHandler_Register_MissingName(t *testing.T) {
	h := newTestNodeHandler(t)

	c := newPostContext(`{"address":"http://192.168.1.10:9090"}`)
	h.Register(nil, c)

	if c.Response.StatusCode() != http.StatusBadRequest {
		t.Fatalf("期望状态码 400, 实际=%d", c.Response.StatusCode())
	}

	resp := parseResponse(t, c.Response.Body())
	if resp["status"] != "error" {
		t.Errorf("期望 status=error, 实际=%v", resp["status"])
	}
	if resp["message"] != "name is required" {
		t.Errorf("期望 message='name is required', 实际=%v", resp["message"])
	}
}

// TestNodeHandler_Register_MissingAddress 验证缺少 address 字段返回 400
func TestNodeHandler_Register_MissingAddress(t *testing.T) {
	h := newTestNodeHandler(t)

	c := newPostContext(`{"name":"agent-1"}`)
	h.Register(nil, c)

	if c.Response.StatusCode() != http.StatusBadRequest {
		t.Fatalf("期望状态码 400, 实际=%d", c.Response.StatusCode())
	}

	resp := parseResponse(t, c.Response.Body())
	if resp["status"] != "error" {
		t.Errorf("期望 status=error, 实际=%v", resp["status"])
	}
	if resp["message"] != "address is required" {
		t.Errorf("期望 message='address is required', 实际=%v", resp["message"])
	}
}

// ========== Heartbeat 测试 ==========

// TestNodeHandler_Heartbeat 验证已注册节点的心跳请求返回 200
func TestNodeHandler_Heartbeat(t *testing.T) {
	h := newTestNodeHandlerWithNodes(t, map[string]string{
		"agent-1": "http://192.168.1.10:9090",
	})

	c := newPostContext(`{"name":"agent-1","status":{"cpu_usage":45.5,"memory_usage_mb":1024}}`)
	h.Heartbeat(nil, c)

	if c.Response.StatusCode() != http.StatusOK {
		t.Fatalf("期望状态码 200, 实际=%d", c.Response.StatusCode())
	}

	resp := parseResponse(t, c.Response.Body())
	if resp["status"] != "ok" {
		t.Errorf("期望 status=ok, 实际=%v", resp["status"])
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("data 字段类型不是 map[string]interface{}")
	}
	// 心跳后节点状态仍为 online
	if data["status"] != "online" {
		t.Errorf("期望 data.status=online, 实际=%v", data["status"])
	}
}

// TestNodeHandler_Heartbeat_NotFound 验证对未注册节点发送心跳返回 404
func TestNodeHandler_Heartbeat_NotFound(t *testing.T) {
	h := newTestNodeHandler(t)

	c := newPostContext(`{"name":"nonexistent","status":{}}`)
	h.Heartbeat(nil, c)

	if c.Response.StatusCode() != http.StatusNotFound {
		t.Fatalf("期望状态码 404, 实际=%d", c.Response.StatusCode())
	}

	resp := parseResponse(t, c.Response.Body())
	if resp["status"] != "error" {
		t.Errorf("期望 status=error, 实际=%v", resp["status"])
	}
	if resp["message"] != "node not found" {
		t.Errorf("期望 message='node not found', 实际=%v", resp["message"])
	}
}

// ========== ListNodes 测试 ==========

// TestNodeHandler_ListNodes 验证返回所有已注册节点
func TestNodeHandler_ListNodes(t *testing.T) {
	h := newTestNodeHandlerWithNodes(t, map[string]string{
		"agent-1": "http://192.168.1.10:9090",
		"agent-2": "http://192.168.1.20:9090",
	})

	c := app.NewContext(0)
	c.Request.SetMethod("GET")
	h.ListNodes(nil, c)

	if c.Response.StatusCode() != http.StatusOK {
		t.Fatalf("期望状态码 200, 实际=%d", c.Response.StatusCode())
	}

	resp := parseResponse(t, c.Response.Body())
	if resp["status"] != "ok" {
		t.Errorf("期望 status=ok, 实际=%v", resp["status"])
	}

	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatal("data 字段类型不是 []interface{}")
	}
	if len(data) != 2 {
		t.Fatalf("期望返回 2 个节点, 实际=%d", len(data))
	}
}

// ========== GetNode 测试 ==========

// TestNodeHandler_GetNode 验证获取已注册节点返回 200 和正确数据
func TestNodeHandler_GetNode(t *testing.T) {
	h := newTestNodeHandlerWithNodes(t, map[string]string{
		"agent-1": "http://192.168.1.10:9090",
	})

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "agent-1"})
	h.GetNode(nil, c)

	if c.Response.StatusCode() != http.StatusOK {
		t.Fatalf("期望状态码 200, 实际=%d", c.Response.StatusCode())
	}

	resp := parseResponse(t, c.Response.Body())
	if resp["status"] != "ok" {
		t.Errorf("期望 status=ok, 实际=%v", resp["status"])
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("data 字段类型不是 map[string]interface{}")
	}
	if data["name"] != "agent-1" {
		t.Errorf("期望 data.name=agent-1, 实际=%v", data["name"])
	}
	if data["address"] != "http://192.168.1.10:9090" {
		t.Errorf("期望 data.address=http://192.168.1.10:9090, 实际=%v", data["address"])
	}
}

// TestNodeHandler_GetNode_NotFound 验证获取不存在的节点返回 404
func TestNodeHandler_GetNode_NotFound(t *testing.T) {
	h := newTestNodeHandler(t)

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "nonexistent"})
	h.GetNode(nil, c)

	if c.Response.StatusCode() != http.StatusNotFound {
		t.Fatalf("期望状态码 404, 实际=%d", c.Response.StatusCode())
	}

	resp := parseResponse(t, c.Response.Body())
	if resp["status"] != "error" {
		t.Errorf("期望 status=error, 实际=%v", resp["status"])
	}
	if resp["message"] != "node not found" {
		t.Errorf("期望 message='node not found', 实际=%v", resp["message"])
	}
}

// ========== DeleteNode 测试 ==========

// TestNodeHandler_DeleteNode 验证删除已注册节点返回 200
func TestNodeHandler_DeleteNode(t *testing.T) {
	h := newTestNodeHandlerWithNodes(t, map[string]string{
		"agent-1": "http://192.168.1.10:9090",
	})

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "agent-1"})
	h.DeleteNode(nil, c)

	if c.Response.StatusCode() != http.StatusOK {
		t.Fatalf("期望状态码 200, 实际=%d", c.Response.StatusCode())
	}

	resp := parseResponse(t, c.Response.Body())
	if resp["status"] != "ok" {
		t.Errorf("期望 status=ok, 实际=%v", resp["status"])
	}

	// 验证删除后再次获取应返回 404
	c2 := newRequestContextWithParams(param.Param{Key: "name", Value: "agent-1"})
	h.GetNode(nil, c2)
	if c2.Response.StatusCode() != http.StatusNotFound {
		t.Errorf("删除后再次获取期望 404, 实际=%d", c2.Response.StatusCode())
	}
}

// TestNodeHandler_DeleteNode_NotFound 验证删除不存在的节点返回 404
func TestNodeHandler_DeleteNode_NotFound(t *testing.T) {
	h := newTestNodeHandler(t)

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "nonexistent"})
	h.DeleteNode(nil, c)

	if c.Response.StatusCode() != http.StatusNotFound {
		t.Fatalf("期望状态码 404, 实际=%d", c.Response.StatusCode())
	}

	resp := parseResponse(t, c.Response.Body())
	if resp["status"] != "error" {
		t.Errorf("期望 status=error, 实际=%v", resp["status"])
	}
	if resp["message"] != "node not found" {
		t.Errorf("期望 message='node not found', 实际=%v", resp["message"])
	}
}
