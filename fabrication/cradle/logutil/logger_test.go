package logutil

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	l := New("test-component")
	if l == nil {
		t.Fatal("New returned nil")
	}
	if l.level != slog.LevelInfo {
		t.Errorf("level = %v, want %v", l.level, slog.LevelInfo)
	}
	if len(l.attrs) != 2 || l.attrs[0] != "component" || l.attrs[1] != "test-component" {
		t.Errorf("attrs = %v, want [component test-component]", l.attrs)
	}
}

func TestNewNop(t *testing.T) {
	l := NewNop()
	if l == nil {
		t.Fatal("NewNop returned nil")
	}
	if l.level != slog.LevelInfo {
		t.Errorf("level = %v, want %v", l.level, slog.LevelInfo)
	}
	if len(l.attrs) != 0 {
		t.Errorf("attrs should be empty, got %v", l.attrs)
	}
}

func TestLoggerInterface_LogLevels(t *testing.T) {
	var buf bytes.Buffer
	l := New("test")
	l.SetOutput(&buf)

	l.Infof("info message %d", 1)
	l.Warnf("warn message %s", "hello")
	l.Errorf("error message: %v", "something")

	output := buf.String()
	if !strings.Contains(output, "info message 1") {
		t.Errorf("output missing info message: %s", output)
	}
	if !strings.Contains(output, "warn message hello") {
		t.Errorf("output missing warn message: %s", output)
	}
	if !strings.Contains(output, "error message: something") {
		t.Errorf("output missing error message: %s", output)
	}
}

func TestSetLevel(t *testing.T) {
	var buf bytes.Buffer
	l := New("test")
	l.SetOutput(&buf)

	l.SetLevel(slog.LevelError)
	l.Infof("should not appear")
	l.Errorf("should appear")

	output := buf.String()
	if strings.Contains(output, "should not appear") {
		t.Error("info message should not appear at error level")
	}
	if !strings.Contains(output, "should appear") {
		t.Error("error message should appear at error level")
	}
}

func TestWithNode(t *testing.T) {
	l1 := New("test")
	l2 := l1.WithNode("node-1")

	if l1 == l2 {
		t.Error("WithNode should return a new instance")
	}

	slogL2 := l2.(*SlogLogger)
	if len(slogL2.attrs) <= len(l1.attrs) {
		t.Error("WithNode should add attrs")
	}
}

func TestWithSession(t *testing.T) {
	l1 := New("test")
	l2 := l1.WithSession("sess-1")

	slogL2 := l2.(*SlogLogger)
	hasSession := false
	for i := 0; i < len(slogL2.attrs); i += 2 {
		if slogL2.attrs[i] == "session_id" && slogL2.attrs[i+1] == "sess-1" {
			hasSession = true
			break
		}
	}
	if !hasSession {
		t.Error("WithSession should add session_id attr")
	}
}

func TestWithAction(t *testing.T) {
	l1 := New("test")
	l2 := l1.WithAction("proxy")

	slogL2 := l2.(*SlogLogger)
	hasAction := false
	for i := 0; i < len(slogL2.attrs); i += 2 {
		if slogL2.attrs[i] == "action" && slogL2.attrs[i+1] == "proxy" {
			hasAction = true
			break
		}
	}
	if !hasAction {
		t.Error("WithAction should add action attr")
	}
}

