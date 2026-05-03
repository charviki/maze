package httputil

import (
	"net/http"
	"strings"
)

// buildOriginSet 将 origin 列表转为小写 map，用于 CORS 和 WebSocket origin 校验。
func buildOriginSet(origins []string) map[string]struct{} {
	set := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		set[strings.ToLower(o)] = struct{}{}
	}
	return set
}

// CORS 返回一个允许所有来源跨域访问的中间件。
// 生产环境应使用 CORSWithOrigins 限制 Access-Control-Allow-Origin 为指定域名。
func CORS() func(http.Handler) http.Handler {
	return CORSWithOrigins(nil)
}

// CORSWithOrigins 返回一个基于允许来源列表的 CORS 中间件。
// allowedOrigins 为 nil 或空时退化为允许所有来源（开发模式兼容），
// 非空时仅允许列表中的 Origin 通过，其余返回 403。
func CORSWithOrigins(allowedOrigins []string) func(http.Handler) http.Handler {
	allowAll := len(allowedOrigins) == 0
	originSet := buildOriginSet(allowedOrigins)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" {
				if _, ok := originSet[strings.ToLower(origin)]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
				} else if r.Method == http.MethodOptions {
					http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
					return
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CheckOrigin 基于 allowedOrigins 列表返回 WebSocket CheckOrigin 函数。
// allowedOrigins 为 nil 或空时始终允许（开发模式兼容），
// 非空时仅允许列表中的 Origin。
func CheckOrigin(allowedOrigins []string) func(*http.Request) bool {
	if len(allowedOrigins) == 0 {
		return func(_ *http.Request) bool { return true }
	}
	originSet := buildOriginSet(allowedOrigins)
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		_, ok := originSet[strings.ToLower(origin)]
		return ok
	}
}
