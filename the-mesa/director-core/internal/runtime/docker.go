package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/maze/the-mesa/director-core/internal/config"
	hostbuilder "github.com/charviki/maze/the-mesa/director-core/internal/hostbuilder"
)

// DockerRuntime 通过 docker CLI 实现的容器运行时
type DockerRuntime struct {
	docker    config.DockerConfig
	workspace config.WorkspaceConfig
	logger    logutil.Logger
}

// NewDockerRuntime 创建 DockerRuntime
func NewDockerRuntime(docker config.DockerConfig, workspace config.WorkspaceConfig, logger logutil.Logger) *DockerRuntime {
	return &DockerRuntime{docker: docker, workspace: workspace, logger: logger}
}

func (d *DockerRuntime) dockerCmd(args ...string) *exec.Cmd {
	//nolint:gosec // docker CLI args are internally constructed
	args = append([]string{"-H", "unix://" + d.docker.SocketPath}, args...)
	//nolint:gosec
	return exec.Command("docker", args...)
}

// imageExistsLocally 检查指定镜像是否已存在于本地 docker 中
func (d *DockerRuntime) imageExistsLocally(imageName string) bool {
	return hostbuilder.ImageExistsLocally(imageName)
}

// tryTagComboImage 尝试将已存在的工具组合镜像 tag 为 Host 专属镜像。
// 会校验缓存镜像的 dockerfile-hash label 是否与当前 Dockerfile 一致，
// 不一致时返回 false 触发重建（供应商镜像更新后自动失效）。
func (d *DockerRuntime) tryTagComboImage(comboTag, imageTag, expectedHash string) bool {
	// 优先检查组合缓存镜像
	if d.imageExistsLocally(comboTag) {
		if d.checkDockerfileHash(comboTag, expectedHash) {
			tagCmd := d.dockerCmd("tag", comboTag, imageTag)
			return tagCmd.Run() == nil
		}
		// hash 不匹配，删除旧缓存触发重建
		_ = d.dockerCmd("rmi", comboTag).Run()
	}
	// 检查 Host 专属镜像
	if d.imageExistsLocally(imageTag) {
		if d.checkDockerfileHash(imageTag, expectedHash) {
			return true
		}
		// hash 不匹配，删除旧镜像触发重建
		_ = d.dockerCmd("rmi", imageTag).Run()
	}
	return false
}

// checkDockerfileHash 从镜像 label 中读取 dockerfile-hash 与期望值比较
func (d *DockerRuntime) checkDockerfileHash(imageName, expectedHash string) bool {
	return hostbuilder.CheckDockerfileHash(imageName, expectedHash)
}

