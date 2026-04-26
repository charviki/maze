package router

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/logger/accesslog"

	"github.com/charviki/maze-cradle/logutil"
	cradlemw "github.com/charviki/maze-cradle/middleware"
	"github.com/charviki/sweetwater-black-ridge/biz/config"
	"github.com/charviki/sweetwater-black-ridge/biz/handler"
	"github.com/charviki/sweetwater-black-ridge/biz/model"
	"github.com/charviki/sweetwater-black-ridge/biz/service"
)

// Register 注册所有 API 路由（创建新的 TmuxService 实例）
func Register(h *server.Hertz, cfg *config.Config, logger logutil.Logger) {
	tmuxService := service.NewTmuxService(&cfg.Tmux, cfg.Workspace.StateDir, logger)
	RegisterWithService(h, cfg, tmuxService, logger)
}

// RegisterWithService 注册所有 API 路由（使用外部注入的 TmuxService）。
// 所有 API 端点（包括 WebSocket）都经过 Auth 中间件保护。
func RegisterWithService(h *server.Hertz, cfg *config.Config, tmuxService service.TmuxService, logger logutil.Logger) {
	templateStore := model.NewTemplateStore("templates.json", logger)
	sessionHandler := handler.NewSessionHandler(tmuxService, templateStore, cfg, logger)
	terminalHandler := handler.NewTerminalHandler(tmuxService, cfg.Terminal.DefaultLines, logger, cfg.Server.AllowedOrigins)

	templateHandler := handler.NewTemplateHandler(templateStore)

	localConfigStore := service.NewLocalConfigStore(cfg.Workspace.RootDir, logger)
	localConfigHandler := handler.NewLocalConfigHandler(localConfigStore)

	h.Use(cradlemw.CORS())
	h.Use(accesslog.New())

	api := h.Group("/api/v1", cradlemw.Auth(cfg.Server.AuthToken))

	// Session CRUD
	api.GET("/sessions", sessionHandler.ListSessions)
	api.POST("/sessions", sessionHandler.CreateSession)
	api.GET("/sessions/:id", sessionHandler.GetSession)
	api.DELETE("/sessions/:id", sessionHandler.DeleteSession)
	api.GET("/sessions/:id/config", sessionHandler.GetSessionConfig)
	api.PUT("/sessions/:id/config", sessionHandler.UpdateSessionConfig)

	// 终端操作
	api.GET("/sessions/:id/output", terminalHandler.GetOutput)
	api.POST("/sessions/:id/input", terminalHandler.SendInput)
	api.POST("/sessions/:id/signal", terminalHandler.SendSignal)
	api.GET("/sessions/:id/env", terminalHandler.GetEnv)
	// WebSocket 端点也纳入 Auth 保护，防止未授权终端访问
	api.GET("/sessions/:id/ws", terminalHandler.HandleWs)

	// 管线保存与恢复
	api.GET("/sessions/saved", sessionHandler.GetSavedSessions)
	api.POST("/sessions/:id/restore", sessionHandler.RestoreSession)
	api.POST("/sessions/save", sessionHandler.SaveSessions)

	// Templates
	api.GET("/templates", templateHandler.ListTemplates)
	api.POST("/templates", templateHandler.CreateTemplate)
	api.GET("/templates/:id", templateHandler.GetTemplate)
	api.PUT("/templates/:id", templateHandler.UpdateTemplate)
	api.DELETE("/templates/:id", templateHandler.DeleteTemplate)
	api.GET("/templates/:id/config", templateHandler.GetTemplateConfig)
	api.PUT("/templates/:id/config", templateHandler.UpdateTemplateConfig)

	// Local Config
	api.GET("/local-config", localConfigHandler.GetConfig)
	api.PUT("/local-config", localConfigHandler.UpdateConfig)

	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
}
