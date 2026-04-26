package httputil

import (
	"context"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

// CORS 返回一个允许所有来源跨域访问的中间件。
// 生产环境应使用 CORSWithOrigins 限制 Access-Control-Allow-Origin 为指定域名。
func CORS() app.HandlerFunc {
	return CORSWithOrigins(nil)
}

// CORSWithOrigins 返回一个基于允许来源列表的 CORS 中间件。
// allowedOrigins 为 nil 或空时退化为允许所有来源（开发模式兼容），
// 非空时仅允许列表中的 Origin 通过，其余返回 403。
func CORSWithOrigins(allowedOrigins []string) app.HandlerFunc {
	allowAll := len(allowedOrigins) == 0
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[strings.ToLower(o)] = struct{}{}
	}

	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.GetHeader("Origin"))
		if allowAll {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin != "" {
			if _, ok := originSet[strings.ToLower(origin)]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			} else if string(c.Method()) == string(http.MethodOptions) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if string(c.Method()) == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next(ctx)
	}
}

// CheckOrigin 基于 allowedOrigins 列表返回 WebSocket CheckOrigin 函数。
// allowedOrigins 为 nil 或空时始终允许（开发模式兼容），
// 非空时仅允许列表中的 Origin。
func CheckOrigin(allowedOrigins []string) func(c *app.RequestContext) bool {
	if len(allowedOrigins) == 0 {
		return func(_ *app.RequestContext) bool { return true }
	}
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[strings.ToLower(o)] = struct{}{}
	}
	return func(c *app.RequestContext) bool {
		origin := string(c.GetHeader("Origin"))
		if origin == "" {
			return true
		}
		_, ok := originSet[strings.ToLower(origin)]
		return ok
	}
}
