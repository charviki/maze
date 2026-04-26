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
  base_dir: "/data/agents"
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
	if cfg.Workspace.BaseDir != "/data/agents" {
		t.Errorf("期望 BaseDir=/data/agents, 实际=%s", cfg.Workspace.BaseDir)
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

	// 未配置 base_dir 时默认为 ~/agents
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "agents")
	if cfg.Workspace.BaseDir != expected {
		t.Errorf("期望默认 BaseDir=%s, 实际=%s", expected, cfg.Workspace.BaseDir)
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