func TestWithFields(t *testing.T) {
	tests := []struct {
		name   string
		fields []string
		check  func(t *testing.T, l *SlogLogger)
	}{
		{
			name:   "empty fields",
			fields: nil,
			check: func(t *testing.T, l *SlogLogger) {
				original := New("test")
				if len(l.attrs) != len(original.attrs) {
					t.Error("empty fields should not change attrs")
				}
			},
		},
		{
			name:   "paired fields",
			fields: []string{"key1", "val1", "key2", "val2"},
			check: func(t *testing.T, l *SlogLogger) {
				hasKey1 := false
				for i := 0; i < len(l.attrs); i += 2 {
					if l.attrs[i] == "key1" && l.attrs[i+1] == "val1" {
						hasKey1 = true
					}
				}
				if !hasKey1 {
					t.Error("paired fields should add attrs")
				}
			},
		},
		{
			name:   "odd number of fields",
			fields: []string{"key1", "val1", "key2"},
			check: func(t *testing.T, l *SlogLogger) {
				original := New("test")
				if len(l.attrs) != len(original.attrs) {
					t.Error("odd fields should not change attrs")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test")
			result := l.WithFields(tt.fields...)
			tt.check(t, result.(*SlogLogger))
		})
	}
}

func TestLogger_JSONOutput(t *testing.T) {
	var buf bytes.Buffer
	l := New("test")
	l.SetOutput(&buf)

	l.WithNode("n1").WithSession("s1").Infof("hello")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if entry["msg"] != "hello" {
		t.Errorf("msg = %v, want hello", entry["msg"])
	}
	if entry["component"] != "test" {
		t.Errorf("component = %v, want test", entry["component"])
	}
}

func TestLogger_CloneIndependence(t *testing.T) {
	l1 := New("test")
	l2 := l1.WithNode("node-a")
	l3 := l1.WithNode("node-b")

	slogL2 := l2.(*SlogLogger)
	slogL3 := l3.(*SlogLogger)

	var nodeALast, nodeBLast string
	for i := len(slogL2.attrs) - 2; i >= 0; i -= 2 {
		if slogL2.attrs[i] == "node_name" {
			nodeALast = slogL2.attrs[i+1].(string)
			break
		}
	}
	for i := len(slogL3.attrs) - 2; i >= 0; i -= 2 {
		if slogL3.attrs[i] == "node_name" {
			nodeBLast = slogL3.attrs[i+1].(string)
			break
		}
	}

	if nodeALast != "node-a" {
		t.Errorf("l2 node_name = %q, want node-a", nodeALast)
	}
	if nodeBLast != "node-b" {
		t.Errorf("l3 node_name = %q, want node-b", nodeBLast)
	}
}

func TestLogger_SetOutput(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	l := New("test")

	l.SetOutput(&buf1)
	l.Infof("to buf1")
	if buf1.Len() == 0 {
		t.Error("message should be written to buf1")
	}

	l.SetOutput(&buf2)
	l.Infof("to buf2")
	if buf2.Len() == 0 {
		t.Error("message should be written to buf2")
	}
}

func TestLogger_TraceDebug(t *testing.T) {
	var buf bytes.Buffer
	l := New("test")
	l.SetOutput(&buf)
	l.SetLevel(slog.LevelDebug)

	l.Tracef("trace msg")
	l.Debugf("debug msg")

	output := buf.String()
	if !strings.Contains(output, "trace msg") {
		t.Error("output missing trace msg")
	}
	if !strings.Contains(output, "debug msg") {
		t.Error("output missing debug msg")
	}
}

func TestLogger_VariadicMethods(t *testing.T) {
	var buf bytes.Buffer
	l := New("test")
	l.SetOutput(&buf)

	l.Info("hello", "world")
	output := buf.String()
	if !strings.Contains(output, "hello") {
		t.Errorf("output missing expected message: %s", output)
	}
}

func TestLogger_CtxMethods(t *testing.T) {
	var buf bytes.Buffer
	l := New("test")
	l.SetOutput(&buf)

	l.CtxInfof(context.Background(), "ctx info")
	if !strings.Contains(buf.String(), "ctx info") {
		t.Errorf("output missing ctx info: %s", buf.String())
	}
}

func TestLogger_Notice(t *testing.T) {
	var buf bytes.Buffer
	l := New("test")
	l.SetOutput(&buf)

	l.Noticef("notice msg")
	if !strings.Contains(buf.String(), "notice msg") {
		t.Errorf("output missing notice msg: %s", buf.String())
	}
}
