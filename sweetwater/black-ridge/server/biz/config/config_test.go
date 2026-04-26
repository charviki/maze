package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoad_ValidConfig 验证 Load 能正确解析合法的 YAML 配置文件
func TestLoad_ValidConfig(t *testing.T) {
	content := `
server:
  listen_addr: ":9090"
  auth_token: "mytoken"
  name: "test-agent"
  external_addr: "http://localhost:9090"
tmux:
  socket_path: "/tmp/test.sock"
  default_shell: "/bin/zsh"
terminal:
  default_lines: 100
controller:
  addr: "http://controller:8080"
  enabled: true
  heartbeat_interval: 30
workspace:
  root_dir: "/home/test"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("写入临时配置文件失败: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load 返回错误: %v", err)
	}

	if cfg.Server.ListenAddr != ":9090" {
		t.Errorf("Server.ListenAddr = %q, 期望 %q", cfg.Server.ListenAddr, ":9090")
	}
	if cfg.Server.AuthToken != "mytoken" {
		t.Errorf("Server.AuthToken = %q, 期望 %q", cfg.Server.AuthToken, "mytoken")
	}
	if cfg.Server.Name != "test-agent" {
		t.Errorf("Server.Name = %q, 期望 %q", cfg.Server.Name, "test-agent")
	}
	if cfg.Server.ExternalAddr != "http://localhost:9090" {
		t.Errorf("Server.ExternalAddr = %q, 期望 %q", cfg.Server.ExternalAddr, "http://localhost:9090")
	}
	if cfg.Tmux.SocketPath != "/tmp/test.sock" {
		t.Errorf("Tmux.SocketPath = %q, 期望 %q", cfg.Tmux.SocketPath, "/tmp/test.sock")
	}
	if cfg.Tmux.DefaultShell != "/bin/zsh" {
		t.Errorf("Tmux.DefaultShell = %q, 期望 %q", cfg.Tmux.DefaultShell, "/bin/zsh")
	}
	if cfg.Terminal.DefaultLines != 100 {
		t.Errorf("Terminal.DefaultLines = %d, 期望 %d", cfg.Terminal.DefaultLines, 100)
	}
	if cfg.Controller.Addr != "http://controller:8080" {
		t.Errorf("Controller.Addr = %q, 期望 %q", cfg.Controller.Addr, "http://controller:8080")
	}
	if cfg.Controller.Enabled != true {
		t.Errorf("Controller.Enabled = %v, 期望 %v", cfg.Controller.Enabled, true)
	}
	if cfg.Controller.HeartbeatInterval != 30 {
		t.Errorf("Controller.HeartbeatInterval = %d, 期望 %d", cfg.Controller.HeartbeatInterval, 30)
	}
	if cfg.Workspace.RootDir != "/home/test" {
		t.Errorf("Workspace.RootDir = %q, 期望 %q", cfg.Workspace.RootDir, "/home/test")
	}
}

// TestLoad_EnvOverride 验证环境变量能覆盖 YAML 文件中的对应配置项
func TestLoad_EnvOverride(t *testing.T) {
	content := `
server:
  listen_addr: ":9090"
  auth_token: "file-token"
  name: "file-name"
tmux:
  default_shell: "/bin/sh"
terminal:
  default_lines: 20
controller:
  addr: "http://original:8080"
  enabled: false
  heartbeat_interval: 5
workspace:
  root_dir: "/original"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("写入临时配置文件失败: %v", err)
	}

	envVars := map[string]string{
		"AGENT_SERVER_LISTEN_ADDR":            ":7070",
		"AGENT_SERVER_AUTH_TOKEN":             "env-token",
		"AGENT_NAME":                          "env-name",
		"AGENT_EXTERNAL_ADDR":                 "http://env:7070",
		"AGENT_TMUX_SOCKET_PATH":              "/tmp/env.sock",
		"AGENT_TMUX_DEFAULT_SHELL":            "/bin/env-shell",
		"AGENT_TERMINAL_DEFAULT_LINES":        "200",
		"AGENT_CONTROLLER_ADDR":               "http://env-controller:8080",
		"AGENT_CONTROLLER_HEARTBEAT_INTERVAL": "60",
		"AGENT_WORKSPACE_ROOT_DIR":            "/env-workspace",
	}
	for k, v := range envVars {
		os.Setenv(k, v)
	}
	defer func() {
		for k := range envVars {
			os.Unsetenv(k)
		}
	}()

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load 返回错误: %v", err)
	}

	if cfg.Server.ListenAddr != ":7070" {
		t.Errorf("Server.ListenAddr = %q, 期望 %q (环境变量覆盖)", cfg.Server.ListenAddr, ":7070")
	}
	if cfg.Server.AuthToken != "env-token" {
		t.Errorf("Server.AuthToken = %q, 期望 %q (环境变量覆盖)", cfg.Server.AuthToken, "env-token")
	}
	if cfg.Server.Name != "env-name" {
		t.Errorf("Server.Name = %q, 期望 %q (环境变量覆盖)", cfg.Server.Name, "env-name")
	}
	if cfg.Server.ExternalAddr != "http://env:7070" {
		t.Errorf("Server.ExternalAddr = %q, 期望 %q (环境变量覆盖)", cfg.Server.ExternalAddr, "http://env:7070")
	}
	if cfg.Tmux.SocketPath != "/tmp/env.sock" {
		t.Errorf("Tmux.SocketPath = %q, 期望 %q (环境变量覆盖)", cfg.Tmux.SocketPath, "/tmp/env.sock")
	}
	if cfg.Tmux.DefaultShell != "/bin/env-shell" {
		t.Errorf("Tmux.DefaultShell = %q, 期望 %q (环境变量覆盖)", cfg.Tmux.DefaultShell, "/bin/env-shell")
	}
	if cfg.Terminal.DefaultLines != 200 {
		t.Errorf("Terminal.DefaultLines = %d, 期望 %d (环境变量覆盖)", cfg.Terminal.DefaultLines, 200)
	}
	if cfg.Controller.Addr != "http://env-controller:8080" {
		t.Errorf("Controller.Addr = %q, 期望 %q (环境变量覆盖)", cfg.Controller.Addr, "http://env-controller:8080")
	}
	if cfg.Controller.Enabled != true {
		t.Errorf("Controller.Enabled = %v, 期望 %v (设置 AGENT_CONTROLLER_ADDR 后应自动启用)", cfg.Controller.Enabled, true)
	}
	if cfg.Controller.HeartbeatInterval != 60 {
		t.Errorf("Controller.HeartbeatInterval = %d, 期望 %d (环境变量覆盖)", cfg.Controller.HeartbeatInterval, 60)
	}
	if cfg.Workspace.RootDir != "/env-workspace" {
		t.Errorf("Workspace.RootDir = %q, 期望 %q (环境变量覆盖)", cfg.Workspace.RootDir, "/env-workspace")
	}
}

