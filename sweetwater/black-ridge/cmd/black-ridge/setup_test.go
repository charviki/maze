package main

import (
	"testing"

	"github.com/charviki/sweetwater-black-ridge/internal/transport"
)

func TestGrpcListenAddrFor(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want string
	}{
		{name: "default", addr: "", want: ":9090"},
		{name: "host with port", addr: "127.0.0.1:19090", want: ":19090"},
		{name: "scheme-less hostname", addr: "agent.example:29090", want: ":29090"},
		{name: "already listen addr", addr: ":39090", want: ":39090"},
		{name: "raw token", addr: "grpc-socket", want: "grpc-socket"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := transport.GRPCListenAddrFor(tt.addr); got != tt.want {
				t.Fatalf("GRPCListenAddrFor(%q) = %q, want %q", tt.addr, got, tt.want)
			}
		})
	}
}
