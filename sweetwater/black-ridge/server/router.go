package main

import (
	"context"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/sweetwater-black-ridge/biz/config"
	bizrouter "github.com/charviki/sweetwater-black-ridge/biz/router"
	"github.com/charviki/sweetwater-black-ridge/biz/model"
	"github.com/charviki/sweetwater-black-ridge/biz/service"
)

// register 注册业务路由和 SPA 静态文件服务。
// 静态文件通过 go:embed 嵌入，NoRoute 处理 SPA fallback：
// 先尝试匹配静态文件路径，不匹配则返回 index.html（支持前端路由）
func register(h *server.Hertz, cfg *config.Config, tmuxService service.TmuxService, logger logutil.Logger) *model.TemplateStore {
	templateStore := bizrouter.RegisterWithService(h, cfg, tmuxService, logger)

	subFS, err := fs.Sub(staticFiles, "web-dist")
	if err != nil {
		logger.Fatalf("embed fs sub: %v", err)
	}

	h.NoRoute(func(ctx context.Context, c *app.RequestContext) {
		path := strings.TrimPrefix(string(c.Path()), "/")
		if path == "" {
			path = "index.html"
		}

		data, err := fs.ReadFile(subFS, path)
		if err != nil {
			data, err = fs.ReadFile(subFS, "index.html")
			if err != nil {
				c.String(http.StatusNotFound, "not found")
				return
			}
			path = "index.html" // Update path to ensure correct Content-Type for SPA fallback
		}

		contentType := "application/octet-stream"
		switch strings.ToLower(filepath.Ext(path)) {
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
