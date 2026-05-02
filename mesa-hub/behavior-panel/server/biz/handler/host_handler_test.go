package handler

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route/param"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/biz/config"
	"github.com/charviki/mesa-hub-behavior-panel/biz/model"
	"github.com/charviki/mesa-hub-behavior-panel/biz/runtime"
	"github.com/charviki/mesa-hub-behavior-panel/biz/service"
)

type mockHostRuntime struct {
	deployErr     error
	deployFn      func(ctx context.Context, spec *protocol.HostDeploySpec, dockerfileContent string) (*protocol.CreateHostResponse, error)
	removeErr     error
	inspectFn     func(ctx context.Context, name string) (*protocol.ContainerInfo, error)
	runtimeLogsFn func(ctx context.Context, name string, tailLines int) (string, error)
	isHealthyFn   func(ctx context.Context, name string) (bool, error)
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

func (m *mockHostRuntime) StopHost(ctx context.Context, name string) error   { return m.removeErr }
func (m *mockHostRuntime) RemoveHost(ctx context.Context, name string) error { return m.removeErr }

func (m *mockHostRuntime) InspectHost(ctx context.Context, name string) (*protocol.ContainerInfo, error) {
	if m.inspectFn != nil {
		return m.inspectFn(ctx, name)
	}
	return &protocol.ContainerInfo{ID: "abc123", Name: name, Status: "running"}, nil
}

func (m *mockHostRuntime) GetRuntimeLogs(ctx context.Context, name string, tailLines int) (string, error) {
	if m.runtimeLogsFn != nil {
		return m.runtimeLogsFn(ctx, name, tailLines)
	}
	return "mock log line 1\nmock log line 2\n", nil
}

func (m *mockHostRuntime) IsHealthy(ctx context.Context, name string) (bool, error) {
	if m.isHealthyFn != nil {
		return m.isHealthyFn(ctx, name)
	}
	return true, nil
}

type testHostHelper struct {
	svc      *service.HostService
	handler  *HostHandler
	registry *model.NodeRegistry
	specMgr  *model.HostSpecManager
	auditLog *AuditLogger
	logDir   string
	rt       runtime.HostRuntime
}

func newTestHostHelper(t *testing.T, rt runtime.HostRuntime) *testHostHelper {
	t.Helper()
	tmpDir := t.TempDir()
	registry := model.NewNodeRegistry(filepath.Join(tmpDir, "nodes.json"), logutil.NewNop())
	specMgr := model.NewHostSpecManager(filepath.Join(tmpDir, "host_specs.json"), logutil.NewNop())
	cfg := &config.Config{
		Server: config.ServerConfig{AuthToken: "test-token"},
		Docker: config.DockerConfig{AgentBaseImage: "test-base:latest"},
	}
	auditLog := NewAuditLogger("", logutil.NewNop())
	logDir := filepath.Join(tmpDir, "host_logs")
	svc := service.NewHostService(registry, specMgr, rt, auditLog, cfg, logutil.NewNop(), logDir)
	return &testHostHelper{
		svc:      svc,
		handler:  NewHostHandler(svc),
		registry: registry,
		specMgr:  specMgr,
		auditLog: auditLog,
		logDir:   logDir,
		rt:       rt,
	}
}

func newTestHostHelperWithNode(t *testing.T, rt runtime.HostRuntime, name, addr string) *testHostHelper {
	t.Helper()
	h := newTestHostHelper(t, rt)
	h.registry.Register(protocol.RegisterRequest{Name: name, Address: addr})
	return h
}

// ========== CreateHost 测试 ==========

func TestHostHandler_CreateHost_Async202(t *testing.T) {
	th := newTestHostHelper(t, &mockHostRuntime{})

	c := newPostContext(`{"name":"test-host","tools":["claude","go"],"resources":{"cpu_limit":"2","memory_limit":"4g"}}`)
	th.handler.CreateHost(context.TODO(), c)

	if c.Response.StatusCode() != http.StatusAccepted {
		t.Fatalf("期望 202, 实际=%d, body=%s", c.Response.StatusCode(), string(c.Response.Body()))
	}

	resp := parseResponse(t, c.Response.Body())
	if resp["status"] != "ok" {
		t.Errorf("期望 status=ok, 实际=%v", resp["status"])
	}

	spec := th.specMgr.Get("test-host")
	if spec == nil {
		t.Fatal("期望 HostSpec 已创建")
	}

	time.Sleep(200 * time.Millisecond)

	spec = th.specMgr.Get("test-host")
	if spec.Status != protocol.HostStatusDeploying {
		t.Errorf("期望 Status=deploying, 实际=%s", spec.Status)
	}
}

func TestHostHandler_CreateHost_MissingName(t *testing.T) {
	th := newTestHostHelper(t, &mockHostRuntime{})

	c := newPostContext(`{"tools":["claude"]}`)
	th.handler.CreateHost(context.TODO(), c)

	if c.Response.StatusCode() != http.StatusBadRequest {
		t.Fatalf("期望 400, 实际=%d", c.Response.StatusCode())
	}
}

func TestHostHandler_CreateHost_NoTools(t *testing.T) {
	th := newTestHostHelper(t, &mockHostRuntime{})

	c := newPostContext(`{"name":"test-host","tools":[]}`)
	th.handler.CreateHost(context.TODO(), c)

	if c.Response.StatusCode() != http.StatusBadRequest {
		t.Fatalf("期望 400, 实际=%d", c.Response.StatusCode())
	}
}

func TestHostHandler_CreateHost_ConflictInSpecMgr(t *testing.T) {
	th := newTestHostHelper(t, &mockHostRuntime{})
	th.specMgr.Create(&protocol.HostSpec{Name: "existing-host", Tools: []string{"claude"}, Status: protocol.HostStatusPending})

	c := newPostContext(`{"name":"existing-host","tools":["claude"]}`)
	th.handler.CreateHost(context.TODO(), c)

	if c.Response.StatusCode() != http.StatusConflict {
		t.Fatalf("期望 409, 实际=%d", c.Response.StatusCode())
	}
}

func TestHostHandler_CreateHost_UnknownTools(t *testing.T) {
	th := newTestHostHelper(t, &mockHostRuntime{})

	c := newPostContext(`{"name":"test-host","tools":["claude","nonexistent"]}`)
	th.handler.CreateHost(context.TODO(), c)

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
	th := newTestHostHelper(t, &mockHostRuntime{
		deployErr: errors.New("docker build failed: some error"),
	})

	c := newPostContext(`{"name":"test-host","tools":["claude"]}`)
	th.handler.CreateHost(context.TODO(), c)

	if c.Response.StatusCode() != http.StatusAccepted {
		t.Fatalf("期望 202（异步创建）, 实际=%d", c.Response.StatusCode())
	}

	time.Sleep(200 * time.Millisecond)

	spec := th.specMgr.Get("test-host")
	if spec == nil {
		t.Fatal("期望 HostSpec 已创建")
	}
	if spec.Status != protocol.HostStatusFailed {
		t.Errorf("期望 Status=failed, 实际=%s", spec.Status)
	}
}

func TestHostHandler_CreateHost_InvalidBody(t *testing.T) {
	th := newTestHostHelper(t, &mockHostRuntime{})

	c := newPostContext(`{invalid json}`)
	th.handler.CreateHost(context.TODO(), c)

	if c.Response.StatusCode() != http.StatusBadRequest {
		t.Fatalf("期望 400, 实际=%d", c.Response.StatusCode())
	}
}

func TestHostHandler_CreateHost_EmptyResources(t *testing.T) {
	th := newTestHostHelper(t, &mockHostRuntime{})

	c := newPostContext(`{"name":"test-host","tools":["claude"]}`)
	th.handler.CreateHost(context.TODO(), c)

	if c.Response.StatusCode() != http.StatusAccepted {
		t.Fatalf("期望 202（resources 可选）, 实际=%d, body=%s", c.Response.StatusCode(), string(c.Response.Body()))
	}

	time.Sleep(200 * time.Millisecond)
}

// ========== ListTools 测试 ==========

func TestHostHandler_ListTools(t *testing.T) {
	th := newTestHostHelper(t, &mockHostRuntime{})

	c := app.NewContext(0)
	c.Request.SetMethod("GET")
	th.handler.ListTools(context.TODO(), c)

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
	th := newTestHostHelperWithNode(t, &mockHostRuntime{}, "test-host", "http://localhost:8080")
	th.specMgr.Create(&protocol.HostSpec{Name: "test-host", Tools: []string{"claude"}, Status: protocol.HostStatusDeploying})

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "test-host"})
	th.handler.DeleteHost(context.TODO(), c)

	if c.Response.StatusCode() != http.StatusOK {
		t.Fatalf("期望 200, 实际=%d", c.Response.StatusCode())
	}

	if th.specMgr.Get("test-host") != nil {
		t.Error("期望 HostSpec 已被删除")
	}
}

