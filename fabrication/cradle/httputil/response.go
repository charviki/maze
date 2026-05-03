// Package httputil 提供标准库 HTTP 工具函数，包括统一 JSON 响应封装和 CORS 中间件。
// gatewayutil 包提供了 grpc-gateway 等价的响应格式包装功能，输出格式与本包的 Success/Error 一致。
package httputil

import (
	"encoding/json"
	"net/http"
)

// Success 返回 HTTP 200 + JSON 成功响应 {status: ok, data: ...}
func Success(w http.ResponseWriter, _ *http.Request, data any) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"data":   data,
	})
}

// Error 返回指定状态码 + JSON 错误响应 {status: error, message: ...}
func Error(w http.ResponseWriter, _ *http.Request, code int, msg string) {
	writeJSON(w, code, map[string]any{
		"status":  "error",
		"message": msg,
	})
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
