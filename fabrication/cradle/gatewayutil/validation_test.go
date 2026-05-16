package gatewayutil

import (
	"context"
	"testing"

	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestValidationInterceptor_RejectsInvalidRequest(t *testing.T) {
	interceptor, err := NewValidationInterceptor()
	if err != nil {
		t.Fatalf("NewValidationInterceptor: %v", err)
	}

	// CreateSessionRequest.name has min_len: 1
	req := &pb.CreateSessionRequest{Name: ""}
	_, err = interceptor(
		context.Background(),
		req,
		&grpc.UnaryServerInfo{},
		func(ctx context.Context, req any) (any, error) { return nil, nil },
	)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if st, _ := status.FromError(err); st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestValidationInterceptor_AllowsValidRequest(t *testing.T) {
	interceptor, err := NewValidationInterceptor()
	if err != nil {
		t.Fatalf("NewValidationInterceptor: %v", err)
	}

	req := &pb.CreateSessionRequest{Name: "session-1"}
	called := false
	_, err = interceptor(
		context.Background(),
		req,
		&grpc.UnaryServerInfo{},
		func(ctx context.Context, req any) (any, error) {
			called = true
			return nil, nil
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestValidationInterceptor_PassesNonProtoMessage(t *testing.T) {
	interceptor, err := NewValidationInterceptor()
	if err != nil {
		t.Fatalf("NewValidationInterceptor: %v", err)
	}

	called := false
	_, err = interceptor(
		context.Background(),
		"not a proto message",
		&grpc.UnaryServerInfo{},
		func(ctx context.Context, req any) (any, error) {
			called = true
			return nil, nil
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("handler was not called for non-proto message")
	}
}
