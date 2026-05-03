package main

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/mesa-hub-behavior-panel/biz/config"
	bizrouter "github.com/charviki/mesa-hub-behavior-panel/biz/router"
)

// register 注册业务路由、健康检查端点和 NoRoute 处理。
// NoRoute 将未匹配的请求转发到 grpc-gateway ServeMux，由其处理 REST API。
func register(h *server.Hertz, cfg *config.Config, logger logutil.Logger, gwmux *runtime.ServeMux) *bizrouter.CleanupResources {
	resources := bizrouter.Register(h, cfg, logger, gwmux)

	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// 将 grpc-gateway ServeMux 包装为 Hertz handler，
	// Hertz 未匹配路由的请求直接转发到 grpc-gateway 处理。
	// adaptor.HertzHandler 负责桥接标准 http.Handler 与 Hertz 的请求/响应类型。
	gatewayHandler := adaptor.HertzHandler(gwmux)
	h.NoRoute(gatewayHandler)

	return resources
}
