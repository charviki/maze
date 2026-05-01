package builder

import (
	"fmt"
	"sort"
	"strings"
)

// ToolsetImageTag 为工具组合生成稳定的镜像标签。
// 排序后用 `-` 连接，确保相同组合产生相同标签。
// 例如 ["claude", "go"] → "maze-agent:claude-go"
func ToolsetImageTag(toolIDs []string) string {
	sorted := make([]string, len(toolIDs))
	copy(sorted, toolIDs)
	sort.Strings(sorted)
	return "maze-agent:" + strings.Join(sorted, "-")
}

// GenerateHostDockerfile 根据工具列表和 agent 基础镜像动态生成 Dockerfile。
// baseImage 是已包含 agent 二进制、entrypoint.sh 和基础运行时的镜像。
// 在此基础上叠加供应商工具链。
// 工具列表排序后再生成，确保相同组合产生相同 Dockerfile，最大化 Docker 层缓存命中。
func GenerateHostDockerfile(toolIDs []string, baseImage string) (string, error) {
	sorted := make([]string, len(toolIDs))
	copy(sorted, toolIDs)
	sort.Strings(sorted)

	var buf strings.Builder

	// 使用 agent 基础镜像（含 agent 二进制、entrypoint、Claude Code 等）
	buf.WriteString(fmt.Sprintf("FROM %s\n", baseImage))

	var allBinPaths []string
	var extraEnvs []string

	for _, id := range sorted {
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
