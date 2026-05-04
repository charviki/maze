package service_test

import (
	"context"
	"errors"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charviki/maze-cradle/configutil"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/internal/config"
	service "github.com/charviki/mesa-hub-behavior-panel/internal/service"
)

type hostServiceRuntimeMock struct {
	mu                   sync.Mutex
	deployDone           chan struct{}
	deployStarted        chan struct{}
	blockDeployUntilDone bool
	deployErr            error
	removeErr            error
	runtimeLogs          string
	runtimeLogsErr       error
	lastDeploySpec       *protocol.HostDeploySpec
	lastDockerfile       string
	removeHostCalls      int
}

func (m *hostServiceRuntimeMock) DeployHost(ctx context.Context, spec *protocol.HostDeploySpec, dockerfileContent string) (*protocol.CreateHostResponse, error) {
	m.mu.Lock()
	m.lastDeploySpec = spec
	m.lastDockerfile = dockerfileContent
	deployDone := m.deployDone
	deployStarted := m.deployStarted
	deployErr := m.deployErr
	m.mu.Unlock()

	if deployDone != nil {
		close(deployDone)
	}
	if deployStarted != nil {
		select {
		case deployStarted <- struct{}{}:
		default:
		}
	}
	if m.blockDeployUntilDone {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	if deployErr != nil {
		return nil, deployErr
	}
	return &protocol.CreateHostResponse{Name: spec.Name, Status: "running"}, nil
}

func (m *hostServiceRuntimeMock) StopHost(ctx context.Context, name string) error {
	return nil
}

func (m *hostServiceRuntimeMock) RemoveHost(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeHostCalls++
	return m.removeErr
}

func (m *hostServiceRuntimeMock) InspectHost(ctx context.Context, name string) (*protocol.ContainerInfo, error) {
	return nil, nil
}

func (m *hostServiceRuntimeMock) GetRuntimeLogs(ctx context.Context, name string, tailLines int) (string, error) {
	return m.runtimeLogs, m.runtimeLogsErr
}

func (m *hostServiceRuntimeMock) IsHealthy(ctx context.Context, name string) (bool, error) {
	return false, nil
}

type auditLoggerRecorder struct {
	mu      sync.Mutex
	entries []protocol.AuditLogEntry
}

func (r *auditLoggerRecorder) Log(entry protocol.AuditLogEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, entry)
}

func (r *auditLoggerRecorder) Entries() []protocol.AuditLogEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]protocol.AuditLogEntry, len(r.entries))
	copy(out, r.entries)
	return out
}

type hostServiceNodeRegistryStub struct {
	mu         sync.RWMutex
	nodes      map[string]*service.Node
	hostTokens map[string]string
}

func newHostServiceNodeRegistryStub() *hostServiceNodeRegistryStub {
	return &hostServiceNodeRegistryStub{
		nodes:      make(map[string]*service.Node),
		hostTokens: make(map[string]string),
	}
}

func (s *hostServiceNodeRegistryStub) StoreHostToken(name, token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hostTokens[name] = token
}

func (s *hostServiceNodeRegistryStub) ValidateHostToken(name, token string) (bool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	expected, ok := s.hostTokens[name]
	if !ok {
		return false, false
	}
	return true, expected == token
}

func (s *hostServiceNodeRegistryStub) RemoveHostToken(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.hostTokens, name)
}

func (s *hostServiceNodeRegistryStub) Register(req protocol.RegisterRequest) *service.Node {
	s.mu.Lock()
	defer s.mu.Unlock()
	node := &service.Node{
		Name:          req.Name,
		Address:       req.Address,
		GrpcAddress:   req.GrpcAddress,
		Status:        service.NodeStatusOnline,
		RegisteredAt:  time.Now(),
		LastHeartbeat: time.Now(),
		Capabilities:  req.Capabilities,
		AgentStatus:   req.Status,
		Metadata:      req.Metadata,
	}
	s.nodes[req.Name] = node
	return node
}

