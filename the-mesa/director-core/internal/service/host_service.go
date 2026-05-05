package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/maze/the-mesa/director-core/internal/config"
	hostbuilder "github.com/charviki/maze/the-mesa/director-core/internal/hostbuilder"
	"github.com/charviki/maze/the-mesa/director-core/internal/runtime"
)

// HostService Host 业务逻辑（Director Core 本地），供 HTTP handler 和 gRPC handler 共用
type HostService struct {
	registry      NodeRegistry
	hostSpecRepo  HostSpecRepository
	txm           HostTxManager
	runtime       runtime.HostRuntime
	auditLog      AuditLogWriter
	cfg           *config.Config
	logger        logutil.Logger
	logDir        string
	deployCancels map[uint64]context.CancelFunc
	nextDeployID  uint64
	deployWg      sync.WaitGroup
	deployMu      sync.Mutex
}

// NewHostService 创建 HostService
func NewHostService(
	registry NodeRegistry,
	hostSpecRepo HostSpecRepository,
	txm HostTxManager,
	rt runtime.HostRuntime,
	auditLog AuditLogWriter,
	cfg *config.Config,
	logger logutil.Logger,
	logDir string,
) *HostService {
	return &HostService{
		registry:      registry,
		hostSpecRepo:  hostSpecRepo,
		txm:           txm,
		runtime:       rt,
		auditLog:      auditLog,
		cfg:           cfg,
		logger:        logger,
		logDir:        logDir,
		deployCancels: make(map[uint64]context.CancelFunc),
	}
}

