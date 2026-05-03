package service

import (
	"context"
	"fmt"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/mesa-hub-behavior-panel/internal/model"
	"google.golang.org/protobuf/types/known/emptypb"
)

// AgentProxyService Agent gRPC 代理 — Session/Template/Config 请求转发到 Agent 节点。
// 通过 ConnectionManager 复用 gRPC 长连接，避免每次请求都建连/断连。
type AgentProxyService struct {
	registry *model.NodeRegistry
	connMgr  *ConnectionManager
	logger   logutil.Logger
}

// NewAgentProxyService 创建 AgentProxyService。
func NewAgentProxyService(registry *model.NodeRegistry, connMgr *ConnectionManager, logger logutil.Logger) *AgentProxyService {
	return &AgentProxyService{
		registry: registry,
		connMgr:  connMgr,
		logger:   logger,
	}
}

// getSessionClient 获取指定节点的 SessionService gRPC client（连接由 ConnectionManager 管理）
func (s *AgentProxyService) getSessionClient(ctx context.Context, nodeName string) (pb.SessionServiceClient, error) {
	addr, err := s.getNodeAddr(nodeName)
	if err != nil {
		return nil, err
	}
	conn, err := s.connMgr.GetConn(ctx, addr)
	if err != nil {
		return nil, err
	}
	return pb.NewSessionServiceClient(conn), nil
}

// getTemplateClient 获取指定节点的 TemplateService gRPC client（连接由 ConnectionManager 管理）
func (s *AgentProxyService) getTemplateClient(ctx context.Context, nodeName string) (pb.TemplateServiceClient, error) {
	addr, err := s.getNodeAddr(nodeName)
	if err != nil {
		return nil, err
	}
	conn, err := s.connMgr.GetConn(ctx, addr)
	if err != nil {
		return nil, err
	}
	return pb.NewTemplateServiceClient(conn), nil
}

// getConfigClient 获取指定节点的 ConfigService gRPC client（连接由 ConnectionManager 管理）
func (s *AgentProxyService) getConfigClient(ctx context.Context, nodeName string) (pb.ConfigServiceClient, error) {
	addr, err := s.getNodeAddr(nodeName)
	if err != nil {
		return nil, err
	}
	conn, err := s.connMgr.GetConn(ctx, addr)
	if err != nil {
		return nil, err
	}
	return pb.NewConfigServiceClient(conn), nil
}

func (s *AgentProxyService) getNodeAddr(nodeName string) (string, error) {
	node := s.registry.Get(nodeName)
	if node == nil {
		return "", fmt.Errorf("node %q not found", nodeName)
	}
	if node.GrpcAddress == "" {
		return "", fmt.Errorf("node %q has no gRPC address", nodeName)
	}
	return node.GrpcAddress, nil
}

// SessionService 转发方法

// ListSessions 查询指定节点的 Session 列表
func (s *AgentProxyService) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
	client, err := s.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.ListSessions(ctx, req)
}

// CreateSession 创建新的 Session
func (s *AgentProxyService) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.Session, error) {
	client, err := s.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.CreateSession(ctx, req)
}

// GetSession 获取 Session 详情
func (s *AgentProxyService) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.Session, error) {
	client, err := s.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.GetSession(ctx, req)
}

// DeleteSession 删除指定 Session
func (s *AgentProxyService) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*emptypb.Empty, error) {
	client, err := s.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.DeleteSession(ctx, req)
}

// GetSessionConfig 获取 Session 配置
func (s *AgentProxyService) GetSessionConfig(ctx context.Context, req *pb.GetSessionConfigRequest) (*pb.SessionConfigView, error) {
	client, err := s.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.GetSessionConfig(ctx, req)
}

// UpdateSessionConfig 更新 Session 配置
func (s *AgentProxyService) UpdateSessionConfig(ctx context.Context, req *pb.UpdateSessionConfigRequest) (*pb.SessionConfigView, error) {
	client, err := s.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.UpdateSessionConfig(ctx, req)
}

