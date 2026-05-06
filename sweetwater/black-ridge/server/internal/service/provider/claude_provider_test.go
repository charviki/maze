package provider

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeProvider_BootstrapTask_MarksProjectTrusted(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(home, ".claude.json")
	initial := map[string]any{
		"projects": map[string]any{
			"/home/agent/existing": map[string]any{
				"hasTrustDialogAccepted":        false,
				"hasCompletedProjectOnboarding": false,
			},
		},
	}
	data, err := json.Marshal(initial)
	if err != nil {
		t.Fatalf("序列化初始配置失败: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("写入初始配置失败: %v", err)
	}

	p := &ClaudeProvider{}
	task := p.BootstrapTask()
	if err := task.Run(TaskContext{HomeDir: home, WorkingDir: "/home/agent/existing"}); err != nil {
		t.Fatalf("BootstrapTask 失败: %v", err)
	}

	updated, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("读取更新后配置失败: %v", err)
	}

	var cfg map[string]any
	if err := json.Unmarshal(updated, &cfg); err != nil {
		t.Fatalf("反序列化更新后配置失败: %v", err)
	}

	projects, ok := cfg["projects"].(map[string]any)
	if !ok {
		t.Fatalf("期望 projects 为对象, 实际 %#v", cfg["projects"])
	}
	entry, ok := projects["/home/agent/existing"].(map[string]any)
	if !ok {
		t.Fatalf("期望项目条目存在, 实际 %#v", projects["/home/agent/existing"])
	}
	if accepted, _ := entry["hasTrustDialogAccepted"].(bool); !accepted {
		t.Fatalf("期望 hasTrustDialogAccepted=true, 实际 %#v", entry)
	}
	if completed, _ := entry["hasCompletedProjectOnboarding"].(bool); !completed {
		t.Fatalf("期望 hasCompletedProjectOnboarding=true, 实际 %#v", entry)
	}
}

func TestClaudeProvider_BootstrapTask_PreservesTopLevelFields(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(home, ".claude.json")
	initial := map[string]any{
		"hasCompletedOnboarding":      false,
		"firstStartTime":              "2026-04-26T09:53:51.986Z",
		"migrationVersion":            12,
		"userID":                      "user-1",
		"opusProMigrationComplete":    true,
		"sonnet1m45MigrationComplete": true,
		"projects":                    map[string]any{},
	}
	data, err := json.Marshal(initial)
	if err != nil {
		t.Fatalf("序列化初始配置失败: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("写入初始配置失败: %v", err)
	}

	p := &ClaudeProvider{}
	task := p.BootstrapTask()
	if err := task.Run(TaskContext{HomeDir: home, WorkingDir: "/home/agent/new"}); err != nil {
		t.Fatalf("BootstrapTask 失败: %v", err)
	}

	updated, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("读取更新后配置失败: %v", err)
	}

	var cfg map[string]any
	if err := json.Unmarshal(updated, &cfg); err != nil {
		t.Fatalf("反序列化更新后配置失败: %v", err)
	}

	if got, _ := cfg["firstStartTime"].(string); got != "2026-04-26T09:53:51.986Z" {
		t.Fatalf("期望保留 firstStartTime, 实际 %#v", cfg["firstStartTime"])
	}
	if got, _ := cfg["userID"].(string); got != "user-1" {
		t.Fatalf("期望保留 userID, 实际 %#v", cfg["userID"])
	}
	if got, _ := cfg["migrationVersion"].(float64); got != 12 {
		t.Fatalf("期望保留 migrationVersion=12, 实际 %#v", cfg["migrationVersion"])
	}
	if onboarding, _ := cfg["hasCompletedOnboarding"].(bool); onboarding {
		t.Fatalf("Bootstrap 不应越界改写全局 onboarding 状态, 实际 %#v", cfg["hasCompletedOnboarding"])
	}
}

func TestClaudeProvider_EntrypointTask_InitClaudeJson(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	p := &ClaudeProvider{}
	tasks := p.EntrypointTasks()
	if len(tasks) < 1 {
		t.Fatal("期望至少 1 个 EntrypointTask")
	}

	if err := tasks[0].Run(TaskContext{HomeDir: home}); err != nil {
		t.Fatalf("init-claude-json 失败: %v", err)
	}

	configPath := filepath.Join(home, ".claude.json")
	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("读取 %s 失败: %v", configPath, err)
	}

	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	if onboarding, _ := cfg["hasCompletedOnboarding"].(bool); !onboarding {
		t.Fatalf("期望 hasCompletedOnboarding=true, 实际 %#v", cfg["hasCompletedOnboarding"])
	}
}

