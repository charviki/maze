package handler

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route/param"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/biz/config"
	"github.com/charviki/mesa-hub-behavior-panel/biz/model"
	"github.com/charviki/mesa-hub-behavior-panel/biz/runtime"
)

// mockHostRuntime HostRuntime 的 mock 实现
type mockHostRuntime struct {
	deployErr error
	deployFn  func(ctx context.Context, spec *protocol.HostDeploySpec, dockerfileContent string) (*protocol.CreateHostResponse, error)
	removeErr error
	inspectFn func(ctx context.Context, name string) (*protocol.ContainerInfo, error)
}

func (m *mockHostRuntime) DeployHost(ctx context.Context, spec *protocol.HostDeploySpec, dockerfileContent string) (*protocol.CreateHostResponse, error) {
	if m.deployFn != nil {
		return m.deployFn(ctx, spec, dockerfileContent)
	}
	if m.deployErr != nil {
		return nil, m.deployErr
	}
	return &protocol.CreateHostResponse{
		Name:        spec.Name,
		Tools:       spec.Tools,
		ImageTag:    "maze-host-" + spec.Name + ":latest",
		ContainerID: "mock-container-id",
		Status:      "running",
	}, nil
}

func (m *mockHostRuntime) RemoveHost(ctx context.Context, name string) error {
	return m.removeErr
}

func (m *mockHostRuntime) InspectHost(ctx context.Context, name string) (*protocol.ContainerInfo, error) {
	if m.inspectFn != nil {
		return m.inspectFn(ctx, name)
	}
	return &protocol.ContainerInfo{ID: "abc123", Name: name, Status: "running"}, nil
}

func newTestHostHandler(t *testing.T, rt runtime.HostRuntime) *HostHandler {
	t.Helper()
	registry := model.NewNodeRegistry(
		filepath.Join(t.TempDir(), "nodes.json"),
		logutil.NewNop(),
	)
	cfg := &config.Config{
		Server: config.ServerConfig{AuthToken: "test-token"},
		Docker: config.DockerConfig{AgentBaseImage: "test-base:latest"},
	}
	auditLog := NewAuditLogger("", logutil.NewNop())
	return NewHostHandler(registry, rt, auditLog, cfg, logutil.NewNop())
}

func newTestHostHandlerWithNode(t *testing.T, rt runtime.HostRuntime, name, addr string) *HostHandler {
	t.Helper()
	h := newTestHostHandler(t, rt)
	h.registry.Register(protocol.RegisterRequest{Name: name, Address: addr})
	return h
}

// ========== CreateHost 测试 ==========

func TestHostHandler_CreateHost_Success(t *testing.T) {
	h := newTestHostHandler(t, &mockHostRuntime{})

	c := newPostContext(`{"name":"test-host","tools":["claude","go"],"resources":{"cpu_limit":"2","memory_limit":"4g"}}`)
	h.CreateHost(nil, c)

	if c.Response.StatusCode() != http.StatusOK {
		t.Fatalf("期望 200, 实际=%d, body=%s", c.Response.StatusCode(), string(c.Response.Body()))
	}

	resp := parseResponse(t, c.Response.Body())
	if resp["status"] != "ok" {
		t.Errorf("期望 status=ok, 实际=%v", resp["status"])
	}
}

func TestHostHandler_CreateHost_MissingName(t *testing.T) {
	h := newTestHostHandler(t, &mockHostRuntime{})

	c := newPostContext(`{"tools":["claude"]}`)
	h.CreateHost(nil, c)

	if c.Response.StatusCode() != http.StatusBadRequest {
		t.Fatalf("期望 400, 实际=%d", c.Response.StatusCode())
	}
}

func TestHostHandler_CreateHost_NoTools(t *testing.T) {
	h := newTestHostHandler(t, &mockHostRuntime{})

	c := newPostContext(`{"name":"test-host","tools":[]}`)
	h.CreateHost(nil, c)

	if c.Response.StatusCode() != http.StatusBadRequest {
		t.Fatalf("期望 400, 实际=%d", c.Response.StatusCode())
	}
}

func TestHostHandler_CreateHost_DuplicateName(t *testing.T) {
	h := newTestHostHandlerWithNode(t, &mockHostRuntime{}, "existing-host", "http://localhost:8080")

	c := newPostContext(`{"name":"existing-host","tools":["claude"]}`)
	h.CreateHost(nil, c)

	if c.Response.StatusCode() != http.StatusConflict {
		t.Fatalf("期望 409, 实际=%d", c.Response.StatusCode())
	}
}

