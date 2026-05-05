//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"time"

	client "github.com/charviki/maze-cradle/api/gen/http"
	"github.com/charviki/maze-integration-tests/kit"
)

type hostPool struct {
	cfg       *kit.TestConfig
	apiClient *client.APIClient
	leases    *kit.LeasePool
	profiles  map[string][]string
}

func newHostPool(cfg *kit.TestConfig) (*hostPool, error) {
	profiles := make(map[string][]string)
	runID := time.Now().UnixMilli()
	if cfg.PoolClaudeSize > 0 {
		names := make([]string, 0, cfg.PoolClaudeSize)
		for i := 1; i <= cfg.PoolClaudeSize; i++ {
			names = append(names, fmt.Sprintf("pool-claude-%d-%d", runID, i))
		}
		profiles["claude"] = names
	}
	if cfg.PoolGoSize > 0 {
		names := make([]string, 0, cfg.PoolGoSize)
		for i := 1; i <= cfg.PoolGoSize; i++ {
			names = append(names, fmt.Sprintf("pool-go-%d-%d", runID, i))
		}
		profiles["go"] = names
	}
	if len(profiles) == 0 {
		return nil, fmt.Errorf("host pool enabled but no profiles configured")
	}
	return &hostPool{
		cfg:       cfg,
		apiClient: kit.NewTestAPIClient(cfg),
		leases:    kit.NewLeasePool(profiles),
		profiles:  profiles,
	}, nil
}

func (p *hostPool) Warmup() error {
	p.logf("WARMUP", "start profiles=%d", len(p.profiles))
	for profile, names := range p.profiles {
		tools, err := toolsForProfile(profile)
		if err != nil {
			return err
		}
		p.logf("WARMUP", "profile=%s size=%d tools=%v", profile, len(names), tools)
		for _, name := range names {
			p.logf("WARMUP", "creating host=%s profile=%s tools=%v", name, profile, tools)
			if err := p.createHostAndWait(name, tools); err != nil {
				return err
			}
		}
	}
	p.logf("WARMUP", "complete")
	return nil
}

func (p *hostPool) Acquire(profile string) (string, error) {
	p.logf("LEASE", "request profile=%s", profile)
	name, err := p.leases.Acquire(profile)
	if err != nil {
		return "", err
	}
	p.logf("LEASE", "acquired profile=%s host=%s", profile, name)
	return name, nil
}

func (p *hostPool) Release(profile, name string) error {
	p.logf("LEASE", "releasing profile=%s host=%s", profile, name)
	if err := p.leases.Release(profile, name); err != nil {
		return err
	}
	p.logf("LEASE", "released profile=%s host=%s", profile, name)
	return nil
}

func (p *hostPool) Cleanup() error {
	var firstErr error
	p.logf("CLEANUP", "start")
	for _, names := range p.profiles {
		for _, name := range names {
			p.logf("CLEANUP", "deleting host=%s", name)
			if err := p.deleteHostAndWait(name); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	if firstErr != nil {
		p.logf("CLEANUP", "finished with error=%v", firstErr)
	} else {
		p.logf("CLEANUP", "complete")
	}
	return firstErr
}

func (p *hostPool) createHostAndWait(name string, tools []string) error {
	nameField := name
	body := client.V1CreateHostRequest{
		Name:  &nameField,
		Tools: tools,
	}
	resp, httpResp, err := p.apiClient.HostServiceAPI.HostServiceCreateHost(context.Background()).Body(body).Execute()
	if err != nil {
		statusCode := 0
		if httpResp != nil {
			statusCode = httpResp.StatusCode
		}
		return fmt.Errorf("create host %s: %v (status=%d)", name, err, statusCode)
	}
	p.logf("WARMUP", "CreateHost response name=%s status=%s", resp.GetName(), resp.GetStatus())
	if err := p.waitForHostStatus(name, "online", 3*time.Minute); err != nil {
		return err
	}
	if err := p.waitForHostAPIReady(name, 30*time.Second); err != nil {
		return err
	}
	if err := p.waitForManagerDataLayout(45 * time.Second); err != nil {
		return err
	}
	return nil
}

func (p *hostPool) waitForHostStatus(name, targetStatus string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	attempt := 0
	for time.Now().Before(deadline) {
		attempt++
		info, _, err := p.apiClient.HostServiceAPI.HostServiceGetHost(context.Background(), name).Execute()
		if err != nil {
			p.logf("WAIT", "host=%s attempt=%d error=%v", name, attempt, err)
			time.Sleep(2 * time.Second)
			continue
		}
		if info.GetStatus() == targetStatus {
			p.logf("WAIT", "host=%s reached %s after %d attempts", name, targetStatus, attempt)
			return nil
		}
		p.logf("WAIT", "host=%s status=%s want=%s attempt=%d", name, info.GetStatus(), targetStatus, attempt)
		time.Sleep(3 * time.Second)
	}
	return fmt.Errorf("wait for host %s status %s: timeout after %v", name, targetStatus, timeout)
}

func (p *hostPool) waitForHostAPIReady(nodeName string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	attempt := 0
	for time.Now().Before(deadline) {
		attempt++
		_, _, err := p.apiClient.SessionServiceAPI.SessionServiceListSessions(context.Background(), nodeName).Execute()
		if err == nil {
			p.logf("WAIT", "host=%s API ready after %d attempts", nodeName, attempt)
			return nil
		}
		p.logf("WAIT", "host=%s API not ready attempt=%d err=%v", nodeName, attempt, err)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("wait for host %s API ready: timeout after %v", nodeName, timeout)
}

func (p *hostPool) waitForManagerDataLayout(timeout time.Duration) error {
	return nil
}

func (p *hostPool) deleteHostAndWait(name string) error {
	_, httpResp, err := p.apiClient.HostServiceAPI.HostServiceDeleteHost(context.Background(), name).Execute()
	if err != nil {
		if httpResp == nil || httpResp.StatusCode != 404 {
			statusCode := 0
			if httpResp != nil {
				statusCode = httpResp.StatusCode
			}
			return fmt.Errorf("delete host %s: %v (status=%d)", name, err, statusCode)
		}
	}
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		_, httpResp, err = p.apiClient.HostServiceAPI.HostServiceGetHost(context.Background(), name).Execute()
		if err != nil && httpResp != nil && httpResp.StatusCode == 404 {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("wait for pooled host %s deletion: timeout", name)
}

func toolsForProfile(profile string) ([]string, error) {
	switch profile {
	case "claude":
		return []string{"claude"}, nil
	case "go":
		return []string{"go"}, nil
	default:
		return nil, fmt.Errorf("unknown host profile %q", profile)
	}
}

func (p *hostPool) logf(stage, format string, args ...any) {
	fmt.Fprintf(os.Stdout, "[POOL/%s] %s\n", stage, fmt.Sprintf(format, args...))
}
