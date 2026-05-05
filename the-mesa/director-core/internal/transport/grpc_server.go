package transport

import (
	"context"
	"math"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/maze/the-mesa/director-core/internal/agentclient"
	"github.com/charviki/maze/the-mesa/director-core/internal/service"
)

// Server Director Core 端 gRPC 服务器
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
	proxy    *agentclient.Proxy
	registry service.NodeRegistry
	logger   logutil.Logger
	// directorCoreAuthToken 复用 director-core server.auth_token，因为 Agent 的 server.auth_token
	// 也是由 Director Core 在部署 Host 时下发；Director Core 主动回调 Agent 的内部 RPC 必须携带它。
	directorCoreAuthToken string

	grpcServer *grpc.Server
}

// NewServer 创建 Director Core gRPC Server
func NewServer(
	hostSvc *service.HostService,
	nodeSvc *service.NodeService,
	auditSvc *service.AuditService,
	proxy *agentclient.Proxy,
	registry service.NodeRegistry,
	directorCoreAuthToken string,
	logger logutil.Logger,
) *Server {
	return &Server{
		hostSvc:          hostSvc,
		nodeSvc:          nodeSvc,
		auditSvc:         auditSvc,
		proxy:            proxy,
		registry:         registry,
		logger:           logger,
		directorCoreAuthToken: directorCoreAuthToken,
	}
}

// RegisterGRPC 将当前服务实现注册到给定 gRPC server。
// 这样 main 层可以用 cradle/lifecycle 统一管理监听与关闭，而业务层只关心接口实现。
func (s *Server) RegisterGRPC(grpcServer *grpc.Server) {
	s.grpcServer = grpcServer
	pb.RegisterHostServiceServer(grpcServer, s)
	pb.RegisterNodeServiceServer(grpcServer, s)
	pb.RegisterAuditServiceServer(grpcServer, s)
	pb.RegisterAgentServiceServer(grpcServer, s)
	pb.RegisterSessionServiceServer(grpcServer, s)
	pb.RegisterTemplateServiceServer(grpcServer, s)
	pb.RegisterConfigServiceServer(grpcServer, s)
}

// AgentService — Register / Heartbeat

// Register Agent 注册 gRPC 接口：校验参数 → 注册到 registry → 异步恢复已保存 session
func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetAddress() == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}

	protoReq := pbRegisterToProtocol(req)
	node, err := s.registry.Register(ctx, protoReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "register node: %v", err)
	}

	// 恢复 session 是后台异步操作，不应阻塞注册响应，故使用独立 context
	go s.restoreAgentSessions(req.GetName(), req.GetGrpcAddress()) //nolint:gosec

	return &pb.RegisterResponse{
		Name:   node.Name,
		Status: node.Status,
	}, nil
}

// Heartbeat Agent 心跳 gRPC 接口：校验参数 → 更新心跳和状态快照
func (s *Server) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	protoReq := pbHeartbeatToProtocol(req)
	node, err := s.registry.Heartbeat(ctx, protoReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update heartbeat: %v", err)
	}
	if node == nil {
		return nil, status.Error(codes.NotFound, "node not found")
	}

	return &pb.HeartbeatResponse{
		Name:   node.Name,
		Status: node.Status,
	}, nil
}

// restoreAgentSessions 在 Agent 注册后异步恢复已保存的 session。
// 通过 gRPC client 调用 Agent 的 GetSavedSessions 和 RestoreSession，
// 跳过 restore_strategy 为 "running" 的 session。
func (s *Server) restoreAgentSessions(name, grpcAddr string) {
	if grpcAddr == "" {
		s.logger.Warnf("[session-restore] node %s has no gRPC address, skip", name)
		return
	}

	conn, err := grpc.NewClient(grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		s.logger.Warnf("[session-restore] dial %s (%s) failed: %v", name, grpcAddr, err)
		return
	}
	defer func() { _ = conn.Close() }()

	client := pb.NewSessionServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ctx = attachBearerToken(ctx, s.directorCoreAuthToken)

	resp, err := client.GetSavedSessions(ctx, &pb.GetSavedSessionsRequest{NodeName: name})
	if err != nil {
		s.logger.Warnf("[session-restore] get saved sessions from %s failed: %v", name, err)
		return
	}

	if len(resp.GetSessions()) == 0 {
		s.logger.Infof("[session-restore] no saved sessions for %s", name)
		return
	}

	restored := 0
	for _, ss := range resp.GetSessions() {
		// "running" 表示 session 仍在运行，无需恢复
		if ss.GetRestoreStrategy() == "running" {
			continue
		}

		restoreCtx, restoreCancel := context.WithTimeout(context.Background(), 10*time.Second)
		restoreCtx = attachBearerToken(restoreCtx, s.directorCoreAuthToken)
		_, err := client.RestoreSession(restoreCtx, &pb.RestoreSessionRequest{
			NodeName: name,
			Id:       ss.GetSessionName(),
		})
		restoreCancel()

		if err != nil {
			s.logger.Warnf("[session-restore] restore session %s/%s failed: %v", name, ss.GetSessionName(), err)
			continue
		}

		restored++
		s.logger.Infof("[session-restore] restored session %s/%s", name, ss.GetSessionName())
	}

	if restored > 0 {
		s.logger.Infof("[session-restore] restored %d sessions for %s", restored, name)
	}
}

