package agentclient

import (
	"context"
	"net"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	filerepo "github.com/charviki/maze/the-mesa/director-core/internal/repository/file"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestProxyGetNodeAddrErrors(t *testing.T) {
	registry := newTestNodeRegistry(t)
	proxy := NewProxy(registry, nil)

	if _, err := proxy.getNodeAddr(context.Background(), "missing"); status.Code(err) != codes.NotFound {
		t.Fatalf("缺失节点 code = %s, want %s (err=%v)", status.Code(err), codes.NotFound, err)
	}

	registry.Register(context.Background(), protocol.RegisterRequest{
		Name:    "node-1",
		Address: "http://node-1:8080",
	})
	if _, err := proxy.getNodeAddr(context.Background(), "node-1"); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("缺少 gRPC 地址 code = %s, want %s (err=%v)", status.Code(err), codes.FailedPrecondition, err)
	}
}

func TestProxyListSessions_MissingNodeReturnsNotFound(t *testing.T) {
	registry := newTestNodeRegistry(t)
	proxy := NewProxy(registry, nil)

	_, err := proxy.ListSessions(t.Context(), &pb.ListSessionsRequest{NodeName: "missing"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("ListSessions code = %s, want %s (err=%v)", status.Code(err), codes.NotFound, err)
	}
}

func TestProxyGetNodeAddrReturnsGrpcAddress(t *testing.T) {
	registry := newTestNodeRegistry(t)
	registerTestNode(registry, "node-1:9090")

	proxy := NewProxy(registry, nil)
	addr, err := proxy.getNodeAddr(context.Background(), "node-1")
	if err != nil {
		t.Fatalf("getNodeAddr 返回错误: %v", err)
	}
	if addr != "node-1:9090" {
		t.Fatalf("addr = %q, want %q", addr, "node-1:9090")
	}
}

type proxyTestAgent struct {
	pb.UnimplementedSessionServiceServer
	pb.UnimplementedTemplateServiceServer
	pb.UnimplementedConfigServiceServer

	mu              sync.Mutex
	authorizations  []string
	listSessionCall atomic.Int32

	listSessionsFn         func(context.Context, *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error)
	createSessionFn        func(context.Context, *pb.CreateSessionRequest) (*pb.Session, error)
	getSessionFn           func(context.Context, *pb.GetSessionRequest) (*pb.Session, error)
	deleteSessionFn        func(context.Context, *pb.DeleteSessionRequest) (*emptypb.Empty, error)
	getSessionConfigFn     func(context.Context, *pb.GetSessionConfigRequest) (*pb.SessionConfigView, error)
	updateSessionConfigFn  func(context.Context, *pb.UpdateSessionConfigRequest) (*pb.SessionConfigView, error)
	restoreSessionFn       func(context.Context, *pb.RestoreSessionRequest) (*emptypb.Empty, error)
	saveSessionsFn         func(context.Context, *pb.SaveSessionsRequest) (*pb.SaveSessionsResponse, error)
	getSavedSessionsFn     func(context.Context, *pb.GetSavedSessionsRequest) (*pb.GetSavedSessionsResponse, error)
	getOutputFn            func(context.Context, *pb.GetOutputRequest) (*pb.TerminalOutput, error)
	sendInputFn            func(context.Context, *pb.SendInputRequest) (*emptypb.Empty, error)
	sendSignalFn           func(context.Context, *pb.SendSignalRequest) (*emptypb.Empty, error)
	getEnvFn               func(context.Context, *pb.GetEnvRequest) (*pb.GetEnvResponse, error)
	listTemplatesFn        func(context.Context, *pb.ListTemplatesRequest) (*pb.ListTemplatesResponse, error)
	createTemplateFn       func(context.Context, *pb.CreateTemplateRequest) (*pb.SessionTemplate, error)
	getTemplateFn          func(context.Context, *pb.GetTemplateRequest) (*pb.SessionTemplate, error)
	updateTemplateFn       func(context.Context, *pb.UpdateTemplateRequest) (*pb.SessionTemplate, error)
	deleteTemplateFn       func(context.Context, *pb.DeleteTemplateRequest) (*emptypb.Empty, error)
	getTemplateConfigFn    func(context.Context, *pb.GetTemplateConfigRequest) (*pb.TemplateConfigView, error)
	updateTemplateConfigFn func(context.Context, *pb.UpdateTemplateConfigRequest) (*pb.TemplateConfigView, error)
	getConfigFn            func(context.Context, *pb.GetConfigRequest) (*pb.LocalAgentConfig, error)
	updateConfigFn         func(context.Context, *pb.UpdateConfigRequest) (*pb.LocalAgentConfig, error)
}

// 按 Session/Template/Config 分组注册 handler，是为了让每条测试只声明自己关心的 RPC，
// 避免大型结构体字面量把无关字段也堆在一起。
type proxySessionHandlers struct {
	listSessions        func(context.Context, *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error)
	createSession       func(context.Context, *pb.CreateSessionRequest) (*pb.Session, error)
	getSession          func(context.Context, *pb.GetSessionRequest) (*pb.Session, error)
	deleteSession       func(context.Context, *pb.DeleteSessionRequest) (*emptypb.Empty, error)
	getSessionConfig    func(context.Context, *pb.GetSessionConfigRequest) (*pb.SessionConfigView, error)
	updateSessionConfig func(context.Context, *pb.UpdateSessionConfigRequest) (*pb.SessionConfigView, error)
	restoreSession      func(context.Context, *pb.RestoreSessionRequest) (*emptypb.Empty, error)
	saveSessions        func(context.Context, *pb.SaveSessionsRequest) (*pb.SaveSessionsResponse, error)
	getSavedSessions    func(context.Context, *pb.GetSavedSessionsRequest) (*pb.GetSavedSessionsResponse, error)
	getOutput           func(context.Context, *pb.GetOutputRequest) (*pb.TerminalOutput, error)
	sendInput           func(context.Context, *pb.SendInputRequest) (*emptypb.Empty, error)
	sendSignal          func(context.Context, *pb.SendSignalRequest) (*emptypb.Empty, error)
	getEnv              func(context.Context, *pb.GetEnvRequest) (*pb.GetEnvResponse, error)
}

type proxyTemplateHandlers struct {
	listTemplates        func(context.Context, *pb.ListTemplatesRequest) (*pb.ListTemplatesResponse, error)
	createTemplate       func(context.Context, *pb.CreateTemplateRequest) (*pb.SessionTemplate, error)
	getTemplate          func(context.Context, *pb.GetTemplateRequest) (*pb.SessionTemplate, error)
	updateTemplate       func(context.Context, *pb.UpdateTemplateRequest) (*pb.SessionTemplate, error)
	deleteTemplate       func(context.Context, *pb.DeleteTemplateRequest) (*emptypb.Empty, error)
	getTemplateConfig    func(context.Context, *pb.GetTemplateConfigRequest) (*pb.TemplateConfigView, error)
	updateTemplateConfig func(context.Context, *pb.UpdateTemplateConfigRequest) (*pb.TemplateConfigView, error)
}

type proxyConfigHandlers struct {
	getConfig    func(context.Context, *pb.GetConfigRequest) (*pb.LocalAgentConfig, error)
	updateConfig func(context.Context, *pb.UpdateConfigRequest) (*pb.LocalAgentConfig, error)
}

func newProxyTestAgent() *proxyTestAgent {
	return &proxyTestAgent{}
}

func (a *proxyTestAgent) withSessionHandlers(h proxySessionHandlers) *proxyTestAgent {
	a.listSessionsFn = h.listSessions
	a.createSessionFn = h.createSession
	a.getSessionFn = h.getSession
	a.deleteSessionFn = h.deleteSession
	a.getSessionConfigFn = h.getSessionConfig
	a.updateSessionConfigFn = h.updateSessionConfig
	a.restoreSessionFn = h.restoreSession
	a.saveSessionsFn = h.saveSessions
	a.getSavedSessionsFn = h.getSavedSessions
	a.getOutputFn = h.getOutput
	a.sendInputFn = h.sendInput
	a.sendSignalFn = h.sendSignal
	a.getEnvFn = h.getEnv
	return a
}

func (a *proxyTestAgent) withTemplateHandlers(h proxyTemplateHandlers) *proxyTestAgent {
	a.listTemplatesFn = h.listTemplates
	a.createTemplateFn = h.createTemplate
	a.getTemplateFn = h.getTemplate
	a.updateTemplateFn = h.updateTemplate
	a.deleteTemplateFn = h.deleteTemplate
	a.getTemplateConfigFn = h.getTemplateConfig
	a.updateTemplateConfigFn = h.updateTemplateConfig
	return a
}

func (a *proxyTestAgent) withConfigHandlers(h proxyConfigHandlers) *proxyTestAgent {
	a.getConfigFn = h.getConfig
	a.updateConfigFn = h.updateConfig
	return a
}

func (a *proxyTestAgent) recordAuthorization(ctx context.Context) {
	md, _ := metadata.FromIncomingContext(ctx)
	a.mu.Lock()
	defer a.mu.Unlock()
	a.authorizations = append(a.authorizations, md.Get("authorization")...)
}

func (a *proxyTestAgent) lastAuthorization() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.authorizations) == 0 {
		return ""
	}
	return a.authorizations[len(a.authorizations)-1]
}