// CreateHost 校验 → 持久化 HostSpec → 后台异步构建部署
func (s *HostService) CreateHost(ctx context.Context, req *protocol.CreateHostRequest) (*protocol.HostSpec, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if len(req.Tools) == 0 {
		return nil, errors.New("at least one tool is required")
	}
	if unknown := hostbuilder.ValidateTools(req.Tools); len(unknown) > 0 {
		return nil, fmt.Errorf("unknown tools: %s. available: %s",
			strings.Join(unknown, ", "),
			strings.Join(func() []string {
				tools := hostbuilder.ListAvailableTools()
				ids := make([]string, len(tools))
				for i, t := range tools {
					ids[i] = t.ID
				}
				return ids
			}(), ", "))
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("generate host token: %w", err)
	}
	hostToken := hex.EncodeToString(tokenBytes)

	spec := &protocol.HostSpec{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Tools:       req.Tools,
		Resources:   req.Resources,
		AuthToken:   hostToken,
		Status:      protocol.HostStatusPending,
	}

	if err := s.txm.WithinTx(ctx, func(txCtx context.Context) error {
		ok, err := s.hostSpecRepo.Create(txCtx, spec)
		if err != nil {
			return fmt.Errorf("create host spec %q: %w", req.Name, err)
		}
		if !ok {
			return fmt.Errorf("host %q already exists", req.Name)
		}
		if err := s.registry.StoreHostToken(txCtx, req.Name, hostToken); err != nil {
			return fmt.Errorf("store host token %q: %w", req.Name, err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// 在启动 goroutine 之前先登记 WaitGroup 和 cancel，避免 Stop 与 Add/注册时序竞争导致漏等后台部署。
	//nolint:gosec // async deployment: 用 WithCancel 包装的 Background，支持取消
	deployCtx, cancel := context.WithCancel(context.Background())
	deployID := atomic.AddUint64(&s.nextDeployID, 1)
	s.deployWg.Add(1)
	s.deployMu.Lock()
	s.deployCancels[deployID] = cancel
	s.deployMu.Unlock()
	go func() {
		defer s.deployWg.Done()
		defer func() {
			s.deployMu.Lock()
			delete(s.deployCancels, deployID)
			s.deployMu.Unlock()
		}()
		s.deployHostAsync(deployCtx, spec)
	}() //nolint:gosec // async deployment OK

	return spec, nil
}

// ListHosts 返回所有 Host（含运行时合并状态）
func (s *HostService) ListHosts(ctx context.Context) ([]*protocol.HostInfo, error) {
	specs, err := s.hostSpecRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list host specs: %w", err)
	}
	infos := make([]*protocol.HostInfo, 0, len(specs))
	for _, spec := range specs {
		node, err := s.registry.Get(ctx, spec.Name)
		if err != nil {
			return nil, fmt.Errorf("get node %q: %w", spec.Name, err)
		}
		infos = append(infos, BuildHostInfo(spec, node))
	}
	return infos, nil
}

// GetHost 返回单个 Host 信息
func (s *HostService) GetHost(ctx context.Context, name string) (*protocol.HostInfo, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	spec, err := s.hostSpecRepo.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get host spec %q: %w", name, err)
	}
	node, err := s.registry.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get node %q: %w", name, err)
	}
	info := BuildHostInfo(spec, node)
	if info == nil {
		return nil, fmt.Errorf("host %q not found", name)
	}
	return info, nil
}

// DeleteHost 销毁 Host：删除 HostSpec + 清理令牌 + 停止容器 + 审计
func (s *HostService) DeleteHost(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("name is required")
	}

	if err := s.runtime.RemoveHost(ctx, name); err != nil {
		// 底层资源没删干净时必须保留控制面记录，避免把"残留资源"伪装成"删除成功"。
		s.logAuditError(ctx, protocol.AuditLogEntry{
			Operator:       "frontend",
			Action:         "delete_host",
			TargetNode:     name,
			PayloadSummary: "container=" + name,
			Result:         "failed",
		})
		return fmt.Errorf("remove host %q: %w", name, err)
	}

	if err := s.txm.WithinTx(ctx, func(txCtx context.Context) error {
		ok, err := s.registry.Delete(txCtx, name)
		if err != nil {
			return fmt.Errorf("delete node %q: %w", name, err)
		}
		if !ok {
			return fmt.Errorf("host %q not found", name)
		}
		if err := s.registry.RemoveHostToken(txCtx, name); err != nil {
			return fmt.Errorf("remove host token %q: %w", name, err)
		}
		ok, err = s.hostSpecRepo.Delete(txCtx, name)
		if err != nil {
			return fmt.Errorf("delete host spec %q: %w", name, err)
		}
		if !ok {
			return fmt.Errorf("host %q not found", name)
		}
		return nil
	}); err != nil {
		return err
	}

	s.logAuditError(ctx, protocol.AuditLogEntry{
		Operator:       "frontend",
		Action:         "delete_host",
		TargetNode:     name,
		PayloadSummary: "container=" + name,
		Result:         "success",
	})

	s.logger.Infof("[host] deleted host %q", name)
	return nil
}

// GetBuildLog 返回构建日志内容
func (s *HostService) GetBuildLog(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", errors.New("name is required")
	}

	spec, err := s.hostSpecRepo.Get(ctx, name)
	if err != nil {
		return "", fmt.Errorf("get host spec %q: %w", name, err)
	}
	if spec == nil {
		return "", fmt.Errorf("host %q not found", name)
	}

	logPath := filepath.Join(s.logDir, name+".log")
	// 异步部署期间日志文件可能尚未创建，此时返回空内容而非错误
	data, err := os.ReadFile(filepath.Clean(logPath))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// GetRuntimeLog 返回运行时日志
func (s *HostService) GetRuntimeLog(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", errors.New("name is required")
	}

	spec, err := s.hostSpecRepo.Get(ctx, name)
	if err != nil {
		return "", fmt.Errorf("get host spec %q: %w", name, err)
	}
	if spec == nil {
		return "", fmt.Errorf("host %q not found", name)
	}

	logs, err := s.runtime.GetRuntimeLogs(ctx, name, 500)
	if err != nil {
		return "", fmt.Errorf("get runtime logs: %w", err)
	}
	return logs, nil
}

