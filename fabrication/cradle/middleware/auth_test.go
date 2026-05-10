package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/charviki/maze/fabrication/cradle/auth"
	"google.golang.org/grpc/codes"
)

const testJWTSecret = "test-secret-key-for-middleware"

func TestAuth_EmptySecret(t *testing.T) {
	called := false
	handler := Auth("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if called {
		t.Error("期望 next 不被调用（空 secret 应拒绝），但被调用了")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("状态码 = %d, 期望 %d", rec.Code, http.StatusUnauthorized)
	}
	assertErrorJSON(t, rec, string(auth.ErrorReasonTokenMissing))
}

func TestAuth_ValidJWT(t *testing.T) {
	called := false
	var gotSubject string
	handler := Auth(testJWTSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if user := auth.GetUserInfo(r.Context()); user != nil {
			gotSubject = user.SubjectKey
		}
	}))

	token, err := auth.GenerateAccessToken(testJWTSecret, auth.DefaultIssuer, "user:alice", 15*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("期望 next 被调用（有效 JWT 应放行），但未调用")
	}
	if gotSubject != "user:alice" {
		t.Errorf("subject = %q, 期望 %q", gotSubject, "user:alice")
	}
}

func TestAuth_ExpiredJWT(t *testing.T) {
	handler := Auth(testJWTSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	token, err := auth.GenerateAccessToken(testJWTSecret, auth.DefaultIssuer, "user:alice", -1*time.Second)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("状态码 = %d, 期望 401", rec.Code)
	}
	if rec.Header().Get("X-Token-Expired") != "true" {
		t.Error("期望 X-Token-Expired: true 头")
	}
	assertErrorJSON(t, rec, string(auth.ErrorReasonTokenExpired))
}

func TestAuth_InvalidJWT(t *testing.T) {
	handler := Auth(testJWTSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("状态码 = %d, 期望 401", rec.Code)
	}
	if rec.Header().Get("X-Token-Expired") != "" {
		t.Error("不期望 X-Token-Expired 头（非过期场景）")
	}
	assertErrorJSON(t, rec, string(auth.ErrorReasonTokenInvalid))
}

func TestAuth_NoHeader(t *testing.T) {
	handler := Auth(testJWTSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Errorf("状态码 = %d, 期望 401", rec.Code)
	}

	assertErrorJSON(t, rec, string(auth.ErrorReasonTokenMissing))
}

func TestAuth_QueryParameterToken(t *testing.T) {
	called := false
	var gotSubject string
	handler := Auth(testJWTSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if user := auth.GetUserInfo(r.Context()); user != nil {
			gotSubject = user.SubjectKey
		}
	}))

	token, err := auth.GenerateAccessToken(testJWTSecret, auth.DefaultIssuer, "user:bob", 15*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/test?token="+token, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("期望 next 被调用（query parameter token 应放行），但未调用")
	}
	if gotSubject != "user:bob" {
		t.Errorf("subject = %q, 期望 %q", gotSubject, "user:bob")
	}
}

func TestAuth_HeaderTakesPrecedenceOverQueryParameter(t *testing.T) {
	var gotSubject string
	handler := Auth(testJWTSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user := auth.GetUserInfo(r.Context()); user != nil {
			gotSubject = user.SubjectKey
		}
	}))

	headerToken, _ := auth.GenerateAccessToken(testJWTSecret, auth.DefaultIssuer, "user:header", 15*time.Minute)
	queryToken, _ := auth.GenerateAccessToken(testJWTSecret, auth.DefaultIssuer, "user:query", 15*time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/test?token="+queryToken, nil)
	req.Header.Set("Authorization", "Bearer "+headerToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if gotSubject != "user:header" {
		t.Errorf("subject = %q, 期望 header 优先（user:header）", gotSubject)
	}
}

func TestAuth_WrongSecret(t *testing.T) {
	handler := Auth(testJWTSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	token, err := auth.GenerateAccessToken("wrong-secret", auth.DefaultIssuer, "user:alice", 15*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("状态码 = %d, 期望 401", rec.Code)
	}
	assertErrorJSON(t, rec, string(auth.ErrorReasonTokenInvalid))
}

type errorResp struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Reason  string `json:"reason,omitempty"`
}

func assertErrorJSON(t *testing.T, rec *httptest.ResponseRecorder, wantReason string) {
	t.Helper()

	var resp errorResp
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("响应体不是合法 JSON: %v, body: %s", err, rec.Body.Bytes())
	}
	if resp.Code != int32(codes.Unauthenticated) {
		t.Errorf("code = %d, 期望 %d", resp.Code, codes.Unauthenticated)
	}
	if resp.Message == "" {
		t.Error("message 为空，期望非空错误描述")
	}
	if resp.Reason != wantReason {
		t.Errorf("reason = %q, 期望 %q", resp.Reason, wantReason)
	}
}
