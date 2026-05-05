package service

import (
	"context"
	"fmt"
	"math"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/sweetwater-black-ridge/internal/config"
)

const (
	// AgentVersion Agent 当前版本号
	AgentVersion = "0.1.0"
	// MaxSessions Agent 最大并行 Session 数
	MaxSessions = 10

	backoffBase       = 10 * time.Second
	backoffMax        = 5 * time.Minute
	backoffMultiplier = 2
)

var supportedTemplates = []string{"claude", "bash"}

// HeartbeatService 心跳服务，通过 gRPC 向 Director Core 注册并定期上报存活状态。
type HeartbeatService struct {
	cfg          *config.Config
	tmuxService  TmuxService
	localConfig  *LocalConfigStore
	grpcConn     *grpc.ClientConn
	agentClient  pb.AgentServiceClient
	registered   bool
	logger       logutil.Logger
	startedAt    time.Time
	currentDelay time.Duration
}

// NewHeartbeatService 创建 HeartbeatService，建立到 Director Core 的 gRPC 连接
func NewHeartbeatService(cfg *config.Config, tmuxService TmuxService, localConfig *LocalConfigStore, logger logutil.Logger) (*HeartbeatService, error) {
	directorCoreAddr := resolveDirectorCoreGRPCAddr(cfg.Controller)
	conn, err := grpc.NewClient(directorCoreAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to director-core %s: %w", directorCoreAddr, err)
	}

	return &HeartbeatService{
		cfg:         cfg,
		tmuxService: tmuxService,
		localConfig: localConfig,
		grpcConn:    conn,
		agentClient: pb.NewAgentServiceClient(conn),
		logger:      logger,
		startedAt:   time.Now(),
	}, nil
}

// resolveDirectorCoreGRPCAddr 解析 Director Core gRPC 地址：优先使用显式配置的 GRPCAddr，
// 否则从 Controller.Addr（HTTP 地址）中提取 host 部分 + 默认 gRPC 端口 9090。
func resolveDirectorCoreGRPCAddr(ctrl config.ControllerConfig) string {
	if ctrl.GRPCAddr != "" {
		return ctrl.GRPCAddr
	}
	// 从 HTTP 地址推导 gRPC 地址：去掉 scheme，取 host:9090
	addr := strings.TrimPrefix(ctrl.Addr, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// 无法解析 host:port，直接使用原始地址（去掉 scheme 后）
		return addr
	}
	return host + ":9090"
}

// Start 启动心跳循环
func (s *HeartbeatService) Start(stopCh <-chan struct{}) {
	if !s.cfg.Controller.Enabled || s.cfg.Controller.Addr == "" {
		s.logger.Infof("[heartbeat] controller not configured, skipping registration")
		return
	}

	name := s.cfg.Server.Name
	if name == "" {
		hostname, _ := os.Hostname()
		name = hostname
	}

	addr := s.cfg.Server.ListenAddr
	externalAddr := s.cfg.Server.ExternalAddr
	if externalAddr == "" {
		externalAddr = "http://localhost" + addr
	}

	baseInterval := time.Duration(s.cfg.Controller.HeartbeatInterval) * time.Second
	s.currentDelay = baseInterval

	for {
		if !s.registered {
			if err := s.register(name, addr, externalAddr); err != nil {
				s.logger.Errorf("[heartbeat] register failed: %v, retry in %v", err, s.currentDelay)
			} else {
				s.registered = true
				s.currentDelay = baseInterval
				s.logger.Infof("[heartbeat] registered as %s", name)
			}
		} else {
			if err := s.heartbeat(name); err != nil {
				s.logger.Errorf("[heartbeat] heartbeat failed: %v, retry in %v", err, s.currentDelay)
				s.registered = false
			} else {
				s.currentDelay = baseInterval
			}
		}

		select {
		case <-stopCh:
			s.logger.Infof("[heartbeat] stopped")
			return
		case <-time.After(s.currentDelay):
		}

		// 失败后指数退避
		if !s.registered {
			s.currentDelay *= backoffMultiplier
			if s.currentDelay > backoffMax {
				s.currentDelay = backoffMax
			}
		}
	}
}

// Stop 关闭 gRPC 连接
func (s *HeartbeatService) Stop() {
	if s.grpcConn != nil {
		_ = s.grpcConn.Close()
	}
}

