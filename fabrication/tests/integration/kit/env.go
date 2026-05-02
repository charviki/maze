package kit

import (
	"context"
	"fmt"
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
	apiClient := NewTestAPIClient(e.cfg)
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

// WaitForManager 轮询 Manager 健康接口，直到超时或服务就绪。
func (e *TestEnv) WaitForManager(timeout time.Duration) error {
	apiClient := NewTestAPIClient(e.cfg)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, _, err := apiClient.HostServiceAPI.HostServiceListHosts(context.TODO()).Execute()
		if err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("manager not available at %s after %v", e.cfg.ManagerURL, timeout)
}
