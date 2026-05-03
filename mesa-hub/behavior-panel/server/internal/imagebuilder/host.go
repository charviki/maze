package builder

import (
	"crypto/sha256"
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
func GenerateHostDockerfile(toolIDs []string, baseImage string) string {
	sorted := make([]string, len(toolIDs))
	copy(sorted, toolIDs)
	sort.Strings(sorted)

	var buf strings.Builder

	// 使用 agent 基础镜像（含 agent 二进制、entrypoint、Claude Code 等）
	fmt.Fprintf(&buf, "FROM %s\n", baseImage)

	var allBinPaths []string
	var extraEnvs []string

	for _, id := range sorted {
		cfg, ok := DefaultToolRegistry[id]
		if !ok {
			continue
		}

		// COPY --from 供应商镜像
		fmt.Fprintf(&buf, "COPY --from=%s %s %s\n", cfg.Image, cfg.SourcePath, cfg.DestPath)
		allBinPaths = append(allBinPaths, cfg.BinPaths...)

		// 工具特定的 ENV（按 key 排序确保确定性）
		if len(cfg.EnvVars) > 0 {
			keys := make([]string, 0, len(cfg.EnvVars))
			for k := range cfg.EnvVars {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				extraEnvs = append(extraEnvs, fmt.Sprintf("ENV %s=%s", k, cfg.EnvVars[k]))
			}
		}
	}

	// 动态拼接 PATH（各工具 bin 目录 + 原始 PATH）
	if len(allBinPaths) > 0 {
		pathValue := strings.Join(allBinPaths, ":") + ":${PATH}"
		fmt.Fprintf(&buf, "ENV PATH=\"%s\"\n", pathValue)
	}

	// 写入额外环境变量
	for _, env := range extraEnvs {
		buf.WriteString(env + "\n")
	}

	// 注入 Dockerfile 内容 hash 作为 LABEL，供应商镜像更新时自动触发重建
	contentHash := DockerfileHash(buf.String())
	fmt.Fprintf(&buf, "LABEL maze.dockerfile-hash=%s\n", contentHash)

	return buf.String()
}

// DockerfileHash 计算 Dockerfile 内容的短 hash，用于判断是否需要重建。
// 供应商镜像更新后，COPY --from 的镜像 digest 变化会导致内容不同，
// 但 hash 基于生成的 Dockerfile 文本，所以需要配合镜像 digest 一起使用。
func DockerfileHash(content string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(content)))[:16]
}
