package service

import (
	"testing"

	"github.com/charviki/sweetwater-black-ridge/internal/config"
)

func TestResolveManagerGRPCAddr_ExplicitGRPCAddr(t *testing.T) {
	cfg := config.ControllerConfig{
		Addr:     "http://myhost:8080",
		GRPCAddr: "myhost:50051",
	}
	got := resolveManagerGRPCAddr(cfg)
	if got != "myhost:50051" {
		t.Errorf("resolveManagerGRPCAddr = %q, want %q", got, "myhost:50051")
	}
}

func TestResolveManagerGRPCAddr_DeriveFromHTTPAddrWithScheme(t *testing.T) {
	cfg := config.ControllerConfig{
		Addr: "http://myhost:8080",
	}
	got := resolveManagerGRPCAddr(cfg)
	if got != "myhost:9090" {
		t.Errorf("resolveManagerGRPCAddr = %q, want %q", got, "myhost:9090")
	}
}

func TestResolveManagerGRPCAddr_DeriveFromHTTPAddrWithoutScheme(t *testing.T) {
	cfg := config.ControllerConfig{
		Addr: "myhost:8080",
	}
	got := resolveManagerGRPCAddr(cfg)
	if got != "myhost:9090" {
		t.Errorf("resolveManagerGRPCAddr = %q, want %q", got, "myhost:9090")
	}
}

func TestResolveManagerGRPCAddr_EmptyAddr(t *testing.T) {
	cfg := config.ControllerConfig{}
	got := resolveManagerGRPCAddr(cfg)
	if got != "" {
		t.Errorf("resolveManagerGRPCAddr = %q, want empty string", got)
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
