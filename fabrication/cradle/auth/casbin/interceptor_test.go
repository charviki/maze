package casbin

import (
	"context"
	"testing"

	mazev1 "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	"github.com/charviki/maze/fabrication/cradle/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewUnaryInterceptor_SessionMethodsSkipCasbin(t *testing.T) {
	enforcer, err := NewEnforcer(nil)
	if err != nil {
		t.Fatalf("NewEnforcer() 返回错误: %v", err)
	}

	interceptor := NewUnaryInterceptor(
		enforcer,
		func(_ context.Context) (*auth.UserInfo, error) {
			t.Fatal("session 生命周期接口不应进入 Casbin 用户提取")
			return nil, nil
		},
		map[string]ResourceAction{},
	)

	tests := []struct {
		name   string
		method string
	}{
		{name: "Login 跳过资源授权", method: mazev1.AuthService_Login_FullMethodName},
		{name: "Refresh 跳过资源授权", method: mazev1.AuthService_Refresh_FullMethodName},
		{name: "Logout 跳过资源授权", method: mazev1.AuthService_Logout_FullMethodName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: tt.method}, func(ctx context.Context, req any) (any, error) {
				called = true
				return "ok", nil
			})
			if err != nil {
				t.Fatalf("interceptor 返回错误: %v", err)
			}
			if !called {
				t.Fatal("handler 未被调用")
			}
			if resp != "ok" {
				t.Fatalf("resp = %v, 期望 ok", resp)
			}
		})
	}
}

func TestNewUnaryInterceptor_ProtectedMethodRequiresResourceMapping(t *testing.T) {
	enforcer, err := NewEnforcer(nil)
	if err != nil {
		t.Fatalf("NewEnforcer() 返回错误: %v", err)
	}

	interceptor := NewUnaryInterceptor(
		enforcer,
		func(_ context.Context) (*auth.UserInfo, error) {
			return &auth.UserInfo{SubjectKey: "user:alice"}, nil
		},
		map[string]ResourceAction{},
	)

	_, err = interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/maze.v1.NodeService/ListNodes"}, func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})
	if err == nil {
		t.Fatal("缺少资源映射时应返回错误，但返回成功")
	}

	st := status.Convert(err)
	if st.Code() != codes.Internal {
		t.Fatalf("gRPC code = %s, 期望 %s", st.Code(), codes.Internal)
	}
}
