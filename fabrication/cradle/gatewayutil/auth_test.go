package gatewayutil

import (
	"context"
	"errors"
	"testing"
	"time"

	mazev1 "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const testGRPJWTSecret = "test-grpc-jwt-secret"

type mockHostTokenValidator struct {
	exists  bool
	matched bool
	err     error
}

func (m *mockHostTokenValidator) ValidateHostToken(_ context.Context, name, token string) (exists, matched bool, err error) {
	return m.exists, m.matched, m.err
}

func callUnaryAuthInterceptor(jwtSecret, authHeader, method string) (any, error) {
	interceptor := UnaryAuthInterceptor(jwtSecret)
	ctx := context.Background()
	if authHeader != "" {
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", authHeader))
	}
	return interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: method}, successHandler)
}

func successHandler(_ context.Context, _ any) (any, error) {
	return "ok", nil
}

func generateTestToken(secret, subject string, ttl time.Duration) string {
	token, err := auth.GenerateAccessToken(secret, auth.DefaultIssuer, subject, ttl)
	if err != nil {
		panic(err)
	}
	return token
}

func TestUnaryAuthInterceptor_EmptySecret(t *testing.T) {
	_, err := callUnaryAuthInterceptor("", "Bearer anything", "/maze.v1.NodeService/ListNode")
	if err == nil {
		t.Fatal("空 secret 应拒绝请求，但放行了")
	}
	assertGRPCStatus(t, err, codes.Unauthenticated, "server misconfiguration: jwt secret not configured")
}

func TestUnaryAuthInterceptor_ValidJWT(t *testing.T) {
	token := generateTestToken(testGRPJWTSecret, "user:alice", 15*time.Minute)
	resp, err := callUnaryAuthInterceptor(testGRPJWTSecret, "Bearer "+token, "/maze.v1.NodeService/ListNode")
	if err != nil {
		t.Fatalf("有效 JWT 应放行，但返回错误: %v", err)
	}
	if resp != "ok" {
		t.Errorf("resp = %v, 期望 %q", resp, "ok")
	}
}

func TestUnaryAuthInterceptor_ExpiredJWT(t *testing.T) {
	token := generateTestToken(testGRPJWTSecret, "user:alice", -1*time.Second)
	_, err := callUnaryAuthInterceptor(testGRPJWTSecret, "Bearer "+token, "/maze.v1.NodeService/ListNode")
	if err == nil {
		t.Fatal("过期 JWT 应被拒绝，但放行了")
	}
	assertGRPCStatus(t, err, codes.Unauthenticated, "unauthorized: token expired")
	assertErrorReason(t, err, auth.ErrorReasonTokenExpired)
}

func TestUnaryAuthInterceptor_InvalidJWT(t *testing.T) {
	_, err := callUnaryAuthInterceptor(testGRPJWTSecret, "Bearer invalid.jwt.token", "/maze.v1.NodeService/ListNode")
	if err == nil {
		t.Fatal("无效 JWT 应被拒绝，但放行了")
	}
	assertGRPCStatus(t, err, codes.Unauthenticated, "unauthorized: invalid authorization header")
	assertErrorReason(t, err, auth.ErrorReasonTokenInvalid)
}

func TestUnaryAuthInterceptor_WrongSecret(t *testing.T) {
	token := generateTestToken("wrong-secret", "user:alice", 15*time.Minute)
	_, err := callUnaryAuthInterceptor(testGRPJWTSecret, "Bearer "+token, "/maze.v1.NodeService/ListNode")
	if err == nil {
		t.Fatal("错误密钥签名的 JWT 应被拒绝，但放行了")
	}
	assertGRPCStatus(t, err, codes.Unauthenticated, "unauthorized: invalid authorization header")
	assertErrorReason(t, err, auth.ErrorReasonTokenInvalid)
}

