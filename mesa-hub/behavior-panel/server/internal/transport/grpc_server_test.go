package transport

import (
	"context"
	"math"
	"path/filepath"
	"testing"
	"time"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/configutil"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/internal/config"
	auditrepo "github.com/charviki/mesa-hub-behavior-panel/internal/repository/audit"
	filerepo "github.com/charviki/mesa-hub-behavior-panel/internal/repository/file"
	"github.com/charviki/mesa-hub-behavior-panel/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAttachBearerToken(t *testing.T) {
	ctx := attachBearerToken(context.Background(), "secret-token")

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}

	got := md.Get("authorization")
	if len(got) != 1 || got[0] != "Bearer secret-token" {
		t.Fatalf("authorization metadata = %v, want [Bearer secret-token]", got)
	}
}

func TestAttachBearerToken_EmptyToken(t *testing.T) {
	base := context.Background()
	ctx := attachBearerToken(base, "")

	if ctx != base {
		t.Fatal("expected empty token to keep original context")
	}
}

type transportRuntimeMock struct {
	deployDone     chan struct{}
	lastDeploySpec *protocol.HostDeploySpec
}

func (m *transportRuntimeMock) DeployHost(ctx context.Context, spec *protocol.HostDeploySpec, dockerfileContent string) (*protocol.CreateHostResponse, error) {
	m.lastDeploySpec = spec
	if m.deployDone != nil {
		close(m.deployDone)
	}
	return &protocol.CreateHostResponse{Name: spec.Name, Status: "running"}, nil
}

func (m *transportRuntimeMock) StopHost(ctx context.Context, name string) error {
	return nil
}

func (m *transportRuntimeMock) RemoveHost(ctx context.Context, name string) error {
	return nil
}

func (m *transportRuntimeMock) InspectHost(ctx context.Context, name string) (*protocol.ContainerInfo, error) {
	return nil, nil
}

func (m *transportRuntimeMock) GetRuntimeLogs(ctx context.Context, name string, tailLines int) (string, error) {
	return "", nil
}

func (m *transportRuntimeMock) IsHealthy(ctx context.Context, name string) (bool, error) {
	return false, nil
}

type transportAuditLoggerStub struct{}

func (transportAuditLoggerStub) Log(_ context.Context, entry protocol.AuditLogEntry) error {
	return nil
}

type transportHostTxManagerStub struct{}

func (transportHostTxManagerStub) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type headerCapturingTransportStream struct {
	method string
	header metadata.MD
}

func (s *headerCapturingTransportStream) Method() string {
	return s.method
}

func (s *headerCapturingTransportStream) SetHeader(md metadata.MD) error {
	s.header = metadata.Join(s.header, md)
	return nil
}

func (s *headerCapturingTransportStream) SendHeader(md metadata.MD) error {
	return s.SetHeader(md)
}

func (s *headerCapturingTransportStream) SetTrailer(md metadata.MD) error {
	return nil
}

func newServerTestEnv(t *testing.T) (*Server, *filerepo.NodeRegistry, *filerepo.HostSpecRepository, *transportRuntimeMock) {
	t.Helper()

	tmpDir := t.TempDir()
	registry := filerepo.NewNodeRegistry(filepath.Join(tmpDir, "nodes.json"), logutil.NewNop())
	specMgr := filerepo.NewHostSpecRepository(filepath.Join(tmpDir, "host_specs.json"), logutil.NewNop())
	rt := &transportRuntimeMock{}
	cfg := &config.Config{
		Server: config.ServerConfig{
			ServerConfig: configutil.ServerConfig{AuthToken: "manager-token"},
		},
		Docker: config.DockerConfig{AgentBaseImage: "maze-agent-base:latest"},
	}
	hostSvc := service.NewHostService(registry, specMgr, transportHostTxManagerStub{}, rt, transportAuditLoggerStub{}, cfg, logutil.NewNop(), filepath.Join(tmpDir, "logs"))
	nodeSvc := service.NewNodeService(registry, logutil.NewNop())
	auditSvc := service.NewAuditService(auditrepo.NewLogger("", logutil.NewNop()))
	server := NewServer(hostSvc, nodeSvc, auditSvc, nil, registry, "manager-token", logutil.NewNop())

	t.Cleanup(specMgr.WaitSave)
	t.Cleanup(registry.WaitSave)
	return server, registry, specMgr, rt
}