func TestHostHandler_CreateHost_UnknownTools(t *testing.T) {
	h := newTestHostHandler(t, &mockHostRuntime{})

	c := newPostContext(`{"name":"test-host","tools":["claude","nonexistent"]}`)
	h.CreateHost(nil, c)

	if c.Response.StatusCode() != http.StatusBadRequest {
		t.Fatalf("期望 400, 实际=%d", c.Response.StatusCode())
	}

	resp := parseResponse(t, c.Response.Body())
	msg, _ := resp["message"].(string)
	if msg == "" {
		t.Error("错误消息不应为空")
	}
}

func TestHostHandler_CreateHost_DeployFailed(t *testing.T) {
	h := newTestHostHandler(t, &mockHostRuntime{
		deployErr: errors.New("docker build failed: some error"),
	})

	c := newPostContext(`{"name":"test-host","tools":["claude"]}`)
	h.CreateHost(nil, c)

	if c.Response.StatusCode() != http.StatusInternalServerError {
		t.Fatalf("期望 500, 实际=%d", c.Response.StatusCode())
	}

	resp := parseResponse(t, c.Response.Body())
	msg, _ := resp["message"].(string)
	if msg == "" {
		t.Error("错误消息不应为空")
	}
}

func TestHostHandler_CreateHost_InvalidBody(t *testing.T) {
	h := newTestHostHandler(t, &mockHostRuntime{})

	c := newPostContext(`{invalid json}`)
	h.CreateHost(nil, c)

	if c.Response.StatusCode() != http.StatusBadRequest {
		t.Fatalf("期望 400, 实际=%d", c.Response.StatusCode())
	}
}

// ========== ListTools 测试 ==========

func TestHostHandler_ListTools(t *testing.T) {
	h := newTestHostHandler(t, &mockHostRuntime{})

	c := app.NewContext(0)
	c.Request.SetMethod("GET")
	h.ListTools(nil, c)

	if c.Response.StatusCode() != http.StatusOK {
		t.Fatalf("期望 200, 实际=%d", c.Response.StatusCode())
	}

	resp := parseResponse(t, c.Response.Body())
	if resp["status"] != "ok" {
		t.Errorf("期望 status=ok, 实际=%v", resp["status"])
	}

	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatal("data 应为数组")
	}
	if len(data) == 0 {
		t.Error("工具列表不应为空")
	}
}

// ========== DeleteHost 测试 ==========

func TestHostHandler_DeleteHost_Success(t *testing.T) {
	h := newTestHostHandlerWithNode(t, &mockHostRuntime{}, "test-host", "http://localhost:8080")

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "test-host"})
	h.DeleteHost(nil, c)

	if c.Response.StatusCode() != http.StatusOK {
		t.Fatalf("期望 200, 实际=%d", c.Response.StatusCode())
	}
}

func TestHostHandler_DeleteHost_RemoveErrorIgnored(t *testing.T) {
	// RemoveHost 失败时 DeleteHost 仍应返回 200（容器可能不存在）
	h := newTestHostHandlerWithNode(t, &mockHostRuntime{
		removeErr: errors.New("container not found"),
	}, "test-host", "http://localhost:8080")

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "test-host"})
	h.DeleteHost(nil, c)

	if c.Response.StatusCode() != http.StatusOK {
		t.Fatalf("期望 200（RemoveHost 错误应被忽略）, 实际=%d", c.Response.StatusCode())
	}
}

func TestHostHandler_CreateHost_EmptyResources(t *testing.T) {
	h := newTestHostHandler(t, &mockHostRuntime{})

	c := newPostContext(`{"name":"test-host","tools":["claude"]}`)
	h.CreateHost(nil, c)

	if c.Response.StatusCode() != http.StatusOK {
		t.Fatalf("期望 200（resources 可选）, 实际=%d, body=%s", c.Response.StatusCode(), string(c.Response.Body()))
	}
}

func TestHostHandler_DeleteHost_EmptyName(t *testing.T) {
	h := newTestHostHandler(t, &mockHostRuntime{})

	c := newRequestContextWithParams(param.Param{Key: "name", Value: ""})
	h.DeleteHost(nil, c)

	if c.Response.StatusCode() != http.StatusBadRequest {
		t.Fatalf("期望 400, 实际=%d", c.Response.StatusCode())
	}
}