func attachBearerToken(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	return metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))
}

// pb → protocol 转换函数

func pbRegisterToProtocol(req *pb.RegisterRequest) protocol.RegisterRequest {
	return protocol.RegisterRequest{
		Name:         req.GetName(),
		Address:      req.GetAddress(),
		ExternalAddr: req.GetExternalAddr(),
		GrpcAddress:  req.GetGrpcAddress(),
		Capabilities: pbCapabilitiesToProtocol(req.GetCapabilities()),
		Status:       pbAgentStatusToProtocol(req.GetStatus()),
		Metadata:     pbMetadataToProtocol(req.GetMetadata()),
	}
}

func pbHeartbeatToProtocol(req *pb.HeartbeatRequest) protocol.HeartbeatRequest {
	return protocol.HeartbeatRequest{
		Name:   req.GetName(),
		Status: pbAgentStatusToProtocol(req.GetStatus()),
	}
}

func pbCapabilitiesToProtocol(c *pb.AgentCapabilities) protocol.AgentCapabilities {
	if c == nil {
		return protocol.AgentCapabilities{}
	}
	return protocol.AgentCapabilities{
		SupportedTemplates: c.GetSupportedTemplates(),
		MaxSessions:        int(c.GetMaxSessions()),
		Tools:              c.GetTools(),
	}
}

func pbAgentStatusToProtocol(s *pb.AgentStatus) protocol.AgentStatus {
	if s == nil {
		return protocol.AgentStatus{}
	}
	details := make([]protocol.SessionDetail, 0, len(s.GetSessionDetails()))
	for _, d := range s.GetSessionDetails() {
		if d == nil {
			continue
		}
		details = append(details, protocol.SessionDetail{
			ID:            d.GetId(),
			Template:      d.GetTemplate(),
			WorkingDir:    d.GetWorkingDir(),
			UptimeSeconds: d.GetUptimeSeconds(),
		})
	}
	var localCfg *protocol.LocalAgentConfig
	if lc := s.GetLocalConfig(); lc != nil {
		localCfg = &protocol.LocalAgentConfig{
			WorkingDir: lc.GetWorkingDir(),
			Env:        lc.GetEnv(),
		}
	}
	return protocol.AgentStatus{
		ActiveSessions: int(s.GetActiveSessions()),
		CPUUsage:       s.GetCpuUsage(),
		MemoryUsageMB:  s.GetMemoryUsageMb(),
		WorkspaceRoot:  s.GetWorkspaceRoot(),
		SessionDetails: details,
		LocalConfig:    localCfg,
	}
}

func pbMetadataToProtocol(m *pb.AgentMetadata) protocol.AgentMetadata {
	if m == nil {
		return protocol.AgentMetadata{}
	}
	startedAt, _ := time.Parse(time.RFC3339, m.GetStartedAt())
	return protocol.AgentMetadata{
		Version:   m.GetVersion(),
		Hostname:  m.GetHostname(),
		StartedAt: startedAt,
	}
}

