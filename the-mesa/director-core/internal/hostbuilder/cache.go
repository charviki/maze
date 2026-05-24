package hostbuilder

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ImageExistsLocally 检查指定镜像是否已存在于本地 docker 中
func ImageExistsLocally(imageName string) bool {
	//nolint:gosec // docker CLI args are internally constructed
	cmd := exec.CommandContext(context.Background(), "docker", "image", "inspect", imageName)
	return cmd.Run() == nil
}

// CheckDockerfileHash 从镜像 label 中读取 dockerfile-hash 与期望值比较
func CheckDockerfileHash(imageName, expectedHash string) bool {
	//nolint:gosec // docker CLI args are internally constructed
	cmd := exec.CommandContext(context.Background(), "docker", "inspect", "--format",
		"{{index .Config.Labels \"maze.dockerfile-hash\"}}", imageName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == expectedHash
}

// ExtractDockerfileHash 从生成的 Dockerfile 内容中提取 maze.dockerfile-hash label 值
func ExtractDockerfileHash(dockerfileContent string) string {
	for _, line := range strings.Split(dockerfileContent, "\n") {
		if strings.HasPrefix(line, "LABEL maze.dockerfile-hash=") {
			return strings.TrimPrefix(line, "LABEL maze.dockerfile-hash=")
		}
	}
	return ""
}

// ResolveContentFingerprint 获取镜像的内容指纹（RootFS.Layers）。
// 使用内容寻址的 layer digest 而非 Image ID，避免构建时间戳等非确定性因素干扰。
// 相同镜像内容 → 相同指纹；代码变更 → 层内容变更 → 指纹变化。
func ResolveContentFingerprint(imageName string) (string, error) {
	//nolint:gosec // docker CLI args are internally constructed
	cmd := exec.CommandContext(context.Background(), "docker", "image", "inspect",
		"--format", "{{json .RootFS.Layers}}", imageName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("resolve fingerprint for %s: %w", imageName, err)
	}
	fingerprint := strings.TrimSpace(string(output))
	if fingerprint == "" {
		return "", fmt.Errorf("resolve fingerprint for %s: empty layers", imageName)
	}
	return fingerprint, nil
}
