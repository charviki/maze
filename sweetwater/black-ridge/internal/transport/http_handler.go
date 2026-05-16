package transport

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/charviki/maze/fabrication/cradle/httputil"
	"github.com/charviki/maze/fabrication/cradle/logutil"
	cradlemw "github.com/charviki/maze/fabrication/cradle/middleware"
	"github.com/charviki/sweetwater-black-ridge/internal/config"
	"github.com/charviki/sweetwater-black-ridge/internal/service"
)

const (
	httpReadTimeout  = 10 * time.Second
	httpWriteTimeout = 30 * time.Second
	httpIdleTimeout  = 120 * time.Second
)

// HTTPHandlerParams 包含构造 black-ridge HTTP 入口所需的全部依赖。
type HTTPHandlerParams struct {
	Config         *config.Config
	TmuxService    service.TmuxService
	Logger         logutil.Logger
	GWMux          *gwruntime.ServeMux
	StaticFiles    fs.FS
	JWTSecret      string
	AllowedOrigins []string
}

// NewHTTPHandler 构造完整的 HTTP 入口，包含路由注册、middleware 编排和 SPA fallback。
// cmd 只需调用此函数并将返回的 handler 交给 http.Server 即可。
func NewHTTPHandler(params HTTPHandlerParams) (http.Handler, *service.TemplateStore) {
	templateStore := service.NewTemplateStore(
		filepath.Join(params.Config.Workspace.StateDir, "templates.json"),
		params.Logger,
	)
	terminalHandler := NewTerminalHandler(
		params.TmuxService,
		params.Config.Terminal.DefaultLines,
		params.Logger,
		params.AllowedOrigins,
	)

	apiHandler := httputil.ChainHTTP(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				params.GWMux.ServeHTTP(w, r)
				return
			}
			serveSPA(params.Logger, params.StaticFiles, w, r)
		}),
		httputil.AccessLogMiddleware(params.Logger),
		httputil.CORSMiddleware(params.AllowedOrigins),
	)
	wsHandler := httputil.ChainHTTP(
		http.HandlerFunc(terminalHandler.HandleWs),
		httputil.AccessLogMiddleware(params.Logger),
		httputil.CORSMiddleware(params.AllowedOrigins),
		cradlemw.Auth(params.JWTSecret),
	)
	mux := http.NewServeMux()
	mux.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	mux.Handle("GET /api/v1/sessions/{id}/ws", wsHandler)
	mux.Handle("/", apiHandler)

	return mux, templateStore
}

// NewHTTPServer 构造完整的 http.Server，供 cmd 直接使用。
func NewHTTPServer(params HTTPHandlerParams) (*http.Server, *service.TemplateStore) {
	handler, templateStore := NewHTTPHandler(params)
	return &http.Server{
		Addr:         params.Config.Server.ListenAddr,
		Handler:      handler,
		ReadTimeout:  httpReadTimeout,
		WriteTimeout: httpWriteTimeout,
		IdleTimeout:  httpIdleTimeout,
	}, templateStore
}

func serveSPA(logger logutil.Logger, staticFiles fs.FS, w http.ResponseWriter, r *http.Request) {
	subFS, err := getStaticFS(staticFiles)
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

	if _, err := fs.Stat(subFS, trimmedPath); err != nil {
		r = r.Clone(r.Context())
		r.URL.Path = "/"
	} else {
		r = r.Clone(r.Context())
		r.URL.Path = "/" + trimmedPath
	}

	http.FileServer(http.FS(subFS)).ServeHTTP(w, r)
}

func getStaticFS(staticFiles fs.FS) (fs.FS, error) {
	if _, statErr := staticFiles.Open("web-dist"); statErr != nil {
		return nil, statErr
	}
	return fs.Sub(staticFiles, "web-dist")
}
