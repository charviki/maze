package agentclient

import (
	"context"
	"fmt"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/mesa-hub-behavior-panel/internal/service"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Proxy 负责把 Session/Template/Config RPC 转发到对应的 Agent 节点。
// 该职责属于 Manager 访问外部 Agent 的客户端层，不应继续混在业务 service 中。
type Proxy struct {
	registry service.NodeRegistry
	connMgr  *ConnectionManager
}

// NewProxy 创建 Agent RPC 代理。
func NewProxy(registry service.NodeRegistry, connMgr *ConnectionManager) *Proxy {
	return &Proxy{
		registry: registry,
		connMgr:  connMgr,
	}
}

// getSessionClient 获取指定节点的 SessionService client。
// 通过连接池复用长连接，是为了避免高频会话操作反复建连。
func (p *Proxy) getSessionClient(ctx context.Context, nodeName string) (pb.SessionServiceClient, error) {
	addr, err := p.getNodeAddr(ctx, nodeName)
	if err != nil {
		return nil, err
	}
	conn, err := p.connMgr.GetConn(ctx, addr)
	if err != nil {
		return nil, err
	}
	return pb.NewSessionServiceClient(conn), nil
}

// getTemplateClient 获取指定节点的 TemplateService client。
func (p *Proxy) getTemplateClient(ctx context.Context, nodeName string) (pb.TemplateServiceClient, error) {
	addr, err := p.getNodeAddr(ctx, nodeName)
	if err != nil {
		return nil, err
	}
	conn, err := p.connMgr.GetConn(ctx, addr)
	if err != nil {
		return nil, err
	}
	return pb.NewTemplateServiceClient(conn), nil
}

// getConfigClient 获取指定节点的 ConfigService client。
func (p *Proxy) getConfigClient(ctx context.Context, nodeName string) (pb.ConfigServiceClient, error) {
	addr, err := p.getNodeAddr(ctx, nodeName)
	if err != nil {
		return nil, err
	}
	conn, err := p.connMgr.GetConn(ctx, addr)
	if err != nil {
		return nil, err
	}
	return pb.NewConfigServiceClient(conn), nil
}

func (p *Proxy) getNodeAddr(ctx context.Context, nodeName string) (string, error) {
	node, err := p.registry.Get(ctx, nodeName)
	if err != nil {
		return "", fmt.Errorf("get node %q: %w", nodeName, err)
	}
	if node == nil {
		return "", fmt.Errorf("node %q not found", nodeName)
	}
	if node.GrpcAddress == "" {
		return "", fmt.Errorf("node %q has no gRPC address", nodeName)
	}
	return node.GrpcAddress, nil
}

// ListSessions 查询指定节点的 Session 列表。
func (p *Proxy) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
	client, err := p.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.ListSessions(ctx, req)
}

// CreateSession 创建新的 Session。
func (p *Proxy) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.Session, error) {
	client, err := p.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.CreateSession(ctx, req)
}

// GetSession 获取 Session 详情。
func (p *Proxy) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.Session, error) {
	client, err := p.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.GetSession(ctx, req)
}

// DeleteSession 删除指定 Session。
func (p *Proxy) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*emptypb.Empty, error) {
	client, err := p.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.DeleteSession(ctx, req)
}

// GetSessionConfig 获取 Session 配置。
func (p *Proxy) GetSessionConfig(ctx context.Context, req *pb.GetSessionConfigRequest) (*pb.SessionConfigView, error) {
	client, err := p.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.GetSessionConfig(ctx, req)
}

// UpdateSessionConfig 更新 Session 配置。
func (p *Proxy) UpdateSessionConfig(ctx context.Context, req *pb.UpdateSessionConfigRequest) (*pb.SessionConfigView, error) {
	client, err := p.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.UpdateSessionConfig(ctx, req)
}

// RestoreSession 恢复已终止的 Session。
func (p *Proxy) RestoreSession(ctx context.Context, req *pb.RestoreSessionRequest) (*emptypb.Empty, error) {
	client, err := p.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.RestoreSession(ctx, req)
}

