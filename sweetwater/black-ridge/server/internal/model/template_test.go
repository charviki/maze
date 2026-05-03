package model

import (
	"path/filepath"
	"testing"

	"github.com/charviki/maze-cradle/configutil"
	"github.com/charviki/maze-cradle/logutil"
)

func TestTemplateStore_EnsureBuiltins(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "templates.json")
	store := NewTemplateStore(path, logutil.NewNop())

	// 验证三个内置模板（claude, codex, bash）都被加载
	builtinIDs := []string{"claude", "codex", "bash"}
	for _, id := range builtinIDs {
		tpl := store.Get(id)
		if tpl == nil {
			t.Errorf("期望内置模板 %s 存在, 实际为 nil", id)
			continue
		}
		if tpl.ID != id {
			t.Errorf("期望 ID=%s, 实际=%s", id, tpl.ID)
		}
		if !tpl.Builtin {
			t.Errorf("期望 Builtin=true for %s", id)
		}
	}

	// 验证内置模板总数
	all := store.List()
	if len(all) < 3 {
		t.Errorf("期望至少 3 个内置模板, 实际=%d", len(all))
	}

	claudeTpl := store.Get("claude")
	if claudeTpl == nil {
		t.Fatal("期望 claude 内置模板存在")
	}
	if claudeTpl.Command != "IS_SANDBOX=1 claude --dangerously-skip-permissions --session-id {session_id}" {
		t.Fatalf("claude 默认命令未对齐自动化权限, 实际 %q", claudeTpl.Command)
	}
	if claudeTpl.RestoreCommand != "IS_SANDBOX=1 claude --dangerously-skip-permissions --resume {session_id}" {
		t.Fatalf("claude 恢复命令不符合预期, 实际 %q", claudeTpl.RestoreCommand)
	}
	expectedGlobalPaths := map[string]bool{
		"~/.claude.json":          false,
		"~/.claude/settings.json": false,
		"~/.claude/CLAUDE.md":     false,
	}
	for _, file := range claudeTpl.Defaults.Files {
		if _, ok := expectedGlobalPaths[file.Path]; ok {
			expectedGlobalPaths[file.Path] = true
		}
	}
	for path, found := range expectedGlobalPaths {
		if !found {
			t.Fatalf("claude 全局固定路径缺失: %s", path)
		}
	}
}

func TestTemplateStore_SetAndGet(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "templates.json")
	store := NewTemplateStore(path, logutil.NewNop())

	custom := &SessionTemplate{
		ID:          "my-template",
		Name:        "自定义模板",
		Command:     "echo hello",
		Description: "测试用自定义模板",
		Icon:        "🔧",
		Builtin:     false,
		Defaults: configutil.ConfigLayer{
			Env:   map[string]string{"FOO": "bar"},
			Files: []configutil.ConfigFile{},
		},
		SessionSchema: configutil.SessionSchema{},
	}

	err := store.Set(custom)
	if err != nil {
		t.Fatalf("Set 失败: %v", err)
	}

	got := store.Get("my-template")
	if got == nil {
		t.Fatal("期望 Get 返回非 nil")
	}
	if got.Name != "自定义模板" {
		t.Errorf("期望 Name=自定义模板, 实际=%s", got.Name)
	}
	if got.Command != "echo hello" {
		t.Errorf("期望 Command=echo hello, 实际=%s", got.Command)
	}
	if got.Builtin {
		t.Error("期望 Builtin=false")
	}
}

func TestTemplateStore_Delete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "templates.json")
	store := NewTemplateStore(path, logutil.NewNop())

	store.Set(&SessionTemplate{
		ID:      "custom-1",
		Name:    "临时模板",
		Command: "true",
		Builtin: false,
	})

	err := store.Delete("custom-1")
	if err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}

	got := store.Get("custom-1")
	if got != nil {
		t.Error("期望删除后 Get 返回 nil")
	}
}

func TestTemplateStore_DeleteBuiltin(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "templates.json")
	store := NewTemplateStore(path, logutil.NewNop())

	// 尝试删除内置模板，应为静默 no-op（不报错也不删除）
	err := store.Delete("claude")
	if err != nil {
		t.Fatalf("删除内置模板不应报错: %v", err)
	}

	got := store.Get("claude")
	if got == nil {
		t.Fatal("内置模板不应被删除")
	}
	if got.ID != "claude" {
		t.Errorf("期望 ID=claude, 实际=%s", got.ID)
	}
}

func TestTemplateStore_Persistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "templates.json")

	// 第一个 store 写入自定义模板
	store1 := NewTemplateStore(path, logutil.NewNop())
	store1.Set(&SessionTemplate{
		ID:      "custom-persist",
		Name:    "持久化测试",
		Command: "echo test",
		Builtin: false,
	})

	// 用相同文件路径创建第二个 store，验证自定义模板和内置模板都存在
	store2 := NewTemplateStore(path, logutil.NewNop())

	// 自定义模板应持久化
	got := store2.Get("custom-persist")
	if got == nil {
		t.Fatal("期望自定义模板被持久化加载")
	}
	if got.Name != "持久化测试" {
		t.Errorf("期望 Name=持久化测试, 实际=%s", got.Name)
	}

	// 内置模板应被 ensureBuiltins 恢复
	builtinIDs := []string{"claude", "codex", "bash"}
	for _, id := range builtinIDs {
		tpl := store2.Get(id)
		if tpl == nil {
			t.Errorf("期望内置模板 %s 被恢复, 实际为 nil", id)
		}
	}
}