func TestHostHandler_DeleteHost_RemoveErrorReturned(t *testing.T) {
	th := newTestHostHelperWithNode(t, &mockHostRuntime{
		removeErr: errors.New("cleanup failed"),
	}, "test-host", "http://localhost:8080")
	th.specMgr.Create(&protocol.HostSpec{Name: "test-host", Tools: []string{"claude"}, Status: protocol.HostStatusOnline})

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "test-host"})
	th.handler.DeleteHost(context.TODO(), c)

	if c.Response.StatusCode() != http.StatusInternalServerError {
		t.Fatalf("期望 500（RemoveHost 错误应向上返回）, 实际=%d", c.Response.StatusCode())
	}
	if th.specMgr.Get("test-host") == nil {
		t.Fatal("期望 HostSpec 保留，避免把底层清理失败伪装成删除成功")
	}
}

func TestHostHandler_DeleteHost_EmptyName(t *testing.T) {
	th := newTestHostHelper(t, &mockHostRuntime{})

	c := newRequestContextWithParams(param.Param{Key: "name", Value: ""})
	th.handler.DeleteHost(context.TODO(), c)

	if c.Response.StatusCode() != http.StatusBadRequest {
		t.Fatalf("期望 400, 实际=%d", c.Response.StatusCode())
	}
}

