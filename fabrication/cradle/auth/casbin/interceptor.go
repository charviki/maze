package casbin

import (
	"context"
	"fmt"

	"github.com/charviki/maze-cradle/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewUnaryInterceptor 返回 Casbin 权限检查的 gRPC UnaryServerInterceptor。
//
// 插入位置: 在 UnaryAuthInterceptor 和 UnaryHostTokenInterceptor 之后，
// UnaryAuditInterceptor 之前。
//
// 首批策略要求所有受保护方法都显式注册映射，避免缺失映射时静默放行。
func NewUnaryInterceptor(
	enforcer *Enforcer,
	extractUser auth.UserInfoExtractor,
	resourceMap map[string]ResourceAction,
) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ra, ok := resourceMap[info.FullMethod]
		if !ok {
			// Agent 注册/心跳只做认证，不进入授权模型。
			if info.FullMethod == "/maze.v1.AgentService/Register" || info.FullMethod == "/maze.v1.AgentService/Heartbeat" {
				return handler(ctx, req)
			}
			return nil, status.Errorf(codes.Internal, "authz resource mapping missing for %s", info.FullMethod)
		}

		user, err := extractUser(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "missing user info")
		}
		if user == nil || user.SubjectKey == "" {
			return nil, status.Error(codes.Unauthenticated, "missing subject key")
		}

		ctx = auth.WithUserInfo(ctx, user)
		allowed, err := enforcer.Enforce(user.SubjectKey, ra.Resource, ra.Action)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "permission check failed: %v", err)
		}
		if !allowed {
			return nil, status.Errorf(codes.PermissionDenied, "access denied for %s on %s", ra.Action, ra.Resource)
		}

		return handler(ctx, req)
	}
}

// ResourceActionForMethod 读取 gRPC 方法对应的资源动作。
func ResourceActionForMethod(resourceMap map[string]ResourceAction, fullMethod string) (ResourceAction, error) {
	ra, ok := resourceMap[fullMethod]
	if !ok {
		return ResourceAction{}, fmt.Errorf("resource mapping missing for %s", fullMethod)
	}
	return ra, nil
}
