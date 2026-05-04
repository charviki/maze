package service

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/charviki/maze-cradle/configutil"
	
)

func TestConfigFileService_ReadGlobalFiles_MissingFileReturnsEmptySnapshot(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

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
	if files[0].Path != "~/.claude/settings.json" {
		t.Fatalf("path = %q, 期望 ~/.claude/settings.json", files[0].Path)
	}
	if files[0].Exists {
		t.Fatal("缺失文件不应标记为存在")
	}
	if files[0].Content != "" {
		t.Fatalf("缺失文件内容应为空, 实际 %q", files[0].Content)
	}
	if files[0].Hash != "md5:empty" {
		t.Fatalf("空内容 hash = %q, 期望 md5:empty", files[0].Hash)
	}
}

func TestConfigFileService_SaveProjectFiles_WritesFileAfterHashCheck(t *testing.T) {
	workingDir := t.TempDir()
	svc := NewConfigFileService()
	defs := []configutil.FileDef{
		{Path: ".claude/settings.json"},
		{Path: "CLAUDE.md"},
	}

	files, err := svc.ReadProjectFiles(workingDir, defs)
	if err != nil {
		t.Fatalf("ReadProjectFiles 失败: %v", err)
	}

	updated, err := svc.SaveProjectFiles(workingDir, defs, []ConfigFileUpdate{
		{
			Path:     ".claude/settings.json",
			Content:  "{\n  \"theme\": \"dark\"\n}",
			BaseHash: files[0].Hash,
		},
	})
	if err != nil {
		t.Fatalf("SaveProjectFiles 失败: %v", err)
	}
	if len(updated) != 2 {
		t.Fatalf("updated 数量 = %d, 期望 2", len(updated))
	}

	data, err := os.ReadFile(filepath.Join(workingDir, ".claude/settings.json"))
	if err != nil {
		t.Fatalf("读取写回文件失败: %v", err)
	}
	if string(data) != "{\n  \"theme\": \"dark\"\n}" {
		t.Fatalf("写回内容不匹配: %q", string(data))
	}
	if !updated[0].Exists {
		t.Fatal("保存后 .claude/settings.json 应标记为存在")
	}
}

func TestConfigFileService_SaveProjectFiles_RejectsConflict(t *testing.T) {
	workingDir := t.TempDir()
	svc := NewConfigFileService()
	defs := []configutil.FileDef{
		{Path: ".claude/settings.json"},
	}

	configPath := filepath.Join(workingDir, ".claude/settings.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(`{"version":1}`), 0644); err != nil {
		t.Fatalf("写入初始文件失败: %v", err)
	}

	files, err := svc.ReadProjectFiles(workingDir, defs)
	if err != nil {
		t.Fatalf("ReadProjectFiles 失败: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(`{"version":2}`), 0644); err != nil {
		t.Fatalf("模拟外部改写失败: %v", err)
	}

	_, err = svc.SaveProjectFiles(workingDir, defs, []ConfigFileUpdate{
		{
			Path:     ".claude/settings.json",
			Content:  `{"version":3}`,
			BaseHash: files[0].Hash,
		},
	})
	if err == nil {
		t.Fatal("期望命中冲突错误")
	}

	var conflictErr *ConfigConflictError
	if !errors.As(err, &conflictErr) {
		t.Fatalf("错误类型 = %T, 期望 *ConfigConflictError", err)
	}
	if len(conflictErr.Conflicts) != 1 {
		t.Fatalf("冲突数量 = %d, 期望 1", len(conflictErr.Conflicts))
	}
	if conflictErr.Conflicts[0].Path != ".claude/settings.json" {
		t.Fatalf("冲突路径 = %q, 期望 .claude/settings.json", conflictErr.Conflicts[0].Path)
	}
}

func TestConfigFileService_SaveGlobalFiles_RejectsUnknownPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	svc := NewConfigFileService()
	_, err := svc.SaveGlobalFiles([]configutil.ConfigFile{
		{Path: "~/.claude/CLAUDE.md"},
	}, []ConfigFileUpdate{
		{
			Path:     "~/.claude/settings.json",
			Content:  "{}",
			BaseHash: "md5:empty",
		},
	})
	if err == nil {
		t.Fatal("期望拒绝未声明的固定路径")
	}
}
