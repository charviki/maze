package builder

import (
	"fmt"
	"strings"
)

// GenerateHostDockerfile 根据工具列表和 agent 基础镜像动态生成 Dockerfile。
// baseImage 是已包含 agent 二进制、entrypoint.sh 和基础运行时的镜像。
// 在此基础上叠加供应商工具链。
func GenerateHostDockerfile(toolIDs []string, baseImage string) (string, error) {
	var buf strings.Builder

	// 使用 agent 基础镜像（含 agent 二进制、entrypoint、Claude Code 等）
	buf.WriteString(fmt.Sprintf("FROM %s\n", baseImage))

	// 收集所有 bin 路径和额外环境变量
	var allBinPaths []string
	var extraEnvs []string

	for _, id := range toolIDs {
		cfg, ok := DefaultToolRegistry[id]
		if !ok {
			continue
		}

		// COPY --from 供应商镜像
		buf.WriteString(fmt.Sprintf("COPY --from=%s %s %s\n", cfg.Image, cfg.SourcePath, cfg.DestPath))
		allBinPaths = append(allBinPaths, cfg.BinPaths...)

		// 工具特定的 ENV
		for k, v := range cfg.EnvVars {
			extraEnvs = append(extraEnvs, fmt.Sprintf("ENV %s=%s", k, v))
		}
	}

	// 动态拼接 PATH（各工具 bin 目录 + 原始 PATH）
	if len(allBinPaths) > 0 {
		pathValue := strings.Join(allBinPaths, ":") + ":${PATH}"
		buf.WriteString(fmt.Sprintf("ENV PATH=\"%s\"\n", pathValue))
	}

	// 写入额外环境变量
	for _, env := range extraEnvs {
		buf.WriteString(env + "\n")
	}

	return buf.String(), nil
}
