package kit

import (
	"fmt"
	"time"
)

type TestEnv struct {
	cfg *TestConfig
}

func NewTestEnv(cfg *TestConfig) *TestEnv {
	return &TestEnv{cfg: cfg}
}

func (e *TestEnv) Setup() error {
	return nil
}

func (e *TestEnv) Teardown() error {
	apiClient := NewTestAPIClient(e.cfg)
	hosts, _, err := apiClient.HostServiceAPI.HostServiceListHosts(nil).Execute()
	if err != nil {
		return nil
	}
	if hosts != nil {
		for _, h := range hosts.GetHosts() {
			apiClient.HostServiceAPI.HostServiceDeleteHost(nil, h.GetName()).Execute()
		}
	}
	return nil
}

func (e *TestEnv) WaitForManager(timeout time.Duration) error {
	apiClient := NewTestAPIClient(e.cfg)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, _, err := apiClient.HostServiceAPI.HostServiceListHosts(nil).Execute()
		if err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("manager not available at %s after %v", e.cfg.ManagerURL, timeout)
}
