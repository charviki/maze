package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAuth_EmptyToken 当 token 为空字符串时，请求应直接放行（开发模式跳过鉴权）
func TestAuth_EmptyToken(t *testing.T) {
	called := false
	handler := Auth("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("期望 next 被调用（空 token 应放行），但未调用")
	}
}

// TestAuth_CorrectToken 当 token 非空且请求携带正确的 Bearer token 时，应放行
func TestAuth_CorrectToken(t *testing.T) {
	called := false
	handler := Auth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("期望 next 被调用（正确 token 应放行），但未调用")
	}
}

// TestAuth_WrongToken 当 token 非空且请求携带错误的 Bearer token 时，应返回 401 + 结构化 JSON
func TestAuth_WrongToken(t *testing.T) {
	handler := Auth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Errorf("状态码 = %d, 期望 401", rec.Code)
	}

	assertErrorJSON(t, rec)
}

// TestAuth_NoHeader 当 token 非空且请求未携带 Authorization 头时，应返回 401 + 结构化 JSON
func TestAuth_NoHeader(t *testing.T) {
	handler := Auth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Errorf("状态码 = %d, 期望 401", rec.Code)
	}

	assertErrorJSON(t, rec)
}

// errorResp 用于反序列化 401 响应体，验证是否为结构化 JSON
type errorResp struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// assertErrorJSON 校验响应体包含合法的结构化 JSON 错误：status=="error" 且 message 非空
func assertErrorJSON(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()

	var resp errorResp
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("响应体不是合法 JSON: %v, body: %s", err, rec.Body.Bytes())
	}
	if resp.Status != "error" {
		t.Errorf("status = %q, 期望 %q", resp.Status, "error")
	}
	if resp.Message == "" {
		t.Error("message 为空，期望非空错误描述")
	}
}
