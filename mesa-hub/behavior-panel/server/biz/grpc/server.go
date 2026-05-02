package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/biz/model"
	"github.com/charviki/mesa-hub-behavior-panel/biz/service"
)

// Server Manager 端 gRPC 服务器
type Server struct {
	pb.UnimplementedHostServiceServer
	pb.UnimplementedNodeServiceServer
	pb.UnimplementedAuditServiceServer
	pb.UnimplementedAgentServiceServer
	pb.UnimplementedSessionServiceServer
	pb.UnimplementedTemplateServiceServer
	pb.UnimplementedConfigServiceServer

	hostSvc  *service.HostService
	nodeSvc  *service.NodeService
	auditSvc *service.AuditService
	proxy    *service.AgentProxyService
	logger   logutil.Logger

	grpcServer *grpc.Server
}

// NewServer 创建 Manager gRPC Server
func NewServer(
	hostSvc *service.HostService,
	nodeSvc *service.NodeService,
	auditSvc *service.AuditService,
	proxy *service.AgentProxyService,
	logger logutil.Logger,
) *Server {
	return &Server{
		hostSvc:  hostSvc,
		nodeSvc:  nodeSvc,
		auditSvc: auditSvc,
		proxy:    proxy,
		logger:   logger,
	}
}

// Start 启动 gRPC server（非阻塞）
func (s *Server) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("grpc listen %s: %w", addr, err)
	}

	s.grpcServer = grpc.NewServer()
	pb.RegisterHostServiceServer(s.grpcServer, s)
	pb.RegisterNodeServiceServer(s.grpcServer, s)
	pb.RegisterAuditServiceServer(s.grpcServer, s)
	pb.RegisterAgentServiceServer(s.grpcServer, s)
	pb.RegisterSessionServiceServer(s.grpcServer, s)
	pb.RegisterTemplateServiceServer(s.grpcServer, s)
	pb.RegisterConfigServiceServer(s.grpcServer, s)

	go func() {
		s.logger.Infof("[grpc] server started on %s", addr)
		if err := s.grpcServer.Serve(lis); err != nil {
			s.logger.Errorf("[grpc] server error: %v", err)
		}
	}()
	return nil
}

// Stop 优雅关闭 gRPC server
func (s *Server) Stop() {
	if s.grpcServer != nil {
		stopped := make(chan struct{})
		go func() {
			s.grpcServer.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
		case <-time.After(5 * time.Second):
			s.grpcServer.Stop()
		}
	}
}

// GrpcServer 返回底层 gRPC Server，供 gateway 进程内直连使用
func (s *Server) GrpcServer() *grpc.Server {
	return s.grpcServer
}

// AgentService — 保持 Unimplemented

func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return nil, status.Error(codes.Unimplemented, "use HTTP POST /api/v1/nodes/register")
}

func (s *Server) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	return nil, status.Error(codes.Unimplemented, "use HTTP POST /api/v1/nodes/heartbeat")
}

// HostService — 调用 service.HostService（内部做 proto ↔ protocol 转换）