func (a *proxyTestAgent) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
	a.recordAuthorization(ctx)
	a.listSessionCall.Add(1)
	if a.listSessionsFn != nil {
		return a.listSessionsFn(ctx, req)
	}
	return &pb.ListSessionsResponse{}, nil
}

func (a *proxyTestAgent) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.Session, error) {
	a.recordAuthorization(ctx)
	if a.createSessionFn != nil {
		return a.createSessionFn(ctx, req)
	}
	return &pb.Session{}, nil
}

func (a *proxyTestAgent) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.Session, error) {
	a.recordAuthorization(ctx)
	if a.getSessionFn != nil {
		return a.getSessionFn(ctx, req)
	}
	return &pb.Session{}, nil
}

func (a *proxyTestAgent) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*emptypb.Empty, error) {
	a.recordAuthorization(ctx)
	if a.deleteSessionFn != nil {
		return a.deleteSessionFn(ctx, req)
	}
	return &emptypb.Empty{}, nil
}

func (a *proxyTestAgent) GetSessionConfig(ctx context.Context, req *pb.GetSessionConfigRequest) (*pb.SessionConfigView, error) {
	a.recordAuthorization(ctx)
	if a.getSessionConfigFn != nil {
		return a.getSessionConfigFn(ctx, req)
	}
	return &pb.SessionConfigView{}, nil
}

