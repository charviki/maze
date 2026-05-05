package configutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadYAML_Basic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "key: value\nnumber: 42\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	var target map[string]interface{}
	if err := LoadYAML(path, &target); err != nil {
		t.Fatalf("LoadYAML failed: %v", err)
	}
	if target["key"] != "value" {
		t.Errorf("key = %v, want value", target["key"])
	}
	if v, ok := target["number"].(int); !ok || v != 42 {
		t.Errorf("number = %v, want 42", target["number"])
	}
}

func TestLoadYAML_NotFound(t *testing.T) {
	var target map[string]interface{}
	err := LoadYAML("/nonexistent/config.yaml", &target)
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadYAML_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte("{invalid"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	var target map[string]interface{}
	err := LoadYAML(path, &target)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestServerConfig_IsDevMode(t *testing.T) {
	tests := []struct {
		name      string
		authToken string
		want      bool
	}{
		{name: "empty token", authToken: "", want: true},
		{name: "has token", authToken: "secret", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ServerConfig{AuthToken: tt.authToken}
			if cfg.IsDevMode() != tt.want {
				t.Errorf("IsDevMode = %v, want %v", cfg.IsDevMode(), tt.want)
			}
		})
	}
}

func TestServerConfig_Origins(t *testing.T) {
	tests := []struct {
		name    string
		origins []string
		wantLen int
	}{
		{name: "nil", origins: nil, wantLen: 0},
		{name: "empty", origins: []string{}, wantLen: 0},
		{name: "single", origins: []string{"http://localhost:3000"}, wantLen: 1},
		{name: "csv split", origins: []string{"http://a.com,http://b.com"}, wantLen: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ServerConfig{AllowedOrigins: tt.origins}
			got := cfg.Origins()
			if len(got) != tt.wantLen {
				t.Errorf("Origins length = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestServerConfig_Origins_TrimsWhitespace(t *testing.T) {
	cfg := ServerConfig{AllowedOrigins: []string{" http://a.com , http://b.com "}}
	got := cfg.Origins()
	if len(got) != 2 {
		t.Errorf("Origins length = %d, want 2: %v", len(got), got)
	}
}
