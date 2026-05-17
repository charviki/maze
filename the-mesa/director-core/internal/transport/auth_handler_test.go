package transport

import (
	"context"
	"errors"
	"testing"

	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	"github.com/charviki/maze/fabrication/cradle/auth"
	"github.com/charviki/maze/fabrication/cradle/errutil"
	"github.com/charviki/maze/the-mesa/director-core/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubAuthService struct {
	logoutSubjectKey string
	logoutToken      string
	logoutErr        error
	loginErr         error
	refreshErr       error
}

func (s *stubAuthService) Login(_ context.Context, _, _ string) (*service.LoginResult, error) {
	if s.loginErr != nil {
		return nil, s.loginErr
	}
	return &service.LoginResult{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 1}, nil
}

func (s *stubAuthService) Refresh(_ context.Context, _ string) (*service.LoginResult, error) {
	if s.refreshErr != nil {
		return nil, s.refreshErr
	}
	return &service.LoginResult{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 1}, nil
}

func (s *stubAuthService) Logout(_ context.Context, subjectKey, refreshToken string) error {
	s.logoutSubjectKey = subjectKey
	s.logoutToken = refreshToken
	return s.logoutErr
}

func TestAuthHandlerLogoutRequiresAuthenticatedSubject(t *testing.T) {
	handler := NewAuthHandler(&stubAuthService{})
	_, err := handler.Logout(context.Background(), &pb.LogoutRequest{RefreshToken: "refresh-token"})
	if err == nil {
		t.Fatal("缺少主体信息时应返回错误，但返回成功")
	}

	st := status.Convert(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("gRPC code = %s, 期望 %s", st.Code(), codes.Unauthenticated)
	}
	if st.Message() != "unauthorized: missing authenticated subject" {
		t.Fatalf("message = %q, 期望 %q", st.Message(), "unauthorized: missing authenticated subject")
	}
}

func TestAuthHandlerLogoutPassesCurrentSubjectToService(t *testing.T) {
	svc := &stubAuthService{}
	handler := NewAuthHandler(svc)
	ctx := auth.WithUserInfo(context.Background(), &auth.UserInfo{SubjectKey: "user:alice"})

	_, err := handler.Logout(ctx, &pb.LogoutRequest{RefreshToken: "refresh-token"})
	if err != nil {
		t.Fatalf("Logout() 返回错误: %v", err)
	}
	if svc.logoutSubjectKey != "user:alice" {
		t.Fatalf("subjectKey = %q, 期望 %q", svc.logoutSubjectKey, "user:alice")
	}
	if svc.logoutToken != "refresh-token" {
		t.Fatalf("refreshToken = %q, 期望 %q", svc.logoutToken, "refresh-token")
	}
}

func TestAuthHandlerInternalErrorReturnsInternal(t *testing.T) {
	svc := &stubAuthService{loginErr: errors.New("database connection lost")}
	handler := NewAuthHandler(svc)
	_, err := handler.Login(context.Background(), &pb.LoginRequest{Username: "admin", Password: "admin"})
	if err == nil {
		t.Fatal("期望返回错误")
	}
	st := status.Convert(err)
	if st.Code() != codes.Internal {
		t.Fatalf("gRPC code = %s, 期望 %s", st.Code(), codes.Internal)
	}
}

func TestAuthHandlerKnownAuthErrorReturnsUnauthenticated(t *testing.T) {
	svc := &stubAuthService{loginErr: service.ErrInvalidCredentials}
	handler := NewAuthHandler(svc)
	_, err := handler.Login(context.Background(), &pb.LoginRequest{Username: "admin", Password: "admin"})
	if err == nil {
		t.Fatal("期望返回错误")
	}
	st := status.Convert(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("gRPC code = %s, 期望 %s", st.Code(), codes.Unauthenticated)
	}
}

func TestAuthStatusError_ReasonMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantCode   codes.Code
		wantReason pb.ErrorReason
	}{
		{"invalid credentials", service.ErrInvalidCredentials, codes.Unauthenticated, pb.ErrorReason_ERROR_REASON_INVALID_CREDENTIALS},
		{"user disabled", service.ErrUserDisabled, codes.Unauthenticated, pb.ErrorReason_ERROR_REASON_USER_DISABLED},
		{"refresh token not found", service.ErrRefreshTokenNotFound, codes.Unauthenticated, pb.ErrorReason_ERROR_REASON_REFRESH_TOKEN_NOT_FOUND},
		{"refresh token revoked", service.ErrRefreshTokenRevoked, codes.Unauthenticated, pb.ErrorReason_ERROR_REASON_REFRESH_TOKEN_REVOKED},
		{"refresh token expired", service.ErrRefreshTokenExpired, codes.Unauthenticated, pb.ErrorReason_ERROR_REASON_REFRESH_TOKEN_EXPIRED},
		{"internal error", errors.New("db connection lost"), codes.Internal, pb.ErrorReason_ERROR_REASON_UNSPECIFIED},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toAuthStatusError(tt.err)
			st, ok := status.FromError(result)
			if !ok {
				t.Fatalf("expected gRPC status error, got %v", result)
			}
			if st.Code() != tt.wantCode {
				t.Errorf("code = %v, want %v", st.Code(), tt.wantCode)
			}
			gotReason := errutil.ReasonFromError(result)
			if gotReason != tt.wantReason {
				t.Errorf("reason = %v, want %v", gotReason, tt.wantReason)
			}
		})
	}
}