// RestoreSession 恢复已终止的 Session
func (s *AgentProxyService) RestoreSession(ctx context.Context, req *pb.RestoreSessionRequest) (*emptypb.Empty, error) {
	client, err := s.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.RestoreSession(ctx, req)
}

// SaveSessions 保存 Session 快照
func (s *AgentProxyService) SaveSessions(ctx context.Context, req *pb.SaveSessionsRequest) (*pb.SaveSessionsResponse, error) {
	client, err := s.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.SaveSessions(ctx, req)
}

// GetSavedSessions 获取已保存的 Session 列表
func (s *AgentProxyService) GetSavedSessions(ctx context.Context, req *pb.GetSavedSessionsRequest) (*pb.GetSavedSessionsResponse, error) {
	client, err := s.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.GetSavedSessions(ctx, req)
}

// GetOutput 获取终端输出
func (s *AgentProxyService) GetOutput(ctx context.Context, req *pb.GetOutputRequest) (*pb.TerminalOutput, error) {
	client, err := s.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.GetOutput(ctx, req)
}

// SendInput 发送终端输入
func (s *AgentProxyService) SendInput(ctx context.Context, req *pb.SendInputRequest) (*emptypb.Empty, error) {
	client, err := s.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.SendInput(ctx, req)
}

// SendSignal 发送终端信号
func (s *AgentProxyService) SendSignal(ctx context.Context, req *pb.SendSignalRequest) (*emptypb.Empty, error) {
	client, err := s.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.SendSignal(ctx, req)
}

// GetEnv 获取 Agent 环境变量
func (s *AgentProxyService) GetEnv(ctx context.Context, req *pb.GetEnvRequest) (*pb.GetEnvResponse, error) {
	client, err := s.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.GetEnv(ctx, req)
}

// TemplateService 转发方法

// ListTemplates 查询模板列表
func (s *AgentProxyService) ListTemplates(ctx context.Context, req *pb.ListTemplatesRequest) (*pb.ListTemplatesResponse, error) {
	client, err := s.getTemplateClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.ListTemplates(ctx, req)
}

// CreateTemplate 创建新模板
func (s *AgentProxyService) CreateTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*pb.SessionTemplate, error) {
	client, err := s.getTemplateClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.CreateTemplate(ctx, req)
}

// GetTemplate 获取模板详情
func (s *AgentProxyService) GetTemplate(ctx context.Context, req *pb.GetTemplateRequest) (*pb.SessionTemplate, error) {
	client, err := s.getTemplateClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.GetTemplate(ctx, req)
}

// UpdateTemplate 更新模板
func (s *AgentProxyService) UpdateTemplate(ctx context.Context, req *pb.UpdateTemplateRequest) (*pb.SessionTemplate, error) {
	client, err := s.getTemplateClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.UpdateTemplate(ctx, req)
}

// DeleteTemplate 删除模板
func (s *AgentProxyService) DeleteTemplate(ctx context.Context, req *pb.DeleteTemplateRequest) (*emptypb.Empty, error) {
	client, err := s.getTemplateClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.DeleteTemplate(ctx, req)
}

// GetTemplateConfig 获取模板配置
func (s *AgentProxyService) GetTemplateConfig(ctx context.Context, req *pb.GetTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	client, err := s.getTemplateClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.GetTemplateConfig(ctx, req)
}

// UpdateTemplateConfig 更新模板配置
func (s *AgentProxyService) UpdateTemplateConfig(ctx context.Context, req *pb.UpdateTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	client, err := s.getTemplateClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.UpdateTemplateConfig(ctx, req)
}

// ConfigService 转发方法

// GetConfig 获取 Agent 本地配置
func (s *AgentProxyService) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.LocalAgentConfig, error) {
	client, err := s.getConfigClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.GetConfig(ctx, req)
}

// UpdateConfig 更新 Agent 本地配置
func (s *AgentProxyService) UpdateConfig(ctx context.Context, req *pb.UpdateConfigRequest) (*pb.LocalAgentConfig, error) {
	client, err := s.getConfigClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.UpdateConfig(ctx, req)
}
