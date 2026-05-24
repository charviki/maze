package service

import (
	"context"
	"fmt"

	"github.com/charviki/maze/fabrication/cradle/protocol"
	"github.com/charviki/maze/the-mesa/director-core/internal/config"
	hostbuilder "github.com/charviki/maze/the-mesa/director-core/internal/hostbuilder"
	"github.com/charviki/maze/the-mesa/director-core/internal/runtime"
)

// BuildAndDeploy 执行 Host 的构建部署：解析镜像指纹 → 生成 Dockerfile → 调用运行时部署。
// 调用方负责状态更新、日志记录和审计日志。
func BuildAndDeploy(ctx context.Context, rt runtime.HostRuntime, spec *protocol.HostSpec, cfg *config.Config) (*protocol.CreateHostResponse, error) {
	imageFingerprints := resolveFingerprints(spec.Tools, cfg.Docker.AgentBaseImage)
	dockerfileContent := hostbuilder.GenerateHostDockerfile(spec.Tools, cfg.Docker.AgentBaseImage, imageFingerprints)

	deploySpec := &protocol.HostDeploySpec{
		Name:            spec.Name,
		Tools:           spec.Tools,
		Resources:       spec.Resources,
		AuthToken:       spec.AuthToken,
		ServerAuthToken: cfg.Server.JWTSecret,
	}

	resp, err := rt.DeployHost(ctx, deploySpec, dockerfileContent)
	if err != nil {
		return nil, fmt.Errorf("deploy host %s: %w", spec.Name, err)
	}
	return resp, nil
}

// resolveFingerprints 解析所有引用镜像的内容指纹（RootFS.Layers）。
// 任一镜像解析失败时跳过该条（不中断流程），全部失败时返回 nil，
// 退化为纯文本 hash。
func resolveFingerprints(toolIDs []string, baseImage string) map[string]string {
	fingerprints := make(map[string]string)

	if fp, err := hostbuilder.ResolveContentFingerprint(baseImage); err == nil {
		fingerprints[baseImage] = fp
	}

	for _, id := range toolIDs {
		cfg, ok := hostbuilder.DefaultToolRegistry[id]
		if !ok {
			continue
		}
		if _, exists := fingerprints[cfg.Image]; exists {
			continue
		}
		if fp, err := hostbuilder.ResolveContentFingerprint(cfg.Image); err == nil {
			fingerprints[cfg.Image] = fp
		}
	}

	if len(fingerprints) == 0 {
		return nil
	}
	return fingerprints
}
