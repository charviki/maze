package hostbuilder

import (
	"os/exec"
	"strings"
)

// ImageExistsLocally 检查指定镜像是否已存在于本地 docker 中
func ImageExistsLocally(imageName string) bool {
	//nolint:gosec // docker CLI args are internally constructed
	cmd := exec.Command("docker", "image", "inspect", imageName)
	return cmd.Run() == nil
}

// CheckDockerfileHash 从镜像 label 中读取 dockerfile-hash 与期望值比较
func CheckDockerfileHash(imageName, expectedHash string) bool {
	//nolint:gosec // docker CLI args are internally constructed
	cmd := exec.Command("docker", "inspect", "--format",
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