func (a *proxyTestAgent) UpdateSessionConfig(ctx context.Context, req *pb.UpdateSessionConfigRequest) (*pb.SessionConfigView, error) {
	a.recordAuthorization(ctx)
	if a.updateSessionConfigFn != nil {
		return a.updateSessionConfigFn(ctx, req)
	}
	return &pb.SessionConfigView{}, nil
}

func (a *proxyTestAgent) RestoreSession(ctx context.Context, req *pb.RestoreSessionRequest) (*emptypb.Empty, error) {
	a.recordAuthorization(ctx)
	if a.restoreSessionFn != nil {
		return a.restoreSessionFn(ctx, req)
	}
	return &emptypb.Empty{}, nil
}

func (a *proxyTestAgent) SaveSessions(ctx context.Context, req *pb.SaveSessionsRequest) (*pb.SaveSessionsResponse, error) {
	a.recordAuthorization(ctx)
	if a.saveSessionsFn != nil {
		return a.saveSessionsFn(ctx, req)
	}
	return &pb.SaveSessionsResponse{}, nil
}

func (a *proxyTestAgent) GetSavedSessions(ctx context.Context, req *pb.GetSavedSessionsRequest) (*pb.GetSavedSessionsResponse, error) {
	a.recordAuthorization(ctx)
	if a.getSavedSessionsFn != nil {
		return a.getSavedSessionsFn(ctx, req)
	}
	return &pb.GetSavedSessionsResponse{}, nil
}

func (a *proxyTestAgent) GetOutput(ctx context.Context, req *pb.GetOutputRequest) (*pb.TerminalOutput, error) {
	a.recordAuthorization(ctx)
	if a.getOutputFn != nil {
		return a.getOutputFn(ctx, req)
	}
	return &pb.TerminalOutput{}, nil
}

func (a *proxyTestAgent) SendInput(ctx context.Context, req *pb.SendInputRequest) (*emptypb.Empty, error) {
	a.recordAuthorization(ctx)
	if a.sendInputFn != nil {
		return a.sendInputFn(ctx, req)
	}
	return &emptypb.Empty{}, nil
}

func (a *proxyTestAgent) SendSignal(ctx context.Context, req *pb.SendSignalRequest) (*emptypb.Empty, error) {
	a.recordAuthorization(ctx)
	if a.sendSignalFn != nil {
		return a.sendSignalFn(ctx, req)
	}
	return &emptypb.Empty{}, nil
}

func (a *proxyTestAgent) GetEnv(ctx context.Context, req *pb.GetEnvRequest) (*pb.GetEnvResponse, error) {
	a.recordAuthorization(ctx)
	if a.getEnvFn != nil {
		return a.getEnvFn(ctx, req)
	}
	return &pb.GetEnvResponse{}, nil
}

func (a *proxyTestAgent) ListTemplates(ctx context.Context, req *pb.ListTemplatesRequest) (*pb.ListTemplatesResponse, error) {
	a.recordAuthorization(ctx)
	if a.listTemplatesFn != nil {
		return a.listTemplatesFn(ctx, req)
	}
	return &pb.ListTemplatesResponse{}, nil
}

func (a *proxyTestAgent) CreateTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*pb.SessionTemplate, error) {
	a.recordAuthorization(ctx)
	if a.createTemplateFn != nil {
		return a.createTemplateFn(ctx, req)
	}
	return &pb.SessionTemplate{}, nil
}

func (a *proxyTestAgent) GetTemplate(ctx context.Context, req *pb.GetTemplateRequest) (*pb.SessionTemplate, error) {
	a.recordAuthorization(ctx)
	if a.getTemplateFn != nil {
		return a.getTemplateFn(ctx, req)
	}
	return &pb.SessionTemplate{}, nil
}

