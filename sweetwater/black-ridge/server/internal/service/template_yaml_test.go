package service

import (
	"path/filepath"
	"testing"

	"github.com/charviki/maze-cradle/configutil"
	"github.com/charviki/maze-cradle/logutil"
)

func TestLoadBuiltinFromYAML_ClaudeTemplate(t *testing.T) {
	tpl, err := loadBuiltinFromYAML("claude.yaml")
	if err != nil {
		t.Fatalf("loadBuiltinFromYAML 失败: %v", err)
	}

	if tpl.ID != "claude" {
		t.Errorf("ID: 期望 claude, 实际 %s", tpl.ID)
	}
	if tpl.Name != "Claude Code" {
		t.Errorf("Name: 期望 Claude Code, 实际 %s", tpl.Name)
	}
	if tpl.Command != "IS_SANDBOX=1 claude --dangerously-skip-permissions --session-id {session_id}" {
		t.Errorf("Command 不匹配, 实际 %s", tpl.Command)
	}
	if tpl.RestoreCommand != "IS_SANDBOX=1 claude --dangerously-skip-permissions --resume {session_id}" {
		t.Errorf("RestoreCommand 不匹配, 实际 %s", tpl.RestoreCommand)
	}
	if tpl.SessionFilePattern != "~/.claude/projects/{encoded_working_dir}/*.jsonl" {
		t.Errorf("SessionFilePattern 不匹配, 实际 %s", tpl.SessionFilePattern)
	}
	if tpl.Description != "Anthropic Claude CLI Agent" {
		t.Errorf("Description: 期望 Anthropic Claude CLI Agent, 实际 %s", tpl.Description)
	}
	if tpl.Icon != "🤖" {
		t.Errorf("Icon: 期望 🤖, 实际 %s", tpl.Icon)
	}
	if !tpl.Builtin {
		t.Error("Builtin: 期望 true")
	}

	if len(tpl.Defaults.Env) != 0 {
		t.Errorf("Defaults.Env: 期望空 map, 实际 %d 项", len(tpl.Defaults.Env))
	}
	if len(tpl.Defaults.Files) != 3 {
		t.Fatalf("Defaults.Files: 期望 3 项, 实际 %d 项", len(tpl.Defaults.Files))
	}

	expectedPaths := map[string]bool{
		"~/.claude.json":          false,
		"~/.claude/settings.json": false,
		"~/.claude/CLAUDE.md":     false,
	}
	for _, f := range tpl.Defaults.Files {
		if _, ok := expectedPaths[f.Path]; ok {
			expectedPaths[f.Path] = true
		}
	}
	for p, found := range expectedPaths {
		if !found {
			t.Errorf("Defaults.Files 缺少 %s", p)
		}
	}

	if len(tpl.SessionSchema.EnvDefs) != 0 {
		t.Errorf("SessionSchema.EnvDefs: 期望 0 项, 实际 %d 项", len(tpl.SessionSchema.EnvDefs))
	}
	if len(tpl.SessionSchema.FileDefs) != 2 {
		t.Fatalf("SessionSchema.FileDefs: 期望 2 项, 实际 %d 项", len(tpl.SessionSchema.FileDefs))
	}

	fileDefPaths := []string{tpl.SessionSchema.FileDefs[0].Path, tpl.SessionSchema.FileDefs[1].Path}
	for _, want := range []string{"CLAUDE.md", ".claude/settings.json"} {
		found := false
		for _, got := range fileDefPaths {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SessionSchema.FileDefs 缺少 %s", want)
		}
	}
}

