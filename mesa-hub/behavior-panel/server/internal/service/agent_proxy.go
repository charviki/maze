package service

import (
	"context"
	"fmt"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/mesa-hub-behavior-panel/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// AgentProxyService Agent gRPC 代理 — Session/Template/Config 请求转发到 Agent 节点
type AgentProxyService struct {
	registry  *model.NodeRegistry
	logger    logutil.Logger
	authToken string
}

// NewAgentProxyService 创建 AgentProxyService。
// authToken 用于代理请求时注入到 gRPC metadata，使 Agent 端认证通过。
func NewAgentProxyService(registry *model.NodeRegistry, authToken string, logger logutil.Logger) *AgentProxyService {
	return &AgentProxyService{
		registry:  registry,
		authToken: authToken,
		logger:    logger,
	}
}

// getSessionClient 获取指定节点的 SessionService gRPC client
func (s *AgentProxyService) getSessionClient(nodeName string) (pb.SessionServiceClient, *grpc.ClientConn, error) {
	node := s.getNode(nodeName)
	if node == nil {
		return nil, nil, fmt.Errorf("node %q not found", nodeName)
	}
	conn, err := s.dial(node.GrpcAddress, nodeName)
	if err != nil {
		return nil, nil, err
	}
	return pb.NewSessionServiceClient(conn), conn, nil
}

// getTemplateClient 获取指定节点的 TemplateService gRPC client
func (s *AgentProxyService) getTemplateClient(nodeName string) (pb.TemplateServiceClient, *grpc.ClientConn, error) {
	node := s.getNode(nodeName)
	if node == nil {
		return nil, nil, fmt.Errorf("node %q not found", nodeName)
	}
	conn, err := s.dial(node.GrpcAddress, nodeName)
	if err != nil {
		return nil, nil, err
	}
	return pb.NewTemplateServiceClient(conn), conn, nil
}

// getConfigClient 获取指定节点的 ConfigService gRPC client
func (s *AgentProxyService) getConfigClient(nodeName string) (pb.ConfigServiceClient, *grpc.ClientConn, error) {
	node := s.getNode(nodeName)
	if node == nil {
		return nil, nil, fmt.Errorf("node %q not found", nodeName)
	}
	conn, err := s.dial(node.GrpcAddress, nodeName)
	if err != nil {
		return nil, nil, err
	}
	return pb.NewConfigServiceClient(conn), conn, nil
}

func (s *AgentProxyService) getNode(nodeName string) *model.Node {
	return s.registry.Get(nodeName)
}

func (s *AgentProxyService) dial(addr, nodeName string) (*grpc.ClientConn, error) {
	if addr == "" {
		return nil, fmt.Errorf("node %q has no gRPC address", nodeName)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Manager 代理请求到 Agent 时注入全局认证 token，
	// 否则 Agent 端的 UnaryAuthInterceptor 会拒绝请求。
	if s.authToken != "" {
		opts = append(opts, grpc.WithUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req interface{},
			reply interface{},
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+s.authToken)
			return invoker(ctx, method, req, reply, cc, opts...)
		}))
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("dial %s for %s: %w", addr, nodeName, err)
	}

	s.logger.Infof("[agent-proxy] connecting to %s (%s)", nodeName, addr)

	return conn, nil
}

// SessionService 转发方法

// ListSessions 查询指定节点的 Session 列表
func (s *AgentProxyService) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
	client, conn, err := s.getSessionClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.ListSessions(ctx, req)
}

// CreateSession 创建新的 Session
func (s *AgentProxyService) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.Session, error) {
	client, conn, err := s.getSessionClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.CreateSession(ctx, req)
}

// GetSession 获取 Session 详情
func (s *AgentProxyService) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.Session, error) {
	client, conn, err := s.getSessionClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.GetSession(ctx, req)
}

// DeleteSession 删除指定 Session
func (s *AgentProxyService) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*emptypb.Empty, error) {
	client, conn, err := s.getSessionClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.DeleteSession(ctx, req)
}

// GetSessionConfig 获取 Session 配置
func (s *AgentProxyService) GetSessionConfig(ctx context.Context, req *pb.GetSessionConfigRequest) (*pb.SessionConfigView, error) {
	client, conn, err := s.getSessionClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.GetSessionConfig(ctx, req)
}

// UpdateSessionConfig 更新 Session 配置
func (s *AgentProxyService) UpdateSessionConfig(ctx context.Context, req *pb.UpdateSessionConfigRequest) (*pb.SessionConfigView, error) {
	client, conn, err := s.getSessionClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.UpdateSessionConfig(ctx, req)
}

