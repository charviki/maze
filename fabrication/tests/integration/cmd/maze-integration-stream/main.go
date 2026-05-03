package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

type testEvent struct {
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Elapsed float64 `json:"Elapsed"`
	Output  string  `json:"Output"`
}

type renderer struct {
	out        io.Writer
	failed     []string
	lastOutput map[string]string
}

func main() {
	r := &renderer{
		out:        os.Stdout,
		lastOutput: make(map[string]string),
	}
	if err := r.run(os.Stdin); err != nil {
		fmt.Fprintf(os.Stderr, "stream test events: %v\n", err)
		os.Exit(1)
	}
}

func (r *renderer) run(input io.Reader) error {
	scanner := bufio.NewScanner(input)
	// go test -json 的单行输出在集成场景下可能较长，显式放大 buffer 避免截断。
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		var event testEvent
		if err := json.Unmarshal(line, &event); err != nil {
			_, _ = fmt.Fprintln(r.out, string(line))
			continue
		}
		r.render(event)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if len(r.failed) > 0 {
		r.renderFailureSummary()
	}
	return nil
}

func (r *renderer) render(event testEvent) {
	switch event.Action {
	case "run":
		if event.Test != "" {
			_, _ = fmt.Fprintf(r.out, "[RUN ] %s\n", event.Test)
		}
	case "pass":
		if event.Test != "" {
			_, _ = fmt.Fprintf(r.out, "[PASS] %s (%.2fs)\n", event.Test, event.Elapsed)
		}
	case "skip":
		if event.Test != "" {
			_, _ = fmt.Fprintf(r.out, "[SKIP] %s (%.2fs)\n", event.Test, event.Elapsed)
		}
	case "fail":
		if event.Test != "" {
			r.trackFailure(event.Test)
			_, _ = fmt.Fprintf(r.out, "[FAIL] %s (%.2fs)\n", event.Test, event.Elapsed)
		}
	case "output":
		r.renderOutput(event)
	}
}

func (r *renderer) renderOutput(event testEvent) {
	line := strings.TrimRight(event.Output, "\n")
	if line == "" {
		return
	}
	if shouldSkipControlLine(line) {
		return
	}
	switch {
	case event.Test != "":
		trimmed := strings.TrimSpace(line)
		r.lastOutput[event.Test] = trimmed
		_, _ = fmt.Fprintf(r.out, "[LOG ] %s | %s\n", event.Test, trimmed)
	case line == "PASS" || line == "FAIL":
		return
	default:
		_, _ = fmt.Fprintln(r.out, line)
	}
}

func (r *renderer) trackFailure(testName string) {
	if slices.Contains(r.failed, testName) {
		return
	}
	r.failed = append(r.failed, testName)
}

func shouldSkipControlLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	for _, prefix := range []string{"=== RUN", "=== PAUSE", "=== CONT", "--- PASS:", "--- FAIL:", "--- SKIP:"} {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	return false
}

func (r *renderer) renderFailureSummary() {
	_, _ = fmt.Fprintf(r.out, "[FAIL] summary: %d test(s) failed\n", len(r.failed))
	for _, testName := range r.failed {
		lastLog, ok := r.lastOutput[testName]
		if ok && lastLog != "" {
			_, _ = fmt.Fprintf(r.out, "[FAIL]   %s | last log: %s\n", testName, lastLog)
			continue
		}
		_, _ = fmt.Fprintf(r.out, "[FAIL]   %s\n", testName)
	}
}
