package reconciler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/internal/config"
	"github.com/charviki/mesa-hub-behavior-panel/internal/runtime"
	"github.com/charviki/mesa-hub-behavior-panel/internal/service"
)

const (
	healthCheckInterval  = 60 * time.Second
	maxRetryCount        = 3
	pendingTimeout       = 5 * time.Minute
	deployingGracePeriod = 5 * time.Minute // deploying 状态保护窗口，避免与 deployHostAsync 并行构建
)

// Reconciler 负责 Host 的启动恢复和定期健康巡检
type Reconciler struct {
	hostSpecRepo service.HostSpecRepository
	registry     service.NodeRegistry
	runtime      runtime.HostRuntime
	cfg          *config.Config
	logger       logutil.Logger
	logDir       string
	stopCh       chan struct{}
	doneCh       chan struct{}
}

// NewReconciler 创建 Reconciler
func NewReconciler(
	hostSpecRepo service.HostSpecRepository,
	registry service.NodeRegistry,
	rt runtime.HostRuntime,
	cfg *config.Config,
	logger logutil.Logger,
	logDir string,
) *Reconciler {
	return &Reconciler{
		hostSpecRepo: hostSpecRepo,
		registry:     registry,
		runtime:      rt,
		cfg:          cfg,
		logger:       logger,
		logDir:       logDir,
		stopCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
	}
}

// RecoverOnStartup 启动时对比 host_specs 与实际运行状态，自动补齐缺失的 Host。
func (r *Reconciler) RecoverOnStartup(ctx context.Context) {
	specs := r.hostSpecRepo.List()
	if len(specs) == 0 {
		r.logger.Infof("[reconciler] no host specs to recover")
		return
	}

	r.logger.Infof("[reconciler] starting recovery for %d host specs", len(specs))
	for _, spec := range specs {
		// 恢复令牌到内存（Manager 重启后令牌丢失，需要从 HostSpec 恢复）
		r.registry.StoreHostToken(spec.Name, spec.AuthToken)
		r.recoverOne(ctx, spec)
	}
	r.logger.Infof("[reconciler] recovery complete")
}

func (r *Reconciler) recoverOne(ctx context.Context, spec *protocol.HostSpec) {
	switch spec.Status {
	case protocol.HostStatusPending, protocol.HostStatusDeploying:
		// 检查运行时是否有对应容器/Pod
		healthy, err := r.runtime.IsHealthy(ctx, spec.Name)
		if err != nil {
			r.logger.Warnf("[reconciler] check health for %s failed: %v", spec.Name, err)
		}
		if healthy {
			// 容器/Pod 存在且健康，等待 Agent 注册
			r.logger.Infof("[reconciler] host %s already running, waiting for agent registration", spec.Name)
			return
		}
		// 不存在，重新部署
		r.logger.Infof("[reconciler] host %s not running, redeploying", spec.Name)
		r.redeploy(ctx, spec)

	case protocol.HostStatusOnline, protocol.HostStatusOffline:
		healthy, err := r.runtime.IsHealthy(ctx, spec.Name)
		if err != nil {
			r.logger.Warnf("[reconciler] check health for %s failed: %v", spec.Name, err)
		}
		if healthy {
			// 等待心跳恢复
			r.logger.Infof("[reconciler] host %s runtime exists, waiting for heartbeat", spec.Name)
			return
		}
		r.logger.Infof("[reconciler] host %s runtime missing, redeploying", spec.Name)
		r.redeploy(ctx, spec)

	case protocol.HostStatusFailed:
		if spec.RetryCount < maxRetryCount {
			r.logger.Infof("[reconciler] host %s failed (retry %d/%d), retrying", spec.Name, spec.RetryCount, maxRetryCount)
			r.hostSpecRepo.IncrementRetry(spec.Name)
			r.redeploy(ctx, r.hostSpecRepo.Get(spec.Name))
		} else {
			r.logger.Infof("[reconciler] host %s failed (retry %d/%d), skipping", spec.Name, spec.RetryCount, maxRetryCount)
		}
	}
}

