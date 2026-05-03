package middleware

import (
	"net/http"
	"strings"

	"github.com/charviki/maze-cradle/httputil"
)

// Auth 返回 Bearer Token 鉴权中间件。
// token 为空时跳过鉴权（开发模式放行所有请求）；
// token 非空时要求请求携带 Authorization: Bearer <token>，
// 校验失败返回结构化 JSON 401 响应而非空 body。
func Auth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}
			auth := r.Header.Get("Authorization")
			bearer := strings.TrimPrefix(auth, "Bearer ")
			if bearer != token {
				httputil.Error(w, r, http.StatusUnauthorized, "unauthorized: invalid or missing authorization header")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
