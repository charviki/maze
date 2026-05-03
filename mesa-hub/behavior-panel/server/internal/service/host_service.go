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

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/internal/config"
	builder "github.com/charviki/mesa-hub-behavior-panel/internal/imagebuilder"
	"github.com/charviki/mesa-hub-behavior-panel/internal/model"
	"github.com/charviki/mesa-hub-behavior-panel/internal/runtime"
)

// HostService Host 业务逻辑（Manager 本地），供 HTTP handler 和 gRPC handler 共用
type HostService struct {
	registry     *model.NodeRegistry
	specMgr      *model.HostSpecManager
	runtime      runtime.HostRuntime
	auditLog     AuditLogger
	cfg          *config.Config
	logger       logutil.Logger
	logDir       string
	deployCancel context.CancelFunc
	deployWg     sync.WaitGroup
	deployMu     sync.Mutex
}

// AuditLogger 审计日志接口，避免循环依赖
type AuditLogger interface {
	Log(entry protocol.AuditLogEntry)
}

// NewHostService 创建 HostService
func NewHostService(
	registry *model.NodeRegistry,
	specMgr *model.HostSpecManager,
	rt runtime.HostRuntime,
	auditLog AuditLogger,
	cfg *config.Config,
	logger logutil.Logger,
	logDir string,
) *HostService {
	return &HostService{
		registry: registry,
		specMgr:  specMgr,
		runtime:  rt,
		auditLog: auditLog,
		cfg:      cfg,
		logger:   logger,
		logDir:   logDir,
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
	if s.specMgr.Get(req.Name) != nil {
		return nil, fmt.Errorf("host %q already exists", req.Name)
	}

	if unknown := builder.ValidateTools(req.Tools); len(unknown) > 0 {
		return nil, fmt.Errorf("unknown tools: %s. available: %s",
			strings.Join(unknown, ", "),
			strings.Join(func() []string {
				tools := builder.ListAvailableTools()
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

	if !s.specMgr.Create(spec) {
		return nil, fmt.Errorf("host %q already exists", req.Name)
	}

	s.registry.StoreHostToken(req.Name, hostToken)

	//nolint:gosec // async deployment: 用 WithCancel 包装的 Background，支持取消
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		s.deployMu.Lock()
		s.deployCancel = cancel
		s.deployMu.Unlock()
		s.deployWg.Add(1)
		defer s.deployWg.Done()
		s.deployHostAsync(ctx, spec)
	}() //nolint:gosec // async deployment OK

	return spec, nil
}

// ListHosts 返回所有 Host（含运行时合并状态）
func (s *HostService) ListHosts(ctx context.Context) ([]*protocol.HostInfo, error) {
	return s.specMgr.ListMerged(s.registry), nil
}

// GetHost 返回单个 Host 信息
func (s *HostService) GetHost(ctx context.Context, name string) (*protocol.HostInfo, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	info := s.specMgr.GetMerged(name, s.registry)
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
		// 底层资源没删干净时必须保留控制面记录，避免把“残留资源”伪装成“删除成功”。
		s.auditLog.Log(protocol.AuditLogEntry{
			Operator:       "frontend",
			Action:         "delete_host",
			TargetNode:     name,
			PayloadSummary: "container=" + name,
			Result:         "failed",
		})
		return fmt.Errorf("remove host %q: %w", name, err)
	}

	s.registry.Delete(name)
	s.registry.RemoveHostToken(name)
	s.specMgr.Delete(name)

	s.auditLog.Log(protocol.AuditLogEntry{
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

	if s.specMgr.Get(name) == nil {
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

	if s.specMgr.Get(name) == nil {
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
	tools := builder.ListAvailableTools()
	return tools, nil
}

// deployHostAsync 后台异步构建部署 Host
func (s *HostService) deployHostAsync(ctx context.Context, spec *protocol.HostSpec) {
	s.specMgr.UpdateStatus(spec.Name, protocol.HostStatusDeploying, "")

	if err := s.runtime.StopHost(ctx, spec.Name); err != nil {
		s.logger.Warnf("[host] stop old container for %s: %v", spec.Name, err)
	}

	if err := os.MkdirAll(s.logDir, 0750); err != nil {
		s.specMgr.UpdateStatus(spec.Name, protocol.HostStatusFailed, fmt.Sprintf("create log dir: %v", err))
		return
	}
	logPath := filepath.Join(s.logDir, spec.Name+".log")
	logFile, err := os.Create(filepath.Clean(logPath))
	if err != nil {
		s.specMgr.UpdateStatus(spec.Name, protocol.HostStatusFailed, fmt.Sprintf("create log file: %v", err))
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
		s.specMgr.UpdateStatus(spec.Name, protocol.HostStatusFailed, errMsg)
		_, _ = fmt.Fprintf(multiWriter, "[ERROR] %s\n", errMsg)

		s.auditLog.Log(protocol.AuditLogEntry{
			Operator:       "system",
			Action:         "create_host",
			TargetNode:     spec.Name,
			PayloadSummary: fmt.Sprintf("tools=%v", spec.Tools),
			Result:         "failed",
		})
		return
	}

	_, _ = fmt.Fprintf(multiWriter, "[INFO] Host %s: deployment complete, waiting for agent registration\n", spec.Name)

	s.auditLog.Log(protocol.AuditLogEntry{
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
	if s.deployCancel != nil {
		s.deployCancel()
	}
	s.deployMu.Unlock()
	s.deployWg.Wait()
}

// writerFunc 将函数适配为 io.Writer
type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) {
	return f(p)
}
