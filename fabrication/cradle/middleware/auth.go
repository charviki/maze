package middleware

import (
	"context"
	"strings"

	"github.com/charviki/maze-cradle/httputil"
	"github.com/cloudwego/hertz/pkg/app"
)

// Auth 返回 Bearer Token 鉴权中间件。
// token 为空时跳过鉴权（开发模式放行所有请求）；
// token 非空时要求请求携带 Authorization: Bearer <token>，
// 校验失败返回结构化 JSON 401 响应而非空 body。
func Auth(token string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if token == "" {
			c.Next(ctx)
			return
		}
		auth := string(c.GetHeader("Authorization"))
		bearer := strings.TrimPrefix(auth, "Bearer ")
		if bearer != token {
			httputil.Error(c, 401, "unauthorized: invalid or missing authorization header")
			c.Abort()
			return
		}
		c.Next(ctx)
	}
}
