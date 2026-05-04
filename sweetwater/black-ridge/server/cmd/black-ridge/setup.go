package main

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/charviki/maze-cradle/httputil"
	"github.com/charviki/maze-cradle/logutil"
	cradlemw "github.com/charviki/maze-cradle/middleware"
	"github.com/charviki/sweetwater-black-ridge/internal/config"
	
	"github.com/charviki/sweetwater-black-ridge/internal/service"
	"github.com/charviki/sweetwater-black-ridge/internal/transport"
	"github.com/charviki/sweetwater-black-ridge/internal/webstatic"
)

const (
	agentHTTPReadTimeout  = 10 * time.Second
	agentHTTPWriteTimeout = 30 * time.Second
	agentHTTPIdleTimeout  = 120 * time.Second
)

func newHTTPServer(cfg *config.Config, tmuxService service.TmuxService, logger logutil.Logger, gwmux *gwruntime.ServeMux) (*http.Server, *service.TemplateStore) {
	templateStore := service.NewTemplateStore(path.Join(cfg.Workspace.StateDir, "templates.json"), logger)
	terminalHandler := transport.NewTerminalHandler(tmuxService, cfg.Terminal.DefaultLines, logger, cfg.Server.AllowedOrigins)

	apiHandler := chainHTTP(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				gwmux.ServeHTTP(w, r)
				return
			}
			serveSPA(logger, w, r)
		}),
		accessLogMiddleware(logger),
		corsMiddleware(cfg.Server.AllowedOrigins),
	)
	wsHandler := chainHTTP(
		http.HandlerFunc(terminalHandler.HandleWs),
		accessLogMiddleware(logger),
		corsMiddleware(cfg.Server.AllowedOrigins),
		cradlemw.Auth(cfg.Server.AuthToken),
	)

	mux := http.NewServeMux()
	mux.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	mux.Handle("GET /api/v1/sessions/{id}/ws", wsHandler)
	mux.Handle("/", apiHandler)

	return &http.Server{
		Addr:         cfg.Server.ListenAddr,
		Handler:      mux,
		ReadTimeout:  agentHTTPReadTimeout,
		WriteTimeout: agentHTTPWriteTimeout,
		IdleTimeout:  agentHTTPIdleTimeout,
	}, templateStore
}

func corsMiddleware(origins []string) func(http.Handler) http.Handler {
	if len(origins) == 0 {
		return cradlemw.CORS()
	}
	return cradlemw.CORSWithOrigins(origins)
}

func chainHTTP(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	wrapped := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}
	return wrapped
}

func accessLogMiddleware(logger logutil.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 访问日志需要记录状态码，但必须保留底层 writer 的 Hijacker 能力，
			// 否则终端 WebSocket 会在升级阶段直接失败。
			recorder := httputil.NewStatusRecorder(w)
			startedAt := time.Now()
			next.ServeHTTP(recorder, r)
			if logger != nil {
				logger.Infof("[http] %s %s status=%d duration=%s", r.Method, r.URL.Path, recorder.Status(), time.Since(startedAt))
			}
		})
	}
}

type backgroundRunner struct {
	name     string
	logger   logutil.Logger
	run      func(<-chan struct{})
	stopOnce sync.Once
	stopCh   chan struct{}
	doneCh   chan struct{}
}

func newBackgroundRunner(name string, logger logutil.Logger, run func(<-chan struct{})) *backgroundRunner {
	return &backgroundRunner{
		name:   name,
		logger: logger,
		run:    run,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

func (r *backgroundRunner) ListenAndServe() error {
	defer close(r.doneCh)
	r.run(r.stopCh)
	return nil
}

func (r *backgroundRunner) Shutdown(ctx context.Context) error {
	r.stopOnce.Do(func() {
		close(r.stopCh)
	})
	select {
	case <-r.doneCh:
		return nil
	case <-ctx.Done():
		if r.logger != nil {
			r.logger.Warnf("[%s] shutdown timed out: %v", r.name, ctx.Err())
		}
		return nil
	}
}

func serveSPA(logger logutil.Logger, w http.ResponseWriter, r *http.Request) {
	subFS, err := getStaticFS()
	if err != nil {
		if logger != nil {
			logger.Errorf("load static fs: %v", err)
		}
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	trimmedPath := strings.TrimPrefix(r.URL.Path, "/")
	if trimmedPath == "" {
		trimmedPath = "index.html"
	}

	// SPA 路由需要在资源不存在时回退到 index.html，但静态文件仍交给标准 FileServer 输出，
	// 这样可以保留 Content-Type/缓存等标准行为，而不是手写整套静态资源响应。
	if _, err := fs.Stat(subFS, trimmedPath); err != nil {
		r = r.Clone(r.Context())
		r.URL.Path = "/index.html"
	} else {
		r = r.Clone(r.Context())
		r.URL.Path = "/" + trimmedPath
	}

	http.FileServer(http.FS(subFS)).ServeHTTP(w, r)
}

func getStaticFS() (fs.FS, error) {
	if _, statErr := webstatic.Files.Open("web-dist"); statErr != nil {
		return nil, statErr
	}
	return fs.Sub(webstatic.Files, "web-dist")
}