func TestLoadBuiltinFromYAML_CodexTemplate(t *testing.T) {
	tpl, err := loadBuiltinFromYAML("codex.yaml")
	if err != nil {
		t.Fatalf("loadBuiltinFromYAML 失败: %v", err)
	}

	if tpl.ID != "codex" {
		t.Errorf("ID: 期望 codex, 实际 %s", tpl.ID)
	}
	if tpl.Name != "Codex" {
		t.Errorf("Name: 期望 Codex, 实际 %s", tpl.Name)
	}
	if tpl.Command != "codex --full-auto" {
		t.Errorf("Command: 期望 codex --full-auto, 实际 %s", tpl.Command)
	}
	if tpl.Description != "OpenAI Codex Agent" {
		t.Errorf("Description: 期望 OpenAI Codex Agent, 实际 %s", tpl.Description)
	}
	if tpl.Icon != "⚡" {
		t.Errorf("Icon: 期望 ⚡, 实际 %s", tpl.Icon)
	}
	if !tpl.Builtin {
		t.Error("Builtin: 期望 true")
	}

	if len(tpl.Defaults.Files) != 2 {
		t.Fatalf("Defaults.Files: 期望 2 项, 实际 %d 项", len(tpl.Defaults.Files))
	}
	filePaths := map[string]bool{
		"~/.codex/config.toml": false,
		"~/AGENTS.md":         false,
	}
	for _, f := range tpl.Defaults.Files {
		if _, ok := filePaths[f.Path]; ok {
			filePaths[f.Path] = true
		}
	}
	for p, found := range filePaths {
		if !found {
			t.Errorf("Defaults.Files 缺少 %s", p)
		}
	}

	if len(tpl.SessionSchema.FileDefs) != 2 {
		t.Fatalf("SessionSchema.FileDefs: 期望 2 项, 实际 %d 项", len(tpl.SessionSchema.FileDefs))
	}
	fileDefPaths := map[string]bool{
		"AGENTS.md":          false,
		".codex/config.toml": false,
	}
	for _, fd := range tpl.SessionSchema.FileDefs {
		if _, ok := fileDefPaths[fd.Path]; ok {
			fileDefPaths[fd.Path] = true
		}
	}
	for p, found := range fileDefPaths {
		if !found {
			t.Errorf("SessionSchema.FileDefs 缺少 %s", p)
		}
	}
}

func TestLoadBuiltinFromYAML_BashTemplate(t *testing.T) {
	tpl, err := loadBuiltinFromYAML("bash.yaml")
	if err != nil {
		t.Fatalf("loadBuiltinFromYAML 失败: %v", err)
	}

	if tpl.ID != "bash" {
		t.Errorf("ID: 期望 bash, 实际 %s", tpl.ID)
	}
	if tpl.Name != "Bash Shell" {
		t.Errorf("Name: 期望 Bash Shell, 实际 %s", tpl.Name)
	}
	if tpl.Command != "" {
		t.Errorf("Command: 期望空字符串, 实际 %s", tpl.Command)
	}
	if tpl.Description != "纯 Bash 终端" {
		t.Errorf("Description: 期望 纯 Bash 终端, 实际 %s", tpl.Description)
	}
	if tpl.Icon != "🖥️" {
		t.Errorf("Icon: 期望 🖥️, 实际 %s", tpl.Icon)
	}
	if !tpl.Builtin {
		t.Error("Builtin: 期望 true")
	}
	if len(tpl.Defaults.Files) != 0 {
		t.Errorf("Defaults.Files: 期望 0 项, 实际 %d 项", len(tpl.Defaults.Files))
	}
	if len(tpl.SessionSchema.EnvDefs) != 0 {
		t.Errorf("SessionSchema.EnvDefs: 期望 0 项, 实际 %d 项", len(tpl.SessionSchema.EnvDefs))
	}
	if len(tpl.SessionSchema.FileDefs) != 0 {
		t.Errorf("SessionSchema.FileDefs: 期望 0 项, 实际 %d 项", len(tpl.SessionSchema.FileDefs))
	}
}

func TestLoadBuiltinFromYAML_InvalidName(t *testing.T) {
	_, err := loadBuiltinFromYAML("nonexistent.yaml")
	if err == nil {
		t.Fatal("期望返回错误, 实际为 nil")
	}
}

func TestTemplateStore_EnsureBuiltins_PreservesCustomTemplates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "templates.json")
	store := NewTemplateStore(path, logutil.NewNop())

	custom := &SessionTemplate{
		ID:          "my-custom",
		Name:        "自定义模板",
		Command:     "echo hello",
		Description: "测试自定义",
		Icon:        "🔧",
		Builtin:     false,
		Defaults: configutil.ConfigLayer{
			Env:   map[string]string{"KEY": "val"},
			Files: []configutil.ConfigFile{},
		},
		SessionSchema: configutil.SessionSchema{},
	}
	if err := store.Set(custom); err != nil {
		t.Fatalf("Set 自定义模板失败: %v", err)
	}

	store.ensureBuiltins()

	got := store.Get("my-custom")
	if got == nil {
		t.Fatal("ensureBuiltins 后自定义模板丢失")
	}
	if got.Name != "自定义模板" {
		t.Errorf("自定义模板 Name: 期望 自定义模板, 实际 %s", got.Name)
	}
	if got.Builtin {
		t.Error("自定义模板 Builtin 应为 false")
	}
}