func (a *proxyTestAgent) UpdateTemplate(ctx context.Context, req *pb.UpdateTemplateRequest) (*pb.SessionTemplate, error) {
	a.recordAuthorization(ctx)
	if a.updateTemplateFn != nil {
		return a.updateTemplateFn(ctx, req)
	}
	return &pb.SessionTemplate{}, nil
}

func (a *proxyTestAgent) DeleteTemplate(ctx context.Context, req *pb.DeleteTemplateRequest) (*emptypb.Empty, error) {
	a.recordAuthorization(ctx)
	if a.deleteTemplateFn != nil {
		return a.deleteTemplateFn(ctx, req)
	}
	return &emptypb.Empty{}, nil
}

func (a *proxyTestAgent) GetTemplateConfig(ctx context.Context, req *pb.GetTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	a.recordAuthorization(ctx)
	if a.getTemplateConfigFn != nil {
		return a.getTemplateConfigFn(ctx, req)
	}
	return &pb.TemplateConfigView{}, nil
}

func (a *proxyTestAgent) UpdateTemplateConfig(ctx context.Context, req *pb.UpdateTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	a.recordAuthorization(ctx)
	if a.updateTemplateConfigFn != nil {
		return a.updateTemplateConfigFn(ctx, req)
	}
	return &pb.TemplateConfigView{}, nil
}

func (a *proxyTestAgent) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.LocalAgentConfig, error) {
	a.recordAuthorization(ctx)
	if a.getConfigFn != nil {
		return a.getConfigFn(ctx, req)
	}
	return &pb.LocalAgentConfig{}, nil
}

func (a *proxyTestAgent) UpdateConfig(ctx context.Context, req *pb.UpdateConfigRequest) (*pb.LocalAgentConfig, error) {
	a.recordAuthorization(ctx)
	if a.updateConfigFn != nil {
		return a.updateConfigFn(ctx, req)
	}
	return &pb.LocalAgentConfig{}, nil
}

func startProxyTestAgent(t *testing.T, agent *proxyTestAgent) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterSessionServiceServer(server, agent)
	pb.RegisterTemplateServiceServer(server, agent)
	pb.RegisterConfigServiceServer(server, agent)

	go func() {
		_ = server.Serve(listener)
	}()

	t.Cleanup(func() {
		server.Stop()
		_ = listener.Close()
	})
	return listener.Addr().String()
}

func newTestNodeRegistry(t *testing.T) *filerepo.NodeRegistry {
	t.Helper()

	registry := filerepo.NewNodeRegistry(filepath.Join(t.TempDir(), "nodes.json"), logutil.NewNop())
	t.Cleanup(registry.WaitSave)
	return registry
}

func registerTestNode(registry *filerepo.NodeRegistry, grpcAddr string) {
	registry.Register(context.Background(), protocol.RegisterRequest{
		Name:        "node-1",
		Address:     "http://node-1:8080",
		GrpcAddress: grpcAddr,
	})
}

func newTestProxy(t *testing.T, grpcAddr string) (*Proxy, *ConnectionManager) {
	t.Helper()
	registry := newTestNodeRegistry(t)
	registerTestNode(registry, grpcAddr)

	connMgr := NewConnectionManager(logutil.NewNop(), "test-token", time.Minute)
	t.Cleanup(connMgr.CloseAll)
	return NewProxy(registry, connMgr), connMgr
}

func newRunningTestProxy(t *testing.T, agent *proxyTestAgent) (*Proxy, *ConnectionManager) {
	t.Helper()
	return newTestProxy(t, startProxyTestAgent(t, agent))
}

func TestProxyListSessions_ForwardsRequestAndAuthHeader(t *testing.T) {
	agent := newProxyTestAgent().withSessionHandlers(proxySessionHandlers{
		listSessions: func(_ context.Context, req *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
			if req.GetNodeName() != "node-1" {
				t.Fatalf("node_name = %q, want node-1", req.GetNodeName())
			}
			return &pb.ListSessionsResponse{
				Sessions: []*pb.Session{{Id: "sess-1", Name: "demo"}},
			}, nil
		},
	})
	proxy, _ := newRunningTestProxy(t, agent)

	resp, err := proxy.ListSessions(t.Context(), &pb.ListSessionsRequest{NodeName: "node-1"})
	if err != nil {
		t.Fatalf("ListSessions 返回错误: %v", err)
	}
	sessions := resp.GetSessions()
	if len(sessions) != 1 || sessions[0].GetId() != "sess-1" {
		t.Fatalf("sessions = %+v, want sess-1", sessions)
	}
	if agent.lastAuthorization() != "Bearer test-token" {
		t.Fatalf("authorization = %q, want Bearer test-token", agent.lastAuthorization())
	}
}