// CreateHost 创建 Host（异步：持久化 HostSpec → Reconciler 构建/启动容器）
func (s *Server) CreateHost(ctx context.Context, req *pb.CreateHostRequest) (*pb.HostSpec, error) {
	protoReq := &protocol.CreateHostRequest{
		Name:        req.GetName(),
		Tools:       req.GetTools(),
		DisplayName: req.GetDisplayName(),
	}
	if req.GetResources() != nil {
		protoReq.Resources = protocol.ResourceLimits{
			CPULimit:    req.GetResources().GetCpuLimit(),
			MemoryLimit: req.GetResources().GetMemoryLimit(),
		}
	}

	spec, err := s.hostSvc.CreateHost(ctx, protoReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// CreateHost 是异步操作（Reconciler 后续构建/启动容器），返回 202 Accepted 与旧 handler 行为一致
	_ = grpc.SetHeader(ctx, metadata.Pairs("x-http-status", "202"))

	return hostSpecToProto(spec), nil
}

// ListHosts 返回所有 Host 信息列表
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

// GetHost 根据 Host 名称获取详情
func (s *Server) GetHost(ctx context.Context, req *pb.GetHostRequest) (*pb.HostInfo, error) {
	info, err := s.hostSvc.GetHost(ctx, req.GetName())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return hostInfoToProto(info), nil
}

// DeleteHost 删除指定 Host 及其关联资源
func (s *Server) DeleteHost(ctx context.Context, req *pb.DeleteHostRequest) (*emptypb.Empty, error) {
	if err := s.hostSvc.DeleteHost(ctx, req.GetName()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

// GetBuildLog 获取 Host 构建日志
func (s *Server) GetBuildLog(ctx context.Context, req *pb.GetBuildLogRequest) (*pb.GetBuildLogResponse, error) {
	log, err := s.hostSvc.GetBuildLog(ctx, req.GetName())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.GetBuildLogResponse{Log: log}, nil
}

// GetRuntimeLog 获取 Host 运行时日志
func (s *Server) GetRuntimeLog(ctx context.Context, req *pb.GetRuntimeLogRequest) (*pb.GetRuntimeLogResponse, error) {
	log, err := s.hostSvc.GetRuntimeLog(ctx, req.GetName())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.GetRuntimeLogResponse{Log: log}, nil
}

// ListTools 返回所有可用的工具配置
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

// ListNodes 返回所有 Agent 节点信息
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

// GetNode 根据节点名称获取节点信息
func (s *Server) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.NodeInfo, error) {
	node, err := s.nodeSvc.GetNode(ctx, req.GetName())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return modelNodeToProto(node), nil
}

// DeleteNode 删除指定节点
func (s *Server) DeleteNode(ctx context.Context, req *pb.DeleteNodeRequest) (*emptypb.Empty, error) {
	if err := s.nodeSvc.DeleteNode(ctx, req.GetName()); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &emptypb.Empty{}, nil
}

// AuditService — 调用 service.AuditService

// GetAuditLogs 分页查询审计日志
func (s *Server) GetAuditLogs(ctx context.Context, req *pb.GetAuditLogsRequest) (*pb.GetAuditLogsResponse, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
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
		Total:    safeInt32(result.Total),
		Page:     safeInt32(result.Page),
		PageSize: safeInt32(result.PageSize),
	}, nil
}

// SessionService — gRPC 转发到 Agent

// ListSessions 查询指定节点的 Session 列表
func (s *Server) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
	return s.proxy.ListSessions(ctx, req)
}

// CreateSession 创建新的 Session
func (s *Server) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.Session, error) {
	return s.proxy.CreateSession(ctx, req)
}

// GetSession 获取 Session 详情
func (s *Server) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.Session, error) {
	return s.proxy.GetSession(ctx, req)
}

// DeleteSession 删除 Session
func (s *Server) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*emptypb.Empty, error) {
	return s.proxy.DeleteSession(ctx, req)
}

// GetSessionConfig 获取 Session 配置
func (s *Server) GetSessionConfig(ctx context.Context, req *pb.GetSessionConfigRequest) (*pb.SessionConfigView, error) {
	return s.proxy.GetSessionConfig(ctx, req)
}

// UpdateSessionConfig 更新 Session 配置
func (s *Server) UpdateSessionConfig(ctx context.Context, req *pb.UpdateSessionConfigRequest) (*pb.SessionConfigView, error) {
	return s.proxy.UpdateSessionConfig(ctx, req)
}

// RestoreSession 恢复已终止的 Session
func (s *Server) RestoreSession(ctx context.Context, req *pb.RestoreSessionRequest) (*emptypb.Empty, error) {
	return s.proxy.RestoreSession(ctx, req)
}