// RestoreSession 恢复已终止的 Session
func (s *AgentProxyService) RestoreSession(ctx context.Context, req *pb.RestoreSessionRequest) (*emptypb.Empty, error) {
	client, conn, err := s.getSessionClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.RestoreSession(ctx, req)
}

// SaveSessions 保存 Session 快照
func (s *AgentProxyService) SaveSessions(ctx context.Context, req *pb.SaveSessionsRequest) (*pb.SaveSessionsResponse, error) {
	client, conn, err := s.getSessionClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.SaveSessions(ctx, req)
}

// GetSavedSessions 获取已保存的 Session 列表
func (s *AgentProxyService) GetSavedSessions(ctx context.Context, req *pb.GetSavedSessionsRequest) (*pb.GetSavedSessionsResponse, error) {
	client, conn, err := s.getSessionClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.GetSavedSessions(ctx, req)
}

// GetOutput 获取终端输出
func (s *AgentProxyService) GetOutput(ctx context.Context, req *pb.GetOutputRequest) (*pb.TerminalOutput, error) {
	client, conn, err := s.getSessionClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.GetOutput(ctx, req)
}

// SendInput 发送终端输入
func (s *AgentProxyService) SendInput(ctx context.Context, req *pb.SendInputRequest) (*emptypb.Empty, error) {
	client, conn, err := s.getSessionClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.SendInput(ctx, req)
}

// SendSignal 发送终端信号
func (s *AgentProxyService) SendSignal(ctx context.Context, req *pb.SendSignalRequest) (*emptypb.Empty, error) {
	client, conn, err := s.getSessionClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.SendSignal(ctx, req)
}

// GetEnv 获取 Agent 环境变量
func (s *AgentProxyService) GetEnv(ctx context.Context, req *pb.GetEnvRequest) (*pb.GetEnvResponse, error) {
	client, conn, err := s.getSessionClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.GetEnv(ctx, req)
}

// TemplateService 转发方法

// ListTemplates 查询模板列表
func (s *AgentProxyService) ListTemplates(ctx context.Context, req *pb.ListTemplatesRequest) (*pb.ListTemplatesResponse, error) {
	client, conn, err := s.getTemplateClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.ListTemplates(ctx, req)
}

// CreateTemplate 创建新模板
func (s *AgentProxyService) CreateTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*pb.SessionTemplate, error) {
	client, conn, err := s.getTemplateClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.CreateTemplate(ctx, req)
}

// GetTemplate 获取模板详情
func (s *AgentProxyService) GetTemplate(ctx context.Context, req *pb.GetTemplateRequest) (*pb.SessionTemplate, error) {
	client, conn, err := s.getTemplateClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.GetTemplate(ctx, req)
}

// UpdateTemplate 更新模板
func (s *AgentProxyService) UpdateTemplate(ctx context.Context, req *pb.UpdateTemplateRequest) (*pb.SessionTemplate, error) {
	client, conn, err := s.getTemplateClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.UpdateTemplate(ctx, req)
}

// DeleteTemplate 删除模板
func (s *AgentProxyService) DeleteTemplate(ctx context.Context, req *pb.DeleteTemplateRequest) (*emptypb.Empty, error) {
	client, conn, err := s.getTemplateClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.DeleteTemplate(ctx, req)
}

// GetTemplateConfig 获取模板配置
func (s *AgentProxyService) GetTemplateConfig(ctx context.Context, req *pb.GetTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	client, conn, err := s.getTemplateClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.GetTemplateConfig(ctx, req)
}

// UpdateTemplateConfig 更新模板配置
func (s *AgentProxyService) UpdateTemplateConfig(ctx context.Context, req *pb.UpdateTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	client, conn, err := s.getTemplateClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.UpdateTemplateConfig(ctx, req)
}

// ConfigService 转发方法

// GetConfig 获取 Agent 本地配置
func (s *AgentProxyService) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.LocalAgentConfig, error) {
	client, conn, err := s.getConfigClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.GetConfig(ctx, req)
}

// UpdateConfig 更新 Agent 本地配置
func (s *AgentProxyService) UpdateConfig(ctx context.Context, req *pb.UpdateConfigRequest) (*pb.LocalAgentConfig, error) {
	client, conn, err := s.getConfigClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	return client.UpdateConfig(ctx, req)
}
