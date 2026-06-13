//go:build integration

package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charviki/maze/fabrication/cradle/logutil"
	"github.com/charviki/maze/fabrication/cradle/pipeline"

	"github.com/charviki/sweetwater-black-ridge/internal/config"
	"github.com/charviki/sweetwater-black-ridge/internal/service/provider"
)

func newTestTmuxService() *tmuxServiceImpl {
	return &tmuxServiceImpl{
		socketPath:   "",
		defaultShell: "/bin/bash",
		stateRepo:    newFileSessionStateRepository("/tmp/test-session-state"),
		logger:       logutil.NewNop(),
		registry:     provider.NewRegistry(),
	}
}

func TestBuildPipeline_WithWorkingDir(t *testing.T) {
	svc := newTestTmuxService()

	configs := []ConfigItem{
		{Type: ConfigTypeEnv, Key: "FOO", Value: "bar"},
		{Type: ConfigTypeFile, Key: "/tmp/test.txt", Value: "hello"},
	}

	pl := svc.BuildPipeline("/home/agent/project", "claude", configs, "claude", "test-session")

	system := pl.SystemSteps()
	if len(system) < 3 {
		t.Fatalf("system 层步骤数 = %d, 至少期望 3 (cd + env + file)", len(system))
	}

	// 第一个步骤应该是 cd
	if system[0].Type != pipeline.StepCD {
		t.Errorf("system[0].Type = %q, 期望 %q", system[0].Type, pipeline.StepCD)
	}
	if system[0].Key != "/home/agent/project" {
		t.Errorf("system[0].Key = %q, 期望 %q", system[0].Key, "/home/agent/project")
	}

	// 应包含 env 步骤
	envSteps := pl.SystemSteps()
	hasEnv := false
	for _, s := range envSteps {
		if s.Type == pipeline.StepEnv && s.Key == "FOO" {
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
		if s.Type == pipeline.StepFile && s.Key == "/tmp/test.txt" {
			hasFile = true
			break
		}
	}
	if !hasFile {
		t.Error("期望包含 file 步骤 /tmp/test.txt")
	}

	// template 层应有 command
	tpl := pl.TemplateSteps()
	if len(tpl) != 1 {
		t.Fatalf("template 层步骤数 = %d, 期望 1", len(tpl))
	}
	if tpl[0].Value != "claude" {
		t.Errorf("template[0].Value = %q, 期望 %q", tpl[0].Value, "claude")
	}
	if tpl[0].Phase != pipeline.PhaseTemplate {
		t.Errorf("template[0].Phase = %q, 期望 %q", tpl[0].Phase, pipeline.PhaseTemplate)
	}
}

func TestBuildPipeline_NoWorkingDir(t *testing.T) {
	svc := newTestTmuxService()
	pl := svc.BuildPipeline("", "bash", nil, "bash", "test-session")

	// 无工作目录时不应有 cd 步骤
	for _, s := range pl {
		if s.Type == pipeline.StepCD {
			t.Error("无工作目录时不应包含 cd 步骤")
		}
	}

	// template 层应有 command
	tpl := pl.TemplateSteps()
	if len(tpl) != 1 || tpl[0].Value != "bash" {
		t.Error("期望包含 bash command 步骤")
	}
}

func TestBuildPipeline_NoCommand(t *testing.T) {
	svc := newTestTmuxService()
	pl := svc.BuildPipeline("/home/agent", "", nil, "", "test-session")

	// 无命令时不应有 template 步骤
	tpl := pl.TemplateSteps()
	if len(tpl) != 0 {
		t.Errorf("无命令时 template 层步骤数 = %d, 期望 0", len(tpl))
	}

	// 应有 cd 步骤
	if len(pl.SystemSteps()) != 1 {
		t.Errorf("只有工作目录时 system 步骤数 = %d, 期望 1", len(pl.SystemSteps()))
	}
}

func TestBuildPipeline_OrderCorrectness(t *testing.T) {
	svc := newTestTmuxService()

	configs := []ConfigItem{
		{Type: ConfigTypeEnv, Key: "A", Value: "1"},
		{Type: ConfigTypeFile, Key: "/tmp/f", Value: "content"},
		{Type: ConfigTypeEnv, Key: "B", Value: "2"},
	}

	pl := svc.BuildPipeline("/home/agent", "claude", configs, "claude", "test-session")
	sorted := pl.Sorted()

	// 验证顺序: cd -> env -> env -> file -> command
	// 注意：有 claude provider 时还会插入 hook env + hook file，这里只验证非 hook 步骤的相对顺序
	cdIdx := -1
	cmdIdx := -1
	for i, s := range sorted {
		if s.Type == pipeline.StepCD {
			cdIdx = i
		}
		if s.Type == pipeline.StepCommand {
			cmdIdx = i
		}
	}
	if cdIdx == -1 {
		t.Fatal("期望包含 cd 步骤")
	}
	if cmdIdx == -1 {
		t.Fatal("期望包含 command 步骤")
	}
	if cdIdx > cmdIdx {
		t.Errorf("cd (index=%d) 应在 command (index=%d) 之前", cdIdx, cmdIdx)
	}
}

func TestBuildPipeline_Empty(t *testing.T) {
	svc := newTestTmuxService()
	pl := svc.BuildPipeline("", "", nil, "", "test-session")
	if len(pl) != 0 {
		t.Errorf("空输入管线步骤数 = %d, 期望 0", len(pl))
	}
}

func TestBuildPipeline_MultipleEnvsAndFiles(t *testing.T) {
	svc := newTestTmuxService()

	configs := []ConfigItem{
		{Type: ConfigTypeEnv, Key: "K1", Value: "V1"},
		{Type: ConfigTypeEnv, Key: "K2", Value: "V2"},
		{Type: ConfigTypeFile, Key: "/tmp/a.txt", Value: "aaa"},
		{Type: ConfigTypeFile, Key: "/tmp/b.txt", Value: "bbb"},
	}

	pl := svc.BuildPipeline("/home/agent", "bash", configs, "bash", "test-session")

	envCount := 0
	fileCount := 0
	for _, s := range pl.SystemSteps() {
		if s.Type == pipeline.StepEnv {
			envCount++
		}
		if s.Type == pipeline.StepFile {
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
		stateRepo: newFileSessionStateRepository(dir),
		logger:    logutil.NewNop(),
	}

	pl := pipeline.Pipeline{
		{ID: "sys-cd", Type: pipeline.StepCD, Phase: pipeline.PhaseSystem, Order: 0, Key: "/home/agent"},
	}

	// 创建 tmux session 以便 CapturePane 和 GetSessionEnv 可以工作
	// 由于没有真实 tmux，直接调用会失败，但 SavePipelineState 内部会忽略这些错误
	err := svc.SavePipelineState(SavePipelineStateOptions{
		SessionName:     "test-session",
		Pipeline:        pl,
		RestoreStrategy: "auto",
	})
	if err != nil {
		t.Fatalf("SavePipelineState 失败: %v", err)
	}

	filePath := filepath.Join(dir, "test-session.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取状态文件失败: %v", err)
	}

	var state SessionState
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
	svc := &tmuxServiceImpl{stateRepo: newFileSessionStateRepository(dir), logger: logutil.NewNop()}

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
	svc := &tmuxServiceImpl{stateRepo: newFileSessionStateRepository(dir), logger: logutil.NewNop()}

	// 删除不存在的文件不应报错
	err := svc.DeleteSessionState("nonexistent")
	if err != nil {
		t.Fatalf("删除不存在的文件不应报错: %v", err)
	}
}

func TestNewTmuxService_StateDir(t *testing.T) {
	cfg := &config.TmuxConfig{}
	svc := NewTmuxService(cfg, "/tmp/test-state", logutil.NewNop(), provider.NewRegistry())
	if svc == nil {
		t.Fatal("NewTmuxService 返回 nil")
	}
	impl := svc.(*tmuxServiceImpl)
	if impl.stateRepo == nil {
		t.Fatal("stateRepo 不应为 nil")
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

func TestBuildPipeline_WithPromptSteps(t *testing.T) {
	svc := newTestTmuxService()
	svc.registry.Register(&provider.ClaudeProvider{})

	configs := []ConfigItem{
		{Type: ConfigTypeEnv, Key: "FOO", Value: "bar"},
		{Type: ConfigTypePrompt, Key: "init", Value: "implement feature X"},
		{Type: ConfigTypePrompt, Key: "verify", Value: "run tests"},
	}

	pl := svc.BuildPipeline("/home/agent", "claude", configs, "claude", "my-session")

	user := pl.UserSteps()
	promptCount := 0
	for _, s := range user {
		if s.Type == pipeline.StepPrompt {
			promptCount++
		}
	}
	if promptCount != 2 {
		t.Errorf("prompt 步骤数 = %d, 期望 2", promptCount)
	}

	for _, s := range user {
		if s.ID == "usr-prompt-init" {
			if s.Value != "implement feature X" {
				t.Errorf("init prompt value = %q, 期望 %q", s.Value, "implement feature X")
			}
			// Key 应为 ready 信号文件路径
			if !strings.Contains(s.Key, "step_ready_") {
				t.Errorf("init prompt Key = %q, 期望包含 'step_ready_'", s.Key)
			}
			// Extra 应为 done 信号文件路径
			if !strings.Contains(s.Extra, "step_done_") {
				t.Errorf("init prompt Extra = %q, 期望包含 'step_done_'", s.Extra)
			}
		}
		if s.ID == "usr-prompt-verify" {
			if s.Value != "run tests" {
				t.Errorf("verify prompt value = %q, 期望 %q", s.Value, "run tests")
			}
			if !strings.Contains(s.Key, "step_ready_") {
				t.Errorf("verify prompt Key = %q, 期望包含 'step_ready_'", s.Key)
			}
			if !strings.Contains(s.Extra, "step_done_") {
				t.Errorf("verify prompt Extra = %q, 期望包含 'step_done_'", s.Extra)
			}
		}
	}
}

func TestBuildPipeline_HookInjectionForClaude(t *testing.T) {
	svc := newTestTmuxService()
	svc.registry.Register(&provider.ClaudeProvider{})

	pl := svc.BuildPipeline("/home/agent", "claude", nil, "claude", "test-sess")

	hasHookFile := false
	for _, s := range pl.SystemSteps() {
		// 验证 hook 配置文件同时包含 SessionStart 和 Stop
		if s.Type == pipeline.StepFile && strings.Contains(s.Key, ".claude") {
			hasHookFile = true
			// 解析 JSON 内容，验证同时包含 SessionStart 和 Stop
			var cfg map[string]any
			if err := json.Unmarshal([]byte(s.Value), &cfg); err != nil {
				t.Fatalf("hook 配置文件 JSON 解析失败: %v", err)
			}
			hooks, ok := cfg["hooks"].(map[string]any)
			if !ok {
				t.Fatal("hook 配置中缺少 hooks 字段")
			}
			if _, ok := hooks["SessionStart"]; !ok {
				t.Error("hook 配置中缺少 SessionStart hook")
			}
			if _, ok := hooks["Stop"]; !ok {
				t.Error("hook 配置中缺少 Stop hook")
			}
		}
	}
	if !hasHookFile {
		t.Error("期望包含 hook 配置文件步骤")
	}
}

func TestBuildPipeline_NoHookInjectionWithoutProvider(t *testing.T) {
	svc := newTestTmuxService()

	pl := svc.BuildPipeline("/home/agent", "bash", nil, "bash", "test-sess")

	for _, s := range pl.SystemSteps() {
		if s.ID == "sys-hook-file" {
			t.Error("无 provider 时不应注入 hook 配置文件步骤")
		}
	}
}

func TestBuildPipeline_PromptStepsAfterCommand(t *testing.T) {
	svc := newTestTmuxService()
	svc.registry.Register(&provider.ClaudeProvider{})

	configs := []ConfigItem{
		{Type: ConfigTypePrompt, Key: "init", Value: "hello"},
	}

	pl := svc.BuildPipeline("/home/agent", "claude", configs, "claude", "test")
	sorted := pl.Sorted()

	cmdIdx := -1
	promptIdx := -1
	for i, s := range sorted {
		if s.Type == pipeline.StepCommand {
			cmdIdx = i
		}
		if s.Type == pipeline.StepPrompt {
			promptIdx = i
		}
	}

	if cmdIdx == -1 {
		t.Fatal("期望包含 command 步骤")
	}
	if promptIdx == -1 {
		t.Fatal("期望包含 prompt 步骤")
	}
	if promptIdx <= cmdIdx {
		t.Errorf("prompt 步骤 (index=%d) 应在 command 步骤 (index=%d) 之后", promptIdx, cmdIdx)
	}
}
