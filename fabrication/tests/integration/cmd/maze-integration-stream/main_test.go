package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRendererRun_StreamEvents(t *testing.T) {
	input := strings.Join([]string{
		`{"Action":"run","Test":"TestSessionSaveAndRestore"}`,
		`{"Action":"output","Test":"TestSessionSaveAndRestore","Output":"=== RUN   TestSessionSaveAndRestore\n"}`,
		`{"Action":"output","Test":"TestSessionSaveAndRestore","Output":"    session_lifecycle_test.go:22: [step] creating session...\n"}`,
		`{"Action":"output","Test":"TestSessionSaveAndRestore","Output":"--- PASS: TestSessionSaveAndRestore (1.25s)\n"}`,
		`{"Action":"pass","Test":"TestSessionSaveAndRestore","Elapsed":1.25}`,
	}, "\n")

	var out bytes.Buffer
	r := &renderer{out: &out, lastOutput: make(map[string]string)}
	if err := r.run(strings.NewReader(input)); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"[RUN ] TestSessionSaveAndRestore",
		"[LOG ] TestSessionSaveAndRestore | session_lifecycle_test.go:22: [step] creating session...",
		"[PASS] TestSessionSaveAndRestore (1.25s)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q, got:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{
		"=== RUN   TestSessionSaveAndRestore",
		"--- PASS: TestSessionSaveAndRestore",
	} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("output should filter %q, got:\n%s", unwanted, got)
		}
	}
}

func TestRendererRun_FailureSummary(t *testing.T) {
	input := strings.Join([]string{
		`{"Action":"run","Test":"TestHostCreateOnline"}`,
		`{"Action":"output","Test":"TestHostCreateOnline","Output":"    host_lifecycle_test.go:44: [wait] host=test-online status=deploying (want online), attempt=3\n"}`,
		`{"Action":"fail","Test":"TestHostCreateOnline","Elapsed":2.5}`,
	}, "\n")

	var out bytes.Buffer
	r := &renderer{out: &out, lastOutput: make(map[string]string)}
	if err := r.run(strings.NewReader(input)); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "[FAIL] TestHostCreateOnline (2.50s)") {
		t.Fatalf("missing fail event, got:\n%s", got)
	}
	if !strings.Contains(got, "[FAIL] summary: 1 test(s) failed") {
		t.Fatalf("missing fail summary, got:\n%s", got)
	}
	if !strings.Contains(got, "[FAIL]   TestHostCreateOnline | last log: host_lifecycle_test.go:44: [wait] host=test-online status=deploying (want online), attempt=3") {
		t.Fatalf("missing fail tail, got:\n%s", got)
	}
}