// ListTools 返回可用工具列表
func (s *HostService) ListTools(ctx context.Context) ([]protocol.ToolConfig, error) {
	tools := hostbuilder.ListAvailableTools()
	return tools, nil
}

// deployHostAsync 后台异步构建部署 Host
func (s *HostService) deployHostAsync(ctx context.Context, spec *protocol.HostSpec) {
	s.updateHostStatus(ctx, spec.Name, protocol.HostStatusDeploying, "")

	if err := s.runtime.StopHost(ctx, spec.Name); err != nil {
		s.logger.Warnf("[host] stop old container for %s: %v", spec.Name, err)
	}

	if err := os.MkdirAll(s.logDir, 0750); err != nil {
		s.updateHostStatus(ctx, spec.Name, protocol.HostStatusFailed, fmt.Sprintf("create log dir: %v", err))
		return
	}
	logPath := filepath.Join(s.logDir, spec.Name+".log")
	logFile, err := os.Create(filepath.Clean(logPath))
	if err != nil {
		s.updateHostStatus(ctx, spec.Name, protocol.HostStatusFailed, fmt.Sprintf("create log file: %v", err))
		return
	}
	defer func() { _ = logFile.Close() }()

	multiWriter := io.MultiWriter(logFile, writerFunc(func(p []byte) (int, error) {
		s.logger.Infof("[host-deploy] %s", string(p))
		return len(p), nil
	}))

	_, _ = fmt.Fprintf(multiWriter, "[INFO] Host %s: starting deployment, tools=%v\n", spec.Name, spec.Tools)

	_, deployErr := BuildAndDeploy(ctx, s.runtime, spec, s.cfg)

	if deployErr != nil {
		errMsg := fmt.Sprintf("deploy failed: %v", deployErr)
		s.updateHostStatus(ctx, spec.Name, protocol.HostStatusFailed, errMsg)
		_, _ = fmt.Fprintf(multiWriter, "[ERROR] %s\n", errMsg)

		s.logAuditError(ctx, protocol.AuditLogEntry{
			Operator:       "system",
			Action:         "create_host",
			TargetNode:     spec.Name,
			PayloadSummary: fmt.Sprintf("tools=%v", spec.Tools),
			Result:         "failed",
		})
		return
	}

	_, _ = fmt.Fprintf(multiWriter, "[INFO] Host %s: deployment complete, waiting for agent registration\n", spec.Name)

	s.logAuditError(ctx, protocol.AuditLogEntry{
		Operator:       "system",
		Action:         "create_host",
		TargetNode:     spec.Name,
		PayloadSummary: fmt.Sprintf("tools=%v", spec.Tools),
		Result:         "success",
	})
}

// Stop 取消进行中的异步部署并等待完成，用于优雅关闭
func (s *HostService) Stop() {
	s.deployMu.Lock()
	cancels := make([]context.CancelFunc, 0, len(s.deployCancels))
	for _, cancel := range s.deployCancels {
		cancels = append(cancels, cancel)
	}
	s.deployMu.Unlock()
	for _, cancel := range cancels {
		cancel()
	}
	s.deployWg.Wait()
}

// writerFunc 将函数适配为 io.Writer
type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) {
	return f(p)
}

func (s *HostService) updateHostStatus(ctx context.Context, name, status, errMsg string) {
	updated, err := s.hostSpecRepo.UpdateStatus(ctx, name, status, errMsg)
	if err != nil {
		s.logger.Errorf("[host] update status for %s -> %s failed: %v", name, status, err)
		return
	}
	if !updated {
		s.logger.Warnf("[host] update status for %s skipped: host spec not found", name)
	}
}

func (s *HostService) logAuditError(ctx context.Context, entry protocol.AuditLogEntry) {
	if err := s.auditLog.Log(ctx, entry); err != nil {
		s.logger.Errorf("[host] write audit log action=%s target=%s failed: %v", entry.Action, entry.TargetNode, err)
	}
}