// DeployHost 部署一个 Host：构建镜像 → 创建持久化目录 → 启动容器
func (d *DockerRuntime) DeployHost(ctx context.Context, spec *protocol.HostDeploySpec, dockerfileContent string) (*protocol.CreateHostResponse, error) {
	deployStart := time.Now()
	imageTag := "maze-agent:" + spec.Name
	comboTag := hostbuilder.ToolsetImageTag(spec.Tools)
	expectedHash := hostbuilder.ExtractDockerfileHash(dockerfileContent)

	cacheStart := time.Now()
	cacheHit := d.tryTagComboImage(comboTag, imageTag, expectedHash)
	if cacheHit {
		d.logger.Infof("[deploy] host=%s cache HIT, combo=%s, took=%v", spec.Name, comboTag, time.Since(cacheStart))
	} else {
		// 获取构建槽位，防止重建风暴
		buildSemaphore <- struct{}{}
		defer func() { <-buildSemaphore }()

		d.logger.Infof("[deploy] host=%s cache MISS, building image (combo=%s)...", spec.Name, comboTag)
		buildStart := time.Now()

		tmpDir, err := os.MkdirTemp("", "maze-build-*")
		if err != nil {
			return nil, fmt.Errorf("create temp dir: %w", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
		if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0600); err != nil {
			return nil, fmt.Errorf("write dockerfile: %w", err)
		}

		buildArgs := []string{"build", "-f", dockerfilePath, "-t", imageTag, tmpDir}
		buildCmd := d.dockerCmd(buildArgs...)
		buildCmd.Env = append(os.Environ(), "DOCKER_BUILDKIT=1")
		var buildOutput strings.Builder
		buildCmd.Stdout = &buildOutput
		buildCmd.Stderr = &buildOutput
		if err := buildCmd.Run(); err != nil {
			return nil, fmt.Errorf("docker build failed: %s", buildOutput.String())
		}

		d.logger.Infof("[deploy] host=%s docker build took=%v", spec.Name, time.Since(buildStart))

		// 构建完成后打上组合标签，供后续相同工具组合的 Host 复用
		tagCmd := d.dockerCmd("tag", imageTag, comboTag)
		_ = tagCmd.Run()
	}

	// 统一目录模型下，workspace 只存放 Director Core 元数据；Agent 数据始终位于 agents/<host>。
	// Director Core 容器通过 MountDir 访问宿主机 bind mount，因此本地目录操作要走容器内可见路径。
	hostMountDir := filepath.Join(d.workspace.MountDir, "agents", spec.Name)
	if err := os.MkdirAll(hostMountDir, 0750); err != nil {
		return nil, fmt.Errorf("create host dir: %w", err)
	}

	runStart := time.Now()
	// 构造容器启动参数
	runArgs := []string{"run", "-d", "--name", spec.Name,
		"--label", "maze-host=true"}
	if d.docker.Network != "" {
		runArgs = append(runArgs, "--network", d.docker.Network)
	}

	// 环境变量：Agent 通过这些变量向 Director Core 注册心跳
	// AGENT_CONTROLLER_AUTH_TOKEN: Host 专属令牌，用于向 Director Core 注册/心跳
	// AGENT_SERVER_AUTH_TOKEN: 全局 auth token，用于 Agent 自身 API 鉴权
	runArgs = append(runArgs,
		"-e", "AGENT_NAME="+spec.Name,
		"-e", "AGENT_EXTERNAL_ADDR=http://"+spec.Name+":8080",
		"-e", "AGENT_ADVERTISED_ADDR=http://"+spec.Name+":8080",
		"-e", fmt.Sprintf("AGENT_GRPC_ADDR=%s:9090", spec.Name),
		"-e", "AGENT_CONTROLLER_ADDR="+d.docker.DirectorCoreAddr,
		"-e", "AGENT_CONTROLLER_GRPC_ADDR="+deriveDirectorCoreGRPCAddr(d.docker.DirectorCoreAddr),
		"-e", "AGENT_SERVER_AUTH_TOKEN="+spec.ServerAuthToken,
		"-e", "AGENT_CONTROLLER_AUTH_TOKEN="+spec.AuthToken,
	)

	// 将宿主机数据目录挂载到 Agent 容器的 /home/agent，使 Agent 工作目录数据持久化到宿主机。
	// -v 格式: host_path:container_path
	hostBaseDir := filepath.Join(d.docker.AgentDataDir, spec.Name)
	runArgs = append(runArgs, "-v", hostBaseDir+":/home/agent")

	// 资源限制
	if spec.Resources.CPULimit != "" {
		runArgs = append(runArgs, "--cpus", spec.Resources.CPULimit)
	}
	if spec.Resources.MemoryLimit != "" {
		runArgs = append(runArgs, "--memory", spec.Resources.MemoryLimit)
	}

	runArgs = append(runArgs, imageTag)

	runCmd := d.dockerCmd(runArgs...)
	output, err := runCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("docker run failed: %s: %w", string(exitErr.Stderr), err)
		}
		return nil, fmt.Errorf("docker run failed: %w", err)
	}

	containerID := strings.TrimSpace(string(output))
	d.logger.Infof("[deploy] host=%s docker run took=%v, container=%s", spec.Name, time.Since(runStart), containerID[:12])
	d.logger.Infof("[deploy] host=%s TOTAL deploy took=%v", spec.Name, time.Since(deployStart))

	return &protocol.CreateHostResponse{
		Name:        spec.Name,
		Tools:       spec.Tools,
		ImageTag:    imageTag,
		ContainerID: containerID,
		Status:      "running",
	}, nil
}