func TestServer_Register_StoresMappedNodeState(t *testing.T) {
	server, registry, _, _ := newServerTestEnv(t)

	resp, err := server.Register(context.Background(), &pb.RegisterRequest{
		Name:    "node-1",
		Address: "http://node-1:8080",
		Capabilities: &pb.AgentCapabilities{
			SupportedTemplates: []string{"claude"},
			MaxSessions:        8,
			Tools:              []string{"tmux"},
		},
		Status: &pb.AgentStatus{
			ActiveSessions: 2,
			CpuUsage:       51.5,
			MemoryUsageMb:  1024,
			LocalConfig: &pb.LocalAgentConfig{
				WorkingDir: "/workspace",
				Env:        map[string]string{"FOO": "bar"},
			},
			SessionDetails: []*pb.SessionDetail{
				nil,
				{Id: "sess-1", Template: "claude", WorkingDir: "/workspace", UptimeSeconds: 90},
			},
		},
		Metadata: &pb.AgentMetadata{
			Version:   "v1.2.3",
			Hostname:  "node-1-host",
			StartedAt: "2026-05-04T12:00:00Z",
		},
	})
	if err != nil {
		t.Fatalf("Register 返回错误: %v", err)
	}
	if resp.GetName() != "node-1" || resp.GetStatus() != service.NodeStatusOnline {
		t.Fatalf("Register response = %#v, want node-1/online", resp)
	}

	node, _ := registry.Get(context.Background(), "node-1")
	if node == nil {
		t.Fatal("Register 后应写入 NodeRegistry")
	}
	if node.Capabilities.MaxSessions != 8 {
		t.Fatalf("MaxSessions = %d, want 8", node.Capabilities.MaxSessions)
	}
	if node.AgentStatus.LocalConfig == nil || node.AgentStatus.LocalConfig.WorkingDir != "/workspace" {
		t.Fatalf("LocalConfig = %#v, want working dir /workspace", node.AgentStatus.LocalConfig)
	}
	if len(node.AgentStatus.SessionDetails) != 1 {
		t.Fatalf("SessionDetails len = %d, want 1", len(node.AgentStatus.SessionDetails))
	}
	if node.Metadata.Hostname != "node-1-host" {
		t.Fatalf("Hostname = %q, want %q", node.Metadata.Hostname, "node-1-host")
	}
}

func TestServer_Heartbeat_NodeNotFound(t *testing.T) {
	server, _, _, _ := newServerTestEnv(t)

	_, err := server.Heartbeat(context.Background(), &pb.HeartbeatRequest{Name: "missing"})
	if err == nil {
		t.Fatal("缺失节点应返回错误")
	}
	if status.Code(err) != codes.NotFound {
		t.Fatalf("status code = %v, want %v", status.Code(err), codes.NotFound)
	}
}

func TestServer_CreateHost_SetsAcceptedHeader(t *testing.T) {
	server, registry, specMgr, rt := newServerTestEnv(t)
	rt.deployDone = make(chan struct{})
	stream := &headerCapturingTransportStream{method: pb.HostService_CreateHost_FullMethodName}
	ctx := grpc.NewContextWithServerTransportStream(context.Background(), stream)

	resp, err := server.CreateHost(ctx, &pb.CreateHostRequest{
		Name:        "host-1",
		DisplayName: "Host One",
		Tools:       []string{"claude", "go"},
		Resources: &pb.ResourceLimits{
			CpuLimit:    "2",
			MemoryLimit: "4Gi",
		},
	})
	if err != nil {
		t.Fatalf("CreateHost 返回错误: %v", err)
	}
	if got := stream.header.Get("x-http-status"); len(got) != 1 || got[0] != "202" {
		t.Fatalf("x-http-status header = %v, want [202]", got)
	}

	select {
	case <-rt.deployDone:
	case <-time.After(2 * time.Second):
		t.Fatal("等待异步部署超时")
	}

	if resp.GetResources().GetMemoryLimit() != "4Gi" {
		t.Fatalf("MemoryLimit = %q, want %q", resp.GetResources().GetMemoryLimit(), "4Gi")
	}
	if got, _ := specMgr.Get(context.Background(), "host-1"); got == nil {
		t.Fatal("CreateHost 后应创建 HostSpec")
	}
	exists, matched, _ := registry.ValidateHostToken(context.Background(), "host-1", resp.GetAuthToken())
	if !exists || !matched {
		t.Fatalf("host token 未正确预存: exists=%v matched=%v", exists, matched)
	}
	if rt.lastDeploySpec == nil || rt.lastDeploySpec.Resources.CPULimit != "2" {
		t.Fatalf("runtime deploy spec = %#v, want cpu limit 2", rt.lastDeploySpec)
	}
}

func TestPbAgentStatusToProtocol_SkipsNilDetailsAndCopiesLocalConfig(t *testing.T) {
	got := pbAgentStatusToProtocol(&pb.AgentStatus{
		ActiveSessions: 3,
		CpuUsage:       70.5,
		MemoryUsageMb:  2048,
		LocalConfig: &pb.LocalAgentConfig{
			WorkingDir: "/tmp/work",
			Env:        map[string]string{"HELLO": "world"},
		},
		SessionDetails: []*pb.SessionDetail{
			nil,
			{Id: "sess-1", Template: "claude", WorkingDir: "/tmp/work", UptimeSeconds: 120},
		},
	})

	if got.ActiveSessions != 3 {
		t.Fatalf("ActiveSessions = %d, want 3", got.ActiveSessions)
	}
	if len(got.SessionDetails) != 1 {
		t.Fatalf("SessionDetails len = %d, want 1", len(got.SessionDetails))
	}
	if got.LocalConfig == nil || got.LocalConfig.Env["HELLO"] != "world" {
		t.Fatalf("LocalConfig = %#v, want copied env", got.LocalConfig)
	}
}

func TestSafeInt32_ClampsOverflow(t *testing.T) {
	if got := safeInt32(math.MaxInt32 + 100); got != math.MaxInt32 {
		t.Fatalf("safeInt32(max overflow) = %d, want %d", got, math.MaxInt32)
	}
	if got := safeInt32(math.MinInt32 - 100); got != math.MinInt32 {
		t.Fatalf("safeInt32(min overflow) = %d, want %d", got, math.MinInt32)
	}
}