// StartHealthCheck 启动定期健康巡检
func (r *Reconciler) StartHealthCheck(ctx context.Context) {
	go func() {
		defer close(r.doneCh)
		ticker := time.NewTicker(healthCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.runHealthCheck(ctx)
			case <-r.stopCh:
				return
			}
		}
	}()
}

// Stop 停止健康巡检
func (r *Reconciler) Stop() {
	close(r.stopCh)
	<-r.doneCh
}

func (r *Reconciler) runHealthCheck(ctx context.Context) {
	specs := r.hostSpecRepo.List()
	for _, spec := range specs {
		r.checkOne(ctx, spec)
	}
}

func (r *Reconciler) checkOne(ctx context.Context, spec *protocol.HostSpec) {
	switch spec.Status {
	case protocol.HostStatusDeploying:
		// 保护窗口：UpdatedAt 在 5 分钟内则跳过，让 deployHostAsync 完成
		if time.Since(spec.UpdatedAt) < deployingGracePeriod {
			return
		}
		// 超过保护期仍未上线，检查运行时健康
		healthy, err := r.runtime.IsHealthy(ctx, spec.Name)
		if err != nil {
			r.logger.Warnf("[health-check] check %s failed: %v", spec.Name, err)
			return
		}
		if !healthy {
			r.logger.Infof("[health-check] host %s deploying timeout, redeploying", spec.Name)
			r.redeploy(ctx, spec)
		}

	case protocol.HostStatusOnline, protocol.HostStatusOffline:
		healthy, err := r.runtime.IsHealthy(ctx, spec.Name)
		if err != nil {
			r.logger.Warnf("[health-check] check %s failed: %v", spec.Name, err)
			return
		}
		if !healthy {
			r.logger.Infof("[health-check] host %s runtime missing, redeploying", spec.Name)
			r.redeploy(ctx, spec)
		}

	case protocol.HostStatusFailed:
		if spec.RetryCount < maxRetryCount {
			r.logger.Infof("[health-check] host %s failed (retry %d/%d), retrying", spec.Name, spec.RetryCount, maxRetryCount)
			r.hostSpecRepo.IncrementRetry(spec.Name)
			r.redeploy(ctx, r.hostSpecRepo.Get(spec.Name))
		}

	case protocol.HostStatusPending:
		// pending 超过 5 分钟视为后台任务丢失
		if time.Since(spec.CreatedAt) > pendingTimeout {
			r.logger.Infof("[health-check] host %s pending timeout, marking as failed", spec.Name)
			r.hostSpecRepo.UpdateStatus(spec.Name, protocol.HostStatusFailed, "pending timeout: background task may be lost")
		}
	}
}

// redeploy 重新部署一个 Host
func (r *Reconciler) redeploy(ctx context.Context, spec *protocol.HostSpec) {
	if spec == nil {
		return
	}

	// 确保令牌已预存（Manager 重启后令牌只在内存中，需要恢复）
	r.registry.StoreHostToken(spec.Name, spec.AuthToken)

	// 清理旧容器/Pod（忽略不存在的错误）
	if err := r.runtime.StopHost(ctx, spec.Name); err != nil {
		r.logger.Warnf("[reconciler] cleanup old runtime for %s failed: %v", spec.Name, err)
	}

	r.hostSpecRepo.UpdateStatus(spec.Name, protocol.HostStatusDeploying, "")

	if err := os.MkdirAll(r.logDir, 0750); err == nil {
		logPath := filepath.Join(r.logDir, spec.Name+".log")
		f, ferr := os.OpenFile(filepath.Clean(logPath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if ferr == nil {
			_, _ = fmt.Fprintf(f, "\n[%s] === RECONCILER REDEPLOY ===\n", time.Now().Format(time.RFC3339))
			func() { _ = f.Close() }()
		}
	}

	_, deployErr := service.BuildAndDeploy(ctx, r.runtime, spec, r.cfg)
	if deployErr != nil {
		errMsg := fmt.Sprintf("redeploy failed: %v", deployErr)
		r.hostSpecRepo.UpdateStatus(spec.Name, protocol.HostStatusFailed, errMsg)
		r.logger.Errorf("[reconciler] %s: %s", spec.Name, errMsg)
		return
	}

	r.logger.Infof("[reconciler] host %s redeployed successfully", spec.Name)
}
