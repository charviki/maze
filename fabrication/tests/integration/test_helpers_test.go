//go:build integration

package integration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	client "github.com/charviki/maze-cradle/api/gen/http"
	"github.com/charviki/maze-integration-tests/kit"
)

type testHelper struct {
	apiClient *client.APIClient
	cfg       *kit.TestConfig
	isolated  []string
	leased    []hostLease
}

type hostLease struct {
	name     string
	profile  string
	snapshot hostSnapshot
}

type hostSnapshot struct {
	sessionIDs  map[string]struct{}
	templateIDs map[string]struct{}
	env         map[string]string
}

func newTestHelper(t *testing.T) *testHelper {
	t.Helper()
	cfg := suite.cfg
	if cfg == nil {
		cfg = kit.LoadTestConfig()
	}
	apiClient := kit.NewTestAPIClient(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, err := apiClient.HostServiceAPI.HostServiceListHosts(ctx).Execute()
	if err != nil {
		t.Fatalf("Manager API not available at %s: %v", cfg.ManagerURL, err)
	}

	return &testHelper{apiClient: apiClient, cfg: cfg}
}

func (h *testHelper) cleanup(t *testing.T) {
	t.Helper()
	var cleanupErr error
	for i := len(h.leased) - 1; i >= 0; i-- {
		lease := h.leased[i]
		t.Logf("[step] cleanup: restoring shared host %s", lease.name)
		restoreErr := h.restoreSharedHostState(t, lease)
		if suite.pool != nil {
			// 即使恢复失败也必须归还租约，否则并行测试会永久阻塞在 Acquire 上。
			if err := suite.pool.Release(lease.profile, lease.name); err != nil {
				cleanupErr = errors.Join(cleanupErr, fmt.Errorf("release shared host %s failed: %w", lease.name, err))
			}
		}
		if restoreErr != nil {
			cleanupErr = errors.Join(cleanupErr, restoreErr)
		}
	}
	t.Logf("[step] cleanup: deleting %d isolated hosts", len(h.isolated))
	for _, name := range h.isolated {
		h.ensureHostRemoved(t, name, 45*time.Second)
	}
	if cleanupErr != nil {
		t.Fatalf("shared host cleanup failed: %v", cleanupErr)
	}
}

func (h *testHelper) trackHost(name string) {
	h.isolated = append(h.isolated, name)
}

func (h *testHelper) acquireHost(t *testing.T, profile string) string {
	t.Helper()
	if suite.pool == nil {
		if h.cfg.EnableHostPool {
			t.Fatalf("shared host pool is enabled but not initialized for profile %s", profile)
		}
		tools, err := toolsForProfile(profile)
		if err != nil {
			t.Fatalf("resolve tools for profile %s: %v", profile, err)
		}
		name := uniqueName("shared-" + profile)
		h.trackHost(name)
		h.createHostAndWait(t, name, tools)
		return name
	}

	name, err := suite.pool.Acquire(profile)
	if err != nil {
		t.Fatalf("acquire shared host for profile %s: %v", profile, err)
	}
	snapshot, err := h.captureHostSnapshot(name)
	if err != nil {
		_ = suite.pool.Release(profile, name)
		t.Fatalf("capture snapshot for shared host %s: %v", name, err)
	}
	h.leased = append(h.leased, hostLease{name: name, profile: profile, snapshot: snapshot})
	t.Logf("[step] leased shared host %s profile=%s", name, profile)
	return name
}

func uniqueName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixMilli())
}

