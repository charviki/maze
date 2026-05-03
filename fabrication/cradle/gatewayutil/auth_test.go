package gatewayutil

import (
	"context"
	"testing"

	mazev1 "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// mockHostTokenValidator 是 HostTokenValidator 的测试替身
type mockHostTokenValidator struct {
	exists  bool
	matched bool
}

func (m *mockHostTokenValidator) ValidateHostToken(name, token string) (exists, matched bool) {
	return m.exists, m.matched
}

// callUnaryAuthInterceptor 封装 UnaryAuthInterceptor 的调用逻辑，减少测试样板代码
func callUnaryAuthInterceptor(globalToken, authHeader, method string) (interface{}, error) {
	interceptor := UnaryAuthInterceptor(globalToken)
	ctx := context.Background()
	if authHeader != "" {
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", authHeader))
	}
	return interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: method}, successHandler)
}

// successHandler 是一个模拟的 downstream handler，总是返回 "ok"
func successHandler(_ context.Context, _ interface{}) (interface{}, error) {
	return "ok", nil
}

func TestUnaryAuthInterceptor_EmptyToken(t *testing.T) {
	// 全局令牌为空 → 开发模式，所有请求放行
	resp, err := callUnaryAuthInterceptor("", "Bearer wrong", "/maze.v1.NodeService/ListNode")
	if err != nil {
		t.Fatalf("空 token 模式应放行，但返回错误: %v", err)
	}
	if resp != "ok" {
		t.Errorf("resp = %v, 期望 %q", resp, "ok")
	}
}

func TestUnaryAuthInterceptor_ValidToken(t *testing.T) {
	resp, err := callUnaryAuthInterceptor("secret", "Bearer secret", "/maze.v1.NodeService/ListNode")
	if err != nil {
		t.Fatalf("正确 token 应放行，但返回错误: %v", err)
	}
	if resp != "ok" {
		t.Errorf("resp = %v, 期望 %q", resp, "ok")
	}
}

func TestUnaryAuthInterceptor_InvalidToken(t *testing.T) {
	_, err := callUnaryAuthInterceptor("secret", "Bearer wrong", "/maze.v1.NodeService/ListNode")
	if err == nil {
		t.Fatal("错误 token 应被拒绝，但放行了")
	}
	assertGRPCStatus(t, err, codes.Unauthenticated, "unauthorized: invalid or missing authorization header")
}

func TestUnaryAuthInterceptor_MissingHeader(t *testing.T) {
	// 传入空 authHeader 不设置 metadata → extractBearerToken 先遇到 missing metadata 错误
	_, err := callUnaryAuthInterceptor("secret", "", "/maze.v1.NodeService/ListNode")
	if err == nil {
		t.Fatal("无 Authorization header 应被拒绝，但放行了")
	}
	assertGRPCStatus(t, err, codes.Unauthenticated, "unauthorized: missing metadata")
}

func TestUnaryAuthInterceptor_InvalidScheme(t *testing.T) {
	// 非Bearer格式（如 Basic xxx）应被拒绝
	_, err := callUnaryAuthInterceptor("secret", "Basic abc123", "/maze.v1.NodeService/ListNode")
	if err == nil {
		t.Fatal("非 Bearer scheme 应被拒绝，但放行了")
	}
	assertGRPCStatus(t, err, codes.Unauthenticated, "unauthorized: invalid authorization scheme")
}

func TestUnaryAuthInterceptor_AgentServiceSkipped(t *testing.T) {
	// AgentService/Register 和 Heartbeat 路径应由 HostTokenInterceptor 处理，AuthInterceptor 直接跳过
	// 即使 token 不匹配也应放行
	tests := []struct {
		name   string
		method string
	}{
		{"Register 跳过", "/maze.v1.AgentService/Register"},
		{"Heartbeat 跳过", "/maze.v1.AgentService/Heartbeat"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := callUnaryAuthInterceptor("secret", "Bearer wrong", tt.method)
			if err != nil {
				t.Fatalf("AgentService 路径应跳过 auth，但返回错误: %v", err)
			}
			if resp != "ok" {
				t.Errorf("resp = %v, 期望 %q", resp, "ok")
			}
		})
	}
}

func TestUnaryHostTokenInterceptor_HostTokenMatched(t *testing.T) {
	// Host 有预存令牌且匹配 → 通过
	validator := &mockHostTokenValidator{exists: true, matched: true}
	interceptor := UnaryHostTokenInterceptor("global-secret", validator)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer host-token"))
	req := &mazev1.RegisterRequest{Name: "agent-1"}

	resp, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/maze.v1.AgentService/Register"}, successHandler)
	if err != nil {
		t.Fatalf("Host 令牌匹配应放行，但返回错误: %v", err)
	}
	if resp != "ok" {
		t.Errorf("resp = %v, 期望 %q", resp, "ok")
	}
}

