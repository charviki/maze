package service

import (
	"context"
	"fmt"

	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/biz/builder"
	"github.com/charviki/mesa-hub-behavior-panel/biz/config"
	"github.com/charviki/mesa-hub-behavior-panel/biz/runtime"
)

// BuildAndDeploy 执行 Host 的构建部署：生成 Dockerfile → 调用运行时部署。
// 调用方负责状态更新、日志记录和审计日志。
func BuildAndDeploy(ctx context.Context, rt runtime.HostRuntime, spec *protocol.HostSpec, cfg *config.Config) (*protocol.CreateHostResponse, error) {
	dockerfileContent := builder.GenerateHostDockerfile(spec.Tools, cfg.Docker.AgentBaseImage)

	deploySpec := &protocol.HostDeploySpec{
		Name:            spec.Name,
		Tools:           spec.Tools,
		Resources:       spec.Resources,
		AuthToken:       spec.AuthToken,
		ServerAuthToken: cfg.Server.AuthToken,
	}

	resp, err := rt.DeployHost(ctx, deploySpec, dockerfileContent)
	if err != nil {
		return nil, fmt.Errorf("deploy host %s: %w", spec.Name, err)
	}
	return resp, nil
}
