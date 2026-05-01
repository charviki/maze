package reconciler

import (
	"context"
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/biz/config"
	"github.com/charviki/mesa-hub-behavior-panel/biz/model"
)

type mockReconcilerRuntime struct {
	healthy    map[string]bool
	deployed   map[string]bool
	deployErr  error
	deployCall int32
}

func newMockReconcilerRuntime() *mockReconcilerRuntime {
	return &mockReconcilerRuntime{
		healthy:  make(map[string]bool),
		deployed: make(map[string]bool),
	}
}

func (m *mockReconcilerRuntime) DeployHost(ctx context.Context, spec *protocol.HostDeploySpec, dockerfileContent string) (*protocol.CreateHostResponse, error) {
	atomic.AddInt32(&m.deployCall, 1)
	if m.deployErr != nil {
		return nil, m.deployErr
	}
	m.deployed[spec.Name] = true
	return &protocol.CreateHostResponse{Name: spec.Name, Status: "running"}, nil
}

func (m *mockReconcilerRuntime) StopHost(ctx context.Context, name string) error {
	return nil
}

func (m *mockReconcilerRuntime) RemoveHost(ctx context.Context, name string) error {
	return nil
}

func (m *mockReconcilerRuntime) InspectHost(ctx context.Context, name string) (*protocol.ContainerInfo, error) {
	return nil, nil
}

func (m *mockReconcilerRuntime) GetRuntimeLogs(ctx context.Context, name string, tailLines int) (string, error) {
	return "", nil
}

func (m *mockReconcilerRuntime) IsHealthy(ctx context.Context, name string) (bool, error) {
	return m.healthy[name], nil
}

func newTestReconciler(t *testing.T, rt *mockReconcilerRuntime) (*Reconciler, *model.HostSpecManager, *model.NodeRegistry) {
	t.Helper()
	tmpDir := t.TempDir()
	specMgr := model.NewHostSpecManager(filepath.Join(tmpDir, "host_specs.json"), logutil.NewNop())
	registry := model.NewNodeRegistry(filepath.Join(tmpDir, "nodes.json"), logutil.NewNop())
	cfg := &config.Config{
		Server: config.ServerConfig{AuthToken: "test-token"},
		Docker: config.DockerConfig{AgentBaseImage: "test-base:latest"},
	}
	logDir := filepath.Join(tmpDir, "host_logs")
	rec := NewReconciler(specMgr, registry, rt, cfg, logutil.NewNop(), logDir)
	return rec, specMgr, registry
}

// ========== RecoverOnStartup 测试 ==========

func TestReconciler_RecoverAllHealthy(t *testing.T) {
	rt := newMockReconcilerRuntime()
	rt.healthy["host-1"] = true
	rt.healthy["host-2"] = true

	rec, specMgr, _ := newTestReconciler(t, rt)
	specMgr.Create(&protocol.HostSpec{Name: "host-1", Tools: []string{"claude"}, Status: protocol.HostStatusOnline, AuthToken: "t1"})
	specMgr.Create(&protocol.HostSpec{Name: "host-2", Tools: []string{"go"}, Status: protocol.HostStatusDeploying, AuthToken: "t2"})

	rec.RecoverOnStartup(context.Background())

	// 不应触发重新部署
	if atomic.LoadInt32(&rt.deployCall) != 0 {
		t.Errorf("期望不触发部署, 实际=%d", rt.deployCall)
	}
}

func TestReconciler_RecoverMissingRuntime(t *testing.T) {
	rt := newMockReconcilerRuntime()
	// host-1 不健康

	rec, specMgr, _ := newTestReconciler(t, rt)
	specMgr.Create(&protocol.HostSpec{Name: "host-1", Tools: []string{"claude"}, Status: protocol.HostStatusOnline, AuthToken: "t1"})

	rec.RecoverOnStartup(context.Background())

	// 应触发重新部署
	if atomic.LoadInt32(&rt.deployCall) != 1 {
		t.Errorf("期望触发 1 次部署, 实际=%d", rt.deployCall)
	}

	// 状态应为 deploying
	spec := specMgr.Get("host-1")
	if spec.Status != protocol.HostStatusDeploying {
		t.Errorf("期望 Status=deploying, 实际=%s", spec.Status)
	}
}