func (s *hostServiceNodeRegistryStub) Heartbeat(req protocol.HeartbeatRequest) *service.Node {
	s.mu.Lock()
	defer s.mu.Unlock()
	node := s.nodes[req.Name]
	if node == nil {
		return nil
	}
	node.LastHeartbeat = time.Now()
	node.AgentStatus = req.Status
	node.Status = service.NodeStatusOnline
	return node
}

func (s *hostServiceNodeRegistryStub) List() []*service.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*service.Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		out = append(out, node)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *hostServiceNodeRegistryStub) Get(name string) *service.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nodes[name]
}

func (s *hostServiceNodeRegistryStub) Delete(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.nodes[name]; !ok {
		return false
	}
	delete(s.nodes, name)
	return true
}

func (s *hostServiceNodeRegistryStub) GetNodeCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.nodes)
}

func (s *hostServiceNodeRegistryStub) GetOnlineCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, node := range s.nodes {
		if node.Status == service.NodeStatusOnline {
			count++
		}
	}
	return count
}

func (s *hostServiceNodeRegistryStub) WaitSave() {}

type hostSpecRepositoryStub struct {
	mu    sync.RWMutex
	specs map[string]*protocol.HostSpec
}

func newHostSpecRepositoryStub() *hostSpecRepositoryStub {
	return &hostSpecRepositoryStub{specs: make(map[string]*protocol.HostSpec)}
}

func (s *hostSpecRepositoryStub) Create(spec *protocol.HostSpec) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.specs[spec.Name]; ok {
		return false
	}
	s.specs[spec.Name] = spec
	return true
}

func (s *hostSpecRepositoryStub) Get(name string) *protocol.HostSpec {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.specs[name]
}

func (s *hostSpecRepositoryStub) List() []*protocol.HostSpec {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*protocol.HostSpec, 0, len(s.specs))
	for _, spec := range s.specs {
		out = append(out, spec)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *hostSpecRepositoryStub) UpdateStatus(name, status, errMsg string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	spec := s.specs[name]
	if spec == nil {
		return false
	}
	spec.Status = status
	spec.ErrorMsg = errMsg
	spec.UpdatedAt = time.Now()
	return true
}

func (s *hostSpecRepositoryStub) Delete(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.specs[name]; !ok {
		return false
	}
	delete(s.specs, name)
	return true
}

func (s *hostSpecRepositoryStub) IncrementRetry(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	spec := s.specs[name]
	if spec == nil {
		return false
	}
	spec.RetryCount++
	spec.UpdatedAt = time.Now()
	return true
}

func (s *hostSpecRepositoryStub) WaitSave() {}

func newHostServiceTestEnv(t *testing.T) (*service.HostService, *hostServiceNodeRegistryStub, *hostSpecRepositoryStub, *hostServiceRuntimeMock, *auditLoggerRecorder) {
	t.Helper()

	tmpDir := t.TempDir()
	registry := newHostServiceNodeRegistryStub()
	hostSpecRepo := newHostSpecRepositoryStub()
	rt := &hostServiceRuntimeMock{}
	auditLog := &auditLoggerRecorder{}
	cfg := &config.Config{
		Server: config.ServerConfig{
			ServerConfig: configutil.ServerConfig{AuthToken: "manager-token"},
		},
		Docker: config.DockerConfig{AgentBaseImage: "maze-agent-base:latest"},
	}
	logDir := filepath.Join(tmpDir, "host-logs")
	svc := service.NewHostService(registry, hostSpecRepo, rt, auditLog, cfg, logutil.NewNop(), logDir)
	return svc, registry, hostSpecRepo, rt, auditLog
}

func waitForCondition(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}

