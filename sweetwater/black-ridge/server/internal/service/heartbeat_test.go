package service

import (
	"errors"
	"testing"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/internal/config"
	"github.com/charviki/sweetwater-black-ridge/internal/service/provider"
)

func TestResolveDirectorCoreGRPCAddr_ExplicitGRPCAddr(t *testing.T) {
	cfg := config.ControllerConfig{
		Addr:     "http://myhost:8080",
		GRPCAddr: "myhost:50051",
	}
	got := resolveDirectorCoreGRPCAddr(cfg)
	if got != "myhost:50051" {
		t.Errorf("resolveDirectorCoreGRPCAddr = %q, want %q", got, "myhost:50051")
	}
}

func TestResolveDirectorCoreGRPCAddr_DeriveFromHTTPAddrWithScheme(t *testing.T) {
	cfg := config.ControllerConfig{
		Addr: "http://myhost:8080",
	}
	got := resolveDirectorCoreGRPCAddr(cfg)
	if got != "myhost:9090" {
		t.Errorf("resolveDirectorCoreGRPCAddr = %q, want %q", got, "myhost:9090")
	}
}

func TestResolveDirectorCoreGRPCAddr_DeriveFromHTTPAddrWithoutScheme(t *testing.T) {
	cfg := config.ControllerConfig{
		Addr: "myhost:8080",
	}
	got := resolveDirectorCoreGRPCAddr(cfg)
	if got != "myhost:9090" {
		t.Errorf("resolveDirectorCoreGRPCAddr = %q, want %q", got, "myhost:9090")
	}
}

func TestResolveDirectorCoreGRPCAddr_EmptyAddr(t *testing.T) {
	cfg := config.ControllerConfig{}
	got := resolveDirectorCoreGRPCAddr(cfg)
	if got != "" {
		t.Errorf("resolveDirectorCoreGRPCAddr = %q, want empty string", got)
	}
}

func TestExtractHostFromAddr_HTTP(t *testing.T) {
	got := extractHostFromAddr("http://myhost:8080")
	if got != "myhost" {
		t.Errorf("extractHostFromAddr = %q, want %q", got, "myhost")
	}
}

func TestExtractHostFromAddr_HTTPS(t *testing.T) {
	got := extractHostFromAddr("https://myhost:9090")
	if got != "myhost" {
		t.Errorf("extractHostFromAddr = %q, want %q", got, "myhost")
	}
}

func TestExtractHostFromAddr_Empty(t *testing.T) {
	got := extractHostFromAddr("")
	hostname := getOwnHostname()
	if got != hostname {
		t.Errorf("extractHostFromAddr empty input = %q, want %q (getOwnHostname fallback)", got, hostname)
	}
}

func TestExtractHostFromAddr_Invalid(t *testing.T) {
	got := extractHostFromAddr("invalid")
	hostname := getOwnHostname()
	if got != hostname {
		t.Errorf("extractHostFromAddr invalid input = %q, want %q (getOwnHostname fallback)", got, hostname)
	}
}

func TestGetOwnHostname_NonEmpty(t *testing.T) {
	got := getOwnHostname()
	if got == "" {
		t.Error("getOwnHostname returned empty string, want non-empty")
	}
}

func TestGetSupportedTemplates_NilRegistry(t *testing.T) {
	svc := &HeartbeatService{registry: nil}
	got := svc.getSupportedTemplates()
	if len(got) != len(defaultSupportedTemplates) {
		t.Errorf("nil registry 应返回 defaultSupportedTemplates, got %v", got)
	}
}

func TestGetSupportedTemplates_AllHealthy(t *testing.T) {
	r := provider.NewRegistry()
	r.Register(&testProvider{id: "claude", healthy: true})
	r.Register(&testProvider{id: "bash", healthy: true})
	svc := &HeartbeatService{registry: r}
	got := svc.getSupportedTemplates()
	if len(got) != 2 {
		t.Errorf("全部健康时应返回 2 个, got %v", got)
	}
}

func TestGetSupportedTemplates_AllUnhealthy(t *testing.T) {
	r := provider.NewRegistry()
	r.Register(&testProvider{id: "claude", healthy: false})
	svc := &HeartbeatService{registry: r}
	got := svc.getSupportedTemplates()
	if len(got) != 0 {
		t.Errorf("全部不健康时应返回空列表, got %v", got)
	}
}

func TestGetSupportedTemplates_EmptyRegistry(t *testing.T) {
	r := provider.NewRegistry()
	svc := &HeartbeatService{registry: r}
	got := svc.getSupportedTemplates()
	if len(got) != 0 {
		t.Errorf("空 registry 应返回空列表, got %v", got)
	}
}

// testProvider 用于测试的 stub Provider
type testProvider struct {
	id      string
	healthy bool
}

func (p *testProvider) ID() string                      { return p.id }
func (p *testProvider) SessionIDPlaceholder() string     { return "" }
func (p *testProvider) RestoreCommandTemplate() string   { return "" }
func (p *testProvider) BootstrapTask() provider.Task     { return provider.Task{Name: "noop"} }
func (p *testProvider) EntrypointTasks() []provider.Task { return nil }
func (p *testProvider) HealthCheckTask() provider.Task {
	return provider.Task{
		Name: "test-hc",
		Run: func(_ provider.TaskContext) error {
			if !p.healthy {
				return errors.New("unhealthy")
			}
			return nil
		},
	}
}

// compile-time check: HeartbeatService unused import guard
var _ = logutil.NewNop