// SaveSessions 保存 Session 快照。
func (p *Proxy) SaveSessions(ctx context.Context, req *pb.SaveSessionsRequest) (*pb.SaveSessionsResponse, error) {
	client, err := p.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.SaveSessions(ctx, req)
}

// GetSavedSessions 获取已保存的 Session 列表。
func (p *Proxy) GetSavedSessions(ctx context.Context, req *pb.GetSavedSessionsRequest) (*pb.GetSavedSessionsResponse, error) {
	client, err := p.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.GetSavedSessions(ctx, req)
}

// GetOutput 获取终端输出。
func (p *Proxy) GetOutput(ctx context.Context, req *pb.GetOutputRequest) (*pb.TerminalOutput, error) {
	client, err := p.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.GetOutput(ctx, req)
}

// SendInput 发送终端输入。
func (p *Proxy) SendInput(ctx context.Context, req *pb.SendInputRequest) (*emptypb.Empty, error) {
	client, err := p.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.SendInput(ctx, req)
}

// SendSignal 发送终端信号。
func (p *Proxy) SendSignal(ctx context.Context, req *pb.SendSignalRequest) (*emptypb.Empty, error) {
	client, err := p.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.SendSignal(ctx, req)
}

// GetEnv 获取 Agent 环境变量。
func (p *Proxy) GetEnv(ctx context.Context, req *pb.GetEnvRequest) (*pb.GetEnvResponse, error) {
	client, err := p.getSessionClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.GetEnv(ctx, req)
}

// ListTemplates 查询模板列表。
func (p *Proxy) ListTemplates(ctx context.Context, req *pb.ListTemplatesRequest) (*pb.ListTemplatesResponse, error) {
	client, err := p.getTemplateClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.ListTemplates(ctx, req)
}

// CreateTemplate 创建新模板。
func (p *Proxy) CreateTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*pb.SessionTemplate, error) {
	client, err := p.getTemplateClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.CreateTemplate(ctx, req)
}

// GetTemplate 获取模板详情。
func (p *Proxy) GetTemplate(ctx context.Context, req *pb.GetTemplateRequest) (*pb.SessionTemplate, error) {
	client, err := p.getTemplateClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.GetTemplate(ctx, req)
}

// UpdateTemplate 更新模板。
func (p *Proxy) UpdateTemplate(ctx context.Context, req *pb.UpdateTemplateRequest) (*pb.SessionTemplate, error) {
	client, err := p.getTemplateClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.UpdateTemplate(ctx, req)
}

// DeleteTemplate 删除模板。
func (p *Proxy) DeleteTemplate(ctx context.Context, req *pb.DeleteTemplateRequest) (*emptypb.Empty, error) {
	client, err := p.getTemplateClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.DeleteTemplate(ctx, req)
}

// GetTemplateConfig 获取模板配置。
func (p *Proxy) GetTemplateConfig(ctx context.Context, req *pb.GetTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	client, err := p.getTemplateClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.GetTemplateConfig(ctx, req)
}

// UpdateTemplateConfig 更新模板配置。
func (p *Proxy) UpdateTemplateConfig(ctx context.Context, req *pb.UpdateTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	client, err := p.getTemplateClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.UpdateTemplateConfig(ctx, req)
}

// GetConfig 获取 Agent 本地配置。
func (p *Proxy) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.LocalAgentConfig, error) {
	client, err := p.getConfigClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.GetConfig(ctx, req)
}

// UpdateConfig 更新 Agent 本地配置。
func (p *Proxy) UpdateConfig(ctx context.Context, req *pb.UpdateConfigRequest) (*pb.LocalAgentConfig, error) {
	client, err := p.getConfigClient(ctx, req.GetNodeName())
	if err != nil {
		return nil, err
	}
	return client.UpdateConfig(ctx, req)
}