func TestReconciler_RecoverK8sDeploymentExists(t *testing.T) {
	rt := newMockReconcilerRuntime()
	rt.healthy["host-1"] = true // K8s Deployment 存在

	rec, specMgr, _ := newTestReconciler(t, rt)
	specMgr.Create(&protocol.HostSpec{Name: "host-1", Tools: []string{"claude"}, Status: protocol.HostStatusOnline, AuthToken: "t1"})

	rec.RecoverOnStartup(context.Background())

	// Deployment 存在 → 跳过，不重新部署
	if atomic.LoadInt32(&rt.deployCall) != 0 {
		t.Errorf("期望不触发部署, 实际=%d", rt.deployCall)
	}
}

func TestReconciler_RecoverFailedWithRetry(t *testing.T) {
	rt := newMockReconcilerRuntime()

	rec, specMgr, _ := newTestReconciler(t, rt)
	specMgr.Create(&protocol.HostSpec{Name: "host-1", Tools: []string{"claude"}, Status: protocol.HostStatusFailed, RetryCount: 1, AuthToken: "t1"})

	rec.RecoverOnStartup(context.Background())

	// failed + RetryCount < 3 → 重试
	if atomic.LoadInt32(&rt.deployCall) != 1 {
		t.Errorf("期望触发 1 次部署, 实际=%d", rt.deployCall)
	}
}

func TestReconciler_RecoverFailedMaxRetry(t *testing.T) {
	rt := newMockReconcilerRuntime()

	rec, specMgr, _ := newTestReconciler(t, rt)
	specMgr.Create(&protocol.HostSpec{Name: "host-1", Tools: []string{"claude"}, Status: protocol.HostStatusFailed, RetryCount: 3, AuthToken: "t1"})

	rec.RecoverOnStartup(context.Background())

	// failed + RetryCount >= 3 → 跳过
	if atomic.LoadInt32(&rt.deployCall) != 0 {
		t.Errorf("期望不触发部署, 实际=%d", rt.deployCall)
	}
}

// ========== HealthCheck 测试 ==========

func TestReconciler_HealthCheck_DeployingGracePeriod(t *testing.T) {
	rt := newMockReconcilerRuntime()
	rt.healthy["host-1"] = false

	rec, specMgr, _ := newTestReconciler(t, rt)
	// deploying 且 UpdatedAt 在 1 分钟前（保护窗口内）
	specMgr.Create(&protocol.HostSpec{
		Name:      "host-1",
		Tools:     []string{"claude"},
		Status:    protocol.HostStatusDeploying,
		AuthToken: "t1",
		UpdatedAt: time.Now().Add(-1 * time.Minute),
	})

	rec.runHealthCheck(context.Background())

	// 保护窗口内不应触发部署
	if atomic.LoadInt32(&rt.deployCall) != 0 {
		t.Errorf("保护窗口内不应触发部署, 实际=%d", rt.deployCall)
	}
}

func TestReconciler_HealthCheck_DeployingGracePeriodExpired(t *testing.T) {
	rt := newMockReconcilerRuntime()
	rt.healthy["host-1"] = false

	rec, specMgr, _ := newTestReconciler(t, rt)
	// deploying 且 UpdatedAt 在 6 分钟前（超过保护窗口）
	specMgr.Create(&protocol.HostSpec{
		Name:      "host-1",
		Tools:     []string{"claude"},
		Status:    protocol.HostStatusDeploying,
		AuthToken: "t1",
		UpdatedAt: time.Now().Add(-6 * time.Minute),
	})

	rec.runHealthCheck(context.Background())

	// 超过保护窗口且不健康 → 触发重建
	if atomic.LoadInt32(&rt.deployCall) != 1 {
		t.Errorf("保护窗口过期应触发 1 次部署, 实际=%d", rt.deployCall)
	}
}

