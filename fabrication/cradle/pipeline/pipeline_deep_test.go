package pipeline

import (
	"context"
	"sync"
	"testing"
)

func TestPipeline_Sorted_Stable(t *testing.T) {
	p := Pipeline{
		{ID: "a1", Order: 1},
		{ID: "a2", Order: 1},
		{ID: "b", Order: 2},
	}
	sorted := p.Sorted()
	if sorted[0].ID != "a1" {
		t.Errorf("sorted[0].ID = %q, want a1 (stable sort preserves order)", sorted[0].ID)
	}
	if sorted[1].ID != "a2" {
		t.Errorf("sorted[1].ID = %q, want a2", sorted[1].ID)
	}
}

func TestPipeline_Sorted_StableEmpty(t *testing.T) {
	p := Pipeline{}
	sorted := p.Sorted()
	if len(sorted) != 0 {
		t.Errorf("empty sorted length = %d, want 0", len(sorted))
	}
}

func TestPipeline_Sorted_DoesNotMutate(t *testing.T) {
	p := Pipeline{
		{ID: "c", Order: 2},
		{ID: "a", Order: 0},
	}
	originalID0 := p[0].ID
	_ = p.Sorted()
	if p[0].ID != originalID0 {
		t.Error("Sorted should not mutate original pipeline")
	}
}

func TestPipeline_SingleStep(t *testing.T) {
	p := Pipeline{{ID: "only", Order: 0}}
	sorted := p.Sorted()
	if sorted[0].ID != "only" {
		t.Errorf("sorted[0].ID = %q, want only", sorted[0].ID)
	}
}

func TestPipeline_ConcurrentFilter(t *testing.T) {
	p := make(Pipeline, 200)
	for i := 0; i < 200; i++ {
		p[i] = PipelineStep{
			ID:    string(rune('a' + i%26)),
			Phase: PhaseSystem,
			Order: i,
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = p.SystemSteps()
			_ = p.TemplateSteps()
			_ = p.UserSteps()
			_ = p.Sorted()
		}()
	}
	wg.Wait()
}

func TestPipeline_ConcurrentSorted(t *testing.T) {
	p := make(Pipeline, 100)
	for i := 0; i < 100; i++ {
		p[i] = PipelineStep{ID: "s", Order: i % 10}
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sorted := p.Sorted()
			if len(sorted) != len(p) {
				t.Errorf("sorted length mismatch: %d vs %d", len(sorted), len(p))
			}
		}()
	}
	wg.Wait()
}

func TestPipeline_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	p := Pipeline{
		{ID: "step1", Phase: PhaseSystem, Order: 0},
		{ID: "step2", Phase: PhaseSystem, Order: 1},
	}

	done := make(chan struct{})
	cancel()

	go func() {
		defer close(done)
		for _, step := range p {
			select {
			case <-ctx.Done():
				return
			default:
				_ = step.ID
			}
		}
	}()

	<-done
}

func TestPipeline_ContextDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()
	<-ctx.Done()
	if ctx.Err() == nil {
		t.Error("context should be expired")
	}
}

func TestPipeline_PhaseConstants(t *testing.T) {
	if PhaseSystem != "system" {
		t.Errorf("PhaseSystem = %q, want system", PhaseSystem)
	}
	if PhaseTemplate != "template" {
		t.Errorf("PhaseTemplate = %q, want template", PhaseTemplate)
	}
	if PhaseUser != "user" {
		t.Errorf("PhaseUser = %q, want user", PhaseUser)
	}
}

func TestPipeline_StepTypeConstants(t *testing.T) {
	if StepCD != "cd" {
		t.Errorf("StepCD = %q, want cd", StepCD)
	}
	if StepEnv != "env" {
		t.Errorf("StepEnv = %q, want env", StepEnv)
	}
	if StepFile != "file" {
		t.Errorf("StepFile = %q, want file", StepFile)
	}
	if StepCommand != "command" {
		t.Errorf("StepCommand = %q, want command", StepCommand)
	}
}
