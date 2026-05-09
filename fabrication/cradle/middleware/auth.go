package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/charviki/maze-cradle/auth"
)

// Auth 返回 JWT 鉴权中间件。
// jwtSecret 为空时拒绝所有请求（配置错误，jwt.secret 为必填项）；
// jwtSecret 非空时按以下优先级提取 token：
//   1. Authorization: Bearer <jwt> header
//   2. URL query parameter token=<jwt>（浏览器 WebSocket 不支持自定义 header，需要此 fallback）
//
// 校验失败返回结构化 JSON 401 响应；过期时额外设置 X-Token-Expired: true 头做兼容。
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 空 secret 视为配置错误，直接拒绝请求
			if jwtSecret == "" {
				auth.WriteHTTPError(w, http.StatusUnauthorized, auth.MissingTokenError("unauthorized: server misconfiguration"))
				return
			}

			token := extractToken(r)
			if token == "" {
				auth.WriteHTTPError(w, http.StatusUnauthorized, auth.MissingTokenError("unauthorized: missing authorization header"))
				return
			}

			claims, err := auth.ValidateAccessToken(jwtSecret, auth.DefaultIssuer, token)
			if err != nil {
				if errors.Is(err, auth.ErrTokenExpired) {
					w.Header().Set("X-Token-Expired", "true")
					auth.WriteHTTPError(w, http.StatusUnauthorized, auth.ExpiredTokenError("unauthorized: token expired"))
					return
				}
				auth.WriteHTTPError(w, http.StatusUnauthorized, auth.InvalidTokenError("unauthorized: invalid authorization header"))
				return
			}

			ctx := auth.WithUserInfo(r.Context(), &auth.UserInfo{SubjectKey: claims.Subject})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken 从 Authorization header 或 URL query parameter 中提取 JWT。
// header 优先；header 不存在时 fallback 到 query parameter "token"。
func extractToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header != "" {
		token := strings.TrimPrefix(header, "Bearer ")
		if token != header && token != "" {
			return token
		}
	}
	if t := r.URL.Query().Get("token"); t != "" {
		return t
	}
	return ""
}
