package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charviki/maze-cradle/configutil"
)

func TestConfigFileService_ReadGlobalFiles_ExistingFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(configDir, 0750); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	configPath := filepath.Join(configDir, "settings.json")
	content := `{"theme":"dark"}`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}

	svc := NewConfigFileService()
	files, err := svc.ReadGlobalFiles([]configutil.ConfigFile{
		{Path: "~/.claude/settings.json"},
	})
	if err != nil {
		t.Fatalf("ReadGlobalFiles 失败: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("files 数量 = %d, 期望 1", len(files))
	}
	f := files[0]
	if f.Path != "~/.claude/settings.json" {
		t.Fatalf("path = %q, 期望 ~/.claude/settings.json", f.Path)
	}
	if !f.Exists {
		t.Fatal("已存在的文件应标记为存在")
	}
	if f.Content != content {
		t.Fatalf("content = %q, 期望 %q", f.Content, content)
	}
	if f.Hash == "md5:empty" {
		t.Fatal("非空内容的 hash 不应为 md5:empty")
	}
}

func TestConfigFileService_SaveGlobalFiles_WritesAndReturnsNewHash(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	defs := []configutil.ConfigFile{
		{Path: "~/.claude/settings.json"},
	}

	svc := NewConfigFileService()

	filesBefore, err := svc.ReadGlobalFiles(defs)
	if err != nil {
		t.Fatalf("初次 ReadGlobalFiles 失败: %v", err)
	}
	if filesBefore[0].Exists {
		t.Fatal("文件不应已存在")
	}

	newContent := `{"theme":"light"}`
	updated, err := svc.SaveGlobalFiles(defs, []ConfigFileUpdate{
		{
			Path:     "~/.claude/settings.json",
			Content:  newContent,
			BaseHash: filesBefore[0].Hash,
		},
	})
	if err != nil {
		t.Fatalf("SaveGlobalFiles 失败: %v", err)
	}
	if len(updated) != 1 {
		t.Fatalf("updated 数量 = %d, 期望 1", len(updated))
	}
	if !updated[0].Exists {
		t.Fatal("保存后应标记为存在")
	}
	if updated[0].Content != newContent {
		t.Fatalf("返回内容 = %q, 期望 %q", updated[0].Content, newContent)
	}
	if updated[0].Hash == filesBefore[0].Hash {
		t.Fatal("保存后 hash 应与保存前不同")
	}

	diskPath := filepath.Join(home, ".claude", "settings.json")
	data, err := os.ReadFile(diskPath)
	if err != nil {
		t.Fatalf("读取磁盘文件失败: %v", err)
	}
	if string(data) != newContent {
		t.Fatalf("磁盘文件内容 = %q, 期望 %q", string(data), newContent)
	}
}

func TestBuildProjectTargets_RejectsPathTraversal(t *testing.T) {
	workingDir := t.TempDir()
	svc := NewConfigFileService()

	_, err := svc.SaveProjectFiles(workingDir, []configutil.FileDef{
		{Path: "../secret"},
	}, []ConfigFileUpdate{
		{
			Path:     "../secret",
			Content:  "leaked",
			BaseHash: "md5:empty",
		},
	})
	if err == nil {
		t.Fatal("期望拒绝路径遍历 ../secret")
	}
}

func TestBuildGlobalTargets_RejectsNonAbsolutePath(t *testing.T) {
	svc := NewConfigFileService()

	_, err := svc.ReadGlobalFiles([]configutil.ConfigFile{
		{Path: "foo/bar"},
	})
	if err == nil {
		t.Fatal("期望拒绝非绝对路径 foo/bar")
	}
}
