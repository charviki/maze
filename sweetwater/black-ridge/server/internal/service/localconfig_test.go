package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/charviki/maze-cradle/logutil"

	"github.com/charviki/maze-cradle/protocol"
)

func TestLocalConfigStore_DefaultConfig(t *testing.T) {
	root := t.TempDir()
	store := NewLocalConfigStore(root, logutil.NewNop())

	cfg := store.Get()

	if cfg.WorkingDir != root {
		t.Errorf("WorkingDir 期望 %s, 实际 %s", root, cfg.WorkingDir)
	}
	if len(cfg.Env) != 0 {
		t.Errorf("Env 期望为空 map, 实际 %v", cfg.Env)
	}
}

func TestLocalConfigStore_PersistAndRecover(t *testing.T) {
	root := t.TempDir()

	store := NewLocalConfigStore(root, logutil.NewNop())
	if err := store.UpdateEnv(map[string]string{"FOO": "bar", "BAZ": "qux"}); err != nil {
		t.Fatalf("UpdateEnv 失败: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	store2 := NewLocalConfigStore(root, logutil.NewNop())
	cfg := store2.Get()

	if cfg.WorkingDir != root {
		t.Errorf("WorkingDir 期望 %s, 实际 %s", root, cfg.WorkingDir)
	}
	if cfg.Env["FOO"] != "bar" {
		t.Errorf("Env[FOO] 期望 bar, 实际 %s", cfg.Env["FOO"])
	}
	if cfg.Env["BAZ"] != "qux" {
		t.Errorf("Env[BAZ] 期望 qux, 实际 %s", cfg.Env["BAZ"])
	}
}

func TestLocalConfigStore_SetEnv(t *testing.T) {
	root := t.TempDir()
	store := NewLocalConfigStore(root, logutil.NewNop())

	if err := store.SetEnv("KEY1", "value1"); err != nil {
		t.Fatalf("SetEnv 失败: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	cfg := store.Get()
	if cfg.Env["KEY1"] != "value1" {
		t.Errorf("Env[KEY1] 期望 value1, 实际 %s", cfg.Env["KEY1"])
	}

	store2 := NewLocalConfigStore(root, logutil.NewNop())
	cfg2 := store2.Get()
	if cfg2.Env["KEY1"] != "value1" {
		t.Errorf("持久化后 Env[KEY1] 期望 value1, 实际 %s", cfg2.Env["KEY1"])
	}
}

func TestLocalConfigStore_UpdateEnvMerge(t *testing.T) {
	root := t.TempDir()
	store := NewLocalConfigStore(root, logutil.NewNop())

	if err := store.UpdateEnv(map[string]string{"A": "1", "B": "2"}); err != nil {
		t.Fatalf("初始 UpdateEnv 失败: %v", err)
	}
	if err := store.UpdateEnv(map[string]string{"C": "3"}); err != nil {
		t.Fatalf("合并 UpdateEnv 失败: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	cfg := store.Get()
	if cfg.Env["A"] != "1" {
		t.Errorf("Env[A] 期望 1, 实际 %s", cfg.Env["A"])
	}
	if cfg.Env["B"] != "2" {
		t.Errorf("Env[B] 期望 2, 实际 %s", cfg.Env["B"])
	}
	if cfg.Env["C"] != "3" {
		t.Errorf("Env[C] 期望 3, 实际 %s", cfg.Env["C"])
	}
}

func TestLocalConfigStore_UpdateEnvDelete(t *testing.T) {
	root := t.TempDir()
	store := NewLocalConfigStore(root, logutil.NewNop())

	if err := store.UpdateEnv(map[string]string{"X": "10", "Y": "20"}); err != nil {
		t.Fatalf("初始 UpdateEnv 失败: %v", err)
	}
	if err := store.UpdateEnv(map[string]string{"X": ""}); err != nil {
		t.Fatalf("删除 UpdateEnv 失败: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	cfg := store.Get()
	if _, ok := cfg.Env["X"]; ok {
		t.Errorf("Env[X] 应已被删除, 但仍存在: %s", cfg.Env["X"])
	}
	if cfg.Env["Y"] != "20" {
		t.Errorf("Env[Y] 期望 20, 实际 %s", cfg.Env["Y"])
	}
}

func TestLocalConfigStore_ConcurrentAccess(t *testing.T) {
	root := t.TempDir()
	store := NewLocalConfigStore(root, logutil.NewNop())

	var wg sync.WaitGroup
	const goroutines = 20

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = store.SetEnv("key", "value")
		}()
	}
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = store.Get()
		}()
	}
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = store.UpdateEnv(map[string]string{"CK": "cv"})
		}()
	}

	wg.Wait()
}

func TestLocalConfigStore_CorruptedFile(t *testing.T) {
	root := t.TempDir()

	configDir := filepath.Join(root, ".maze")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("创建配置目录失败: %v", err)
	}
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, []byte("{{invalid json}}"), 0644); err != nil {
		t.Fatalf("写入损坏文件失败: %v", err)
	}

	store := NewLocalConfigStore(root, logutil.NewNop())

	cfg := store.Get()
	if cfg.WorkingDir != root {
		t.Errorf("损坏文件时 WorkingDir 应回退为 %s, 实际 %s", root, cfg.WorkingDir)
	}
	if len(cfg.Env) != 0 {
		t.Errorf("损坏文件时 Env 应为空 map, 实际 %v", cfg.Env)
	}
}

func TestLocalConfigStore_GetReturnsCopy(t *testing.T) {
	root := t.TempDir()
	store := NewLocalConfigStore(root, logutil.NewNop())

	_ = store.SetEnv("ORIGINAL", "value")

	cfg := store.Get()
	cfg.Env["ORIGINAL"] = "tampered"
	cfg.WorkingDir = "/tampered"

	inner := store.Get()
	if inner.Env["ORIGINAL"] != "value" {
		t.Errorf("Get 返回非独立副本，内部状态被污染: 期望 value, 实际 %s", inner.Env["ORIGINAL"])
	}
	if inner.WorkingDir != root {
		t.Errorf("Get 返回非独立副本，WorkingDir 被污染: 期望 %s, 实际 %s", root, inner.WorkingDir)
	}
}

func TestLocalConfigStore_LoadExistingFile(t *testing.T) {
	root := t.TempDir()

	configDir := filepath.Join(root, ".maze")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("创建配置目录失败: %v", err)
	}
	existing := protocol.LocalAgentConfig{
		WorkingDir: "/custom/dir",
		Env:        map[string]string{"PRE": "set"},
	}
	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		t.Fatalf("序列化配置失败: %v", err)
	}
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	store := NewLocalConfigStore(root, logutil.NewNop())
	cfg := store.Get()

	if cfg.WorkingDir != root {
		t.Errorf("只读基础工作目录应固定为 %s, 实际 %s", root, cfg.WorkingDir)
	}
	if cfg.Env["PRE"] != "set" {
		t.Errorf("Env[PRE] 期望 set, 实际 %s", cfg.Env["PRE"])
	}
}
