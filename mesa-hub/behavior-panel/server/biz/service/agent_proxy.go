package service

import (
	"context"
	"fmt"
	"time"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/mesa-hub-behavior-panel/biz/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

// AgentProxyService Agent gRPC 代理 — Session/Template/Config 请求转发到 Agent 节点
type AgentProxyService struct {
	registry *model.NodeRegistry
	logger   logutil.Logger
}

// NewAgentProxyService 创建 AgentProxyService
func NewAgentProxyService(registry *model.NodeRegistry, logger logutil.Logger) *AgentProxyService {
	return &AgentProxyService{
		registry: registry,
		logger:   logger,
	}
}

// getSessionClient 获取指定节点的 SessionService gRPC client
func (s *AgentProxyService) getSessionClient(ctx context.Context, nodeName string) (pb.SessionServiceClient, *grpc.ClientConn, error) {
	node := s.getNode(nodeName)
	if node == nil {
		return nil, nil, fmt.Errorf("node %q not found", nodeName)
	}
	conn, err := s.dial(ctx, node.GrpcAddress, nodeName)
	if err != nil {
		return nil, nil, err
	}
	return pb.NewSessionServiceClient(conn), conn, nil
}

// getTemplateClient 获取指定节点的 TemplateService gRPC client
func (s *AgentProxyService) getTemplateClient(ctx context.Context, nodeName string) (pb.TemplateServiceClient, *grpc.ClientConn, error) {
	node := s.getNode(nodeName)
	if node == nil {
		return nil, nil, fmt.Errorf("node %q not found", nodeName)
	}
	conn, err := s.dial(ctx, node.GrpcAddress, nodeName)
	if err != nil {
		return nil, nil, err
	}
	return pb.NewTemplateServiceClient(conn), conn, nil
}

// getConfigClient 获取指定节点的 ConfigService gRPC client
func (s *AgentProxyService) getConfigClient(ctx context.Context, nodeName string) (pb.ConfigServiceClient, *grpc.ClientConn, error) {
	node := s.getNode(nodeName)
	if node == nil {
		return nil, nil, fmt.Errorf("node %q not found", nodeName)
	}
	conn, err := s.dial(ctx, node.GrpcAddress, nodeName)
	if err != nil {
		return nil, nil, err
	}
	return pb.NewConfigServiceClient(conn), conn, nil
}

func (s *AgentProxyService) getNode(nodeName string) *model.Node {
	return s.registry.Get(nodeName)
}

func (s *AgentProxyService) dial(ctx context.Context, addr, nodeName string) (*grpc.ClientConn, error) {
	if addr == "" {
		return nil, fmt.Errorf("node %q has no gRPC address", nodeName)
	}

	// 设置连接超时
	dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("dial %s for %s: %w", addr, nodeName, err)
	}

	// 等待连接就绪
	if conn.GetState() != connectivity.Ready {
		s.logger.Infof("[agent-proxy] connecting to %s (%s) state=%v", nodeName, addr, conn.GetState())
	}

	return conn, nil
}

// SessionService 转发方法

func (s *AgentProxyService) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
	client, conn, err := s.getSessionClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.ListSessions(ctx, req)
}

func (s *AgentProxyService) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.Session, error) {
	client, conn, err := s.getSessionClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.CreateSession(ctx, req)
}

func (s *AgentProxyService) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.Session, error) {
	client, conn, err := s.getSessionClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.GetSession(ctx, req)
}

func (s *AgentProxyService) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*emptypb.Empty, error) {
	client, conn, err := s.getSessionClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.DeleteSession(ctx, req)
}

func (s *AgentProxyService) GetSessionConfig(ctx context.Context, req *pb.GetSessionConfigRequest) (*pb.SessionConfigView, error) {
	client, conn, err := s.getSessionClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.GetSessionConfig(ctx, req)
}

func (s *AgentProxyService) UpdateSessionConfig(ctx context.Context, req *pb.UpdateSessionConfigRequest) (*pb.SessionConfigView, error) {
	client, conn, err := s.getSessionClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.UpdateSessionConfig(ctx, req)
}

func (s *AgentProxyService) RestoreSession(ctx context.Context, req *pb.RestoreSessionRequest) (*emptypb.Empty, error) {
	client, conn, err := s.getSessionClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.RestoreSession(ctx, req)
}

func (s *AgentProxyService) SaveSessions(ctx context.Context, req *pb.SaveSessionsRequest) (*pb.SaveSessionsResponse, error) {
	client, conn, err := s.getSessionClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.SaveSessions(ctx, req)
}

func (s *AgentProxyService) GetSavedSessions(ctx context.Context, req *pb.GetSavedSessionsRequest) (*pb.GetSavedSessionsResponse, error) {
	client, conn, err := s.getSessionClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.GetSavedSessions(ctx, req)
}

func (s *AgentProxyService) GetOutput(ctx context.Context, req *pb.GetOutputRequest) (*pb.TerminalOutput, error) {
	client, conn, err := s.getSessionClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.GetOutput(ctx, req)
}

func (s *AgentProxyService) SendInput(ctx context.Context, req *pb.SendInputRequest) (*emptypb.Empty, error) {
	client, conn, err := s.getSessionClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.SendInput(ctx, req)
}

func (s *AgentProxyService) SendSignal(ctx context.Context, req *pb.SendSignalRequest) (*emptypb.Empty, error) {
	client, conn, err := s.getSessionClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.SendSignal(ctx, req)
}

func (s *AgentProxyService) GetEnv(ctx context.Context, req *pb.GetEnvRequest) (*pb.GetEnvResponse, error) {
	client, conn, err := s.getSessionClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.GetEnv(ctx, req)
}

// TemplateService 转发方法

func (s *AgentProxyService) ListTemplates(ctx context.Context, req *pb.ListTemplatesRequest) (*pb.ListTemplatesResponse, error) {
	client, conn, err := s.getTemplateClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.ListTemplates(ctx, req)
}

func (s *AgentProxyService) CreateTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*pb.SessionTemplate, error) {
	client, conn, err := s.getTemplateClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.CreateTemplate(ctx, req)
}

func (s *AgentProxyService) GetTemplate(ctx context.Context, req *pb.GetTemplateRequest) (*pb.SessionTemplate, error) {
	client, conn, err := s.getTemplateClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.GetTemplate(ctx, req)
}

func (s *AgentProxyService) UpdateTemplate(ctx context.Context, req *pb.UpdateTemplateRequest) (*pb.SessionTemplate, error) {
	client, conn, err := s.getTemplateClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.UpdateTemplate(ctx, req)
}

func (s *AgentProxyService) DeleteTemplate(ctx context.Context, req *pb.DeleteTemplateRequest) (*emptypb.Empty, error) {
	client, conn, err := s.getTemplateClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.DeleteTemplate(ctx, req)
}

func (s *AgentProxyService) GetTemplateConfig(ctx context.Context, req *pb.GetTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	client, conn, err := s.getTemplateClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.GetTemplateConfig(ctx, req)
}

func (s *AgentProxyService) UpdateTemplateConfig(ctx context.Context, req *pb.UpdateTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	client, conn, err := s.getTemplateClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.UpdateTemplateConfig(ctx, req)
}

// ConfigService 转发方法

func (s *AgentProxyService) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.LocalAgentConfig, error) {
	client, conn, err := s.getConfigClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.GetConfig(ctx, req)
}

func (s *AgentProxyService) UpdateConfig(ctx context.Context, req *pb.UpdateConfigRequest) (*pb.LocalAgentConfig, error) {
	client, conn, err := s.getConfigClient(ctx, req.NodeName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return client.UpdateConfig(ctx, req)
}
