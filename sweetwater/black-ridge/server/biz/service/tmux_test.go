package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charviki/maze-cradle/logutil"

	"github.com/charviki/sweetwater-black-ridge/biz/config"
	"github.com/charviki/sweetwater-black-ridge/biz/model"
)

func newTestTmuxService() *tmuxServiceImpl {
	return &tmuxServiceImpl{
		socketPath:   "",
		defaultShell: "/bin/bash",
		stateDir:     "/tmp/test-session-state",
		logger:       logutil.NewNop(),
	}
}

func TestBuildPipeline_WithWorkingDir(t *testing.T) {
	svc := newTestTmuxService()

	configs := []model.ConfigItem{
		{Type: "env", Key: "FOO", Value: "bar"},
		{Type: "file", Key: "/tmp/test.txt", Value: "hello"},
	}

	pipeline := svc.BuildPipeline("/home/agent/project", "claude", configs)

	system := pipeline.SystemSteps()
	if len(system) < 3 {
		t.Fatalf("system 层步骤数 = %d, 至少期望 3 (cd + env + file)", len(system))
	}

	// 第一个步骤应该是 cd
	if system[0].Type != model.StepCD {
		t.Errorf("system[0].Type = %q, 期望 %q", system[0].Type, model.StepCD)
	}
	if system[0].Key != "/home/agent/project" {
		t.Errorf("system[0].Key = %q, 期望 %q", system[0].Key, "/home/agent/project")
	}

	// 应包含 env 步骤
	envSteps := pipeline.SystemSteps()
	hasEnv := false
	for _, s := range envSteps {
		if s.Type == model.StepEnv && s.Key == "FOO" {
			hasEnv = true
			break
		}
	}
	if !hasEnv {
		t.Error("期望包含 env 步骤 FOO=bar")
	}

	// 应包含 file 步骤
	hasFile := false
	for _, s := range envSteps {
		if s.Type == model.StepFile && s.Key == "/tmp/test.txt" {
			hasFile = true
			break
		}
	}
	if !hasFile {
		t.Error("期望包含 file 步骤 /tmp/test.txt")
	}

	// template 层应有 command
	template := pipeline.TemplateSteps()
	if len(template) != 1 {
		t.Fatalf("template 层步骤数 = %d, 期望 1", len(template))
	}
	if template[0].Value != "claude" {
		t.Errorf("template[0].Value = %q, 期望 %q", template[0].Value, "claude")
	}
	if template[0].Phase != model.PhaseTemplate {
		t.Errorf("template[0].Phase = %q, 期望 %q", template[0].Phase, model.PhaseTemplate)
	}
}

func TestBuildPipeline_NoWorkingDir(t *testing.T) {
	svc := newTestTmuxService()
	pipeline := svc.BuildPipeline("", "bash", nil)

	// 无工作目录时不应有 cd 步骤
	for _, s := range pipeline {
		if s.Type == model.StepCD {
			t.Error("无工作目录时不应包含 cd 步骤")
		}
	}

	// template 层应有 command
	template := pipeline.TemplateSteps()
	if len(template) != 1 || template[0].Value != "bash" {
		t.Error("期望包含 bash command 步骤")
	}
}

func TestBuildPipeline_NoCommand(t *testing.T) {
	svc := newTestTmuxService()
	pipeline := svc.BuildPipeline("/home/agent", "", nil)

	// 无命令时不应有 template 步骤
	template := pipeline.TemplateSteps()
	if len(template) != 0 {
		t.Errorf("无命令时 template 层步骤数 = %d, 期望 0", len(template))
	}

	// 应有 cd 步骤
	if len(pipeline.SystemSteps()) != 1 {
		t.Errorf("只有工作目录时 system 步骤数 = %d, 期望 1", len(pipeline.SystemSteps()))
	}
}

func TestBuildPipeline_OrderCorrectness(t *testing.T) {
	svc := newTestTmuxService()

	configs := []model.ConfigItem{
		{Type: "env", Key: "A", Value: "1"},
		{Type: "file", Key: "/tmp/f", Value: "content"},
		{Type: "env", Key: "B", Value: "2"},
	}

	pipeline := svc.BuildPipeline("/home/agent", "claude", configs)
	sorted := pipeline.Sorted()

	// 验证顺序: cd -> env -> env -> file -> command
	expectedOrder := []model.PipelineStepType{
		model.StepCD, model.StepEnv, model.StepEnv, model.StepFile, model.StepCommand,
	}
	if len(sorted) != len(expectedOrder) {
		t.Fatalf("步骤数 = %d, 期望 %d", len(sorted), len(expectedOrder))
	}
	for i, step := range sorted {
		if step.Type != expectedOrder[i] {
			t.Errorf("sorted[%d].Type = %q, 期望 %q", i, step.Type, expectedOrder[i])
		}
	}
}

func TestBuildPipeline_Empty(t *testing.T) {
	svc := newTestTmuxService()
	pipeline := svc.BuildPipeline("", "", nil)
	if len(pipeline) != 0 {
		t.Errorf("空输入管线步骤数 = %d, 期望 0", len(pipeline))
	}
}