// collectStatus 收集当前 Agent 运行状态快照
func (s *HeartbeatService) collectStatus() protocol.AgentStatus {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	var localConfig *protocol.LocalAgentConfig
	if s.localConfig != nil {
		cfg := s.localConfig.Get()
		localConfig = &cfg
	}

	status := protocol.AgentStatus{
		MemoryUsageMB:  float64(memStats.Alloc) / 1024 / 1024,
		WorkspaceRoot:  s.cfg.Workspace.RootDir,
		SessionDetails: s.collectSessionDetails(),
		LocalConfig:    localConfig,
	}
	status.ActiveSessions = len(status.SessionDetails)
	return status
}

// collectSessionDetails 收集所有活跃 tmux Session 的详细信息
func (s *HeartbeatService) collectSessionDetails() []protocol.SessionDetail {
	sessions, err := s.tmuxService.ListSessions()
	if err != nil || sessions == nil {
		return nil
	}

	savedStates := make(map[string]*SessionState)
	if saved, err := s.tmuxService.GetSavedSessions(); err == nil {
		for i := range saved {
			savedStates[saved[i].SessionName] = &saved[i]
		}
	}

	now := time.Now()
	details := make([]protocol.SessionDetail, 0, len(sessions))
	for _, sess := range sessions {
		detail := protocol.SessionDetail{
			ID: sess.ID,
		}

		if state, ok := savedStates[sess.Name]; ok {
			detail.Template = state.TemplateID
			detail.WorkingDir = state.WorkingDir
		}

		if t, err := time.Parse("2006-01-02 15:04:05", sess.CreatedAt); err == nil {
			detail.UptimeSeconds = int64(now.Sub(t).Seconds())
		}

		details = append(details, detail)
	}
	return details
}

// register 通过 gRPC 向 Director Core 发送注册请求
func (s *HeartbeatService) register(name, addr, externalAddr string) error {
	registerAddr := s.cfg.Server.AdvertisedAddr
	if registerAddr == "" {
		registerAddr = fmt.Sprintf("http://%s%s", getOwnHostname(), addr)
	}

	grpcAddr := s.cfg.Server.GRPCAddr
	if grpcAddr != "" && strings.HasPrefix(grpcAddr, ":") {
		grpcAddr = extractHostFromAddr(s.cfg.Server.AdvertisedAddr) + grpcAddr
	}

	status := s.collectStatus()
	hostname, _ := os.Hostname()

	req := &pb.RegisterRequest{
		Name:         name,
		Address:      registerAddr,
		ExternalAddr: externalAddr,
		GrpcAddress:  grpcAddr,
		Capabilities: &pb.AgentCapabilities{
			SupportedTemplates: supportedTemplates,
			MaxSessions:        int32(MaxSessions),
			Tools:              []string{"tmux", "filesystem"},
		},
		Status: &pb.AgentStatus{
			CpuUsage:       status.CPUUsage,
			MemoryUsageMb:  status.MemoryUsageMB,
			WorkspaceRoot:  status.WorkspaceRoot,
			ActiveSessions: int32(min(status.ActiveSessions, math.MaxInt32)), //nolint:gosec
		},
		Metadata: &pb.AgentMetadata{
			Version:   AgentVersion,
			Hostname:  hostname,
			StartedAt: s.startedAt.Format(time.RFC3339),
		},
	}

	ctx := s.withAuth(context.Background())
	_, err := s.agentClient.Register(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC register: %w", err)
	}
	return nil
}

// heartbeat 通过 gRPC 向 Director Core 发送心跳
func (s *HeartbeatService) heartbeat(name string) error {
	status := s.collectStatus()

	req := &pb.HeartbeatRequest{
		Name: name,
		Status: &pb.AgentStatus{
			CpuUsage:       status.CPUUsage,
			MemoryUsageMb:  status.MemoryUsageMB,
			WorkspaceRoot:  status.WorkspaceRoot,
			ActiveSessions: int32(min(status.ActiveSessions, math.MaxInt32)), //nolint:gosec
		},
	}

	ctx := s.withAuth(context.Background())
	_, err := s.agentClient.Heartbeat(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC heartbeat: %w", err)
	}
	return nil
}

// withAuth 创建带认证 metadata 的 context
func (s *HeartbeatService) withAuth(ctx context.Context) context.Context {
	if s.cfg.Controller.AuthToken != "" {
		md := metadata.Pairs("authorization", "Bearer "+s.cfg.Controller.AuthToken)
		return metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

// getOwnHostname 获取本机主机名，失败时回退到 localhost
func getOwnHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return hostname
}

// extractHostFromAddr 从 AdvertisedAddr 中提取 hostname
func extractHostFromAddr(advertisedAddr string) string {
	addr := strings.TrimPrefix(advertisedAddr, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return getOwnHostname()
	}
	return host
}
