package service

import (
	"strings"
	"testing"

	"github.com/charviki/maze-cradle/pipeline"
)

func TestSessionState_ToJSON_ContainsAllFields(t *testing.T) {
	s := SessionState{
		SessionName:      "test-session",
		Pipeline:         pipeline.Pipeline{},
		RestoreStrategy:  "tmux",
		WorkingDir:       "/home/agent",
		EnvSnapshot:      map[string]string{},
		TerminalSnapshot: "",
		SavedAt:          "2025-01-01T00:00:00Z",
	}

	data, err := s.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON 失败: %v", err)
	}

	alwaysPresentKeys := []string{
		`"session_name"`,
		`"pipeline"`,
		`"restore_strategy"`,
		`"working_dir"`,
		`"env_snapshot"`,
		`"terminal_snapshot"`,
		`"saved_at"`,
	}
	raw := string(data)
	for _, key := range alwaysPresentKeys {
		if !strings.Contains(raw, key) {
			t.Errorf("ToJSON 输出缺少 JSON key: %s", key)
		}
	}

	omitemptyKeys := []string{`"restore_command"`, `"template_id"`, `"cli_session_id"`}
	for _, key := range omitemptyKeys {
		if strings.Contains(raw, key) {
			t.Errorf("omitempty 字段零值时不应该出现: %s", key)
		}
	}
}

func TestSessionState_ToJSON_OmitemptyFieldsPresent(t *testing.T) {
	s := SessionState{
		SessionName:      "test-session",
		Pipeline:         pipeline.Pipeline{},
		RestoreStrategy:  "tmux",
		RestoreCommand:   "tmux new-session -d -s test",
		WorkingDir:       "/home/agent",
		TemplateID:       "tpl-001",
		CLISessionID:     "cli-abc123",
		EnvSnapshot:      map[string]string{},
		TerminalSnapshot: "",
		SavedAt:          "2025-01-01T00:00:00Z",
	}

	data, err := s.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON 失败: %v", err)
	}

	raw := string(data)
	omitemptyKeys := []string{`"restore_command"`, `"template_id"`, `"cli_session_id"`}
	for _, key := range omitemptyKeys {
		if !strings.Contains(raw, key) {
			t.Errorf("omitempty 字段有值时应该出现: %s", key)
		}
	}
}

func TestSessionState_FromJSON_InvalidJSON(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{name: "完全无效", input: "{{not-json"},
		{name: "空字符串", input: ""},
		{name: "截断JSON", input: `{"session_name": "test"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s SessionState
			err := s.FromJSON([]byte(tc.input))
			if err == nil {
				t.Error("期望返回错误，实际返回 nil")
			}
		})
	}
}

func TestSessionState_RoundTrip(t *testing.T) {
	original := SessionState{
		SessionName: "round-trip-session",
		Pipeline: pipeline.Pipeline{
			pipeline.PipelineStep{ID: "sys-cd", Type: pipeline.StepCD, Phase: pipeline.PhaseSystem, Order: 0, Key: "/home/agent"},
			pipeline.PipelineStep{ID: "sys-env-FOO", Type: pipeline.StepEnv, Phase: pipeline.PhaseSystem, Order: 1, Key: "FOO", Value: "bar"},
			pipeline.PipelineStep{ID: "tpl-command", Type: pipeline.StepCommand, Phase: pipeline.PhaseTemplate, Order: 2, Value: "claude --dangerously-skip-permissions"},
		},
		RestoreStrategy:  "tmux",
		RestoreCommand:   "tmux new-session -d -s test",
		WorkingDir:       "/home/agent/project",
		TemplateID:       "tpl-001",
		CLISessionID:     "cli-abc123",
		EnvSnapshot:      map[string]string{"PATH": "/usr/bin", "HOME": "/home/agent"},
		TerminalSnapshot: "last visible content",
		SavedAt:          "2025-06-15T12:30:00Z",
	}

	data, err := original.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON 失败: %v", err)
	}

	var restored SessionState
	if err := restored.FromJSON(data); err != nil {
		t.Fatalf("FromJSON 失败: %v", err)
	}

	if restored.SessionName != original.SessionName {
		t.Errorf("SessionName: 期望 %q, 实际 %q", original.SessionName, restored.SessionName)
	}
	if len(restored.Pipeline) != len(original.Pipeline) {
		t.Fatalf("Pipeline 长度: 期望 %d, 实际 %d", len(original.Pipeline), len(restored.Pipeline))
	}
	for i, step := range original.Pipeline {
		if restored.Pipeline[i] != step {
			t.Errorf("Pipeline[%d]: 期望 %+v, 实际 %+v", i, step, restored.Pipeline[i])
		}
	}
	if restored.RestoreStrategy != original.RestoreStrategy {
		t.Errorf("RestoreStrategy: 期望 %q, 实际 %q", original.RestoreStrategy, restored.RestoreStrategy)
	}
	if restored.RestoreCommand != original.RestoreCommand {
		t.Errorf("RestoreCommand: 期望 %q, 实际 %q", original.RestoreCommand, restored.RestoreCommand)
	}
	if restored.WorkingDir != original.WorkingDir {
		t.Errorf("WorkingDir: 期望 %q, 实际 %q", original.WorkingDir, restored.WorkingDir)
	}
	if restored.TemplateID != original.TemplateID {
		t.Errorf("TemplateID: 期望 %q, 实际 %q", original.TemplateID, restored.TemplateID)
	}
	if restored.CLISessionID != original.CLISessionID {
		t.Errorf("CLISessionID: 期望 %q, 实际 %q", original.CLISessionID, restored.CLISessionID)
	}
	for k, v := range original.EnvSnapshot {
		if restored.EnvSnapshot[k] != v {
			t.Errorf("EnvSnapshot[%q]: 期望 %q, 实际 %q", k, v, restored.EnvSnapshot[k])
		}
	}
	if len(restored.EnvSnapshot) != len(original.EnvSnapshot) {
		t.Errorf("EnvSnapshot 长度: 期望 %d, 实际 %d", len(original.EnvSnapshot), len(restored.EnvSnapshot))
	}
	if restored.TerminalSnapshot != original.TerminalSnapshot {
		t.Errorf("TerminalSnapshot: 期望 %q, 实际 %q", original.TerminalSnapshot, restored.TerminalSnapshot)
	}
	if restored.SavedAt != original.SavedAt {
		t.Errorf("SavedAt: 期望 %q, 实际 %q", original.SavedAt, restored.SavedAt)
	}
}