func (s *Server) CreateHost(ctx context.Context, req *pb.CreateHostRequest) (*pb.HostSpec, error) {
	protoReq := &protocol.CreateHostRequest{
		Name:        req.Name,
		Tools:       req.Tools,
		DisplayName: req.DisplayName,
	}
	if req.Resources != nil {
		protoReq.Resources = protocol.ResourceLimits{
			CPULimit:    req.Resources.CpuLimit,
			MemoryLimit: req.Resources.MemoryLimit,
		}
	}

	spec, err := s.hostSvc.CreateHost(ctx, protoReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return hostSpecToProto(spec), nil
}

func (s *Server) ListHosts(ctx context.Context, req *pb.ListHostsRequest) (*pb.ListHostsResponse, error) {
	hosts, err := s.hostSvc.ListHosts(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	pbHosts := make([]*pb.HostInfo, len(hosts))
	for i, h := range hosts {
		pbHosts[i] = hostInfoToProto(h)
	}
	return &pb.ListHostsResponse{Hosts: pbHosts}, nil
}

func (s *Server) GetHost(ctx context.Context, req *pb.GetHostRequest) (*pb.HostInfo, error) {
	info, err := s.hostSvc.GetHost(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return hostInfoToProto(info), nil
}

func (s *Server) DeleteHost(ctx context.Context, req *pb.DeleteHostRequest) (*emptypb.Empty, error) {
	if err := s.hostSvc.DeleteHost(ctx, req.Name); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetBuildLog(ctx context.Context, req *pb.GetBuildLogRequest) (*pb.GetBuildLogResponse, error) {
	log, err := s.hostSvc.GetBuildLog(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.GetBuildLogResponse{Log: log}, nil
}

func (s *Server) GetRuntimeLog(ctx context.Context, req *pb.GetRuntimeLogRequest) (*pb.GetRuntimeLogResponse, error) {
	log, err := s.hostSvc.GetRuntimeLog(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.GetRuntimeLogResponse{Log: log}, nil
}

func (s *Server) ListTools(ctx context.Context, req *pb.ListToolsRequest) (*pb.ListToolsResponse, error) {
	tools, err := s.hostSvc.ListTools(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	pbTools := make([]*pb.ToolConfig, len(tools))
	for i, t := range tools {
		pbTools[i] = toolConfigToProto(t)
	}
	return &pb.ListToolsResponse{Tools: pbTools}, nil
}

// NodeService — 调用 service.NodeService

func (s *Server) ListNodes(ctx context.Context, req *pb.ListNodesRequest) (*pb.ListNodesResponse, error) {
	nodes, err := s.nodeSvc.ListNodes(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	pbNodes := make([]*pb.NodeInfo, len(nodes))
	for i, n := range nodes {
		pbNodes[i] = modelNodeToProto(n)
	}
	return &pb.ListNodesResponse{Nodes: pbNodes}, nil
}

func (s *Server) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.NodeInfo, error) {
	node, err := s.nodeSvc.GetNode(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return modelNodeToProto(node), nil
}

func (s *Server) DeleteNode(ctx context.Context, req *pb.DeleteNodeRequest) (*emptypb.Empty, error) {
	if err := s.nodeSvc.DeleteNode(ctx, req.Name); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &emptypb.Empty{}, nil
}

// AuditService — 调用 service.AuditService

func (s *Server) GetAuditLogs(ctx context.Context, req *pb.GetAuditLogsRequest) (*pb.GetAuditLogsResponse, error) {
	page := int(req.Page)
	pageSize := int(req.PageSize)
	result, err := s.auditSvc.GetAuditLogs(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	pbLogs := make([]*pb.AuditLogEntry, len(result.Logs))
	for i, l := range result.Logs {
		pbLogs[i] = auditEntryToProto(l)
	}
	return &pb.GetAuditLogsResponse{
		Logs:     pbLogs,
		Total:    int32(result.Total),
		Page:     int32(result.Page),
		PageSize: int32(result.PageSize),
	}, nil
}

// SessionService — gRPC 转发到 Agent

func (s *Server) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
	return s.proxy.ListSessions(ctx, req)
}

func (s *Server) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.Session, error) {
	return s.proxy.CreateSession(ctx, req)
}

func (s *Server) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.Session, error) {
	return s.proxy.GetSession(ctx, req)
}

func (s *Server) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*emptypb.Empty, error) {
	return s.proxy.DeleteSession(ctx, req)
}

func (s *Server) GetSessionConfig(ctx context.Context, req *pb.GetSessionConfigRequest) (*pb.SessionConfigView, error) {
	return s.proxy.GetSessionConfig(ctx, req)
}

func (s *Server) UpdateSessionConfig(ctx context.Context, req *pb.UpdateSessionConfigRequest) (*pb.SessionConfigView, error) {
	return s.proxy.UpdateSessionConfig(ctx, req)
}

func (s *Server) RestoreSession(ctx context.Context, req *pb.RestoreSessionRequest) (*emptypb.Empty, error) {
	return s.proxy.RestoreSession(ctx, req)
}

func (s *Server) SaveSessions(ctx context.Context, req *pb.SaveSessionsRequest) (*pb.SaveSessionsResponse, error) {
	return s.proxy.SaveSessions(ctx, req)
}

func (s *Server) GetSavedSessions(ctx context.Context, req *pb.GetSavedSessionsRequest) (*pb.GetSavedSessionsResponse, error) {
	return s.proxy.GetSavedSessions(ctx, req)
}

func (s *Server) GetOutput(ctx context.Context, req *pb.GetOutputRequest) (*pb.TerminalOutput, error) {
	return s.proxy.GetOutput(ctx, req)
}

func (s *Server) SendInput(ctx context.Context, req *pb.SendInputRequest) (*emptypb.Empty, error) {
	return s.proxy.SendInput(ctx, req)
}

func (s *Server) SendSignal(ctx context.Context, req *pb.SendSignalRequest) (*emptypb.Empty, error) {
	return s.proxy.SendSignal(ctx, req)
}

func (s *Server) GetEnv(ctx context.Context, req *pb.GetEnvRequest) (*pb.GetEnvResponse, error) {
	return s.proxy.GetEnv(ctx, req)
}

// TemplateService — gRPC 转发到 Agent

func (s *Server) ListTemplates(ctx context.Context, req *pb.ListTemplatesRequest) (*pb.ListTemplatesResponse, error) {
	return s.proxy.ListTemplates(ctx, req)
}

func (s *Server) CreateTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*pb.SessionTemplate, error) {
	return s.proxy.CreateTemplate(ctx, req)
}

func (s *Server) GetTemplate(ctx context.Context, req *pb.GetTemplateRequest) (*pb.SessionTemplate, error) {
	return s.proxy.GetTemplate(ctx, req)
}

func (s *Server) UpdateTemplate(ctx context.Context, req *pb.UpdateTemplateRequest) (*pb.SessionTemplate, error) {
	return s.proxy.UpdateTemplate(ctx, req)
}

func (s *Server) DeleteTemplate(ctx context.Context, req *pb.DeleteTemplateRequest) (*emptypb.Empty, error) {
	return s.proxy.DeleteTemplate(ctx, req)
}

func (s *Server) GetTemplateConfig(ctx context.Context, req *pb.GetTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	return s.proxy.GetTemplateConfig(ctx, req)
}

func (s *Server) UpdateTemplateConfig(ctx context.Context, req *pb.UpdateTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	return s.proxy.UpdateTemplateConfig(ctx, req)
}

// ConfigService — gRPC 转发到 Agent

func (s *Server) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.LocalAgentConfig, error) {
	return s.proxy.GetConfig(ctx, req)
}

func (s *Server) UpdateConfig(ctx context.Context, req *pb.UpdateConfigRequest) (*pb.LocalAgentConfig, error) {
	return s.proxy.UpdateConfig(ctx, req)
}

// 转换函数

func modelNodeToProto(n *model.Node) *pb.NodeInfo {
	info := &pb.NodeInfo{
		Name:          n.Name,
		Address:       n.Address,
		ExternalAddr:  n.ExternalAddr,
		GrpcAddress:   n.GrpcAddress,
		Status:        n.Status,
		RegisteredAt:  n.RegisteredAt.Format(time.RFC3339),
		LastHeartbeat: n.LastHeartbeat.Format(time.RFC3339),
	}
	info.Capabilities = protocolCapabilitiesToProto(n.Capabilities)
	info.Metadata = protocolMetadataToProto(n.Metadata)
	return info
}

func protocolCapabilitiesToProto(c protocol.AgentCapabilities) *pb.AgentCapabilities {
	return &pb.AgentCapabilities{
		SupportedTemplates: c.SupportedTemplates,
		MaxSessions:        int32(c.MaxSessions),
		Tools:              c.Tools,
	}
}

func protocolMetadataToProto(m protocol.AgentMetadata) *pb.AgentMetadata {
	return &pb.AgentMetadata{
		Version:   m.Version,
		Hostname:  m.Hostname,
		StartedAt: m.StartedAt.Format(time.RFC3339),
	}
}

func hostInfoToProto(info *protocol.HostInfo) *pb.HostInfo {
	return &pb.HostInfo{
		Name:          info.Name,
		DisplayName:   info.DisplayName,
		Tools:         info.Tools,
		Resources:     resourceLimitsToProto(info.Resources),
		CreatedAt:     info.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     info.UpdatedAt.Format(time.RFC3339),
		Status:        info.Status,
		ErrorMsg:      info.ErrorMsg,
		RetryCount:    int32(info.RetryCount),
		Address:       info.Address,
		SessionCount:  int32(info.SessionCount),
		LastHeartbeat: info.LastHeartbeat,
	}
}

func toolConfigToProto(t protocol.ToolConfig) *pb.ToolConfig {
	return &pb.ToolConfig{
		Id:          t.ID,
		Image:       t.Image,
		EnvVars:     t.EnvVars,
		Description: t.Description,
		Category:    t.Category,
		SourcePath:  t.SourcePath,
		DestPath:    t.DestPath,
		BinPaths:    t.BinPaths,
	}
}

func auditEntryToProto(e protocol.AuditLogEntry) *pb.AuditLogEntry {
	return &pb.AuditLogEntry{
		Id:             e.ID,
		Timestamp:      e.Timestamp.Format(time.RFC3339),
		Operator:       e.Operator,
		TargetNode:     e.TargetNode,
		Action:         e.Action,
		PayloadSummary: e.PayloadSummary,
		Result:         e.Result,
		StatusCode:     int32(e.StatusCode),
	}
}

func hostSpecToProto(spec *protocol.HostSpec) *pb.HostSpec {
	if spec == nil {
		return nil
	}
	return &pb.HostSpec{
		Name:        spec.Name,
		DisplayName: spec.DisplayName,
		Tools:       spec.Tools,
		Resources:   resourceLimitsToProto(spec.Resources),
		AuthToken:   spec.AuthToken,
		CreatedAt:   spec.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   spec.UpdatedAt.Format(time.RFC3339),
		Status:      spec.Status,
		ErrorMsg:    spec.ErrorMsg,
		RetryCount:  int32(spec.RetryCount),
	}
}

func resourceLimitsToProto(limits protocol.ResourceLimits) *pb.ResourceLimits {
	if limits.CPULimit == "" && limits.MemoryLimit == "" {
		return nil
	}
	return &pb.ResourceLimits{
		CpuLimit:    limits.CPULimit,
		MemoryLimit: limits.MemoryLimit,
	}
}

var (
	_ pb.HostServiceServer     = (*Server)(nil)
	_ pb.NodeServiceServer     = (*Server)(nil)
	_ pb.AuditServiceServer    = (*Server)(nil)
	_ pb.AgentServiceServer    = (*Server)(nil)
	_ pb.SessionServiceServer  = (*Server)(nil)
	_ pb.TemplateServiceServer = (*Server)(nil)
	_ pb.ConfigServiceServer   = (*Server)(nil)
)
