package kit

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

// TestEnv 管理集成测试的生命周期（setup/teardown/等待就绪）。
type TestEnv struct {
	cfg *TestConfig
}

// NewTestEnv 创建测试环境实例。
func NewTestEnv(cfg *TestConfig) *TestEnv {
	return &TestEnv{cfg: cfg}
}

// Setup 初始化测试环境（当前为空操作，环境由 Makefile test-integration 管理）。
func (e *TestEnv) Setup() error {
	return nil
}

// Teardown 清理测试数据：列出所有 Host 并逐一删除。
func (e *TestEnv) Teardown() error {
	apiClient, err := NewTestAPIClient(e.cfg)
	if err != nil {
		return fmt.Errorf("create teardown client: %w", err)
	}
	hosts, _, err := apiClient.HostServiceAPI.HostServiceListHosts(context.TODO()).Execute()
	if err != nil {
		return fmt.Errorf("list hosts for teardown: %w", err)
	}
	if hosts != nil {
		for _, h := range hosts.GetHosts() {
			if _, _, err := apiClient.HostServiceAPI.HostServiceDeleteHost(context.TODO(), h.GetName()).Execute(); err != nil {
				return fmt.Errorf("delete host %s: %w", h.GetName(), err)
			}
		}
	}
	return nil
}

// WaitForDirectorCore 轮询 Director Core 直到就绪。
// 先探测 /health（免鉴权）确认进程存活，再尝试登录并直接访问受保护 API 确认 JWT 链路畅通。
// 这里刻意使用原始 HTTP 探测，而不是依赖生成的 SDK client，避免把“服务是否就绪”和
// “测试侧 SDK 初始化是否成功”耦合到一起，导致启动阶段出现误报。
// 支持 SIGINT/SIGTERM 信号中断等待。
func (e *TestEnv) WaitForDirectorCore(timeout time.Duration) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	deadline := time.Now().Add(timeout)
	httpClient := &http.Client{Timeout: 5 * time.Second}
	lastFailure := "no probe executed"

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return fmt.Errorf("interrupted while waiting for director-core at %s", e.cfg.DirectorCoreURL)
		default:
		}

		resp, err := httpClient.Get(e.cfg.DirectorCoreURL + "/health")
		if err != nil || resp.StatusCode != http.StatusOK {
			if err != nil {
				lastFailure = fmt.Sprintf("/health request failed: %v", err)
			} else {
				lastFailure = fmt.Sprintf("/health returned status %d", resp.StatusCode)
			}
			if resp != nil {
				_ = resp.Body.Close()
			}
			if !sleep(ctx) {
				return fmt.Errorf("interrupted while waiting for director-core at %s", e.cfg.DirectorCoreURL)
			}
			continue
		}
		_ = resp.Body.Close()

		loginResult, err := LoginAdmin(context.Background(), e.cfg)
		if err != nil {
			lastFailure = fmt.Sprintf("login failed: %v", err)
			if !sleep(ctx) {
				return fmt.Errorf("interrupted while waiting for director-core at %s", e.cfg.DirectorCoreURL)
			}
			continue
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.cfg.DirectorCoreURL+"/api/v1/hosts", nil)
		if err != nil {
			if !sleep(ctx) {
				return fmt.Errorf("interrupted while waiting for director-core at %s", e.cfg.DirectorCoreURL)
			}
			continue
		}
		req.Header.Set("Authorization", "Bearer "+loginResult.AccessToken)
		protectedResp, err := httpClient.Do(req)
		if err != nil || protectedResp.StatusCode != http.StatusOK {
			if err != nil {
				lastFailure = fmt.Sprintf("protected request failed: %v", err)
			} else {
				lastFailure = fmt.Sprintf("protected request returned status %d", protectedResp.StatusCode)
			}
			if protectedResp != nil {
				_ = protectedResp.Body.Close()
			}
			if !sleep(ctx) {
				return fmt.Errorf("interrupted while waiting for director-core at %s", e.cfg.DirectorCoreURL)
			}
			continue
		}
		_ = protectedResp.Body.Close()
		return nil
	}
	return fmt.Errorf("director-core not available at %s after %v (last failure: %s)", e.cfg.DirectorCoreURL, timeout, lastFailure)
}

// sleep 等待 2 秒，收到信号时立即返回 false。
func sleep(ctx context.Context) bool {
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
