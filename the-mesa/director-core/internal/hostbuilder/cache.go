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

// ResolveImageDigest 获取本地镜像的唯一标识符。
// 优先使用 RepoDigest（registry 拉取的镜像），回退到 Image ID（本地构建的镜像）。
func ResolveImageDigest(imageName string) (string, error) {
	//nolint:gosec // docker CLI args are internally constructed
	cmd := exec.CommandContext(context.Background(), "docker", "image", "inspect",
		"--format", "{{if .RepoDigests}}{{index .RepoDigests 0}}{{end}}", imageName)
	output, err := cmd.Output()
	if err == nil {
		if digest := strings.TrimSpace(string(output)); digest != "" {
			return digest, nil
		}
	}

	// 本地构建的镜像无 RepoDigests，回退到 Image ID
	//nolint:gosec
	cmd = exec.CommandContext(context.Background(), "docker", "image", "inspect",
		"--format", "{{.Id}}", imageName)
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("resolve digest for %s: image not found", imageName)
	}
	if id := strings.TrimSpace(string(output)); id != "" {
		return id, nil
	}
	return "", fmt.Errorf("resolve digest for %s: empty id", imageName)
}
