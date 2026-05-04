package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charviki/maze-cradle/pipeline"
)

func TestFileSessionStateRepository_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	repo := newFileSessionStateRepository(dir)

	state := &SessionState{
		SessionName:     "test-session",
		Pipeline:        pipeline.Pipeline{{ID: "sys-cd", Type: pipeline.StepCD, Phase: pipeline.PhaseSystem, Order: 0, Key: "/home/agent"}},
		RestoreStrategy: "auto",
		WorkingDir:      "/home/agent",
		EnvSnapshot:     map[string]string{"PATH": "/usr/bin"},
		SavedAt:         time.Now().Format(time.RFC3339),
	}

	if err := repo.Save(state); err != nil {
		t.Fatalf("Save 失败: %v", err)
	}

	loaded, err := repo.Load("test-session")
	if err != nil {
		t.Fatalf("Load 失败: %v", err)
	}
	if loaded.SessionName != "test-session" {
		t.Errorf("SessionName = %q, 期望 %q", loaded.SessionName, "test-session")
	}
	if loaded.RestoreStrategy != "auto" {
		t.Errorf("RestoreStrategy = %q, 期望 %q", loaded.RestoreStrategy, "auto")
	}
	if len(loaded.Pipeline) != 1 {
		t.Errorf("Pipeline 长度 = %d, 期望 1", len(loaded.Pipeline))
	}
	if loaded.EnvSnapshot["PATH"] != "/usr/bin" {
		t.Errorf("EnvSnapshot[PATH] = %q, 期望 %q", loaded.EnvSnapshot["PATH"], "/usr/bin")
	}
}

func TestFileSessionStateRepository_LoadNotExist(t *testing.T) {
	dir := t.TempDir()
	repo := newFileSessionStateRepository(dir)

	_, err := repo.Load("nonexistent")
	if err == nil {
		t.Fatal("期望返回错误，实际为 nil")
	}
}

func TestFileSessionStateRepository_Delete(t *testing.T) {
	dir := t.TempDir()
	repo := newFileSessionStateRepository(dir)

	state := &SessionState{SessionName: "to-delete", SavedAt: time.Now().Format(time.RFC3339)}
	repo.Save(state)

	if err := repo.Delete("to-delete"); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}

	filePath := filepath.Join(dir, "to-delete.json")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("期望文件已被删除")
	}
}

func TestFileSessionStateRepository_DeleteNotExist(t *testing.T) {
	dir := t.TempDir()
	repo := newFileSessionStateRepository(dir)

	if err := repo.Delete("nonexistent"); err != nil {
		t.Fatalf("删除不存在的文件不应报错: %v", err)
	}
}

func TestFileSessionStateRepository_List(t *testing.T) {
	dir := t.TempDir()
	repo := newFileSessionStateRepository(dir)

	state1 := &SessionState{SessionName: "session-a", SavedAt: time.Now().Format(time.RFC3339)}
	state2 := &SessionState{SessionName: "session-b", SavedAt: time.Now().Format(time.RFC3339)}
	repo.Save(state1)
	repo.Save(state2)

	states, err := repo.List()
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if len(states) != 2 {
		t.Errorf("返回 %d 个状态, 期望 2", len(states))
	}
}

func TestFileSessionStateRepository_ListEmptyDir(t *testing.T) {
	dir := t.TempDir()
	repo := newFileSessionStateRepository(dir)

	states, err := repo.List()
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if len(states) != 0 {
		t.Errorf("空目录返回 %d 个状态, 期望 0", len(states))
	}
}

func TestFileSessionStateRepository_ListNonexistentDir(t *testing.T) {
	repo := newFileSessionStateRepository("/tmp/nonexistent-dir-12345")

	states, err := repo.List()
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if len(states) != 0 {
		t.Errorf("不存在的目录返回 %d 个状态, 期望 0", len(states))
	}
}

func TestFileSessionStateRepository_ListSkipsTemplatesJSON(t *testing.T) {
	dir := t.TempDir()
	repo := newFileSessionStateRepository(dir)

	// 写入 templates.json（应被跳过）
	os.WriteFile(filepath.Join(dir, "templates.json"), []byte("{}"), 0644)

	// 写入有效 session state
	state := &SessionState{SessionName: "session-a", SavedAt: time.Now().Format(time.RFC3339)}
	repo.Save(state)

	states, err := repo.List()
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if len(states) != 1 {
		t.Errorf("返回 %d 个状态, 期望 1 (templates.json 应被跳过)", len(states))
	}
	if states[0].SessionName != "session-a" {
		t.Errorf("SessionName = %q, 期望 %q", states[0].SessionName, "session-a")
	}
}

func TestFileSessionStateRepository_ListSkipsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	repo := newFileSessionStateRepository(dir)

	os.WriteFile(filepath.Join(dir, "bad.json"), []byte("not json"), 0644)

	state := &SessionState{SessionName: "good-session", SavedAt: time.Now().Format(time.RFC3339)}
	repo.Save(state)

	states, err := repo.List()
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if len(states) != 1 {
		t.Errorf("返回 %d 个状态, 期望 1 (无效 JSON 应被跳过)", len(states))
	}
}