// SaveSessions 保存 Session 快照
func (s *Server) SaveSessions(ctx context.Context, req *pb.SaveSessionsRequest) (*pb.SaveSessionsResponse, error) {
	return s.proxy.SaveSessions(ctx, req)
}

// GetSavedSessions 获取已保存的 Session 列表
func (s *Server) GetSavedSessions(ctx context.Context, req *pb.GetSavedSessionsRequest) (*pb.GetSavedSessionsResponse, error) {
	return s.proxy.GetSavedSessions(ctx, req)
}

// GetOutput 获取终端输出
func (s *Server) GetOutput(ctx context.Context, req *pb.GetOutputRequest) (*pb.TerminalOutput, error) {
	return s.proxy.GetOutput(ctx, req)
}

// SendInput 发送终端输入
func (s *Server) SendInput(ctx context.Context, req *pb.SendInputRequest) (*emptypb.Empty, error) {
	return s.proxy.SendInput(ctx, req)
}

// SendSignal 发送终端信号
func (s *Server) SendSignal(ctx context.Context, req *pb.SendSignalRequest) (*emptypb.Empty, error) {
	return s.proxy.SendSignal(ctx, req)
}

// GetEnv 获取 Agent 环境变量
func (s *Server) GetEnv(ctx context.Context, req *pb.GetEnvRequest) (*pb.GetEnvResponse, error) {
	return s.proxy.GetEnv(ctx, req)
}

// TemplateService — gRPC 转发到 Agent

// ListTemplates 查询模板列表
func (s *Server) ListTemplates(ctx context.Context, req *pb.ListTemplatesRequest) (*pb.ListTemplatesResponse, error) {
	return s.proxy.ListTemplates(ctx, req)
}

// CreateTemplate 创建新模板
func (s *Server) CreateTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*pb.SessionTemplate, error) {
	return s.proxy.CreateTemplate(ctx, req)
}

// GetTemplate 获取模板详情
func (s *Server) GetTemplate(ctx context.Context, req *pb.GetTemplateRequest) (*pb.SessionTemplate, error) {
	return s.proxy.GetTemplate(ctx, req)
}

// UpdateTemplate 更新模板
func (s *Server) UpdateTemplate(ctx context.Context, req *pb.UpdateTemplateRequest) (*pb.SessionTemplate, error) {
	return s.proxy.UpdateTemplate(ctx, req)
}

// DeleteTemplate 删除模板
func (s *Server) DeleteTemplate(ctx context.Context, req *pb.DeleteTemplateRequest) (*emptypb.Empty, error) {
	return s.proxy.DeleteTemplate(ctx, req)
}

// GetTemplateConfig 获取模板配置
func (s *Server) GetTemplateConfig(ctx context.Context, req *pb.GetTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	return s.proxy.GetTemplateConfig(ctx, req)
}

// UpdateTemplateConfig 更新模板配置
func (s *Server) UpdateTemplateConfig(ctx context.Context, req *pb.UpdateTemplateConfigRequest) (*pb.TemplateConfigView, error) {
	return s.proxy.UpdateTemplateConfig(ctx, req)
}

// ConfigService — gRPC 转发到 Agent

// GetConfig 获取 Agent 本地配置
func (s *Server) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.LocalAgentConfig, error) {
	return s.proxy.GetConfig(ctx, req)
}

// UpdateConfig 更新 Agent 本地配置
func (s *Server) UpdateConfig(ctx context.Context, req *pb.UpdateConfigRequest) (*pb.LocalAgentConfig, error) {
	return s.proxy.UpdateConfig(ctx, req)
}

// 转换函数

func modelNodeToProto(n *service.Node) *pb.NodeInfo {
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
		MaxSessions:        safeInt32(c.MaxSessions),
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
		RetryCount:    safeInt32(info.RetryCount),
		Address:       info.Address,
		SessionCount:  safeInt32(info.SessionCount),
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
		StatusCode:     safeInt32(e.StatusCode),
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
		RetryCount:  safeInt32(spec.RetryCount),
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

func safeInt32(n int) int32 {
	if n > math.MaxInt32 {
		return math.MaxInt32
	}
	if n < math.MinInt32 {
		return math.MinInt32
	}
	return int32(n)
}