func TestReconciler_HealthCheck_CrashRebuild(t *testing.T) {
	rt := newMockReconcilerRuntime()
	rt.healthy["host-1"] = false // 容器崩溃

	rec, specMgr, _ := newTestReconciler(t, rt)
	specMgr.Create(&protocol.HostSpec{Name: "host-1", Tools: []string{"claude"}, Status: protocol.HostStatusOnline, AuthToken: "t1"})

	rec.runHealthCheck(context.Background())

	if atomic.LoadInt32(&rt.deployCall) != 1 {
		t.Errorf("期望触发 1 次部署, 实际=%d", rt.deployCall)
	}
}

func TestReconciler_HealthCheck_PendingTimeout(t *testing.T) {
	rt := newMockReconcilerRuntime()

	rec, specMgr, _ := newTestReconciler(t, rt)
	// 创建一个 6 分钟前创建的 pending Host
	spec := &protocol.HostSpec{
		Name:      "host-1",
		Tools:     []string{"claude"},
		Status:    protocol.HostStatusPending,
		AuthToken: "t1",
		CreatedAt: time.Now().Add(-6 * time.Minute),
		UpdatedAt: time.Now().Add(-6 * time.Minute),
	}
	specMgr.Create(spec)

	rec.runHealthCheck(context.Background())

	// 应被标记为 failed
	got := specMgr.Get("host-1")
	if got.Status != protocol.HostStatusFailed {
		t.Errorf("期望 Status=failed, 实际=%s", got.Status)
	}
}

func TestReconciler_HealthCheck_PendingNotTimeout(t *testing.T) {
	rt := newMockReconcilerRuntime()

	rec, specMgr, _ := newTestReconciler(t, rt)
	spec := &protocol.HostSpec{
		Name:      "host-1",
		Tools:     []string{"claude"},
		Status:    protocol.HostStatusPending,
		AuthToken: "t1",
		CreatedAt: time.Now().Add(-3 * time.Minute),
		UpdatedAt: time.Now().Add(-3 * time.Minute),
	}
	specMgr.Create(spec)

	rec.runHealthCheck(context.Background())

	// 不应改变状态
	got := specMgr.Get("host-1")
	if got.Status != protocol.HostStatusPending {
		t.Errorf("期望 Status=pending, 实际=%s", got.Status)
	}
}

func TestReconciler_HealthCheck_FailedRetry(t *testing.T) {
	rt := newMockReconcilerRuntime()

	rec, specMgr, _ := newTestReconciler(t, rt)
	specMgr.Create(&protocol.HostSpec{Name: "host-1", Tools: []string{"claude"}, Status: protocol.HostStatusFailed, RetryCount: 0, AuthToken: "t1"})

	rec.runHealthCheck(context.Background())

	if atomic.LoadInt32(&rt.deployCall) != 1 {
		t.Errorf("期望触发 1 次部署, 实际=%d", rt.deployCall)
	}
}

func TestReconciler_HealthCheck_DeployFailed(t *testing.T) {
	rt := newMockReconcilerRuntime()
	rt.deployErr = errors.New("build error")

	rec, specMgr, _ := newTestReconciler(t, rt)
	specMgr.Create(&protocol.HostSpec{Name: "host-1", Tools: []string{"claude"}, Status: protocol.HostStatusOnline, AuthToken: "t1"})
	rt.healthy["host-1"] = false

	rec.runHealthCheck(context.Background())

	// 部署失败，状态应为 failed
	got := specMgr.Get("host-1")
	if got.Status != protocol.HostStatusFailed {
		t.Errorf("期望 Status=failed, 实际=%s", got.Status)
	}
}
