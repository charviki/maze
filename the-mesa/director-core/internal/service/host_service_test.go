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
	"github.com/charviki/maze/the-mesa/director-core/internal/config"
	service "github.com/charviki/maze/the-mesa/director-core/internal/service"
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

func (r *auditLoggerRecorder) Log(_ context.Context, entry protocol.AuditLogEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, entry)
	return nil
}

func (r *auditLoggerRecorder) Entries() []protocol.AuditLogEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]protocol.AuditLogEntry, len(r.entries))
	copy(out, r.entries)
	return out
}

type hostServiceNodeRegistryStub struct {
	mu                 sync.RWMutex
	nodes              map[string]*service.Node
	hostTokens         map[string]string
	storeHostTokenErr  error
	removeHostTokenErr error
	getErr             error
	deleteErr          error
}

func newHostServiceNodeRegistryStub() *hostServiceNodeRegistryStub {
	return &hostServiceNodeRegistryStub{
		nodes:      make(map[string]*service.Node),
		hostTokens: make(map[string]string),
	}
}

func (s *hostServiceNodeRegistryStub) StoreHostToken(_ context.Context, name, token string) error {
	if s.storeHostTokenErr != nil {
		return s.storeHostTokenErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hostTokens[name] = token
	return nil
}

func (s *hostServiceNodeRegistryStub) ValidateHostToken(_ context.Context, name, token string) (bool, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	expected, ok := s.hostTokens[name]
	if !ok {
		return false, false, nil
	}
	return true, expected == token, nil
}

func (s *hostServiceNodeRegistryStub) RemoveHostToken(_ context.Context, name string) error {
	if s.removeHostTokenErr != nil {
		return s.removeHostTokenErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.hostTokens, name)
	return nil
}

func (s *hostServiceNodeRegistryStub) Register(_ context.Context, req protocol.RegisterRequest) (*service.Node, error) {
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
	return node, nil
}

func (s *hostServiceNodeRegistryStub) Heartbeat(_ context.Context, req protocol.HeartbeatRequest) (*service.Node, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	node := s.nodes[req.Name]
	if node == nil {
		return nil, nil
	}
	node.LastHeartbeat = time.Now()
	node.AgentStatus = req.Status
	node.Status = service.NodeStatusOnline
	return node, nil
}

func (s *hostServiceNodeRegistryStub) List(_ context.Context) ([]*service.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*service.Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		out = append(out, node)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s *hostServiceNodeRegistryStub) Get(_ context.Context, name string) (*service.Node, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nodes[name], nil
}

func (s *hostServiceNodeRegistryStub) Delete(_ context.Context, name string) (bool, error) {
	if s.deleteErr != nil {
		return false, s.deleteErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.nodes[name]; !ok {
		return false, nil
	}
	delete(s.nodes, name)
	return true, nil
}

func (s *hostServiceNodeRegistryStub) GetNodeCount(_ context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.nodes), nil
}

func (s *hostServiceNodeRegistryStub) GetOnlineCount(_ context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, node := range s.nodes {
		if node.Status == service.NodeStatusOnline {
			count++
		}
	}
	return count, nil
}

type hostSpecRepositoryStub struct {
	mu        sync.RWMutex
	specs     map[string]*protocol.HostSpec
	createErr error
	getErr    error
	listErr   error
	deleteErr error
}

func newHostSpecRepositoryStub() *hostSpecRepositoryStub {
	return &hostSpecRepositoryStub{specs: make(map[string]*protocol.HostSpec)}
}

func (s *hostSpecRepositoryStub) Create(_ context.Context, spec *protocol.HostSpec) (bool, error) {
	if s.createErr != nil {
		return false, s.createErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.specs[spec.Name]; ok {
		return false, nil
	}
	s.specs[spec.Name] = spec
	return true, nil
}

func (s *hostSpecRepositoryStub) Get(_ context.Context, name string) (*protocol.HostSpec, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.specs[name], nil
}

func (s *hostSpecRepositoryStub) List(_ context.Context) ([]*protocol.HostSpec, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*protocol.HostSpec, 0, len(s.specs))
	for _, spec := range s.specs {
		out = append(out, spec)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s *hostSpecRepositoryStub) UpdateStatus(_ context.Context, name, status, errMsg string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	spec := s.specs[name]
	if spec == nil {
		return false, nil
	}
	spec.Status = status
	spec.ErrorMsg = errMsg
	spec.UpdatedAt = time.Now()
	return true, nil
}

func (s *hostSpecRepositoryStub) Delete(_ context.Context, name string) (bool, error) {
	if s.deleteErr != nil {
		return false, s.deleteErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.specs[name]; !ok {
		return false, nil
	}
	delete(s.specs, name)
	return true, nil
}

func (s *hostSpecRepositoryStub) IncrementRetry(_ context.Context, name string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	spec := s.specs[name]
	if spec == nil {
		return false, nil
	}
	spec.RetryCount++
	spec.UpdatedAt = time.Now()
	return true, nil
}

type hostServiceTxManagerStub struct {
	registry *hostServiceNodeRegistryStub
	specRepo *hostSpecRepositoryStub
}

func (m *hostServiceTxManagerStub) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	specSnapshot := m.specRepo.snapshot()
	nodeSnapshot, tokenSnapshot := m.registry.snapshot()
	if err := fn(ctx); err != nil {
		m.specRepo.restore(specSnapshot)
		m.registry.restore(nodeSnapshot, tokenSnapshot)
		return err
	}
	return nil
}

func (s *hostServiceNodeRegistryStub) snapshot() (map[string]*service.Node, map[string]string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	nodes := make(map[string]*service.Node, len(s.nodes))
	for name, node := range s.nodes {
		if node == nil {
			nodes[name] = nil
			continue
		}
		cloned := *node
		nodes[name] = &cloned
	}
	tokens := make(map[string]string, len(s.hostTokens))
	for name, token := range s.hostTokens {
		tokens[name] = token
	}
	return nodes, tokens
}

func (s *hostServiceNodeRegistryStub) restore(nodes map[string]*service.Node, tokens map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodes = nodes
	s.hostTokens = tokens
}

func (s *hostSpecRepositoryStub) snapshot() map[string]*protocol.HostSpec {
	s.mu.RLock()
	defer s.mu.RUnlock()
	specs := make(map[string]*protocol.HostSpec, len(s.specs))
	for name, spec := range s.specs {
		if spec == nil {
			specs[name] = nil
			continue
		}
		cloned := *spec
		specs[name] = &cloned
	}
	return specs
}

func (s *hostSpecRepositoryStub) restore(specs map[string]*protocol.HostSpec) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.specs = specs
}

func newHostServiceTestEnv(t *testing.T) (*service.HostService, *hostServiceNodeRegistryStub, *hostSpecRepositoryStub, *hostServiceRuntimeMock, *auditLoggerRecorder) {
	t.Helper()

	tmpDir := t.TempDir()
	registry := newHostServiceNodeRegistryStub()
	hostSpecRepo := newHostSpecRepositoryStub()
	rt := &hostServiceRuntimeMock{}
	auditLog := &auditLoggerRecorder{}
	txm := &hostServiceTxManagerStub{registry: registry, specRepo: hostSpecRepo}
	cfg := &config.Config{
		Server: config.ServerConfig{
			ServerConfig: configutil.ServerConfig{AuthToken: "director-core-token"},
		},
		Docker: config.DockerConfig{AgentBaseImage: "maze-agent-base:latest"},
	}
	logDir := filepath.Join(tmpDir, "host-logs")
	svc := service.NewHostService(registry, hostSpecRepo, txm, rt, auditLog, cfg, logutil.NewNop(), logDir)
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
		got, _ := specMgr.Get(context.Background(), "host-1")
		return got != nil && got.Status == protocol.HostStatusDeploying
	})

	saved, _ := specMgr.Get(context.Background(), "host-1")
	if saved == nil {
		t.Fatal("CreateHost 后应持久化 HostSpec")
	}
	if saved.DisplayName != "Host One" {
		t.Fatalf("DisplayName = %q, want %q", saved.DisplayName, "Host One")
	}

	exists, matched, _ := registry.ValidateHostToken(context.Background(), "host-1", spec.AuthToken)
	if !exists || !matched {
		t.Fatalf("host token 未正确预存: exists=%v matched=%v", exists, matched)
	}

	if rt.lastDeploySpec == nil {
		t.Fatal("异步部署应调用 runtime.DeployHost")
	}
	if rt.lastDeploySpec.ServerAuthToken != "director-core-token" {
		t.Fatalf("ServerAuthToken = %q, want %q", rt.lastDeploySpec.ServerAuthToken, "director-core-token")
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

func TestHostService_CreateHost_StoreTokenFailureRollsBackSpec(t *testing.T) {
	svc, registry, specMgr, _, _ := newHostServiceTestEnv(t)
	registry.storeHostTokenErr = errors.New("token store down")

	_, err := svc.CreateHost(context.Background(), &protocol.CreateHostRequest{
		Name:  "host-rollback",
		Tools: []string{"claude"},
	})
	if err == nil {
		t.Fatal("StoreHostToken 失败时应返回错误")
	}
	if !strings.Contains(err.Error(), "store host token") {
		t.Fatalf("error = %q, want contains %q", err.Error(), "store host token")
	}
	if got, _ := specMgr.Get(context.Background(), "host-rollback"); got != nil {
		t.Fatal("事务失败后不应残留 HostSpec")
	}
}

func TestHostService_DeleteHost_RemovesStateAndWritesSuccessAudit(t *testing.T) {
	svc, registry, specMgr, rt, auditLog := newHostServiceTestEnv(t)
	spec := testHostSpec("host-delete-ok")

	if ok, _ := specMgr.Create(context.Background(), spec); !ok {
		t.Fatal("预置 HostSpec 失败")
	}
	_ = registry.StoreHostToken(context.Background(), spec.Name, spec.AuthToken)
	_, _ = registry.Register(context.Background(), protocol.RegisterRequest{Name: spec.Name, Address: "http://host-delete-ok"})

	if err := svc.DeleteHost(context.Background(), spec.Name); err != nil {
		t.Fatalf("DeleteHost 返回错误: %v", err)
	}
	if got, _ := specMgr.Get(context.Background(), spec.Name); got != nil {
		t.Fatal("DeleteHost 成功后应删除 HostSpec")
	}
	if got, _ := registry.Get(context.Background(), spec.Name); got != nil {
		t.Fatal("DeleteHost 成功后应删除节点注册信息")
	}
	exists, _, _ := registry.ValidateHostToken(context.Background(), spec.Name, spec.AuthToken)
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

	if ok, _ := specMgr.Create(context.Background(), spec); !ok {
		t.Fatal("预置 HostSpec 失败")
	}
	_ = registry.StoreHostToken(context.Background(), spec.Name, spec.AuthToken)
	_, _ = registry.Register(context.Background(), protocol.RegisterRequest{Name: spec.Name, Address: "http://host-delete-failed"})

	err := svc.DeleteHost(context.Background(), spec.Name)
	if err == nil {
		t.Fatal("底层删除失败时应返回错误")
	}
	if !strings.Contains(err.Error(), "remove host") {
		t.Fatalf("error = %q, want contains %q", err.Error(), "remove host")
	}

	if got, _ := specMgr.Get(context.Background(), spec.Name); got == nil {
		t.Fatal("删除失败时应保留 HostSpec，避免失去控制面记录")
	}
	if got, _ := registry.Get(context.Background(), spec.Name); got == nil {
		t.Fatal("删除失败时应保留节点记录")
	}
	exists, matched, _ := registry.ValidateHostToken(context.Background(), spec.Name, spec.AuthToken)
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

func TestHostService_DeleteHost_PersistFailureRollsBackState(t *testing.T) {
	svc, registry, specMgr, _, auditLog := newHostServiceTestEnv(t)
	spec := testHostSpec("host-delete-rollback")
	specMgr.deleteErr = errors.New("delete spec failed")

	if ok, _ := specMgr.Create(context.Background(), spec); !ok {
		t.Fatal("预置 HostSpec 失败")
	}
	_ = registry.StoreHostToken(context.Background(), spec.Name, spec.AuthToken)
	_, _ = registry.Register(context.Background(), protocol.RegisterRequest{Name: spec.Name, Address: "http://host-delete-rollback"})

	err := svc.DeleteHost(context.Background(), spec.Name)
	if err == nil {
		t.Fatal("持久化删除失败时应返回错误")
	}
	if !strings.Contains(err.Error(), "delete host spec") {
		t.Fatalf("error = %q, want contains %q", err.Error(), "delete host spec")
	}
	if got, _ := specMgr.Get(context.Background(), spec.Name); got == nil {
		t.Fatal("事务回滚后应保留 HostSpec")
	}
	if got, _ := registry.Get(context.Background(), spec.Name); got == nil {
		t.Fatal("事务回滚后应恢复节点记录")
	}
	exists, matched, _ := registry.ValidateHostToken(context.Background(), spec.Name, spec.AuthToken)
	if !exists || !matched {
		t.Fatalf("事务回滚后应恢复 host token: exists=%v matched=%v", exists, matched)
	}
	if len(auditLog.Entries()) != 0 {
		t.Fatal("持久化删除失败时不应写入成功审计")
	}
}

func TestHostService_GetBuildLog_MissingFileReturnsEmpty(t *testing.T) {
	svc, _, specMgr, _, _ := newHostServiceTestEnv(t)

	if ok, _ := specMgr.Create(context.Background(), testHostSpec("host-build-log")); !ok {
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

	if ok, _ := specMgr.Create(context.Background(), testHostSpec("host-runtime-log")); !ok {
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
