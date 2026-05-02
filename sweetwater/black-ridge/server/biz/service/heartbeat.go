package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/sweetwater-black-ridge/biz/config"
	"github.com/charviki/sweetwater-black-ridge/biz/model"
)

const (
	// AgentVersion Agent 当前版本号
	AgentVersion = "0.1.0"
	// MaxSessions Agent 最大并行 Session 数
	MaxSessions  = 10

	backoffBase       = 10 * time.Second
	backoffMax        = 5 * time.Minute
	backoffMultiplier = 2
)

var supportedTemplates = []string{"claude", "bash"}

// HeartbeatService 心跳服务，负责向 Agent Manager 注册并定期上报存活状态。
// 状态机逻辑：未注册 → 注册 → 已注册 → 心跳。
// 失败时采用指数退避策略避免对 Manager 造成持续压力。
type HeartbeatService struct {
	cfg          *config.Config
	tmuxService  TmuxService
	localConfig  *LocalConfigStore
	client       *http.Client
	registered   bool
	logger       logutil.Logger
	startedAt    time.Time
	currentDelay time.Duration
}

// NewHeartbeatService 创建 HeartbeatService，使用 5 秒超时的 HTTP 客户端
func NewHeartbeatService(cfg *config.Config, tmuxService TmuxService, localConfig *LocalConfigStore, logger logutil.Logger) *HeartbeatService {
	return &HeartbeatService{
		cfg:         cfg,
		tmuxService: tmuxService,
		localConfig: localConfig,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		logger:    logger,
		startedAt: time.Now(),
	}
}

// Start 启动心跳循环。状态机逻辑：
// 1. 未注册 → 发送注册请求，失败则下个周期重试（指数退避）
// 2. 已注册 → 发送心跳，失败则标记为未注册，下个周期重新注册（指数退避）
// 通过 stopCh 实现优雅停止
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

// collectStatus 收集当前 Agent 运行状态快照（CPU、内存、Session 详情、本地配置）
func (s *HeartbeatService) collectStatus() protocol.AgentStatus {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 将本地配置作为只读视图上报给 Manager
	var localConfig *protocol.LocalAgentConfig
	if s.localConfig != nil {
		cfg := s.localConfig.Get()
		localConfig = &cfg
	}

	status := protocol.AgentStatus{
		CPUUsage:       0, // Go 标准库不直接提供 CPU 使用率，需要 cgo 或外部工具；原型阶段填 0
		MemoryUsageMB:  float64(memStats.Alloc) / 1024 / 1024,
		WorkspaceRoot:  s.cfg.Workspace.RootDir,
		SessionDetails: s.collectSessionDetails(),
		LocalConfig:    localConfig,
	}
	status.ActiveSessions = len(status.SessionDetails)
	return status
}

// collectSessionDetails 收集所有活跃 tmux Session 的详细信息。
// template 和 working_dir 从 .session-state/*.json 状态文件读取，
// 运行时长基于 tmux session 创建时间推导。
func (s *HeartbeatService) collectSessionDetails() []protocol.SessionDetail {
	sessions, err := s.tmuxService.ListSessions()
	if err != nil || sessions == nil {
		return nil
	}

	// 从状态文件构建 session name → SessionState 映射
	savedStates := make(map[string]*model.SessionState)
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

		// 从 session 创建时间推导运行时长
		if t, err := time.Parse("2006-01-02 15:04:05", sess.CreatedAt); err == nil {
			detail.UptimeSeconds = int64(now.Sub(t).Seconds())
		}

		details = append(details, detail)
	}
	return details
}

// register 向 Manager 发送注册请求（携带 capabilities、status、metadata）
func (s *HeartbeatService) register(name, addr, externalAddr string) error {
	hostname, _ := os.Hostname()

	// AdvertisedAddr 优先：Docker 环境下 os.Hostname() 返回容器 ID 而非容器名，
	// Docker DNS 无法解析容器 ID，因此需要显式配置可被 Manager 解析的地址
	registerAddr := s.cfg.Server.AdvertisedAddr
	if registerAddr == "" {
		registerAddr = fmt.Sprintf("http://%s%s", getOwnHostname(), addr)
	}

	// GrpcAddress: 优先使用完整配置值，若仅有端口则拼接与 HTTP 相同的 hostname
	grpcAddr := s.cfg.Server.GRPCAddr
	if grpcAddr != "" && strings.HasPrefix(grpcAddr, ":") {
		grpcAddr = extractHostFromAddr(s.cfg.Server.AdvertisedAddr) + grpcAddr
	}

	reqBody := protocol.RegisterRequest{
		Name:         name,
		Address:      registerAddr,
		ExternalAddr: externalAddr,
		GrpcAddress:  grpcAddr,
		Capabilities: protocol.AgentCapabilities{
			SupportedTemplates: supportedTemplates,
			MaxSessions:        MaxSessions,
			Tools:              []string{"tmux", "filesystem"},
		},
		Status: s.collectStatus(),
		Metadata: protocol.AgentMetadata{
			Version:   AgentVersion,
			Hostname:  hostname,
			StartedAt: s.startedAt,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal register body: %w", err)
	}

	url := s.cfg.Controller.Addr + "/api/v1/nodes/register"
	return s.doRequest(url, body)
}

// heartbeat 向 Manager 发送心跳请求（携带完整状态快照）
func (s *HeartbeatService) heartbeat(name string) error {
	reqBody := protocol.HeartbeatRequest{
		Name:   name,
		Status: s.collectStatus(),
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal heartbeat body: %w", err)
	}

	url := s.cfg.Controller.Addr + "/api/v1/nodes/heartbeat"
	return s.doRequest(url, body)
}

// doRequest 向 Manager 发送带 Auth header 的 JSON POST 请求
func (s *HeartbeatService) doRequest(url string, body []byte) error {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.Controller.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.Controller.AuthToken)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("post request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == 401 {
		return errors.New("authentication failed: check controller.auth_token config")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("request returned status %d", resp.StatusCode)
	}
	return nil
}

// getOwnHostname 获取本机主机名，失败时回退到 localhost
func getOwnHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return hostname
}

// extractHostFromAddr 从 AdvertisedAddr（如 http://host:8080）中提取 hostname。
// AdvertisedAddr 在 Docker 环境中被设为容器名，Docker DNS 可解析。
func extractHostFromAddr(advertisedAddr string) string {
	addr := strings.TrimPrefix(advertisedAddr, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return getOwnHostname()
	}
	return host
}
