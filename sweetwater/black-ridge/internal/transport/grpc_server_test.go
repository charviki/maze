package transport

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/charviki/sweetwater-black-ridge/internal/service"
)

func TestErrToStatus_Nil(t *testing.T) {
	if err := errToStatus(nil); err != nil {
		t.Fatalf("nil 输入应返回 nil, 实际: %v", err)
	}
}

func TestErrToStatus_SessionNotFound(t *testing.T) {
	err := errToStatus(service.ErrSessionNotFound)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, 期望 NotFound", st.Code())
	}
}

func TestErrToStatus_ConfigConflictError(t *testing.T) {
	conflicts := []service.ConfigConflict{
		{Path: "test.txt", CurrentHash: "md5:abc"},
	}
	confErr := &service.ConfigConflictError{Conflicts: conflicts}
	err := errToStatus(confErr)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("code = %v, 期望 FailedPrecondition", st.Code())
	}
	msg := st.Message()
	if msg == "" {
		t.Error("消息不应为空")
	}
}

func TestErrToStatus_GenericError(t *testing.T) {
	err := errToStatus(errors.New("some internal error"))
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.Internal {
		t.Errorf("code = %v, 期望 Internal", st.Code())
	}
}

func TestGRPCListenAddrFor(t *testing.T) {
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
			if got := GRPCListenAddrFor(tt.addr); got != tt.want {
				t.Fatalf("GRPCListenAddrFor(%q) = %q, want %q", tt.addr, got, tt.want)
			}
		})
	}
}

func TestErrToStatus_ConfigConflict(t *testing.T) {
	conflicts := []service.ConfigConflict{
		{Path: "CLAUDE.md", CurrentHash: "md5:xyz"},
	}
	confErr := &service.ConfigConflictError{Conflicts: conflicts}
	err := errToStatus(confErr)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("code = %v, 期望 FailedPrecondition", st.Code())
	}
}