func TestUnaryAuthInterceptor_MissingHeader(t *testing.T) {
	_, err := callUnaryAuthInterceptor(testGRPJWTSecret, "", "/maze.v1.NodeService/ListNode")
	if err == nil {
		t.Fatal("无 Authorization header 应被拒绝，但放行了")
	}
	assertGRPCStatus(t, err, codes.Unauthenticated, "unauthorized: missing authorization header")
	assertErrorReason(t, err, auth.ErrorReasonTokenMissing)
}

func TestUnaryAuthInterceptor_InvalidScheme(t *testing.T) {
	_, err := callUnaryAuthInterceptor(testGRPJWTSecret, "Basic abc123", "/maze.v1.NodeService/ListNode")
	if err == nil {
		t.Fatal("非 Bearer scheme 应被拒绝，但放行了")
	}
	assertGRPCStatus(t, err, codes.Unauthenticated, "unauthorized: invalid authorization scheme")
	assertErrorReason(t, err, auth.ErrorReasonTokenInvalid)
}

func TestUnaryAuthInterceptor_AnonymousAuthMethods(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{name: "Login 匿名放行", method: mazev1.AuthService_Login_FullMethodName},
		{name: "Refresh 匿名放行", method: mazev1.AuthService_Refresh_FullMethodName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := callUnaryAuthInterceptor(testGRPJWTSecret, "", tt.method)
			if err != nil {
				t.Fatalf("匿名 auth 方法应放行，但返回错误: %v", err)
			}
			if resp != "ok" {
				t.Errorf("resp = %v, 期望 %q", resp, "ok")
			}
		})
	}
}

func TestUnaryAuthInterceptor_LogoutRequiresJWT(t *testing.T) {
	_, err := callUnaryAuthInterceptor(testGRPJWTSecret, "", mazev1.AuthService_Logout_FullMethodName)
	if err == nil {
		t.Fatal("Logout 缺少 JWT 应被拒绝，但放行了")
	}
	assertGRPCStatus(t, err, codes.Unauthenticated, "unauthorized: missing authorization header")
	assertErrorReason(t, err, auth.ErrorReasonTokenMissing)
}

func TestUnaryAuthInterceptor_AgentServiceSkipped(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{"Register 跳过", "/maze.v1.AgentService/Register"},
		{"Heartbeat 跳过", "/maze.v1.AgentService/Heartbeat"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := callUnaryAuthInterceptor(testGRPJWTSecret, "Bearer invalid", tt.method)
			if err != nil {
				t.Fatalf("AgentService 路径应跳过 auth，但返回错误: %v", err)
			}
			if resp != "ok" {
				t.Errorf("resp = %v, 期望 %q", resp, "ok")
			}
		})
	}
}

func TestUnaryAuthInterceptor_InjectsUserInfo(t *testing.T) {
	token := generateTestToken(testGRPJWTSecret, "user:bob", 15*time.Minute)
	interceptor := UnaryAuthInterceptor(testGRPJWTSecret)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))

	var gotSubject string
	_, _ = interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/maze.v1.NodeService/ListNode"}, func(ctx context.Context, _ any) (any, error) {
		if user := auth.GetUserInfo(ctx); user != nil {
			gotSubject = user.SubjectKey
		}
		return "ok", nil
	})

	if gotSubject != "user:bob" {
		t.Errorf("subject = %q, 期望 %q", gotSubject, "user:bob")
	}
}

func TestUnaryHostTokenInterceptor_HostTokenMatched(t *testing.T) {
	validator := &mockHostTokenValidator{exists: true, matched: true}
	interceptor := UnaryHostTokenInterceptor(testGRPJWTSecret, validator)

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
	validator := &mockHostTokenValidator{exists: true, matched: false}
	interceptor := UnaryHostTokenInterceptor(testGRPJWTSecret, validator)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer wrong-token"))
	req := &mazev1.RegisterRequest{Name: "agent-1"}

	_, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/maze.v1.AgentService/Register"}, successHandler)
	if err == nil {
		t.Fatal("Host 令牌不匹配应被拒绝，但放行了")
	}
	assertGRPCStatus(t, err, codes.Unauthenticated, "unauthorized: invalid host token")
}