func TestProxyListSessions_ConcurrentCallsReuseConnection(t *testing.T) {
	agent := newProxyTestAgent().withSessionHandlers(proxySessionHandlers{
		listSessions: func(_ context.Context, _ *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
			return &pb.ListSessionsResponse{Sessions: []*pb.Session{{Id: "sess"}}}, nil
		},
	})
	proxy, connMgr := newRunningTestProxy(t, agent)

	const calls = 8
	var wg sync.WaitGroup
	errCh := make(chan error, calls)
	for range calls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := proxy.ListSessions(t.Context(), &pb.ListSessionsRequest{NodeName: "node-1"})
			errCh <- err
		}()
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("并发调用返回错误: %v", err)
		}
	}

	if got := agent.listSessionCall.Load(); got != calls {
		t.Fatalf("ListSessions 调用次数 = %d, want %d", got, calls)
	}

	connMgr.mu.RLock()
	defer connMgr.mu.RUnlock()
	if len(connMgr.conns) != 1 {
		t.Fatalf("连接池条目数 = %d, want 1", len(connMgr.conns))
	}
}

func TestProxyGetSession_HonorsContextCancellation(t *testing.T) {
	agent := newProxyTestAgent().withSessionHandlers(proxySessionHandlers{
		getSession: func(ctx context.Context, req *pb.GetSessionRequest) (*pb.Session, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
	})
	proxy, _ := newRunningTestProxy(t, agent)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := proxy.GetSession(ctx, &pb.GetSessionRequest{NodeName: "node-1", Id: "sess-1"})
	if status.Code(err) != codes.Canceled {
		t.Fatalf("code = %s, want %s (err=%v)", status.Code(err), codes.Canceled, err)
	}
}

func TestProxyListSessions_ConnectionTimeout(t *testing.T) {
	proxy, _ := newTestProxy(t, "203.0.113.1:65535")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := proxy.ListSessions(ctx, &pb.ListSessionsRequest{NodeName: "node-1"})
	if status.Code(err) != codes.DeadlineExceeded {
		t.Fatalf("code = %s, want %s (err=%v)", status.Code(err), codes.DeadlineExceeded, err)
	}
}

func TestProxyListSessions_UsesUpdatedNodeAddress(t *testing.T) {
	firstAgent := newProxyTestAgent().withSessionHandlers(proxySessionHandlers{
		listSessions: func(_ context.Context, _ *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
			return &pb.ListSessionsResponse{Sessions: []*pb.Session{{Id: "old"}}}, nil
		},
	})
	secondAgent := newProxyTestAgent().withSessionHandlers(proxySessionHandlers{
		listSessions: func(_ context.Context, _ *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
			return &pb.ListSessionsResponse{Sessions: []*pb.Session{{Id: "new"}}}, nil
		},
	})

	registry := newTestNodeRegistry(t)
	connMgr := NewConnectionManager(logutil.NewNop(), "test-token", time.Minute)
	t.Cleanup(connMgr.CloseAll)
	proxy := NewProxy(registry, connMgr)

	registerTestNode(registry, startProxyTestAgent(t, firstAgent))

	firstResp, err := proxy.ListSessions(t.Context(), &pb.ListSessionsRequest{NodeName: "node-1"})
	if err != nil {
		t.Fatalf("首次 ListSessions 返回错误: %v", err)
	}
	if firstResp.GetSessions()[0].GetId() != "old" {
		t.Fatalf("首次命中结果 = %q, want old", firstResp.GetSessions()[0].GetId())
	}

	registerTestNode(registry, startProxyTestAgent(t, secondAgent))

	secondResp, err := proxy.ListSessions(t.Context(), &pb.ListSessionsRequest{NodeName: "node-1"})
	if err != nil {
		t.Fatalf("地址更新后 ListSessions 返回错误: %v", err)
	}
	if secondResp.GetSessions()[0].GetId() != "new" {
		t.Fatalf("地址更新后结果 = %q, want new", secondResp.GetSessions()[0].GetId())
	}
}

func TestProxy_TemplateAndConfigClients(t *testing.T) {
	agent := newProxyTestAgent().
		withTemplateHandlers(proxyTemplateHandlers{
			listTemplates: func(_ context.Context, req *pb.ListTemplatesRequest) (*pb.ListTemplatesResponse, error) {
				if req.GetNodeName() != "node-1" {
					t.Fatalf("node_name = %q, want node-1", req.GetNodeName())
				}
				return &pb.ListTemplatesResponse{
					Templates: []*pb.SessionTemplate{{Id: "tpl-1", Name: "Claude"}},
				}, nil
			},
		}).
		withConfigHandlers(proxyConfigHandlers{
			getConfig: func(_ context.Context, req *pb.GetConfigRequest) (*pb.LocalAgentConfig, error) {
				if req.GetNodeName() != "node-1" {
					t.Fatalf("node_name = %q, want node-1", req.GetNodeName())
				}
				return &pb.LocalAgentConfig{
					WorkingDir: "/workspace",
					Env:        map[string]string{"FOO": "bar"},
				}, nil
			},
		})
	proxy, _ := newRunningTestProxy(t, agent)

	templates, err := proxy.ListTemplates(t.Context(), &pb.ListTemplatesRequest{NodeName: "node-1"})
	if err != nil {
		t.Fatalf("ListTemplates 返回错误: %v", err)
	}
	templateItems := templates.GetTemplates()
	if len(templateItems) != 1 || templateItems[0].GetId() != "tpl-1" {
		t.Fatalf("templates = %+v, want tpl-1", templateItems)
	}

	cfg, err := proxy.GetConfig(t.Context(), &pb.GetConfigRequest{NodeName: "node-1"})
	if err != nil {
		t.Fatalf("GetConfig 返回错误: %v", err)
	}
	if cfg.GetWorkingDir() != "/workspace" || cfg.GetEnv()["FOO"] != "bar" {
		t.Fatalf("cfg = %+v, want working_dir=/workspace env.FOO=bar", cfg)
	}
}

func TestProxy_SessionWrappers(t *testing.T) {
	agent := newProxyTestAgent().withSessionHandlers(proxySessionHandlers{
		createSession: func(_ context.Context, req *pb.CreateSessionRequest) (*pb.Session, error) {
			if req.GetNodeName() != "node-1" || req.GetName() != "new-session" {
				t.Fatalf("CreateSession req = %+v", req)
			}
			return &pb.Session{Id: "created", Name: req.GetName()}, nil
		},
		deleteSession: func(_ context.Context, req *pb.DeleteSessionRequest) (*emptypb.Empty, error) {
			if req.GetNodeName() != "node-1" || req.GetId() != "sess-1" {
				t.Fatalf("DeleteSession req = %+v", req)
			}
			return &emptypb.Empty{}, nil
		},
		getSessionConfig: func(_ context.Context, req *pb.GetSessionConfigRequest) (*pb.SessionConfigView, error) {
			if req.GetId() != "sess-1" {
				t.Fatalf("GetSessionConfig req = %+v", req)
			}
			return &pb.SessionConfigView{SessionId: req.GetId()}, nil
		},
		updateSessionConfig: func(_ context.Context, req *pb.UpdateSessionConfigRequest) (*pb.SessionConfigView, error) {
			if req.GetId() != "sess-1" || len(req.GetFiles()) != 1 {
				t.Fatalf("UpdateSessionConfig req = %+v", req)
			}
			return &pb.SessionConfigView{SessionId: req.GetId()}, nil
		},
		restoreSession: func(_ context.Context, req *pb.RestoreSessionRequest) (*emptypb.Empty, error) {
			if req.GetId() != "sess-1" {
				t.Fatalf("RestoreSession req = %+v", req)
			}
			return &emptypb.Empty{}, nil
		},
		saveSessions: func(_ context.Context, req *pb.SaveSessionsRequest) (*pb.SaveSessionsResponse, error) {
			if req.GetNodeName() != "node-1" {
				t.Fatalf("SaveSessions req = %+v", req)
			}
			return &pb.SaveSessionsResponse{SavedAt: "now"}, nil
		},
		getSavedSessions: func(_ context.Context, req *pb.GetSavedSessionsRequest) (*pb.GetSavedSessionsResponse, error) {
			if req.GetNodeName() != "node-1" {
				t.Fatalf("GetSavedSessions req = %+v", req)
			}
			return &pb.GetSavedSessionsResponse{}, nil
		},
		getOutput: func(_ context.Context, req *pb.GetOutputRequest) (*pb.TerminalOutput, error) {
			if req.GetId() != "sess-1" || req.GetLines() != 50 {
				t.Fatalf("GetOutput req = %+v", req)
			}
			return &pb.TerminalOutput{SessionId: req.GetId(), Output: "hello"}, nil
		},
		sendInput: func(_ context.Context, req *pb.SendInputRequest) (*emptypb.Empty, error) {
			if req.GetCommand() != "pwd" {
				t.Fatalf("SendInput req = %+v", req)
			}
			return &emptypb.Empty{}, nil
		},
		sendSignal: func(_ context.Context, req *pb.SendSignalRequest) (*emptypb.Empty, error) {
			if req.GetSignal() != "SIGINT" {
				t.Fatalf("SendSignal req = %+v", req)
			}
			return &emptypb.Empty{}, nil
		},
		getEnv: func(_ context.Context, req *pb.GetEnvRequest) (*pb.GetEnvResponse, error) {
			if req.GetId() != "sess-1" {
				t.Fatalf("GetEnv req = %+v", req)
			}
			return &pb.GetEnvResponse{Env: map[string]string{"PATH": "/usr/bin"}}, nil
		},
	})
	proxy, _ := newRunningTestProxy(t, agent)

	created, err := proxy.CreateSession(t.Context(), &pb.CreateSessionRequest{NodeName: "node-1", Name: "new-session"})
	if err != nil || created.GetId() != "created" {
		t.Fatalf("CreateSession got = %+v err = %v", created, err)
	}

	if _, err := proxy.DeleteSession(t.Context(), &pb.DeleteSessionRequest{NodeName: "node-1", Id: "sess-1"}); err != nil {
		t.Fatalf("DeleteSession 返回错误: %v", err)
	}

	sessionCfg, err := proxy.GetSessionConfig(t.Context(), &pb.GetSessionConfigRequest{NodeName: "node-1", Id: "sess-1"})
	if err != nil || sessionCfg.GetSessionId() != "sess-1" {
		t.Fatalf("GetSessionConfig got = %+v err = %v", sessionCfg, err)
	}

	updatedCfg, err := proxy.UpdateSessionConfig(t.Context(), &pb.UpdateSessionConfigRequest{
		NodeName: "node-1",
		Id:       "sess-1",
		Files:    []*pb.ConfigFileUpdate{{Path: ".env", Content: "FOO=bar"}},
	})
	if err != nil || updatedCfg.GetSessionId() != "sess-1" {
		t.Fatalf("UpdateSessionConfig got = %+v err = %v", updatedCfg, err)
	}

	if _, err := proxy.RestoreSession(t.Context(), &pb.RestoreSessionRequest{NodeName: "node-1", Id: "sess-1"}); err != nil {
		t.Fatalf("RestoreSession 返回错误: %v", err)
	}

	saved, err := proxy.SaveSessions(t.Context(), &pb.SaveSessionsRequest{NodeName: "node-1"})
	if err != nil || saved.GetSavedAt() != "now" {
		t.Fatalf("SaveSessions got = %+v err = %v", saved, err)
	}

	if _, err := proxy.GetSavedSessions(t.Context(), &pb.GetSavedSessionsRequest{NodeName: "node-1"}); err != nil {
		t.Fatalf("GetSavedSessions 返回错误: %v", err)
	}

	output, err := proxy.GetOutput(t.Context(), &pb.GetOutputRequest{NodeName: "node-1", Id: "sess-1", Lines: 50})
	if err != nil || output.GetOutput() != "hello" {
		t.Fatalf("GetOutput got = %+v err = %v", output, err)
	}

	if _, err := proxy.SendInput(t.Context(), &pb.SendInputRequest{NodeName: "node-1", Id: "sess-1", Command: "pwd"}); err != nil {
		t.Fatalf("SendInput 返回错误: %v", err)
	}

	if _, err := proxy.SendSignal(t.Context(), &pb.SendSignalRequest{NodeName: "node-1", Id: "sess-1", Signal: "SIGINT"}); err != nil {
		t.Fatalf("SendSignal 返回错误: %v", err)
	}

	env, err := proxy.GetEnv(t.Context(), &pb.GetEnvRequest{NodeName: "node-1", Id: "sess-1"})
	if err != nil || env.GetEnv()["PATH"] != "/usr/bin" {
		t.Fatalf("GetEnv got = %+v err = %v", env, err)
	}
}

func TestProxy_TemplateAndConfigMutations(t *testing.T) {
	agent := newProxyTestAgent().
		withTemplateHandlers(proxyTemplateHandlers{
			createTemplate: func(_ context.Context, req *pb.CreateTemplateRequest) (*pb.SessionTemplate, error) {
				if req.GetNodeName() != "node-1" || req.GetTemplate().GetName() != "Claude" {
					t.Fatalf("CreateTemplate req = %+v", req)
				}
				return &pb.SessionTemplate{Id: "tpl-created", Name: req.GetTemplate().GetName()}, nil
			},
			getTemplate: func(_ context.Context, req *pb.GetTemplateRequest) (*pb.SessionTemplate, error) {
				if req.GetId() != "tpl-1" {
					t.Fatalf("GetTemplate req = %+v", req)
				}
				return &pb.SessionTemplate{Id: req.GetId(), Name: "Claude"}, nil
			},
			updateTemplate: func(_ context.Context, req *pb.UpdateTemplateRequest) (*pb.SessionTemplate, error) {
				if req.GetId() != "tpl-1" || req.GetTemplate().GetName() != "Claude 2" {
					t.Fatalf("UpdateTemplate req = %+v", req)
				}
				return &pb.SessionTemplate{Id: req.GetId(), Name: req.GetTemplate().GetName()}, nil
			},
			deleteTemplate: func(_ context.Context, req *pb.DeleteTemplateRequest) (*emptypb.Empty, error) {
				if req.GetId() != "tpl-1" {
					t.Fatalf("DeleteTemplate req = %+v", req)
				}
				return &emptypb.Empty{}, nil
			},
			getTemplateConfig: func(_ context.Context, req *pb.GetTemplateConfigRequest) (*pb.TemplateConfigView, error) {
				if req.GetId() != "tpl-1" {
					t.Fatalf("GetTemplateConfig req = %+v", req)
				}
				return &pb.TemplateConfigView{TemplateId: req.GetId()}, nil
			},
			updateTemplateConfig: func(_ context.Context, req *pb.UpdateTemplateConfigRequest) (*pb.TemplateConfigView, error) {
				if req.GetId() != "tpl-1" || len(req.GetFiles()) != 1 {
					t.Fatalf("UpdateTemplateConfig req = %+v", req)
				}
				return &pb.TemplateConfigView{TemplateId: req.GetId()}, nil
			},
		}).
		withConfigHandlers(proxyConfigHandlers{
			updateConfig: func(_ context.Context, req *pb.UpdateConfigRequest) (*pb.LocalAgentConfig, error) {
				if req.GetNodeName() != "node-1" || req.GetWorkingDir() != "/workspace-2" || req.GetEnv()["FOO"] != "baz" {
					t.Fatalf("UpdateConfig req = %+v", req)
				}
				return &pb.LocalAgentConfig{WorkingDir: req.GetWorkingDir(), Env: req.GetEnv()}, nil
			},
		})
	proxy, _ := newRunningTestProxy(t, agent)

	created, err := proxy.CreateTemplate(t.Context(), &pb.CreateTemplateRequest{
		NodeName: "node-1",
		Template: &pb.SessionTemplate{Name: "Claude"},
	})
	if err != nil || created.GetId() != "tpl-created" {
		t.Fatalf("CreateTemplate got = %+v err = %v", created, err)
	}

	gotTemplate, err := proxy.GetTemplate(t.Context(), &pb.GetTemplateRequest{NodeName: "node-1", Id: "tpl-1"})
	if err != nil || gotTemplate.GetId() != "tpl-1" {
		t.Fatalf("GetTemplate got = %+v err = %v", gotTemplate, err)
	}

	updatedTemplate, err := proxy.UpdateTemplate(t.Context(), &pb.UpdateTemplateRequest{
		NodeName: "node-1",
		Id:       "tpl-1",
		Template: &pb.SessionTemplate{Name: "Claude 2"},
	})
	if err != nil || updatedTemplate.GetName() != "Claude 2" {
		t.Fatalf("UpdateTemplate got = %+v err = %v", updatedTemplate, err)
	}

	if _, err := proxy.DeleteTemplate(t.Context(), &pb.DeleteTemplateRequest{NodeName: "node-1", Id: "tpl-1"}); err != nil {
		t.Fatalf("DeleteTemplate 返回错误: %v", err)
	}

	templateCfg, err := proxy.GetTemplateConfig(t.Context(), &pb.GetTemplateConfigRequest{NodeName: "node-1", Id: "tpl-1"})
	if err != nil || templateCfg.GetTemplateId() != "tpl-1" {
		t.Fatalf("GetTemplateConfig got = %+v err = %v", templateCfg, err)
	}

	updatedTemplateCfg, err := proxy.UpdateTemplateConfig(t.Context(), &pb.UpdateTemplateConfigRequest{
		NodeName: "node-1",
		Id:       "tpl-1",
		Files:    []*pb.ConfigFileUpdate{{Path: "template.yaml", Content: "foo: bar"}},
	})
	if err != nil || updatedTemplateCfg.GetTemplateId() != "tpl-1" {
		t.Fatalf("UpdateTemplateConfig got = %+v err = %v", updatedTemplateCfg, err)
	}

	updatedCfg, err := proxy.UpdateConfig(t.Context(), &pb.UpdateConfigRequest{
		NodeName:   "node-1",
		WorkingDir: "/workspace-2",
		Env:        map[string]string{"FOO": "baz"},
	})
	if err != nil || updatedCfg.GetWorkingDir() != "/workspace-2" || updatedCfg.GetEnv()["FOO"] != "baz" {
		t.Fatalf("UpdateConfig got = %+v err = %v", updatedCfg, err)
	}
}

func TestProxy_GetTemplateAndConfigClientErrors(t *testing.T) {
	proxy := NewProxy(filerepo.NewNodeRegistry(filepath.Join(t.TempDir(), "nodes.json"), logutil.NewNop()), NewConnectionManager(logutil.NewNop(), "", time.Minute))

	if _, err := proxy.ListTemplates(t.Context(), &pb.ListTemplatesRequest{NodeName: "missing"}); err == nil {
		t.Fatal("缺失节点时 ListTemplates 应返回错误")
	}
	if _, err := proxy.GetConfig(t.Context(), &pb.GetConfigRequest{NodeName: "missing"}); err == nil {
		t.Fatal("缺失节点时 GetConfig 应返回错误")
	}
}
