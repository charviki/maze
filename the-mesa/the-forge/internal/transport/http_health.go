package transport

import (
	"net/http"
	"time"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/charviki/maze/fabrication/cradle/httputil"
	"github.com/charviki/maze/fabrication/cradle/logutil"
	cradlemw "github.com/charviki/maze/fabrication/cradle/middleware"

	"github.com/charviki/maze/the-mesa/the-forge/internal/config"
	"github.com/charviki/maze/the-mesa/the-forge/internal/service"
)

const (
	httpReadTimeout  = 10 * time.Second
	httpWriteTimeout = 30 * time.Second
	httpIdleTimeout  = 120 * time.Second
)

// HTTPServerParams 包含创建 HTTP server 所需的所有参数。
type HTTPServerParams struct {
	Config        *config.Config
	HealthService *service.HealthService
	ChatHandler   *ChatHandler
	FileHandler   *FileHandler
	GWMux         *gwruntime.ServeMux
	Logger        logutil.Logger
}

// NewHTTPServer 创建 HTTP server，包含健康检查、grpc-gateway、Chat SSE 和文件路由。
func NewHTTPServer(params HTTPServerParams) *http.Server {
	mux := http.NewServeMux()

	// 健康检查
	mux.Handle("GET /health", NewHealthHandler(params.HealthService))

	// grpc-gateway 代理所有 /api/v1/ 路由（Knowledge + Directive）
	if params.GWMux != nil {
		mux.Handle("/api/v1/", params.GWMux)
	}

	// Chat SSE 和文件路由
	if params.ChatHandler != nil {
		params.ChatHandler.RegisterRoutes(mux)
	}
	if params.FileHandler != nil {
		params.FileHandler.RegisterRoutes(mux)
	}

	handler := chainHTTP(
		mux,
		accessLogMiddleware(params.Logger),
		corsMiddleware(params.Config.Server.Origins()),
		cradlemw.Auth(params.Config.Server.JWTSecret),
	)

	return &http.Server{
		Addr:         params.Config.Server.ListenAddr,
		Handler:      handler,
		ReadTimeout:  httpReadTimeout,
		WriteTimeout: httpWriteTimeout,
		IdleTimeout:  httpIdleTimeout,
	}
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
			recorder := httputil.NewStatusRecorder(w)
			startedAt := time.Now()
			next.ServeHTTP(recorder, r)
			if logger != nil {
				logger.Infof("[http] %s %s status=%d duration=%s", r.Method, r.URL.Path, recorder.Status(), time.Since(startedAt))
			}
		})
	}
}
