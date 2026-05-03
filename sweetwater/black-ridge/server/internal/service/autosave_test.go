package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charviki/sweetwater-black-ridge/internal/model"
	"github.com/charviki/maze-cradle/logutil"
)

func TestGetSavedSessions_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	svc := &tmuxServiceImpl{stateDir: dir, logger: logutil.NewNop()}

	states, err := svc.GetSavedSessions()
	if err != nil {
		t.Fatalf("GetSavedSessions 失败: %v", err)
	}
	if len(states) != 0 {
		t.Errorf("空目录返回 %d 个状态, 期望 0", len(states))
	}
}

func TestGetSavedSessions_NonexistentDir(t *testing.T) {
	svc := &tmuxServiceImpl{stateDir: "/tmp/nonexistent-dir-12345", logger: logutil.NewNop()}

	states, err := svc.GetSavedSessions()
	if err != nil {
		t.Fatalf("GetSavedSessions 失败: %v", err)
	}
	if len(states) != 0 {
		t.Errorf("不存在的目录返回 %d 个状态, 期望 0", len(states))
	}
}

func TestGetSavedSessions_WithFiles(t *testing.T) {
	dir := t.TempDir()
	svc := &tmuxServiceImpl{stateDir: dir}

	// 写入两个状态文件
	state1 := model.SessionState{
		SessionName:     "session-a",
		Pipeline:        model.Pipeline{{ID: "sys-cd", Type: model.StepCD, Phase: model.PhaseSystem, Order: 0, Key: "/home/agent"}},
		RestoreStrategy: "auto",
		SavedAt:         time.Now().Format(time.RFC3339),
	}
	state2 := model.SessionState{
		SessionName:     "session-b",
		Pipeline:        model.Pipeline{{ID: "tpl-cmd", Type: model.StepCommand, Phase: model.PhaseTemplate, Order: 0, Value: "bash"}},
		RestoreStrategy: "manual",
		SavedAt:         time.Now().Format(time.RFC3339),
	}

	data1, _ := state1.ToJSON()
	data2, _ := state2.ToJSON()
	os.WriteFile(filepath.Join(dir, "session-a.json"), data1, 0644)
	os.WriteFile(filepath.Join(dir, "session-b.json"), data2, 0644)

	states, err := svc.GetSavedSessions()
	if err != nil {
		t.Fatalf("GetSavedSessions 失败: %v", err)
	}
	if len(states) != 2 {
		t.Errorf("返回 %d 个状态, 期望 2", len(states))
	}

	// 验证能找到两个 session
	names := map[string]bool{}
	for _, s := range states {
		names[s.SessionName] = true
	}
	if !names["session-a"] {
		t.Error("期望包含 session-a")
	}
	if !names["session-b"] {
		t.Error("期望包含 session-b")
	}
}

func TestGetSavedSessions_IgnoresInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	svc := &tmuxServiceImpl{stateDir: dir}

	// 写入一个无效 JSON 文件
	os.WriteFile(filepath.Join(dir, "bad.json"), []byte("not json"), 0644)

	// 写入一个有效文件
	state := model.SessionState{
		SessionName:     "good-session",
		RestoreStrategy: "auto",
		SavedAt:         time.Now().Format(time.RFC3339),
	}
	data, _ := state.ToJSON()
	os.WriteFile(filepath.Join(dir, "good-session.json"), data, 0644)

	states, err := svc.GetSavedSessions()
	if err != nil {
		t.Fatalf("GetSavedSessions 失败: %v", err)
	}
	if len(states) != 1 {
		t.Errorf("返回 %d 个状态, 期望 1 (无效 JSON 应被跳过)", len(states))
	}
	if states[0].SessionName != "good-session" {
		t.Errorf("SessionName = %q, 期望 %q", states[0].SessionName, "good-session")
	}
}

func TestSavePipelineState_FileContent(t *testing.T) {
	dir := t.TempDir()
	svc := &tmuxServiceImpl{stateDir: dir}

	pipeline := model.Pipeline{
		{ID: "sys-cd", Type: model.StepCD, Phase: model.PhaseSystem, Order: 0, Key: "/home/agent"},
		{ID: "tpl-cmd", Type: model.StepCommand, Phase: model.PhaseTemplate, Order: 1, Value: "claude --dangerously-skip-permissions"},
	}

	err := svc.SavePipelineState("my-session", pipeline, "auto", "", "", "")
	if err != nil {
		t.Fatalf("SavePipelineState 失败: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "my-session.json"))
	if err != nil {
		t.Fatalf("读取状态文件失败: %v", err)
	}

	var state model.SessionState
	if err := state.FromJSON(data); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证所有必要字段
	if state.SessionName != "my-session" {
		t.Errorf("SessionName = %q, 期望 %q", state.SessionName, "my-session")
	}
	if len(state.Pipeline) != 2 {
		t.Errorf("Pipeline 长度 = %d, 期望 2", len(state.Pipeline))
	}
	if state.RestoreStrategy != "auto" {
		t.Errorf("RestoreStrategy = %q, 期望 %q", state.RestoreStrategy, "auto")
	}
	if state.EnvSnapshot == nil {
		t.Error("EnvSnapshot 不应为 nil")
	}
	if state.SavedAt == "" {
		t.Error("SavedAt 不应为空")
	}
}

func TestAutoSaveService_DefaultInterval(t *testing.T) {
	svc := NewAutoSaveService(nil, 0, logutil.NewNop())
	if svc.interval != 60*time.Second {
		t.Errorf("默认 interval = %v, 期望 %v", svc.interval, 60*time.Second)
	}
}

func TestAutoSaveService_CustomInterval(t *testing.T) {
	svc := NewAutoSaveService(nil, 30, logutil.NewNop())
	if svc.interval != 30*time.Second {
		t.Errorf("自定义 interval = %v, 期望 %v", svc.interval, 30*time.Second)
	}
}

func TestAutoSaveService_NegativeInterval(t *testing.T) {
	svc := NewAutoSaveService(nil, -1, logutil.NewNop())
	if svc.interval != 60*time.Second {
		t.Errorf("负数 interval 应回退到默认 60s, 实际 = %v", svc.interval)
	}
}
