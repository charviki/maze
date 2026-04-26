package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEntrypointInitializesClaudeDefaultsInIndependentPhase(t *testing.T) {
	agentHome := t.TempDir()
	runEntrypoint(t, agentHome)

	claudeJSON := readJSONFile(t, filepath.Join(agentHome, ".claude.json"))
	if got, _ := claudeJSON["hasCompletedOnboarding"].(bool); !got {
		t.Fatalf("期望 hasCompletedOnboarding=true, 实际 %#v", claudeJSON["hasCompletedOnboarding"])
	}
	if got, _ := claudeJSON["migrationVersion"].(float64); got != 11 {
		t.Fatalf("期望 migrationVersion=11, 实际 %#v", claudeJSON["migrationVersion"])
	}
	if _, ok := claudeJSON["projects"].(map[string]any); !ok {
		t.Fatalf("期望 projects 为对象, 实际 %#v", claudeJSON["projects"])
	}

	settings := readJSONFile(t, filepath.Join(agentHome, ".claude", "settings.json"))
	if got, _ := settings["skipDangerousModePermissionPrompt"].(bool); !got {
		t.Fatalf("期望顶层 skipDangerousModePermissionPrompt=true, 实际 %#v", settings["skipDangerousModePermissionPrompt"])
	}
	permissions, ok := settings["permissions"].(map[string]any)
	if !ok {
		t.Fatalf("期望 permissions 为对象, 实际 %#v", settings["permissions"])
	}
	if got, _ := permissions["skipDangerousModePermissionPrompt"].(bool); !got {
		t.Fatalf("期望 permissions.skipDangerousModePermissionPrompt=true, 实际 %#v", permissions["skipDangerousModePermissionPrompt"])
	}
	if got, _ := settings["theme"].(string); got != "dark" {
		t.Fatalf("期望 theme=dark, 实际 %#v", settings["theme"])
	}
}

func TestEntrypointPreservesExistingClaudeStateWhilePatchingRequiredDefaults(t *testing.T) {
	agentHome := t.TempDir()
	claudeDir := filepath.Join(agentHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("创建 ~/.claude 失败: %v", err)
	}

	writeJSONFile(t, filepath.Join(agentHome, ".claude.json"), map[string]any{
		"hasCompletedOnboarding": false,
		"userID":                 "user-1",
		"projects": map[string]any{
			"/tmp/existing": map[string]any{
				"hasTrustDialogAccepted": true,
			},
		},
	})
	writeJSONFile(t, filepath.Join(claudeDir, "settings.json"), map[string]any{
		"theme": "light",
		"permissions": map[string]any{
			"allow": []string{"Bash(*)"},
		},
	})

	runEntrypoint(t, agentHome)

	claudeJSON := readJSONFile(t, filepath.Join(agentHome, ".claude.json"))
	if got, _ := claudeJSON["userID"].(string); got != "user-1" {
		t.Fatalf("期望保留 userID, 实际 %#v", claudeJSON["userID"])
	}
	if got, _ := claudeJSON["hasCompletedOnboarding"].(bool); !got {
		t.Fatalf("期望修复 hasCompletedOnboarding=true, 实际 %#v", claudeJSON["hasCompletedOnboarding"])
	}

	settings := readJSONFile(t, filepath.Join(claudeDir, "settings.json"))
	if got, _ := settings["theme"].(string); got != "light" {
		t.Fatalf("期望保留已有 theme=light, 实际 %#v", settings["theme"])
	}
	permissions := settings["permissions"].(map[string]any)
	if got, _ := permissions["skipDangerousModePermissionPrompt"].(bool); !got {
		t.Fatalf("期望补齐 permissions.skipDangerousModePermissionPrompt=true, 实际 %#v", permissions["skipDangerousModePermissionPrompt"])
	}
}

func runEntrypoint(t *testing.T, agentHome string) {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("无法定位测试文件路径")
	}
	entrypointPath := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "entrypoint.sh"))

	cmd := exec.Command("bash", entrypointPath, "true")
	cmd.Env = append(os.Environ(),
		"AGENT_HOME="+agentHome,
		"SKIP_TMUX_INIT=1",
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("执行 entrypoint 失败: %v\n输出:\n%s", err, string(output))
	}
}

func readJSONFile(t *testing.T, path string) map[string]any {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取 %s 失败: %v", path, err)
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("解析 %s 失败: %v", path, err)
	}
	return value
}

func writeJSONFile(t *testing.T, path string, value map[string]any) {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("序列化 %s 失败: %v", path, err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("写入 %s 失败: %v", path, err)
	}
}
