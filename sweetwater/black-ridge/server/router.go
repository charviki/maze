package main

import (
	"context"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/biz/config"
	"github.com/charviki/sweetwater-black-ridge/biz/model"
	bizrouter "github.com/charviki/sweetwater-black-ridge/biz/router"
	"github.com/charviki/sweetwater-black-ridge/biz/service"
)

// register 注册业务路由和 SPA 静态文件服务。
// API 路由通过 grpc-gateway（Hertz NoRoute 转发）处理，
// 非静态文件路径 fallback 到 index.html（支持前端 SPA 路由）。
func register(h *server.Hertz, cfg *config.Config, tmuxService service.TmuxService, logger logutil.Logger, gwmux *gwruntime.ServeMux) *model.TemplateStore {
	templateStore := bizrouter.RegisterWithService(h, cfg, tmuxService, logger, gwmux)

	subFS, err := fs.Sub(staticFiles, "web-dist")
	if err != nil {
		logger.Fatalf("embed fs sub: %v", err)
	}

	// grpc-gateway handler 包装为 Hertz handler，用于 NoRoute 转发
	gatewayHandler := adaptor.HertzHandler(gwmux)

	h.NoRoute(func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Path())

		// /api/ 前缀的请求转发到 grpc-gateway 处理
		if strings.HasPrefix(path, "/api/") {
			gatewayHandler(ctx, c)
			return
		}

		// SPA fallback：先尝试匹配静态文件，不匹配则返回 index.html
		trimmedPath := strings.TrimPrefix(path, "/")
		if trimmedPath == "" {
			trimmedPath = "index.html"
		}

		data, err := fs.ReadFile(subFS, trimmedPath)
		if err != nil {
			data, err = fs.ReadFile(subFS, "index.html")
			if err != nil {
				c.String(http.StatusNotFound, "not found")
				return
			}
			trimmedPath = "index.html"
		}

		contentType := "application/octet-stream"
		switch strings.ToLower(filepath.Ext(trimmedPath)) {
		case ".html":
			contentType = "text/html; charset=utf-8"
		case ".js":
			contentType = "application/javascript; charset=utf-8"
		case ".css":
			contentType = "text/css; charset=utf-8"
		case ".json":
			contentType = "application/json; charset=utf-8"
		case ".svg":
			contentType = "image/svg+xml"
		case ".png":
			contentType = "image/png"
		case ".ico":
			contentType = "image/x-icon"
		}
		c.Data(http.StatusOK, contentType, data)
	})

	return templateStore
}
