package kit

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// TestEnv 管理测试环境的生命周期
type TestEnv struct {
	cfg *TestConfig
}

// NewTestEnv 创建测试环境管理器
func NewTestEnv(cfg *TestConfig) *TestEnv {
	return &TestEnv{cfg: cfg}
}

// Setup 初始化测试环境（创建数据目录等）
func (e *TestEnv) Setup() error {
	dir := e.dataDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create test data dir %s: %w", dir, err)
	}
	return nil
}

// Teardown 清理测试环境
func (e *TestEnv) Teardown() error {
	// Docker 环境：停止 docker-compose 服务
	if e.cfg.Env == "docker" {
		return e.teardownDocker()
	}
	// K8s 环境：删除测试 namespace
	if e.cfg.Env == "kubernetes" {
		return e.teardownK8s()
	}
	return nil
}

// WaitForManager 等待 Manager API 可用
func (e *TestEnv) WaitForManager(timeout time.Duration) error {
	client := NewAPIClient(e.cfg.ManagerURL, e.cfg.AuthToken)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, err := client.ListHosts()
		if err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("manager not available at %s after %v", e.cfg.ManagerURL, timeout)
}

func (e *TestEnv) dataDir() string {
	return e.cfg.DataDir + "/" + e.cfg.Env
}

func (e *TestEnv) teardownDocker() error {
	// 清理所有测试 Host（通过 API 删除）
	client := NewAPIClient(e.cfg.ManagerURL, e.cfg.AuthToken)
	hosts, err := client.ListHosts()
	if err != nil {
		return nil // Manager 可能已停止
	}
	for _, h := range hosts {
		client.DeleteHost(h.Name)
	}
	return nil
}

func (e *TestEnv) teardownK8s() error {
	// 清理测试 Host
	client := NewAPIClient(e.cfg.ManagerURL, e.cfg.AuthToken)
	hosts, err := client.ListHosts()
	if err != nil {
		return nil
	}
	for _, h := range hosts {
		client.DeleteHost(h.Name)
	}

	// 可选：删除 K8s 测试 namespace
	cmd := exec.Command("kubectl", "delete", "namespace", e.cfg.Namespace, "--ignore-not-found=true")
	cmd.Run()
	return nil
}