func TestClaudeProvider_EntrypointTask_InitClaudeSettings(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	p := &ClaudeProvider{}
	tasks := p.EntrypointTasks()
	if len(tasks) < 2 {
		t.Fatal("期望至少 2 个 EntrypointTask")
	}

	if err := tasks[1].Run(TaskContext{HomeDir: home}); err != nil {
		t.Fatalf("init-claude-settings 失败: %v", err)
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")
	raw, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("读取 %s 失败: %v", settingsPath, err)
	}

	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	permissions, _ := cfg["permissions"].(map[string]any)
	if permissions == nil {
		t.Fatal("期望 permissions 字段存在")
	}
	if skip, _ := permissions["skipDangerousModePermissionPrompt"].(bool); !skip {
		t.Fatalf("期望 skipDangerousModePermissionPrompt=true")
	}
	if theme, _ := cfg["theme"].(string); theme != "dark" {
		t.Fatalf("期望 theme=dark, 实际 %q", theme)
	}
}

func TestClaudeProvider_EntrypointTask_PreservesExistingState(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(home, ".claude.json")
	existing := map[string]any{
		"hasCompletedOnboarding": false,
		"userID":                 "existing-user",
		"projects":               map[string]any{},
	}
	data, _ := json.Marshal(existing)
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("写入失败: %v", err)
	}

	p := &ClaudeProvider{}
	tasks := p.EntrypointTasks()
	if err := tasks[0].Run(TaskContext{HomeDir: home}); err != nil {
		t.Fatalf("init-claude-json 失败: %v", err)
	}

	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("读取失败: %v", err)
	}

	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	if got, _ := cfg["userID"].(string); got != "existing-user" {
		t.Fatalf("期望保留 userID=existing-user, 实际 %q", got)
	}
	if onboarding, _ := cfg["hasCompletedOnboarding"].(bool); !onboarding {
		t.Fatalf("期望强制 hasCompletedOnboarding=true, 实际 %#v", cfg["hasCompletedOnboarding"])
	}
}

func TestClaudeProvider_DataMethods(t *testing.T) {
	p := &ClaudeProvider{}
	if p.ID() != "claude" {
		t.Errorf("ID() = %q, want %q", p.ID(), "claude")
	}
	if p.SessionIDPlaceholder() != "{session_id}" {
		t.Errorf("SessionIDPlaceholder() = %q, want %q", p.SessionIDPlaceholder(), "{session_id}")
	}
	if p.RestoreCommandTemplate() != "" {
		t.Errorf("RestoreCommandTemplate() = %q, want empty", p.RestoreCommandTemplate())
	}
}

func TestClaudeProvider_HealthCheckTask_NoBinary(t *testing.T) {
	home := t.TempDir()
	t.Setenv("AGENT_HOME", home)
	t.Setenv("PATH", home)

	p := &ClaudeProvider{}
	task := p.HealthCheckTask()
	if err := task.Run(TaskContext{}); err == nil {
		t.Error("期望 PATH 中无 claude 时 HealthCheck 返回错误")
	}
}

// fakeClaudeBinary 在 dir 下创建一个名为 "claude" 的可执行文件，用于 LookPath 测试
func fakeClaudeBinary(t *testing.T, dir string) {
	t.Helper()
	p := filepath.Join(dir, "claude")
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatalf("创建假 claude 二进制失败: %v", err)
	}
}

func TestClaudeProvider_HealthCheckTask_BinaryExistsButConfigMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("AGENT_HOME", home)
	t.Setenv("PATH", home)

	// 放入假 claude 二进制，确保 LookPath 通过
	fakeClaudeBinary(t, home)

	p := &ClaudeProvider{}
	task := p.HealthCheckTask()
	err := task.Run(TaskContext{})
	if err == nil {
		t.Fatal("期望配置文件缺失时 HealthCheck 返回错误")
	}
	if !strings.Contains(err.Error(), "config invalid") {
		t.Errorf("错误应包含 'config invalid', 实际: %v", err)
	}
}

func TestClaudeProvider_HealthCheckTask_BinaryExistsButConfigCorrupted(t *testing.T) {
	home := t.TempDir()
	t.Setenv("AGENT_HOME", home)
	t.Setenv("PATH", home)

	fakeClaudeBinary(t, home)

	// 写入合法的 ~/.claude.json
	configData, _ := json.Marshal(map[string]any{"hasCompletedOnboarding": true})
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), configData, 0600); err != nil {
		t.Fatalf("写入 claude.json 失败: %v", err)
	}

	// 写入损坏的 ~/.claude/settings.json（不是合法 JSON）
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0750); err != nil {
		t.Fatalf("创建 .claude 目录失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".claude", "settings.json"), []byte("{invalid json}"), 0600); err != nil {
		t.Fatalf("写入损坏的 settings.json 失败: %v", err)
	}

	p := &ClaudeProvider{}
	task := p.HealthCheckTask()
	err := task.Run(TaskContext{})
	if err == nil {
		t.Fatal("期望配置文件损坏时 HealthCheck 返回错误")
	}
	if !strings.Contains(err.Error(), "settings invalid") {
		t.Errorf("错误应包含 'settings invalid', 实际: %v", err)
	}
}