func testHostSpec(name string) *protocol.HostSpec {
	return &protocol.HostSpec{
		Name:        name,
		DisplayName: strings.ToUpper(name),
		Tools:       []string{"claude"},
		AuthToken:   name + "-token",
		Status:      protocol.HostStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func TestHostService_CreateHost_PersistsSpecAndToken(t *testing.T) {
	svc, registry, specMgr, rt, _ := newHostServiceTestEnv(t)
	rt.deployDone = make(chan struct{})

	spec, err := svc.CreateHost(context.Background(), &protocol.CreateHostRequest{
		Name:        "host-1",
		DisplayName: "Host One",
		Tools:       []string{"claude", "go"},
		Resources: protocol.ResourceLimits{
			CPULimit:    "2",
			MemoryLimit: "4Gi",
		},
	})
	if err != nil {
		t.Fatalf("CreateHost 返回错误: %v", err)
	}
	if spec.AuthToken == "" {
		t.Fatal("CreateHost 应生成 host token")
	}

	select {
	case <-rt.deployDone:
	case <-time.After(2 * time.Second):
		t.Fatal("等待异步部署超时")
	}

	waitForCondition(t, time.Second, func() bool {
		got := specMgr.Get("host-1")
		return got != nil && got.Status == protocol.HostStatusDeploying
	})

	saved := specMgr.Get("host-1")
	if saved == nil {
		t.Fatal("CreateHost 后应持久化 HostSpec")
	}
	if saved.DisplayName != "Host One" {
		t.Fatalf("DisplayName = %q, want %q", saved.DisplayName, "Host One")
	}

	exists, matched := registry.ValidateHostToken("host-1", spec.AuthToken)
	if !exists || !matched {
		t.Fatalf("host token 未正确预存: exists=%v matched=%v", exists, matched)
	}

	if rt.lastDeploySpec == nil {
		t.Fatal("异步部署应调用 runtime.DeployHost")
	}
	if rt.lastDeploySpec.ServerAuthToken != "manager-token" {
		t.Fatalf("ServerAuthToken = %q, want %q", rt.lastDeploySpec.ServerAuthToken, "manager-token")
	}
	if rt.lastDeploySpec.Resources.MemoryLimit != "4Gi" {
		t.Fatalf("MemoryLimit = %q, want %q", rt.lastDeploySpec.Resources.MemoryLimit, "4Gi")
	}
	if !strings.Contains(rt.lastDockerfile, "maze-agent-base:latest") {
		t.Fatalf("dockerfile 未使用测试基础镜像: %q", rt.lastDockerfile)
	}
}

func TestHostService_CreateHost_RejectsUnknownTools(t *testing.T) {
	svc, _, _, _, _ := newHostServiceTestEnv(t)

	_, err := svc.CreateHost(context.Background(), &protocol.CreateHostRequest{
		Name:  "host-1",
		Tools: []string{"unknown-tool"},
	})
	if err == nil {
		t.Fatal("未知工具应返回错误")
	}
	if !strings.Contains(err.Error(), "unknown tools") {
		t.Fatalf("error = %q, want contains %q", err.Error(), "unknown tools")
	}
}

func TestHostService_DeleteHost_RemovesStateAndWritesSuccessAudit(t *testing.T) {
	svc, registry, specMgr, rt, auditLog := newHostServiceTestEnv(t)
	spec := testHostSpec("host-delete-ok")

	if !specMgr.Create(spec) {
		t.Fatal("预置 HostSpec 失败")
	}
	registry.StoreHostToken(spec.Name, spec.AuthToken)
	registry.Register(protocol.RegisterRequest{Name: spec.Name, Address: "http://host-delete-ok"})

	if err := svc.DeleteHost(context.Background(), spec.Name); err != nil {
		t.Fatalf("DeleteHost 返回错误: %v", err)
	}
	if specMgr.Get(spec.Name) != nil {
		t.Fatal("DeleteHost 成功后应删除 HostSpec")
	}
	if registry.Get(spec.Name) != nil {
		t.Fatal("DeleteHost 成功后应删除节点注册信息")
	}
	exists, _ := registry.ValidateHostToken(spec.Name, spec.AuthToken)
	if exists {
		t.Fatal("DeleteHost 成功后应清理 host token")
	}

	entries := auditLog.Entries()
	if len(entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(entries))
	}
	if entries[0].Result != "success" {
		t.Fatalf("audit result = %q, want %q", entries[0].Result, "success")
	}
	if rt.removeHostCalls != 1 {
		t.Fatalf("RemoveHost calls = %d, want 1", rt.removeHostCalls)
	}
}

func TestHostService_DeleteHost_RemoveFailureKeepsState(t *testing.T) {
	svc, registry, specMgr, rt, auditLog := newHostServiceTestEnv(t)
	spec := testHostSpec("host-delete-failed")
	rt.removeErr = errors.New("runtime down")

	if !specMgr.Create(spec) {
		t.Fatal("预置 HostSpec 失败")
	}
	registry.StoreHostToken(spec.Name, spec.AuthToken)
	registry.Register(protocol.RegisterRequest{Name: spec.Name, Address: "http://host-delete-failed"})

	err := svc.DeleteHost(context.Background(), spec.Name)
	if err == nil {
		t.Fatal("底层删除失败时应返回错误")
	}
	if !strings.Contains(err.Error(), "remove host") {
		t.Fatalf("error = %q, want contains %q", err.Error(), "remove host")
	}

	if specMgr.Get(spec.Name) == nil {
		t.Fatal("删除失败时应保留 HostSpec，避免失去控制面记录")
	}
	if registry.Get(spec.Name) == nil {
		t.Fatal("删除失败时应保留节点记录")
	}
	exists, matched := registry.ValidateHostToken(spec.Name, spec.AuthToken)
	if !exists || !matched {
		t.Fatalf("删除失败时应保留 host token: exists=%v matched=%v", exists, matched)
	}

	entries := auditLog.Entries()
	if len(entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(entries))
	}
	if entries[0].Result != "failed" {
		t.Fatalf("audit result = %q, want %q", entries[0].Result, "failed")
	}
}