// TestValidate_Defaults 验证 validate 对空/零值字段填充默认值
func TestValidate_Defaults(t *testing.T) {
	cfg := &Config{}

	if err := validate(cfg); err != nil {
		t.Fatalf("validate 返回错误: %v", err)
	}

	if cfg.Server.ListenAddr != ":8080" {
		t.Errorf("默认 Server.ListenAddr = %q, 期望 %q", cfg.Server.ListenAddr, ":8080")
	}
	if cfg.Tmux.DefaultShell != "/bin/bash" {
		t.Errorf("默认 Tmux.DefaultShell = %q, 期望 %q", cfg.Tmux.DefaultShell, "/bin/bash")
	}
	if cfg.Terminal.DefaultLines != 50 {
		t.Errorf("默认 Terminal.DefaultLines = %d, 期望 %d", cfg.Terminal.DefaultLines, 50)
	}
	if cfg.Controller.HeartbeatInterval != 10 {
		t.Errorf("默认 Controller.HeartbeatInterval = %d, 期望 %d", cfg.Controller.HeartbeatInterval, 10)
	}
	if cfg.Workspace.RootDir != "/home/agent" {
		t.Errorf("默认 Workspace.RootDir = %q, 期望 %q", cfg.Workspace.RootDir, "/home/agent")
	}
}

// TestLoadFromExe_NotFound 验证找不到配置文件时返回错误
func TestLoadFromExe_NotFound(t *testing.T) {
	_, err := LoadFromExe("nonexistent_config_file_12345.yaml")
	if err == nil {
		t.Fatal("期望返回错误，但 LoadFromExe 返回 nil")
	}
}