// ========== ListHosts 测试 ==========

func TestHostHandler_ListHosts(t *testing.T) {
	th := newTestHostHelper(t, &mockHostRuntime{})
	th.specMgr.Create(&protocol.HostSpec{Name: "host-1", Tools: []string{"claude"}, Status: protocol.HostStatusPending})
	th.specMgr.Create(&protocol.HostSpec{Name: "host-2", Tools: []string{"go"}, Status: protocol.HostStatusDeploying})

	c := app.NewContext(0)
	c.Request.SetMethod("GET")
	th.handler.ListHosts(context.TODO(), c)

	if c.Response.StatusCode() != http.StatusOK {
		t.Fatalf("期望 200, 实际=%d", c.Response.StatusCode())
	}

	resp := parseResponse(t, c.Response.Body())
	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatal("data 应为数组")
	}
	if len(data) != 2 {
		t.Errorf("期望 2 个 Host, 实际=%d", len(data))
	}
}

// ========== GetHost 测试 ==========

func TestHostHandler_GetHost(t *testing.T) {
	th := newTestHostHelper(t, &mockHostRuntime{})
	th.specMgr.Create(&protocol.HostSpec{Name: "host-1", Tools: []string{"claude"}, Status: protocol.HostStatusPending})

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "host-1"})
	th.handler.GetHost(context.TODO(), c)

	if c.Response.StatusCode() != http.StatusOK {
		t.Fatalf("期望 200, 实际=%d", c.Response.StatusCode())
	}
}

func TestHostHandler_GetHost_NotFound(t *testing.T) {
	th := newTestHostHelper(t, &mockHostRuntime{})

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "nonexistent"})
	th.handler.GetHost(context.TODO(), c)

	if c.Response.StatusCode() != http.StatusNotFound {
		t.Fatalf("期望 404, 实际=%d", c.Response.StatusCode())
	}
}

// ========== GetBuildLog 测试 ==========

func TestHostHandler_GetBuildLog(t *testing.T) {
	th := newTestHostHelper(t, &mockHostRuntime{})
	th.specMgr.Create(&protocol.HostSpec{Name: "host-1", Tools: []string{"claude"}, Status: protocol.HostStatusDeploying})

	buildSvc := service.NewHostService(th.registry, th.specMgr, th.rt, th.auditLog,
		&config.Config{
			Server: config.ServerConfig{AuthToken: "test-token"},
			Docker: config.DockerConfig{AgentBaseImage: "test-base:latest"},
		}, logutil.NewNop(), th.logDir)
	buildSvc.CreateHost(context.TODO(), &protocol.CreateHostRequest{Name: "host-1", Tools: []string{"claude"}})
	time.Sleep(200 * time.Millisecond)

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "host-1"})
	th.handler.GetBuildLog(context.TODO(), c)

	if c.Response.StatusCode() != http.StatusOK {
		t.Fatalf("期望 200, 实际=%d", c.Response.StatusCode())
	}
}

func TestHostHandler_GetBuildLog_NotFound(t *testing.T) {
	th := newTestHostHelper(t, &mockHostRuntime{})

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "nonexistent"})
	th.handler.GetBuildLog(context.TODO(), c)

	if c.Response.StatusCode() != http.StatusNotFound {
		t.Fatalf("期望 404, 实际=%d", c.Response.StatusCode())
	}
}

// ========== GetRuntimeLog 测试 ==========

func TestHostHandler_GetRuntimeLog(t *testing.T) {
	th := newTestHostHelper(t, &mockHostRuntime{})
	th.specMgr.Create(&protocol.HostSpec{Name: "host-1", Tools: []string{"claude"}, Status: protocol.HostStatusDeploying})

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "host-1"})
	th.handler.GetRuntimeLog(context.TODO(), c)

	if c.Response.StatusCode() != http.StatusOK {
		t.Fatalf("期望 200, 实际=%d, body=%s", c.Response.StatusCode(), string(c.Response.Body()))
	}
}

func TestHostHandler_GetRuntimeLog_NotFound(t *testing.T) {
	th := newTestHostHelper(t, &mockHostRuntime{})

	c := newRequestContextWithParams(param.Param{Key: "name", Value: "nonexistent"})
	th.handler.GetRuntimeLog(context.TODO(), c)

	if c.Response.StatusCode() != http.StatusNotFound {
		t.Fatalf("期望 404, 实际=%d", c.Response.StatusCode())
	}
}
