package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charviki/maze-cradle/protocol"
	"github.com/charviki/mesa-hub-behavior-panel/biz/config"
)

// DockerRuntime 通过 docker CLI 实现的容器运行时
type DockerRuntime struct {
	docker    config.DockerConfig
	workspace config.WorkspaceConfig
}

// NewDockerRuntime 创建 DockerRuntime
func NewDockerRuntime(docker config.DockerConfig, workspace config.WorkspaceConfig) *DockerRuntime {
	return &DockerRuntime{docker: docker, workspace: workspace}
}

func (d *DockerRuntime) dockerCmd(args ...string) *exec.Cmd {
	args = append([]string{"-H", "unix://" + d.docker.SocketPath}, args...)
	return exec.Command("docker", args...)
}

// DeployHost 部署一个 Host：构建镜像 → 创建持久化目录 → 启动容器
func (d *DockerRuntime) DeployHost(ctx context.Context, spec *protocol.HostDeploySpec, dockerfileContent string) (*protocol.CreateHostResponse, error) {
	imageTag := fmt.Sprintf("maze-host-%s:latest", strings.ReplaceAll(spec.Name, "_", "-"))

	// 构建镜像
	tmpDir, err := os.MkdirTemp("", "maze-build-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return nil, fmt.Errorf("write dockerfile: %w", err)
	}

	buildArgs := []string{"build", "-f", dockerfilePath, "-t", imageTag, tmpDir}
	buildCmd := d.dockerCmd(buildArgs...)
	var buildOutput strings.Builder
	buildCmd.Stdout = &buildOutput
	buildCmd.Stderr = &buildOutput
	if err := buildCmd.Run(); err != nil {
		return nil, fmt.Errorf("docker build failed: %s", buildOutput.String())
	}

	// 创建持久化目录（容器内挂载路径）
	hostMountDir := filepath.Join(d.workspace.MountDir, spec.Name)
	if err := os.MkdirAll(hostMountDir, 0755); err != nil {
		return nil, fmt.Errorf("create host dir: %w", err)
	}

	// 构造容器启动参数
	runArgs := []string{"run", "-d", "--name", spec.Name}
	if d.docker.Network != "" {
		runArgs = append(runArgs, "--network", d.docker.Network)
	}

	// 环境变量：Agent 通过这些变量向 Manager 注册心跳
	// AGENT_CONTROLLER_AUTH_TOKEN: Host 专属令牌，用于向 Manager 注册/心跳
	// AGENT_SERVER_AUTH_TOKEN: 全局 auth token，用于 Agent 自身 API 鉴权
	runArgs = append(runArgs,
		"-e", fmt.Sprintf("AGENT_NAME=%s", spec.Name),
		"-e", fmt.Sprintf("AGENT_EXTERNAL_ADDR=http://%s:8080", spec.Name),
		"-e", fmt.Sprintf("AGENT_ADVERTISED_ADDR=http://%s:8080", spec.Name),
		"-e", fmt.Sprintf("AGENT_CONTROLLER_ADDR=%s", d.docker.ManagerAddr),
		"-e", fmt.Sprintf("AGENT_SERVER_AUTH_TOKEN=%s", spec.ServerAuthToken),
		"-e", fmt.Sprintf("AGENT_CONTROLLER_AUTH_TOKEN=%s", spec.AuthToken),
	)

	// 卷挂载（宿主机路径）
	hostBaseDir := filepath.Join(d.workspace.BaseDir, spec.Name)
	runArgs = append(runArgs, "-v", fmt.Sprintf("%s:/home/agent", hostBaseDir))

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

	return &protocol.CreateHostResponse{
		Name:        spec.Name,
		Tools:       spec.Tools,
		ImageTag:    imageTag,
		ContainerID: containerID,
		Status:      "running",
	}, nil
}

// RemoveHost 先尝试停止容器（忽略已停止的错误），再强制移除，最后清理镜像和持久化目录
func (d *DockerRuntime) RemoveHost(ctx context.Context, name string) error {
	stopCmd := d.dockerCmd("stop", name)
	_ = stopCmd.Run()

	rmCmd := d.dockerCmd("rm", "-f", name)
	if err := rmCmd.Run(); err != nil {
		return fmt.Errorf("docker rm failed: %w", err)
	}

	// 清理构建镜像
	imageTag := fmt.Sprintf("maze-host-%s:latest", strings.ReplaceAll(name, "_", "-"))
	rmiCmd := d.dockerCmd("rmi", "-f", imageTag)
	_ = rmiCmd.Run()

	// 清理持久化目录（容器内挂载路径）
	hostMountDir := filepath.Join(d.workspace.MountDir, name)
	os.RemoveAll(hostMountDir)

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
