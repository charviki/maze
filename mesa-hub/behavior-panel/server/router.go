package main

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/mesa-hub-behavior-panel/biz/config"
	bizrouter "github.com/charviki/mesa-hub-behavior-panel/biz/router"
)

// 注册业务路由、健康检查端点和 NoRoute 处理（重定向到 /）
func register(h *server.Hertz, cfg *config.Config, logger logutil.Logger) *bizrouter.CleanupResources {
	resources := bizrouter.Register(h, cfg, logger)

	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	h.NoRoute(func(ctx context.Context, c *app.RequestContext) {
		c.Redirect(http.StatusFound, []byte("/"))
	})

	return resources
}
