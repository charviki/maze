package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	"github.com/charviki/maze/fabrication/cradle/errutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMissingTokenError(t *testing.T) {
	err := MissingTokenError("no token")
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("code = %v, want Unauthenticated", st.Code())
	}
	if got := errutil.ReasonFromError(err); got != pb.ErrorReason_ERROR_REASON_TOKEN_MISSING {
		t.Errorf("reason = %v, want TOKEN_MISSING", got)
	}
}

func TestInvalidTokenError(t *testing.T) {
	err := InvalidTokenError("bad token")
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("code = %v, want Unauthenticated", st.Code())
	}
	if got := errutil.ReasonFromError(err); got != pb.ErrorReason_ERROR_REASON_TOKEN_INVALID {
		t.Errorf("reason = %v, want TOKEN_INVALID", got)
	}
}

func TestExpiredTokenError(t *testing.T) {
	err := ExpiredTokenError("token expired")
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("code = %v, want Unauthenticated", st.Code())
	}
	if got := errutil.ReasonFromError(err); got != pb.ErrorReason_ERROR_REASON_TOKEN_EXPIRED {
		t.Errorf("reason = %v, want TOKEN_EXPIRED", got)
	}
}

func TestErrorResponseFromError(t *testing.T) {
	err := ExpiredTokenError("session ended")
	resp := ErrorResponseFromError(err)
	if resp.Code != int32(codes.Unauthenticated) {
		t.Errorf("code = %d, want %d", resp.Code, codes.Unauthenticated)
	}
	if resp.Message != "session ended" {
		t.Errorf("message = %q, want %q", resp.Message, "session ended")
	}
	if resp.Reason != "TOKEN_EXPIRED" {
		t.Errorf("reason = %q, want %q", resp.Reason, "TOKEN_EXPIRED")
	}
}

func TestErrorReasonFromError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ErrorReason
	}{
		{"missing", MissingTokenError("x"), ErrorReasonTokenMissing},
		{"invalid", InvalidTokenError("x"), ErrorReasonTokenInvalid},
		{"expired", ExpiredTokenError("x"), ErrorReasonTokenExpired},
		{"other", status.Error(codes.Internal, "boom"), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrorReasonFromError(tt.err); got != tt.want {
				t.Errorf("reason = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteHTTPError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteHTTPError(rec, http.StatusUnauthorized, ExpiredTokenError("token expired"))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json; charset=utf-8" {
		t.Errorf("content-type = %q, want %q", ct, "application/json; charset=utf-8")
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if resp.Code != int32(codes.Unauthenticated) {
		t.Errorf("code = %d, want %d", resp.Code, codes.Unauthenticated)
	}
	if resp.Message != "token expired" {
		t.Errorf("message = %q, want %q", resp.Message, "token expired")
	}
	if resp.Reason != "TOKEN_EXPIRED" {
		t.Errorf("reason = %q, want %q", resp.Reason, "TOKEN_EXPIRED")
	}
}