func TestUnaryHostTokenInterceptor_HostTokenMismatched(t *testing.T) {
	// Host 有预存令牌但不匹配 → 拒绝
	validator := &mockHostTokenValidator{exists: true, matched: false}
	interceptor := UnaryHostTokenInterceptor("global-secret", validator)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer wrong-token"))
	req := &mazev1.RegisterRequest{Name: "agent-1"}

	_, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/maze.v1.AgentService/Register"}, successHandler)
	if err == nil {
		t.Fatal("Host 令牌不匹配应被拒绝，但放行了")
	}
	assertGRPCStatus(t, err, codes.Unauthenticated, "unauthorized: invalid host token")
}

func TestUnaryHostTokenInterceptor_FallbackGlobal(t *testing.T) {
	// Host 无预存令牌 → 回退到全局令牌校验
	validator := &mockHostTokenValidator{exists: false, matched: false}
	interceptor := UnaryHostTokenInterceptor("global-secret", validator)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer global-secret"))
	req := &mazev1.HeartbeatRequest{Name: "agent-1"}

	resp, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/maze.v1.AgentService/Heartbeat"}, successHandler)
	if err != nil {
		t.Fatalf("全局令牌匹配应放行，但返回错误: %v", err)
	}
	if resp != "ok" {
		t.Errorf("resp = %v, 期望 %q", resp, "ok")
	}
}

func TestUnaryHostTokenInterceptor_FallbackGlobalFailed(t *testing.T) {
	// Host 无预存令牌，全局令牌也不匹配 → 拒绝
	validator := &mockHostTokenValidator{exists: false, matched: false}
	interceptor := UnaryHostTokenInterceptor("global-secret", validator)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer wrong"))
	req := &mazev1.RegisterRequest{Name: "agent-1"}

	_, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/maze.v1.AgentService/Register"}, successHandler)
	if err == nil {
		t.Fatal("全局令牌不匹配应被拒绝，但放行了")
	}
	assertGRPCStatus(t, err, codes.Unauthenticated, "unauthorized: invalid authorization header")
}

func TestUnaryHostTokenInterceptor_NonAgentMethod(t *testing.T) {
	// 非 AgentService 路径应直接放行，不走 Host 令牌校验
	validator := &mockHostTokenValidator{exists: false, matched: false}
	interceptor := UnaryHostTokenInterceptor("global-secret", validator)

	resp, err := interceptor(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/maze.v1.NodeService/ListNode"},
		successHandler,
	)
	if err != nil {
		t.Fatalf("非 AgentService 路径应放行，但返回错误: %v", err)
	}
	if resp != "ok" {
		t.Errorf("resp = %v, 期望 %q", resp, "ok")
	}
}

func TestUnaryHostTokenInterceptor_EmptyAgentName(t *testing.T) {
	// AgentService 路径但 agent name 为空 → InvalidArgument 错误
	validator := &mockHostTokenValidator{exists: false, matched: false}
	interceptor := UnaryHostTokenInterceptor("global-secret", validator)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer some-token"))
	// RegisterRequest 的 Name 字段为零值
	req := &mazev1.RegisterRequest{}

	_, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/maze.v1.AgentService/Register"}, successHandler)
	if err == nil {
		t.Fatal("agent name 为空应被拒绝，但放行了")
	}
	assertGRPCStatus(t, err, codes.InvalidArgument, "agent name is required")
}

func TestUnaryHostTokenInterceptor_EmptyGlobalToken(t *testing.T) {
	// 全局令牌为空 → 开发模式，所有请求放行
	validator := &mockHostTokenValidator{exists: false, matched: false}
	interceptor := UnaryHostTokenInterceptor("", validator)

	resp, err := interceptor(
		context.Background(),
		&mazev1.RegisterRequest{Name: "agent-1"},
		&grpc.UnaryServerInfo{FullMethod: "/maze.v1.AgentService/Register"},
		successHandler,
	)
	if err != nil {
		t.Fatalf("空全局令牌模式应放行，但返回错误: %v", err)
	}
	if resp != "ok" {
		t.Errorf("resp = %v, 期望 %q", resp, "ok")
	}
}

// assertGRPCStatus 校验错误是否为预期的 gRPC status code 和 message
func assertGRPCStatus(t *testing.T, err error, wantCode codes.Code, wantMsg string) {
	t.Helper()
	s, ok := status.FromError(err)
	if !ok {
		t.Fatalf("错误不是 gRPC status: %v", err)
	}
	if s.Code() != wantCode {
		t.Errorf("gRPC code = %v, 期望 %v", s.Code(), wantCode)
	}
	if s.Message() != wantMsg {
		t.Errorf("message = %q, 期望 %q", s.Message(), wantMsg)
	}
}
