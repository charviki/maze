package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/mesa-hub-behavior-panel/biz/config"
)

func main() {
	logger := logutil.New("manager")

	// 用 slog 替换 Hertz 框架默认 logger，统一 JSON 输出
	hlog.SetLogger(logger)

	cfg, err := config.LoadFromExe()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	h := server.Default(server.WithHostPorts(cfg.Server.ListenAddr))

	resources := register(h, cfg, logger)

	logger.Infof("agent-manager controller started on %s", cfg.Server.ListenAddr)
	if cfg.IsDevMode() {
		logger.Warnf("[security] running in DEV mode: auth_token is empty, all API endpoints are open")
	}
	if len(cfg.AllowedOrigins()) == 0 {
		logger.Warnf("[security] running in DEV mode: CORS and WebSocket allow all origins")
	}

	// 优雅关闭：监听 SIGINT/SIGTERM，依次停止 HTTP 服务 → 刷盘节点数据 → 关闭审计日志文件
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Infof("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.Shutdown(ctx); err != nil {
			logger.Errorf("shutdown error: %v", err)
		}
		// 刷盘 NodeRegistry 脏数据，确保最后 30 秒内的注册/心跳不丢失
		resources.Registry.WaitSave()
		// 关闭审计日志文件句柄，确保 last write 刷盘
		resources.AuditLog.Close()
	}()

	h.Spin()
}
