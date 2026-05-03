package gatewayutil

import (
	"context"
	"strings"

	mazev1 "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// HostTokenValidator 定义 Host 专属令牌校验接口。
// 与 model.NodeRegistry.ValidateHostToken 签名一致，Manager 可直接注入而无需适配器。
type HostTokenValidator interface {
	ValidateHostToken(name, token string) (exists, matched bool)
}

// UnaryAuthInterceptor 返回 gRPC UnaryServerInterceptor，对非 Agent 注册/心跳路径执行全局 Bearer Token 校验。
// globalToken 为空时放行所有请求（开发模式），非空时要求请求携带匹配的 Bearer token。
// Agent 注册/心跳路径由 UnaryHostTokenInterceptor 单独处理，此处跳过。
func UnaryAuthInterceptor(globalToken string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Agent 注册/心跳路径跳过，由分层令牌 interceptor 处理
		if isAgentServiceMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// 开发模式：全局令牌为空时放行所有请求
		if globalToken == "" {
			return handler(ctx, req)
		}

		token, err := extractBearerToken(ctx)
		if err != nil {
			return nil, err
		}
		if token != globalToken {
			return nil, status.Error(codes.Unauthenticated, "unauthorized: invalid or missing authorization header")
		}

		return handler(ctx, req)
	}
}

// UnaryHostTokenInterceptor 返回 gRPC UnaryServerInterceptor，仅拦截 Agent 注册/心跳路径，
// 执行分层令牌校验：优先检查 Host 专属令牌，不匹配再回退到全局令牌。
func UnaryHostTokenInterceptor(globalToken string, registry HostTokenValidator) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 仅拦截 Agent 注册/心跳路径，其他路径直接放行
		if !isAgentServiceMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// 开发模式：全局令牌为空时放行
		if globalToken == "" {
			return handler(ctx, req)
		}

		token, err := extractBearerToken(ctx)
		if err != nil {
			return nil, err
		}

		// 从请求中提取 agent name（Register 和 Heartbeat 请求都有 Name 字段）
		agentName := extractAgentName(req)
		if agentName == "" {
			return nil, status.Error(codes.InvalidArgument, "agent name is required")
		}

		// 分层校验：先检查 Host 专属令牌，再回退全局令牌
		exists, matched := registry.ValidateHostToken(agentName, token)
		if exists {
			// Host 有预存令牌，必须精确匹配
			if !matched {
				return nil, status.Error(codes.Unauthenticated, "unauthorized: invalid host token")
			}
			return handler(ctx, req)
		}

		// 无预存令牌的 Host 走全局 auth 校验
		if token != globalToken {
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
		return "", status.Error(codes.Unauthenticated, "unauthorized: missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "unauthorized: missing authorization header")
	}

	token := strings.TrimPrefix(values[0], "Bearer ")
	// TrimPrefix 不匹配时原样返回，说明不是 Bearer 格式
	if token == values[0] {
		return "", status.Error(codes.Unauthenticated, "unauthorized: invalid authorization scheme")
	}

	return token, nil
}

// isAgentServiceMethod 判断是否是 AgentService 的 RPC 方法（注册/心跳），
// 这些路径由 UnaryHostTokenInterceptor 单独处理。
func isAgentServiceMethod(method string) bool {
	return method == "/maze.v1.AgentService/Register" ||
		method == "/maze.v1.AgentService/Heartbeat"
}

// extractAgentName 从 gRPC 请求中提取 agent name，
// 支持 RegisterRequest 和 HeartbeatRequest 两种类型。
func extractAgentName(req interface{}) string {
	switch r := req.(type) {
	case *mazev1.RegisterRequest:
		return r.GetName()
	case *mazev1.HeartbeatRequest:
		return r.GetName()
	default:
		return ""
	}
}