func TestHostService_GetBuildLog_MissingFileReturnsEmpty(t *testing.T) {
	svc, _, specMgr, _, _ := newHostServiceTestEnv(t)

	if !specMgr.Create(testHostSpec("host-build-log")) {
		t.Fatal("预置 HostSpec 失败")
	}

	log, err := svc.GetBuildLog(context.Background(), "host-build-log")
	if err != nil {
		t.Fatalf("GetBuildLog 返回错误: %v", err)
	}
	if log != "" {
		t.Fatalf("build log = %q, want empty", log)
	}
}

func TestHostService_GetRuntimeLog_WrapsRuntimeError(t *testing.T) {
	svc, _, specMgr, rt, _ := newHostServiceTestEnv(t)
	rt.runtimeLogsErr = errors.New("logs unavailable")

	if !specMgr.Create(testHostSpec("host-runtime-log")) {
		t.Fatal("预置 HostSpec 失败")
	}

	_, err := svc.GetRuntimeLog(context.Background(), "host-runtime-log")
	if err == nil {
		t.Fatal("runtime 日志失败时应返回错误")
	}
	if !strings.Contains(err.Error(), "get runtime logs") {
		t.Fatalf("error = %q, want contains %q", err.Error(), "get runtime logs")
	}
}

func TestHostService_StopCancelsAllInFlightDeployments(t *testing.T) {
	svc, _, _, rt, _ := newHostServiceTestEnv(t)
	rt.blockDeployUntilDone = true
	rt.deployStarted = make(chan struct{}, 2)

	if _, err := svc.CreateHost(context.Background(), &protocol.CreateHostRequest{
		Name:  "host-stop-1",
		Tools: []string{"claude"},
	}); err != nil {
		t.Fatalf("create first host: %v", err)
	}
	if _, err := svc.CreateHost(context.Background(), &protocol.CreateHostRequest{
		Name:  "host-stop-2",
		Tools: []string{"go"},
	}); err != nil {
		t.Fatalf("create second host: %v", err)
	}

	for range 2 {
		select {
		case <-rt.deployStarted:
		case <-time.After(2 * time.Second):
			t.Fatal("waiting for async deployment start timed out")
		}
	}

	done := make(chan struct{})
	go func() {
		svc.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop should cancel and wait for all in-flight deployments")
	}
}