func TestBuildPipeline_MultipleEnvsAndFiles(t *testing.T) {
	svc := newTestTmuxService()

	configs := []model.ConfigItem{
		{Type: "env", Key: "K1", Value: "V1"},
		{Type: "env", Key: "K2", Value: "V2"},
		{Type: "file", Key: "/tmp/a.txt", Value: "aaa"},
		{Type: "file", Key: "/tmp/b.txt", Value: "bbb"},
	}

	pipeline := svc.BuildPipeline("/home/agent", "bash", configs)

	envCount := 0
	fileCount := 0
	for _, s := range pipeline.SystemSteps() {
		if s.Type == model.StepEnv {
			envCount++
		}
		if s.Type == model.StepFile {
			fileCount++
		}
	}
	if envCount != 2 {
		t.Errorf("env 步骤数 = %d, 期望 2", envCount)
	}
	if fileCount != 2 {
		t.Errorf("file 步骤数 = %d, 期望 2", fileCount)
	}
}

func TestSavePipelineState_WritesFile(t *testing.T) {
	dir := t.TempDir()
	svc := &tmuxServiceImpl{
		stateDir: dir,
		logger:   logutil.NewNop(),
	}

	pipeline := model.Pipeline{
		{ID: "sys-cd", Type: model.StepCD, Phase: model.PhaseSystem, Order: 0, Key: "/home/agent"},
	}

	// 创建 tmux session 以便 CapturePane 和 GetSessionEnv 可以工作
	// 由于没有真实 tmux，直接调用会失败，但 SavePipelineState 内部会忽略这些错误
	err := svc.SavePipelineState("test-session", pipeline, "auto", "", "", "")
	if err != nil {
		t.Fatalf("SavePipelineState 失败: %v", err)
	}

	filePath := filepath.Join(dir, "test-session.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取状态文件失败: %v", err)
	}

	var state model.SessionState
	if err := state.FromJSON(data); err != nil {
		t.Fatalf("反序列化状态文件失败: %v", err)
	}

	if state.SessionName != "test-session" {
		t.Errorf("SessionName = %q, 期望 %q", state.SessionName, "test-session")
	}
	if state.RestoreStrategy != "auto" {
		t.Errorf("RestoreStrategy = %q, 期望 %q", state.RestoreStrategy, "auto")
	}
	if len(state.Pipeline) != 1 {
		t.Errorf("Pipeline 长度 = %d, 期望 1", len(state.Pipeline))
	}
	if state.SavedAt == "" {
		t.Error("SavedAt 不应为空")
	}
}

func TestDeleteSessionState(t *testing.T) {
	dir := t.TempDir()
	svc := &tmuxServiceImpl{stateDir: dir, logger: logutil.NewNop()}

	// 先创建文件
	filePath := filepath.Join(dir, "test.json")
	os.WriteFile(filePath, []byte("{}"), 0644)

	err := svc.DeleteSessionState("test")
	if err != nil {
		t.Fatalf("DeleteSessionState 失败: %v", err)
	}

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("期望状态文件已被删除")
	}
}

func TestDeleteSessionState_NotExist(t *testing.T) {
	dir := t.TempDir()
	svc := &tmuxServiceImpl{stateDir: dir, logger: logutil.NewNop()}

	// 删除不存在的文件不应报错
	err := svc.DeleteSessionState("nonexistent")
	if err != nil {
		t.Fatalf("删除不存在的文件不应报错: %v", err)
	}
}

func TestNewTmuxService_StateDir(t *testing.T) {
	cfg := &config.TmuxConfig{}
	svc := NewTmuxService(cfg, "/tmp/test-state", logutil.NewNop())
	if svc == nil {
		t.Fatal("NewTmuxService 返回 nil")
	}
	impl := svc.(*tmuxServiceImpl)
	if impl.stateDir != "/tmp/test-state" {
		t.Fatalf("stateDir = %q, want /tmp/test-state", impl.stateDir)
	}
}

func TestResizeSession_ZeroSizeNoop(t *testing.T) {
	svc := newTestTmuxService()

	if err := svc.ResizeSession("demo", 0, 120); err != nil {
		t.Fatalf("rows=0 时应直接忽略, 实际错误: %v", err)
	}
	if err := svc.ResizeSession("demo", 40, 0); err != nil {
		t.Fatalf("cols=0 时应直接忽略, 实际错误: %v", err)
	}
}

func TestResizeSession_ReturnsCommandContextOnFailure(t *testing.T) {
	svc := newTestTmuxService()

	err := svc.ResizeSession("session-under-test", 40, 120)
	if err == nil {
		t.Skip("当前环境存在 tmux，无法稳定断言失败路径，跳过该检查")
	}
	if !strings.Contains(err.Error(), "resize tmux window session-under-test:0 to 120x40") {
		t.Fatalf("错误信息未包含目标窗口和尺寸, 实际: %v", err)
	}
}
