package transport

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	"github.com/charviki/maze/fabrication/cradle/errutil"
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
	if gotReason := errutil.ReasonFromError(err); gotReason != pb.ErrorReason_ERROR_REASON_SESSION_NOT_FOUND {
		t.Errorf("reason = %v, want SESSION_NOT_FOUND", gotReason)
	}
}

func TestErrToStatus_ConfigConflictError(t *testing.T) {
	confErr := &service.ConfigConflictError{
		Conflicts: []service.ConfigConflict{
			{Path: "/path/to/file", CurrentHash: "abc123"},
		},
	}
	result := errToStatus(confErr)

	st, ok := status.FromError(result)
	if !ok {
		t.Fatal("应返回 gRPC status")
	}
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("code = %v, 期望 FailedPrecondition", st.Code())
	}
	if st.Message() != "config conflict" {
		t.Errorf("message = %v, want 'config conflict'", st.Message())
	}

	if gotReason := errutil.ReasonFromError(result); gotReason != pb.ErrorReason_ERROR_REASON_CONFIG_CONFLICT {
		t.Errorf("reason = %v, want CONFIG_CONFLICT", gotReason)
	}

	violations := errutil.PreconditionViolationsFromError(result)
	if len(violations) != 1 {
		t.Fatalf("violations count = %d, want 1", len(violations))
	}
	if violations[0].Type != "CONFIG_CONFLICT" {
		t.Errorf("type = %v, want CONFIG_CONFLICT", violations[0].Type)
	}
	if violations[0].Subject != "/path/to/file" {
		t.Errorf("subject = %v, want /path/to/file", violations[0].Subject)
	}
	if violations[0].Description != "abc123" {
		t.Errorf("description = %v, want abc123", violations[0].Description)
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
	if gotReason := errutil.ReasonFromError(err); gotReason != pb.ErrorReason_ERROR_REASON_UNSPECIFIED {
		t.Errorf("reason = %v, want UNSPECIFIED", gotReason)
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
