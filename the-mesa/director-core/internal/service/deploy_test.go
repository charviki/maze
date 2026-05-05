package service

import (
	"context"
	"errors"
	"testing"

	"github.com/charviki/maze-cradle/configutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/maze/the-mesa/director-core/internal/config"
	hostbuilder "github.com/charviki/maze/the-mesa/director-core/internal/hostbuilder"
)

type mockHostRuntime struct {
	deployFn func(ctx context.Context, spec *protocol.HostDeploySpec, dockerfile string) (*protocol.CreateHostResponse, error)
}

func (m *mockHostRuntime) DeployHost(ctx context.Context, spec *protocol.HostDeploySpec, dockerfile string) (*protocol.CreateHostResponse, error) {
	if m.deployFn != nil {
		return m.deployFn(ctx, spec, dockerfile)
	}
	return &protocol.CreateHostResponse{Name: spec.Name}, nil
}
func (m *mockHostRuntime) StopHost(ctx context.Context, name string) error   { return nil }
func (m *mockHostRuntime) RemoveHost(ctx context.Context, name string) error { return nil }
func (m *mockHostRuntime) InspectHost(ctx context.Context, name string) (*protocol.ContainerInfo, error) {
	return nil, nil
}
func (m *mockHostRuntime) GetRuntimeLogs(ctx context.Context, name string, tailLines int) (string, error) {
	return "", nil
}
func (m *mockHostRuntime) IsHealthy(ctx context.Context, name string) (bool, error) { return true, nil }

func TestBuildAndDeploy_Success(t *testing.T) {
	rt := &mockHostRuntime{}
	cfg := &config.Config{
		Docker: config.DockerConfig{AgentBaseImage: "maze-agent:latest"},
		Server: config.ServerConfig{
			ServerConfig: configutil.ServerConfig{AuthToken: "tok"},
		},
	}
	spec := &protocol.HostSpec{Name: "host-1", Tools: []string{"claude"}, AuthToken: "host-tok"}

	resp, err := BuildAndDeploy(context.Background(), rt, spec, cfg)
	if err != nil {
		t.Fatalf("BuildAndDeploy error: %v", err)
	}
	if resp.Name != "host-1" {
		t.Errorf("Name = %q, want host-1", resp.Name)
	}
}

func TestBuildAndDeploy_DeployError(t *testing.T) {
	rt := &mockHostRuntime{
		deployFn: func(ctx context.Context, spec *protocol.HostDeploySpec, dockerfile string) (*protocol.CreateHostResponse, error) {
			return nil, errors.New("deploy failed")
		},
	}
	cfg := &config.Config{
		Docker: config.DockerConfig{AgentBaseImage: "maze-agent:latest"},
	}
	spec := &protocol.HostSpec{Name: "host-1"}

	_, err := BuildAndDeploy(context.Background(), rt, spec, cfg)
	if err == nil {
		t.Error("expected error from deploy failure")
	}
}

func TestBuildAndDeploy_DockerfileGenerated(t *testing.T) {
	var generatedDockerfile string
	rt := &mockHostRuntime{
		deployFn: func(ctx context.Context, spec *protocol.HostDeploySpec, dockerfile string) (*protocol.CreateHostResponse, error) {
			generatedDockerfile = dockerfile
			return &protocol.CreateHostResponse{Name: spec.Name}, nil
		},
	}
	cfg := &config.Config{
		Docker: config.DockerConfig{AgentBaseImage: "maze-agent:latest"},
		Server: config.ServerConfig{
			ServerConfig: configutil.ServerConfig{AuthToken: "tok"},
		},
	}
	spec := &protocol.HostSpec{Name: "host-1", Tools: []string{"claude"}, AuthToken: "host-tok"}

	_, _ = BuildAndDeploy(context.Background(), rt, spec, cfg)

	expectedContent := hostbuilder.GenerateHostDockerfile(spec.Tools, cfg.Docker.AgentBaseImage)
	if generatedDockerfile != expectedContent {
		t.Errorf("dockerfile mismatch:\ngot:  %s\nwant: %s", generatedDockerfile, expectedContent)
	}
}

func TestBuildAndDeploy_ServerAuthToken(t *testing.T) {
	var deploySpec *protocol.HostDeploySpec
	rt := &mockHostRuntime{
		deployFn: func(ctx context.Context, spec *protocol.HostDeploySpec, dockerfile string) (*protocol.CreateHostResponse, error) {
			deploySpec = spec
			return &protocol.CreateHostResponse{Name: spec.Name}, nil
		},
	}
	cfg := &config.Config{
		Docker: config.DockerConfig{AgentBaseImage: "maze-agent:latest"},
		Server: config.ServerConfig{
			ServerConfig: configutil.ServerConfig{AuthToken: "server-token"},
		},
	}
	spec := &protocol.HostSpec{Name: "host-1"}

	_, err := BuildAndDeploy(context.Background(), rt, spec, cfg)
	if err != nil {
		t.Fatalf("BuildAndDeploy error: %v", err)
	}
	if deploySpec.ServerAuthToken != "server-token" {
		t.Errorf("ServerAuthToken = %q, want server-token", deploySpec.ServerAuthToken)
	}
}
