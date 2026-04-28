package pipeline

import (
	"encoding/json"
	"testing"
)

func TestPipelineStep_JSONRoundTrip(t *testing.T) {
	original := PipelineStep{
		ID:    "sys-env-API_KEY",
		Type:  StepEnv,
		Phase: PhaseSystem,
		Order: 2,
		Key:   "API_KEY",
		Value: "sk-12345678",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal 失败: %v", err)
	}

	var decoded PipelineStep
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal 失败: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID: 期望 %q, 实际 %q", original.ID, decoded.ID)
	}
	if decoded.Type != original.Type {
		t.Errorf("Type: 期望 %q, 实际 %q", original.Type, decoded.Type)
	}
	if decoded.Phase != original.Phase {
		t.Errorf("Phase: 期望 %q, 实际 %q", original.Phase, decoded.Phase)
	}
	if decoded.Order != original.Order {
		t.Errorf("Order: 期望 %d, 实际 %d", original.Order, decoded.Order)
	}
	if decoded.Key != original.Key {
		t.Errorf("Key: 期望 %q, 实际 %q", original.Key, decoded.Key)
	}
	if decoded.Value != original.Value {
		t.Errorf("Value: 期望 %q, 实际 %q", original.Value, decoded.Value)
	}
}

func TestPipelineStep_JSONRoundTrip_AllTypes(t *testing.T) {
	steps := []PipelineStep{
		{ID: "cd", Type: StepCD, Phase: PhaseSystem, Order: 0, Key: "/home/agent", Value: ""},
		{ID: "env", Type: StepEnv, Phase: PhaseSystem, Order: 1, Key: "FOO", Value: "bar"},
		{ID: "file", Type: StepFile, Phase: PhaseSystem, Order: 2, Key: "/tmp/test.txt", Value: "hello"},
		{ID: "cmd", Type: StepCommand, Phase: PhaseTemplate, Order: 3, Key: "", Value: "claude"},
	}

	for _, step := range steps {
		data, err := json.Marshal(step)
		if err != nil {
			t.Fatalf("Marshal %s 步骤失败: %v", step.Type, err)
		}
		var decoded PipelineStep
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal %s 步骤失败: %v", step.Type, err)
		}
		if decoded != step {
			t.Errorf("步骤 %s round-trip 不一致: 期望 %+v, 实际 %+v", step.Type, step, decoded)
		}
	}
}

func TestPipeline_Sorted(t *testing.T) {
	pipeline := Pipeline{
		{ID: "c", Order: 2, Value: "third"},
		{ID: "a", Order: 0, Value: "first"},
		{ID: "b", Order: 1, Value: "second"},
	}

	sorted := pipeline.Sorted()

	if sorted[0].ID != "a" {
		t.Errorf("sorted[0].ID = %q, 期望 %q", sorted[0].ID, "a")
	}
	if sorted[1].ID != "b" {
		t.Errorf("sorted[1].ID = %q, 期望 %q", sorted[1].ID, "b")
	}
	if sorted[2].ID != "c" {
		t.Errorf("sorted[2].ID = %q, 期望 %q", sorted[2].ID, "c")
	}

	// 原始管线不应被修改
	if pipeline[0].ID != "c" {
		t.Error("Sorted 不应修改原始管线")
	}
}

func TestPipeline_Sorted_Empty(t *testing.T) {
	pipeline := Pipeline{}
	sorted := pipeline.Sorted()
	if len(sorted) != 0 {
		t.Errorf("空管线排序后长度 = %d, 期望 0", len(sorted))
	}
}

func TestPipeline_SystemSteps(t *testing.T) {
	pipeline := Pipeline{
		{ID: "sys-1", Phase: PhaseSystem, Order: 0},
		{ID: "tpl-1", Phase: PhaseTemplate, Order: 1},
		{ID: "sys-2", Phase: PhaseSystem, Order: 2},
		{ID: "usr-1", Phase: PhaseUser, Order: 3},
	}

	system := pipeline.SystemSteps()
	if len(system) != 2 {
		t.Fatalf("SystemSteps 长度 = %d, 期望 2", len(system))
	}
	for _, s := range system {
		if s.Phase != PhaseSystem {
			t.Errorf("SystemSteps 包含非 system 步骤: %s (phase=%s)", s.ID, s.Phase)
		}
	}
}

func TestPipeline_TemplateSteps(t *testing.T) {
	pipeline := Pipeline{
		{ID: "sys-1", Phase: PhaseSystem, Order: 0},
		{ID: "tpl-1", Phase: PhaseTemplate, Order: 1},
		{ID: "usr-1", Phase: PhaseUser, Order: 2},
	}

	template := pipeline.TemplateSteps()
	if len(template) != 1 {
		t.Fatalf("TemplateSteps 长度 = %d, 期望 1", len(template))
	}
	if template[0].ID != "tpl-1" {
		t.Errorf("TemplateSteps[0].ID = %q, 期望 %q", template[0].ID, "tpl-1")
	}
}

func TestPipeline_UserSteps(t *testing.T) {
	pipeline := Pipeline{
		{ID: "sys-1", Phase: PhaseSystem, Order: 0},
		{ID: "tpl-1", Phase: PhaseTemplate, Order: 1},
		{ID: "usr-1", Phase: PhaseUser, Order: 2},
		{ID: "usr-2", Phase: PhaseUser, Order: 3},
	}

	user := pipeline.UserSteps()
	if len(user) != 2 {
		t.Fatalf("UserSteps 长度 = %d, 期望 2", len(user))
	}
}

func TestPipeline_FilterByPhase_Empty(t *testing.T) {
	pipeline := Pipeline{}
	result := pipeline.filterByPhase(PhaseSystem)
	if len(result) != 0 {
		t.Errorf("空管线 filterByPhase 长度 = %d, 期望 0", len(result))
	}
}