func TestClaudeProvider_HealthCheckTask_AllValid(t *testing.T) {
	home := t.TempDir()
	t.Setenv("AGENT_HOME", home)
	t.Setenv("PATH", home)

	fakeClaudeBinary(t, home)

	// 写入合法的配置文件
	configData, _ := json.Marshal(map[string]any{"hasCompletedOnboarding": true})
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), configData, 0600); err != nil {
		t.Fatalf("写入 claude.json 失败: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0750); err != nil {
		t.Fatalf("创建 .claude 目录失败: %v", err)
	}
	settingsData, _ := json.Marshal(map[string]any{"permissions": map[string]any{}})
	if err := os.WriteFile(filepath.Join(home, ".claude", "settings.json"), settingsData, 0600); err != nil {
		t.Fatalf("写入 settings.json 失败: %v", err)
	}

	p := &ClaudeProvider{}
	task := p.HealthCheckTask()
	if err := task.Run(TaskContext{}); err != nil {
		t.Errorf("所有条件满足时 HealthCheck 应通过, 实际: %v", err)
	}
}

func TestClaudeProvider_HealthCheckTask_NullConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("AGENT_HOME", home)
	t.Setenv("PATH", home)

	fakeClaudeBinary(t, home)

	// ~/.claude.json 写入 null（合法 JSON 但不是 object）
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), []byte("null"), 0600); err != nil {
		t.Fatalf("写入 null claude.json 失败: %v", err)
	}

	p := &ClaudeProvider{}
	task := p.HealthCheckTask()
	err := task.Run(TaskContext{})
	if err == nil {
		t.Fatal("期望 null 配置文件时 HealthCheck 返回错误")
	}
	if !strings.Contains(err.Error(), "config invalid") {
		t.Errorf("错误应包含 'config invalid', 实际: %v", err)
	}
}

func TestValidateJSONFile_NullContent(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.json")
	if err := os.WriteFile(p, []byte("null"), 0600); err != nil {
		t.Fatalf("写入失败: %v", err)
	}
	err := validateJSONFile(p)
	if err == nil {
		t.Fatal("期望 null 内容返回错误")
	}
	if !strings.Contains(err.Error(), "got null") {
		t.Errorf("错误应包含 'got null', 实际: %v", err)
	}
}

func TestValidateJSONFile_ValidObject(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.json")
	if err := os.WriteFile(p, []byte(`{"key": "value"}`), 0600); err != nil {
		t.Fatalf("写入失败: %v", err)
	}
	if err := validateJSONFile(p); err != nil {
		t.Errorf("合法 JSON object 应通过, 实际: %v", err)
	}
}

func TestCodexProvider_DataMethods(t *testing.T) {
	p := &CodexProvider{}
	if p.ID() != "codex" {
		t.Errorf("ID() = %q, want %q", p.ID(), "codex")
	}
	if p.SessionIDPlaceholder() != "" {
		t.Errorf("SessionIDPlaceholder() = %q, want empty", p.SessionIDPlaceholder())
	}
	if p.RestoreCommandTemplate() != "" {
		t.Errorf("RestoreCommandTemplate() = %q, want empty", p.RestoreCommandTemplate())
	}
}

func TestCodexProvider_BootstrapTask_NoOp(t *testing.T) {
	p := &CodexProvider{}
	task := p.BootstrapTask()
	if task.Run != nil {
		t.Error("Codex BootstrapTask.Run should be nil")
	}
}

func TestCodexProvider_HealthCheckTask_NoBinary(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", home)

	p := &CodexProvider{}
	task := p.HealthCheckTask()
	if err := task.Run(TaskContext{}); err == nil {
		t.Error("期望 PATH 中无 codex 时 HealthCheck 返回错误")
	}
}

func TestBashProvider(t *testing.T) {
	p := &BashProvider{}
	if p.ID() != "bash" {
		t.Errorf("ID() = %q, want %q", p.ID(), "bash")
	}
	if p.SessionIDPlaceholder() != "" {
		t.Errorf("SessionIDPlaceholder() = %q, want empty", p.SessionIDPlaceholder())
	}
	task := p.BootstrapTask()
	if task.Run != nil {
		t.Error("Bash BootstrapTask.Run should be nil")
	}
	if len(p.EntrypointTasks()) != 0 {
		t.Error("Bash EntrypointTasks should be empty")
	}
	hcTask := p.HealthCheckTask()
	if hcTask.Run != nil {
		t.Error("Bash HealthCheckTask.Run should be nil")
	}
}
