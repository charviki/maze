// Package httputil 提供 Hertz HTTP 工具函数，包括统一 JSON 响应封装和 CORS 中间件。
// gatewayutil 包提供了 grpc-gateway 等价的响应格式包装功能，输出格式与本包的 Success/Error 一致。
package httputil

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// Success 返回 HTTP 200 + JSON 成功响应 {status: ok, data: ...}
func Success(c *app.RequestContext, data interface{}) {
	c.JSON(consts.StatusOK, utils.H{
		"status": "ok",
		"data":   data,
	})
}

// Error 返回指定状态码 + JSON 错误响应 {status: error, message: ...}
func Error(c *app.RequestContext, code int, msg string) {
	c.JSON(code, utils.H{
		"status":  "error",
		"message": msg,
	})
}
