package router

import (
	"context"
	"net/http"
	"path/filepath"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hertz-contrib/logger/accesslog"

	"github.com/charviki/maze-cradle/logutil"
	cradlemw "github.com/charviki/maze-cradle/middleware"
	"github.com/charviki/sweetwater-black-ridge/biz/config"
	"github.com/charviki/sweetwater-black-ridge/biz/handler"
	"github.com/charviki/sweetwater-black-ridge/biz/model"
	"github.com/charviki/sweetwater-black-ridge/biz/service"
)

// Register 注册所有 API 路由（便捷入口，内部创建 TmuxService）
func Register(h *server.Hertz, cfg *config.Config, logger logutil.Logger, gwmux *gwruntime.ServeMux) *model.TemplateStore {
	tmuxService := service.NewTmuxService(&cfg.Tmux, cfg.Workspace.StateDir, logger)
	return RegisterWithService(h, cfg, tmuxService, logger, gwmux)
}

// RegisterWithService 注册路由：中间件 + WebSocket + 健康检查。
// REST API 路由由 grpc-gateway（通过 Hertz NoRoute 转发）处理。
func RegisterWithService(h *server.Hertz, cfg *config.Config, tmuxService service.TmuxService, logger logutil.Logger, _ *gwruntime.ServeMux) *model.TemplateStore {
	templateStore := model.NewTemplateStore(filepath.Join(cfg.Workspace.StateDir, "templates.json"), logger)

	terminalHandler := handler.NewTerminalHandler(tmuxService, cfg.Terminal.DefaultLines, logger, cfg.Server.AllowedOrigins)

	// Hertz 中间件：CORS 和 AccessLog 仍由 Hertz 处理
	h.Use(cradlemw.CORS())
	h.Use(accesslog.New())

	// WebSocket 端点由 Hertz 直接管理（gRPC 不支持 WebSocket 升级）
	protected := h.Group("/api/v1", cradlemw.Auth(cfg.Server.AuthToken))
	protected.GET("/sessions/:id/ws", terminalHandler.HandleWs)

	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	return templateStore
}
