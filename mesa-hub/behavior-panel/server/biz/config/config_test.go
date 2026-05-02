package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	// 写入有效的 YAML 配置文件
	content := `server:
  listen_addr: ":9090"
  auth_token: "my-secret-token"
workspace:
  base_dir: "/data"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load 失败: %v", err)
	}

	if cfg.Server.ListenAddr != ":9090" {
		t.Errorf("期望 ListenAddr=:9090, 实际=%s", cfg.Server.ListenAddr)
	}
	if cfg.Server.AuthToken != "my-secret-token" {
		t.Errorf("期望 AuthToken=my-secret-token, 实际=%s", cfg.Server.AuthToken)
	}
	if cfg.Workspace.BaseDir != "/data" {
		t.Errorf("期望 BaseDir=/data, 实际=%s", cfg.Workspace.BaseDir)
	}
}

func TestValidate_Defaults(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	// 写入空配置，验证默认值填充
	content := `{}`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load 失败: %v", err)
	}

	// 未配置 listen_addr 时默认为 :8080
	if cfg.Server.ListenAddr != ":8080" {
		t.Errorf("期望默认 ListenAddr=:8080, 实际=%s", cfg.Server.ListenAddr)
	}

	// 未配置 base_dir 时默认为 ~/.maze/docker，agents/ 子目录留给 Docker Agent 工作目录。
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".maze", "docker")
	if cfg.Workspace.BaseDir != expected {
		t.Errorf("期望默认 BaseDir=%s, 实际=%s", expected, cfg.Workspace.BaseDir)
	}
	if cfg.Docker.AgentDataDir != filepath.Join(expected, "agents") {
		t.Errorf("期望默认 AgentDataDir=%s, 实际=%s", filepath.Join(expected, "agents"), cfg.Docker.AgentDataDir)
	}
}

func TestApplyEnvOverrides_AllowPrivateNetworks(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	// YAML 中启用私有网络
	content := `server:
  allow_private_networks: true
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	// 环境变量显式关闭
	t.Setenv("AGENT_MANAGER_ALLOW_PRIVATE_NETWORKS", "false")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load 失败: %v", err)
	}

	if cfg.Server.AllowPrivateNetworks {
		t.Error("期望 AllowPrivateNetworks=false, 实际=true")
	}

	// 环境变量显式开启
	t.Setenv("AGENT_MANAGER_ALLOW_PRIVATE_NETWORKS", "true")

	cfg2, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load 失败: %v", err)
	}

	if !cfg2.Server.AllowPrivateNetworks {
		t.Error("期望 AllowPrivateNetworks=true, 实际=false")
	}
}

func TestDockerConfig_Defaults(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `{}`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load 失败: %v", err)
	}

	if cfg.Docker.SocketPath != "/var/run/docker.sock" {
		t.Errorf("期望默认 SocketPath=/var/run/docker.sock, 实际=%s", cfg.Docker.SocketPath)
	}
	if cfg.Docker.ManagerAddr != "http://agent-manager:8080" {
		t.Errorf("期望默认 ManagerAddr=http://agent-manager:8080, 实际=%s", cfg.Docker.ManagerAddr)
	}
}

func TestDockerConfig_EnvOverrides(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `{}`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	t.Setenv("AGENT_MANAGER_DOCKER_SOCKET_PATH", "/custom/docker.sock")
	t.Setenv("AGENT_MANAGER_DOCKER_NETWORK", "my-network")
	t.Setenv("AGENT_MANAGER_DOCKER_MANAGER_ADDR", "http://my-manager:9090")
	t.Setenv("AGENT_MANAGER_DOCKER_AGENT_BASE_IMAGE", "custom-agent:latest")
	t.Setenv("AGENT_MANAGER_DOCKER_AGENT_DATA_DIR", "/custom/agents")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load 失败: %v", err)
	}

	if cfg.Docker.SocketPath != "/custom/docker.sock" {
		t.Errorf("期望 SocketPath=/custom/docker.sock, 实际=%s", cfg.Docker.SocketPath)
	}
	if cfg.Docker.Network != "my-network" {
		t.Errorf("期望 Network=my-network, 实际=%s", cfg.Docker.Network)
	}
	if cfg.Docker.ManagerAddr != "http://my-manager:9090" {
		t.Errorf("期望 ManagerAddr=http://my-manager:9090, 实际=%s", cfg.Docker.ManagerAddr)
	}
	if cfg.Docker.AgentBaseImage != "custom-agent:latest" {
		t.Errorf("期望 AgentBaseImage=custom-agent:latest, 实际=%s", cfg.Docker.AgentBaseImage)
	}
	if cfg.Docker.AgentDataDir != "/custom/agents" {
		t.Errorf("期望 AgentDataDir=/custom/agents, 实际=%s", cfg.Docker.AgentDataDir)
	}
}
