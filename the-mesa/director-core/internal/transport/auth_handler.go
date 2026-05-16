package transport

import (
	"context"
	"errors"

	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	"github.com/charviki/maze/fabrication/cradle/auth"
	"github.com/charviki/maze/the-mesa/director-core/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// AuthService 定义 transport 层依赖的认证业务能力。
type AuthService interface {
	Login(ctx context.Context, username, password string) (*service.LoginResult, error)
	Refresh(ctx context.Context, refreshToken string) (*service.LoginResult, error)
	Logout(ctx context.Context, subjectKey, refreshToken string) error
}

// AuthHandler 实现 AuthService gRPC 接口。
type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	service AuthService
}

// NewAuthHandler 创建 AuthHandler。
func NewAuthHandler(svc AuthService) *AuthHandler {
	return &AuthHandler{service: svc}
}

// Login 处理用户登录请求。
func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	result, err := h.service.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, toAuthStatusError(err)
	}

	return &pb.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	}, nil
}

// Refresh 处理令牌刷新请求。
func (h *AuthHandler) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	result, err := h.service.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, toAuthStatusError(err)
	}

	return &pb.RefreshResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	}, nil
}

// Logout 处理用户登出请求。
func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*emptypb.Empty, error) {
	userInfo := auth.GetUserInfo(ctx)
	if userInfo == nil || userInfo.SubjectKey == "" {
		// Logout 应由认证层保证已有主体；这里额外兜底，防止绕过拦截器时退化为匿名撤销。
		return nil, status.Error(codes.Unauthenticated, "unauthorized: missing authenticated subject")
	}

	if err := h.service.Logout(ctx, userInfo.SubjectKey, req.GetRefreshToken()); err != nil {
		return nil, toAuthStatusError(err)
	}

	return &emptypb.Empty{}, nil
}

func toAuthStatusError(err error) error {
	if err == nil {
		return nil
	}
	if st, ok := status.FromError(err); ok {
		return st.Err()
	}
	switch {
	case errors.Is(err, service.ErrInvalidCredentials),
		errors.Is(err, service.ErrUserDisabled),
		errors.Is(err, service.ErrRefreshTokenNotFound),
		errors.Is(err, service.ErrRefreshTokenRevoked),
		errors.Is(err, service.ErrRefreshTokenExpired):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

var _ pb.AuthServiceServer = (*AuthHandler)(nil)
