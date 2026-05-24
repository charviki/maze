package gatewayutil

import (
	"context"
	"errors"
	"strings"

	mazev1 "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	"github.com/charviki/maze/fabrication/cradle/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// HostTokenValidator 定义 Host 专属令牌校验接口。
// 与 service.NodeRegistry.ValidateHostToken 签名一致，Manager 可直接注入而无需适配器。
type HostTokenValidator interface {
	ValidateHostToken(ctx context.Context, name, token string) (exists, matched bool, err error)
}

// UnaryAuthInterceptor 返回 gRPC UnaryServerInterceptor，对非 Agent 注册/心跳路径执行 JWT 校验。
// jwtSecret 为空时拒绝所有请求（配置错误），非空时验证 Bearer token 中的 JWT 签名和有效期，
// 并将 claims.Subject 注入到 context 的 auth.UserInfo 中。
// Agent 注册/心跳路径由 UnaryHostTokenInterceptor 单独处理，此处跳过。
func UnaryAuthInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if isAgentServiceMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		if isAuthServiceMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// 空 secret 视为配置错误，直接拒绝请求
		if jwtSecret == "" {
			return nil, status.Error(codes.Unauthenticated, "server misconfiguration: jwt secret not configured")
		}

		token, err := extractBearerToken(ctx)
		if err != nil {
			return nil, err
		}

		claims, err := auth.ValidateAccessToken(jwtSecret, auth.DefaultIssuer, token)
		if err != nil {
			if errors.Is(err, auth.ErrTokenExpired) {
				return nil, auth.ExpiredTokenError("unauthorized: token expired")
			}
			return nil, auth.InvalidTokenError("unauthorized: invalid authorization header")
		}

		ctx = auth.WithUserInfo(ctx, &auth.UserInfo{SubjectKey: claims.Subject})
		return handler(ctx, req)
	}
}

// UnaryHostTokenInterceptor 返回 gRPC UnaryServerInterceptor，仅拦截 Agent 注册/心跳路径，
// 执行分层令牌校验：优先检查 Host 专属令牌，不匹配再回退到 jwtSecret 对应的 JWT 校验。
func UnaryHostTokenInterceptor(jwtSecret string, registry HostTokenValidator) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if !isAgentServiceMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// 空 secret 视为配置错误，直接拒绝请求
		if jwtSecret == "" {
			return nil, status.Error(codes.Unauthenticated, "server misconfiguration: jwt secret not configured")
		}

		token, err := extractBearerToken(ctx)
		if err != nil {
			return nil, err
		}

		agentName := extractAgentName(req)
		if agentName == "" {
			return nil, status.Error(codes.InvalidArgument, "agent name is required")
		}

		exists, matched, err := registry.ValidateHostToken(ctx, agentName, token)
		if err != nil {
			return nil, status.Error(codes.Unavailable, "service unavailable: host token validation failed")
		}
		if exists {
			if !matched {
				return nil, status.Error(codes.Unauthenticated, "unauthorized: invalid host token")
			}
			return handler(ctx, req)
		}

		// 无预存令牌的 Host 走 JWT 校验
		if _, jwtErr := auth.ValidateAccessToken(jwtSecret, auth.DefaultIssuer, token); jwtErr != nil {
			return nil, status.Error(codes.Unauthenticated, "unauthorized: invalid authorization header")
		}

		return handler(ctx, req)
	}
}

// extractBearerToken 从 gRPC metadata 中提取 Authorization Bearer token。
// grpc-gateway 会自动将 HTTP header 透传到 gRPC metadata。
func extractBearerToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", auth.MissingTokenError("unauthorized: missing authorization header")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", auth.MissingTokenError("unauthorized: missing authorization header")
	}

	token := strings.TrimPrefix(values[0], "Bearer ")
	if token == values[0] {
		return "", auth.InvalidTokenError("unauthorized: invalid authorization scheme")
	}

	return token, nil
}

// isAgentServiceMethod 判断是否是 AgentService 的 RPC 方法（注册/心跳/配置拉取），
// 这些路径由 UnaryHostTokenInterceptor 单独处理。
func isAgentServiceMethod(method string) bool {
	return method == mazev1.AgentService_Register_FullMethodName ||
		method == mazev1.AgentService_Heartbeat_FullMethodName ||
		method == mazev1.AgentService_GetHostConfig_FullMethodName
}

// isAuthServiceMethod 判断是否是 AuthService 的免鉴权方法（Login/Refresh 不需要 JWT）。
// Logout 需要认证但不在此列表中，走正常 JWT 校验。
func isAuthServiceMethod(method string) bool {
	return method == mazev1.AuthService_Login_FullMethodName ||
		method == mazev1.AuthService_Refresh_FullMethodName
}

// extractAgentName 从 gRPC 请求中提取 agent name，
// 支持 RegisterRequest、HeartbeatRequest 和 GetHostConfigRequest。
func extractAgentName(req any) string {
	switch r := req.(type) {
	case *mazev1.RegisterRequest:
		return r.GetName()
	case *mazev1.HeartbeatRequest:
		return r.GetName()
	case *mazev1.GetHostConfigRequest:
		return r.GetName()
	default:
		return ""
	}
}
