package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestClaudeTrustBootstrapper_TrustDirMarksProjectTrusted(t *testing.T) {
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

	bootstrapper := &ClaudeTrustBootstrapper{}
	if err := bootstrapper.TrustDir("/home/agent/existing"); err != nil {
		t.Fatalf("TrustDir 失败: %v", err)
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

func TestClaudeTrustBootstrapper_TrustDirPreservesTopLevelFields(t *testing.T) {
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

	bootstrapper := &ClaudeTrustBootstrapper{}
	if err := bootstrapper.TrustDir("/home/agent/new"); err != nil {
		t.Fatalf("TrustDir 失败: %v", err)
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
		t.Fatalf("TrustDir 不应越界改写全局 onboarding 状态, 实际 %#v", cfg["hasCompletedOnboarding"])
	}
}
