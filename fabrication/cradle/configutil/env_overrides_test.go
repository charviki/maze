package configutil

import (
	"os"
	"reflect"
	"testing"
)

type envOverrideConfig struct {
	Server struct {
		ServerConfig `yaml:",inline"`
		Enabled      bool `yaml:"enabled"`
	} `yaml:"server"`
	Worker struct {
		Concurrency int `yaml:"concurrency"`
	} `yaml:"worker"`
}

func TestApplyEnvOverrides(t *testing.T) {
	t.Setenv("MAZE_SERVER_LISTEN_ADDR", ":9090")
	t.Setenv("MAZE_SERVER_AUTH_TOKEN", "token")
	t.Setenv("MAZE_SERVER_ALLOWED_ORIGINS", "http://a.test, http://b.test")
	t.Setenv("MAZE_SERVER_ENABLED", "true")
	t.Setenv("MAZE_WORKER_CONCURRENCY", "8")

	var cfg envOverrideConfig
	if err := ApplyEnvOverrides("MAZE", &cfg); err != nil {
		t.Fatalf("ApplyEnvOverrides() error = %v", err)
	}

	if cfg.Server.ListenAddr != ":9090" {
		t.Fatalf("ListenAddr = %q, 期望 :9090", cfg.Server.ListenAddr)
	}
	if cfg.Server.AuthToken != "token" {
		t.Fatalf("AuthToken = %q, 期望 token", cfg.Server.AuthToken)
	}
	if !cfg.Server.Enabled {
		t.Fatal("Enabled = false, 期望 true")
	}
	if cfg.Worker.Concurrency != 8 {
		t.Fatalf("Concurrency = %d, 期望 8", cfg.Worker.Concurrency)
	}
	expectedOrigins := []string{"http://a.test", "http://b.test"}
	if !reflect.DeepEqual(cfg.Server.AllowedOrigins, expectedOrigins) {
		t.Fatalf("AllowedOrigins = %#v, 期望 %#v", cfg.Server.AllowedOrigins, expectedOrigins)
	}
}

func TestApplyEnvOverrides_InvalidInput(t *testing.T) {
	var cfg envOverrideConfig

	if err := ApplyEnvOverrides("MAZE", cfg); err == nil {
		t.Fatal("对非指针调用 ApplyEnvOverrides() 未报错，期望报错")
	}
	if err := ApplyEnvOverrides("MAZE", nil); err == nil {
		t.Fatal("对 nil 调用 ApplyEnvOverrides() 未报错，期望报错")
	}
}

func TestExpandHomePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}

	got := ExpandHomePath("~/maze/data")
	want := home + "/maze/data"
	if got != want {
		t.Fatalf("ExpandHomePath() = %q, 期望 %q", got, want)
	}
}

func TestServerConfigHelpers(t *testing.T) {
	cfg := ServerConfig{
		AuthToken:      "",
		AllowedOrigins: []string{" http://a.test ", "http://b.test"},
	}

	if !cfg.IsDevMode() {
		t.Fatal("IsDevMode() = false, 期望 true")
	}

	expected := []string{"http://a.test", "http://b.test"}
	if !reflect.DeepEqual(cfg.Origins(), expected) {
		t.Fatalf("Origins() = %#v, 期望 %#v", cfg.Origins(), expected)
	}
}