func TestUnaryHostTokenInterceptor_FallbackJWT(t *testing.T) {
	validator := &mockHostTokenValidator{exists: false, matched: false}
	interceptor := UnaryHostTokenInterceptor(testGRPJWTSecret, validator)

	token := generateTestToken(testGRPJWTSecret, "user:admin", 15*time.Minute)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))
	req := &mazev1.HeartbeatRequest{Name: "agent-1"}

	resp, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/maze.v1.AgentService/Heartbeat"}, successHandler)
	if err != nil {
		t.Fatalf("JWT 匹配应放行，但返回错误: %v", err)
	}
	if resp != "ok" {
		t.Errorf("resp = %v, 期望 %q", resp, "ok")
	}
}

func TestUnaryHostTokenInterceptor_FallbackJWTFailed(t *testing.T) {
	validator := &mockHostTokenValidator{exists: false, matched: false}
	interceptor := UnaryHostTokenInterceptor(testGRPJWTSecret, validator)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer wrong"))
	req := &mazev1.RegisterRequest{Name: "agent-1"}

	_, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/maze.v1.AgentService/Register"}, successHandler)
	if err == nil {
		t.Fatal("无效 JWT 应被拒绝，但放行了")
	}
	assertGRPCStatus(t, err, codes.Unauthenticated, "unauthorized: invalid authorization header")
}

func TestUnaryHostTokenInterceptor_NonAgentMethod(t *testing.T) {
	validator := &mockHostTokenValidator{exists: false, matched: false}
	interceptor := UnaryHostTokenInterceptor(testGRPJWTSecret, validator)

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
	validator := &mockHostTokenValidator{exists: false, matched: false}
	interceptor := UnaryHostTokenInterceptor(testGRPJWTSecret, validator)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer some-token"))
	req := &mazev1.RegisterRequest{}

	_, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/maze.v1.AgentService/Register"}, successHandler)
	if err == nil {
		t.Fatal("agent name 为空应被拒绝，但放行了")
	}
	assertGRPCStatus(t, err, codes.InvalidArgument, "agent name is required")
}

func TestUnaryHostTokenInterceptor_EmptySecret(t *testing.T) {
	validator := &mockHostTokenValidator{exists: false, matched: false}
	interceptor := UnaryHostTokenInterceptor("", validator)

	_, err := interceptor(
		context.Background(),
		&mazev1.RegisterRequest{Name: "agent-1"},
		&grpc.UnaryServerInfo{FullMethod: "/maze.v1.AgentService/Register"},
		successHandler,
	)
	if err == nil {
		t.Fatal("空 secret 应拒绝请求，但放行了")
	}
	assertGRPCStatus(t, err, codes.Unauthenticated, "server misconfiguration: jwt secret not configured")
}

func TestUnaryHostTokenInterceptor_RegistryError(t *testing.T) {
	validator := &mockHostTokenValidator{exists: false, matched: false, err: errors.New("db connection lost")}
	interceptor := UnaryHostTokenInterceptor(testGRPJWTSecret, validator)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer some-token"))
	req := &mazev1.RegisterRequest{Name: "agent-1"}

	_, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/maze.v1.AgentService/Register"}, successHandler)
	if err == nil {
		t.Fatal("registry 返回 error 时应拒绝请求，但放行了")
	}
	assertGRPCStatus(t, err, codes.Unavailable, "service unavailable: host token validation failed")
}

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

func assertErrorReason(t *testing.T, err error, want auth.ErrorReason) {
	t.Helper()
	if got := auth.ErrorReasonFromError(err); got != want {
		t.Errorf("reason = %q, 期望 %q", got, want)
	}
}