// StopHost 停止并移除容器，保留持久化数据和镜像
func (d *DockerRuntime) StopHost(ctx context.Context, name string) error {
	stopCmd := d.dockerCmd("stop", name)
	_ = stopCmd.Run()

	rmCmd := d.dockerCmd("rm", "-f", name)
	if err := rmCmd.Run(); err != nil {
		return fmt.Errorf("docker rm failed: %w", err)
	}
	return nil
}

// RemoveHost 销毁 Host：停止容器 + 清理镜像 + 清理持久化数据
func (d *DockerRuntime) RemoveHost(ctx context.Context, name string) error {
	_ = d.StopHost(ctx, name)

	imageTag := "maze-agent:" + name
	rmiCmd := d.dockerCmd("rmi", "-f", imageTag)
	_ = rmiCmd.Run()

	// bind mount 的宿主机目录不会随着容器删除自动回收；Director Core 必须显式删除 agents/<host>。
	hostMountDir := filepath.Join(d.workspace.MountDir, "agents", name)
	if err := os.RemoveAll(hostMountDir); err != nil {
		return fmt.Errorf("remove host dir %s: %w", hostMountDir, err)
	}

	return nil
}

// InspectHost 通过 docker inspect 查询容器信息
func (d *DockerRuntime) InspectHost(ctx context.Context, name string) (*protocol.ContainerInfo, error) {
	cmd := d.dockerCmd("inspect", "--format", "{{json .}}", name)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("docker inspect failed: %w", err)
	}

	var results []struct {
		ID    string `json:"Id"`
		Name  string `json:"Name"`
		State struct {
			Status string `json:"Status"`
		} `json:"State"`
		Config struct {
			Image string `json:"Image"`
		} `json:"Config"`
		Created string `json:"Created"`
	}

	if err := json.Unmarshal(output, &results); err != nil {
		return nil, fmt.Errorf("parse inspect result: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("container %s not found", name)
	}

	r := results[0]
	containerName := strings.TrimPrefix(r.Name, "/")

	createdAt, err := time.Parse(time.RFC3339Nano, r.Created)
	if err != nil {
		createdAt = time.Time{}
	}

	return &protocol.ContainerInfo{
		ID:        r.ID[:12],
		Name:      containerName,
		Status:    r.State.Status,
		Image:     r.Config.Image,
		CreatedAt: createdAt,
	}, nil
}

// GetRuntimeLogs 通过 docker logs 获取容器运行日志
func (d *DockerRuntime) GetRuntimeLogs(ctx context.Context, name string, tailLines int) (string, error) {
	args := make([]string, 0, 4)
	args = append(args, "logs", "--tail", strconv.Itoa(tailLines))
	cmd := d.dockerCmd(append(args, name)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker logs failed: %w", err)
	}
	return string(output), nil
}

// IsHealthy 检查 Docker 容器是否存在且 running
func (d *DockerRuntime) IsHealthy(ctx context.Context, name string) (bool, error) {
	info, err := d.InspectHost(ctx, name)
	if err != nil {
		// 容器不存在视为不健康，不返回错误
		return false, err
	}
	return info.Status == "running", nil
}

// deriveDirectorCoreGRPCAddr 从 Director Core HTTP 地址推导 gRPC 地址：去掉 scheme，将端口替换为 9090
func deriveDirectorCoreGRPCAddr(httpAddr string) string {
	addr := strings.TrimPrefix(httpAddr, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr + ":9090"
	}
	return host + ":9090"
}
