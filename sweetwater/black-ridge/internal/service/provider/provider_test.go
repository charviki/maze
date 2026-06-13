package provider

import (
	"errors"
	"os"
	"testing"

	"github.com/charviki/maze/fabrication/cradle/logutil"
)

type stubProvider struct {
	id              string
	sessionIDHolder string
	restoreCmdTmpl  string
	bootstrapTask   Task
	entrypointTasks []Task
	healthCheckTask Task
}

func (s *stubProvider) ID() string                  { return s.id }
func (s *stubProvider) SessionIDPlaceholder() string { return s.sessionIDHolder }
func (s *stubProvider) RestoreCommandTemplate() string { return s.restoreCmdTmpl }
func (s *stubProvider) BootstrapTask() Task          { return s.bootstrapTask }
func (s *stubProvider) EntrypointTasks() []Task      { return s.entrypointTasks }
func (s *stubProvider) HealthCheckTask() Task        { return s.healthCheckTask }
func (s *stubProvider) CompletionHookConfig(_ string, _ string) *CompletionHookConfig { return nil }

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	p1 := &stubProvider{id: "alpha"}
	p2 := &stubProvider{id: "beta"}
	r.Register(p1)
	r.Register(p2)

	if got := r.Get("alpha"); got != p1 {
		t.Error("Get(alpha) should return p1")
	}
	if got := r.Get("beta"); got != p2 {
		t.Error("Get(beta) should return p2")
	}
}

func TestRegistry_GetNotFound(t *testing.T) {
	r := NewRegistry()
	if got := r.Get("nonexistent"); got != nil {
		t.Error("Get(nonexistent) should return nil")
	}
}

func TestRegistry_ListAvailable_FiltersUnhealthy(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubProvider{
		id: "healthy",
		healthCheckTask: Task{Name: "hc", Run: func(_ TaskContext) error { return nil }},
	})
	r.Register(&stubProvider{
		id: "unhealthy",
		healthCheckTask: Task{Name: "hc", Run: func(_ TaskContext) error { return errors.New("not found") }},
	})

	available := r.ListAvailable()
	if len(available) != 1 {
		t.Fatalf("ListAvailable length = %d, want 1", len(available))
	}
	if available[0] != "healthy" {
		t.Errorf("available[0] = %q, want %q", available[0], "healthy")
	}
}

func TestRegistry_ListAvailable_AllHealthy(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubProvider{
		id: "a",
		healthCheckTask: Task{Name: "hc", Run: func(_ TaskContext) error { return nil }},
	})
	r.Register(&stubProvider{
		id: "b",
		healthCheckTask: Task{Name: "hc", Run: func(_ TaskContext) error { return nil }},
	})

	available := r.ListAvailable()
	if len(available) != 2 {
		t.Fatalf("ListAvailable length = %d, want 2", len(available))
	}
}

func TestResolveHomeDir_AgentHome(t *testing.T) {
	t.Setenv("AGENT_HOME", "/custom/agent/home")
	if got := ResolveHomeDir(); got != "/custom/agent/home" {
		t.Errorf("ResolveHomeDir() = %q, want %q", got, "/custom/agent/home")
	}
}

func TestResolveHomeDir_Default(t *testing.T) {
	t.Setenv("AGENT_HOME", "")
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("os.UserHomeDir() failed")
	}
	if got := ResolveHomeDir(); got != home {
		t.Errorf("ResolveHomeDir() = %q, want %q", got, home)
	}
}

func TestRunTask_NilRun(t *testing.T) {
	task := Task{Name: "noop", Run: nil}
	if err := RunTask(logutil.NewNop(), task, TaskContext{}); err != nil {
		t.Errorf("nil Run should not error, got: %v", err)
	}
}

func TestRunTask_RunError(t *testing.T) {
	expected := errors.New("boom")
	task := Task{
		Name: "fail-task",
		Run:  func(_ TaskContext) error { return expected },
	}
	err := RunTask(logutil.NewNop(), task, TaskContext{})
	if err == nil {
		t.Fatal("should return error")
	}
	if !errors.Is(err, expected) {
		t.Errorf("error = %v, want wrapped %v", err, expected)
	}
}

func TestRegistry_ListAvailable_StableOrder(t *testing.T) {
	r := NewRegistry()
	for _, id := range []string{"codex", "bash", "claude"} {
		r.Register(&stubProvider{
			id:              id,
			healthCheckTask: Task{Name: "hc", Run: func(_ TaskContext) error { return nil }},
		})
	}

	got := r.ListAvailable()
	if len(got) != 3 {
		t.Fatalf("ListAvailable length = %d, want 3", len(got))
	}
	// 多次调用结果应一致且按字典序
	for i := 0; i < 5; i++ {
		result := r.ListAvailable()
		if result[0] != "bash" || result[1] != "claude" || result[2] != "codex" {
			t.Errorf("iteration %d: ListAvailable = %v, want [bash claude codex]", i, result)
		}
	}
}