func (h *testHelper) createHostAndWait(t *testing.T, name string, tools []string) {
	t.Helper()

	t.Logf("[step] creating host %s with tools=%v...", name, tools)
	nameField := name
	body := client.V1CreateHostRequest{
		Name:  &nameField,
		Tools: tools,
	}

	ctx := context.Background()
	resp, httpResp, err := h.apiClient.HostServiceAPI.HostServiceCreateHost(ctx).Body(body).Execute()
	if err != nil {
		t.Fatalf("create host failed: %v (status=%d)", err, httpResp.StatusCode)
	}
	t.Logf("[step] CreateHost response: name=%s status=%s", resp.GetName(), resp.GetStatus())
	if gotTools := resp.GetTools(); len(gotTools) != len(tools) {
		t.Fatalf("expected CreateHost to echo %d tools, got %d", len(tools), len(gotTools))
	}

	h.waitForHostStatus(t, name, "online", 3*time.Minute)
	h.waitForHostAPIReady(t, name, 30*time.Second)
	h.waitForManagerDataLayout(t, 45*time.Second)
}

func (h *testHelper) waitForHostStatus(t *testing.T, name, targetStatus string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	attempt := 0
	for time.Now().Before(deadline) {
		attempt++
		info, _, err := h.apiClient.HostServiceAPI.HostServiceGetHost(context.Background(), name).Execute()
		if err != nil {
			t.Logf("    [wait] host=%s attempt=%d error=%v", name, attempt, err)
			time.Sleep(2 * time.Second)
			continue
		}
		if info.GetStatus() == targetStatus {
			t.Logf("    [wait] host=%s reached %s after %d attempts", name, targetStatus, attempt)
			return
		}
		t.Logf("    [wait] host=%s status=%s (want %s), attempt=%d", name, info.GetStatus(), targetStatus, attempt)
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("wait for host %s status %s: timeout after %v", name, targetStatus, timeout)
}

func (h *testHelper) waitForHostAPIReady(t *testing.T, nodeName string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	attempt := 0
	for time.Now().Before(deadline) {
		attempt++
		_, _, err := h.apiClient.SessionServiceAPI.SessionServiceListSessions(context.Background(), nodeName).Execute()
		if err == nil {
			t.Logf("    [wait] host=%s API ready after %d attempts", nodeName, attempt)
			return
		}
		// K8s 中节点可能已标记 online，但代理到 Agent 的 HTTP 服务还在完成最后启动。
		t.Logf("    [wait] host=%s API not ready yet, attempt=%d err=%v", nodeName, attempt, err)
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("wait for host %s API ready: timeout after %v", nodeName, timeout)
}

func (h *testHelper) ensureHostRemoved(t *testing.T, name string, timeout time.Duration) {
	t.Helper()
	gone, err := h.isHostGone(context.Background(), name)
	if err != nil {
		t.Fatalf("check host %s before cleanup failed: %v", name, err)
	}
	if !gone {
		// 测试显式删除 Host 时不会再走这里；cleanup 只负责兜底并严格校验结果。
		_, httpResp, err := h.apiClient.HostServiceAPI.HostServiceDeleteHost(context.Background(), name).Execute()
		if err != nil {
			statusCode := 0
			if httpResp != nil {
				statusCode = httpResp.StatusCode
			}
			t.Fatalf("cleanup delete host %s failed: %v (status=%d)", name, err, statusCode)
		}
	}
	h.waitForHostGone(t, name, timeout)
	if h.shouldAssertHostDataRemoval() {
		h.waitForHostDataRemoved(t, name, timeout)
	} else {
		// PVC 等非宿主机可见存储没有稳定文件路径可查，这里只验证对外可观察到的 Host 删除结果。
		t.Logf("    [wait] host=%s skips host data dir assertion for storage backend=%s", name, h.cfg.AgentStorageBackend)
	}
}

func (h *testHelper) waitForHostGone(t *testing.T, name string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		gone, err := h.isHostGone(context.Background(), name)
		if err == nil && gone {
			t.Logf("    [wait] host=%s deleted", name)
			return
		}
		if err != nil {
			t.Logf("    [wait] host=%s delete check error=%v", name, err)
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("wait for host %s deletion: timeout after %v", name, timeout)
}

func (h *testHelper) waitForHostDataRemoved(t *testing.T, name string, timeout time.Duration) {
	t.Helper()
	hostDir := h.hostDataDir(name)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(hostDir); os.IsNotExist(err) {
			t.Logf("    [wait] host=%s data dir removed: %s", name, hostDir)
			return
		} else if err != nil {
			t.Fatalf("stat host data dir %s: %v", hostDir, err)
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("wait for host %s data dir removal: timeout after %v (dir=%s)", name, timeout, hostDir)
}

func (h *testHelper) isHostGone(ctx context.Context, name string) (bool, error) {
	_, httpResp, err := h.apiClient.HostServiceAPI.HostServiceGetHost(ctx, name).Execute()
	if err == nil {
		return false, nil
	}
	if httpResp != nil && httpResp.StatusCode == 404 {
		return true, nil
	}
	return false, err
}

func (h *testHelper) hostDataDir(name string) string {
	// DataDir 指向测试根目录；不同运行时各自放在 docker/kubernetes 子目录下，避免路径硬编码散落在用例里。
	return filepath.Join(h.cfg.DataDir, h.cfg.Env, "agents", name)
}

func (h *testHelper) shouldAssertHostDataRemoval() bool {
	switch h.cfg.AgentStorageBackend {
	case "bind", "hostpath":
		return true
	default:
		return false
	}
}

func (h *testHelper) waitForManagerDataLayout(t *testing.T, timeout time.Duration) {
	t.Helper()
}

func (h *testHelper) createSession(t *testing.T, nodeName, sessionName string) string {
	return h.createSessionWithTemplate(t, nodeName, sessionName, "")
}

func (h *testHelper) createSessionWithTemplate(t *testing.T, nodeName, sessionName, templateID string) string {
	t.Helper()
	nameField := sessionName
	workingDir := "/home/agent/project"
	body := client.SessionServiceCreateSessionBody{
		Name:       &nameField,
		WorkingDir: &workingDir,
	}
	if templateID != "" {
		// 配置类用例必须绑定模板，服务端才知道应暴露哪些固定文件。
		body.TemplateId = &templateID
	}
	ctx := context.Background()
	session, httpResp, err := h.apiClient.SessionServiceAPI.SessionServiceCreateSession(ctx, nodeName).Body(body).Execute()
	if err != nil {
		t.Fatalf("create session failed: %v (status=%d)", err, httpResp.StatusCode)
	}
	sid := session.GetId()
	t.Logf("[step] created session %s on %s → id=%s", sessionName, nodeName, sid)
	return sid
}

func ptrStr(s string) *string {
	return &s
}

func (h *testHelper) captureHostSnapshot(nodeName string) (hostSnapshot, error) {
	snapshot := hostSnapshot{
		sessionIDs:  make(map[string]struct{}),
		templateIDs: make(map[string]struct{}),
		env:         make(map[string]string),
	}

	sessions, _, err := h.apiClient.SessionServiceAPI.SessionServiceListSessions(context.Background(), nodeName).Execute()
	if err != nil {
		return hostSnapshot{}, fmt.Errorf("list sessions on %s: %w", nodeName, err)
	}
	for _, session := range sessions.GetSessions() {
		snapshot.sessionIDs[session.GetId()] = struct{}{}
	}

	templates, _, err := h.apiClient.TemplateServiceAPI.TemplateServiceListTemplates(context.Background(), nodeName).Execute()
	if err != nil {
		return hostSnapshot{}, fmt.Errorf("list templates on %s: %w", nodeName, err)
	}
	for _, tmpl := range templates.GetTemplates() {
		snapshot.templateIDs[tmpl.GetId()] = struct{}{}
	}

	configView, _, err := h.apiClient.ConfigServiceAPI.ConfigServiceGetConfig(context.Background(), nodeName).Execute()
	if err != nil {
		return hostSnapshot{}, fmt.Errorf("get local config on %s: %w", nodeName, err)
	}
	for key, value := range configView.GetEnv() {
		snapshot.env[key] = value
	}

	return snapshot, nil
}

func (h *testHelper) restoreSharedHostState(t *testing.T, lease hostLease) error {
	t.Helper()
	var restoreErr error
	if err := h.deleteExtraSessions(t, lease); err != nil {
		restoreErr = errors.Join(restoreErr, err)
	}
	if err := h.deleteExtraTemplates(t, lease); err != nil {
		restoreErr = errors.Join(restoreErr, err)
	}
	if err := h.restoreLocalConfigEnv(t, lease); err != nil {
		restoreErr = errors.Join(restoreErr, err)
	}
	return restoreErr
}

func (h *testHelper) deleteExtraSessions(t *testing.T, lease hostLease) error {
	t.Helper()
	sessions, _, err := h.apiClient.SessionServiceAPI.SessionServiceListSessions(context.Background(), lease.name).Execute()
	if err != nil {
		return fmt.Errorf("list sessions for shared host %s cleanup failed: %w", lease.name, err)
	}
	for _, session := range sessions.GetSessions() {
		if _, ok := lease.snapshot.sessionIDs[session.GetId()]; ok {
			continue
		}
		t.Logf("[step] cleanup: deleting shared-host session %s on %s", session.GetId(), lease.name)
		_, _, err := h.apiClient.SessionServiceAPI.SessionServiceDeleteSession(context.Background(), lease.name, session.GetId()).Execute()
		if err != nil {
			return fmt.Errorf("delete session %s on shared host %s failed: %w", session.GetId(), lease.name, err)
		}
	}
	return nil
}

func (h *testHelper) deleteExtraTemplates(t *testing.T, lease hostLease) error {
	t.Helper()
	templates, _, err := h.apiClient.TemplateServiceAPI.TemplateServiceListTemplates(context.Background(), lease.name).Execute()
	if err != nil {
		return fmt.Errorf("list templates for shared host %s cleanup failed: %w", lease.name, err)
	}
	for _, tmpl := range templates.GetTemplates() {
		if tmpl.GetBuiltin() {
			continue
		}
		if _, ok := lease.snapshot.templateIDs[tmpl.GetId()]; ok {
			continue
		}
		t.Logf("[step] cleanup: deleting shared-host template %s on %s", tmpl.GetId(), lease.name)
		_, _, err := h.apiClient.TemplateServiceAPI.TemplateServiceDeleteTemplate(context.Background(), lease.name, tmpl.GetId()).Execute()
		if err != nil {
			return fmt.Errorf("delete template %s on shared host %s failed: %w", tmpl.GetId(), lease.name, err)
		}
	}
	return nil
}

func (h *testHelper) restoreLocalConfigEnv(t *testing.T, lease hostLease) error {
	t.Helper()
	configView, _, err := h.apiClient.ConfigServiceAPI.ConfigServiceGetConfig(context.Background(), lease.name).Execute()
	if err != nil {
		return fmt.Errorf("get local config for shared host %s cleanup failed: %w", lease.name, err)
	}
	patch := diffEnv(configView.GetEnv(), lease.snapshot.env)
	if len(patch) == 0 {
		return nil
	}
	t.Logf("[step] cleanup: restoring %d env keys on shared host %s", len(patch), lease.name)
	_, _, err = h.apiClient.ConfigServiceAPI.ConfigServiceUpdateConfig(context.Background(), lease.name).
		Body(client.ConfigServiceUpdateConfigBody{Env: &patch}).Execute()
	if err != nil {
		return fmt.Errorf("restore local config for shared host %s failed: %w", lease.name, err)
	}
	return nil
}

func diffEnv(current, target map[string]string) map[string]string {
	patch := make(map[string]string)
	for key, currentValue := range current {
		targetValue, ok := target[key]
		if !ok {
			patch[key] = ""
			continue
		}
		if currentValue != targetValue {
			patch[key] = targetValue
		}
	}
	for key, targetValue := range target {
		if currentValue, ok := current[key]; !ok || currentValue != targetValue {
			patch[key] = targetValue
		}
	}
	return patch
}
