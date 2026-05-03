package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSuccess 验证 Success 返回 200 状态码和正确的 JSON 响应体
func TestSuccess(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	Success(rec, req, map[string]string{"key": "value"})

	if rec.Code != http.StatusOK {
		t.Errorf("状态码 = %d, 期望 200", rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("status = %v, 期望 %q", body["status"], "ok")
	}

	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatal("data 字段类型不是 map[string]interface{}")
	}
	if data["key"] != "value" {
		t.Errorf("data.key = %v, 期望 %q", data["key"], "value")
	}
}

// TestError 验证 Error 返回指定状态码和正确的错误 JSON 响应体
func TestError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	Error(rec, req, 404, "not found")

	if rec.Code != 404 {
		t.Errorf("状态码 = %d, 期望 404", rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}

	if body["status"] != "error" {
		t.Errorf("status = %v, 期望 %q", body["status"], "error")
	}
	if body["message"] != "not found" {
		t.Errorf("message = %v, 期望 %q", body["message"], "not found")
	}
}
