package transport

import (
	"encoding/json"
	"net/http"
	"time"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/charviki/maze-cradle/httputil"
	"github.com/charviki/maze-cradle/logutil"
	cradlemw "github.com/charviki/maze-cradle/middleware"
	"github.com/charviki/maze/the-mesa/director-core/internal/config"
)

const (
	httpReadTimeout  = 10 * time.Second
	httpWriteTimeout = 30 * time.Second
	httpIdleTimeout  = 120 * time.Second
)

// HTTPHandlerParams 包含构造 director-core HTTP 入口所需的全部依赖。
type HTTPHandlerParams struct {
	Config             *config.Config
	Logger             logutil.Logger
	GWMux              *gwruntime.ServeMux
	SessionProxyHandler *SessionProxyHandler
	AuthToken          string
	AllowedOrigins     []string
}

// NewHTTPHandler 构造完整的 HTTP 入口，包含路由注册和 middleware 编排。
func NewHTTPHandler(params HTTPHandlerParams) http.Handler {
	apiHandler := chainHTTP(
		params.GWMux,
		accessLogMiddleware(params.Logger),
		corsMiddleware(params.AllowedOrigins),
		cradlemw.Auth(params.AuthToken),
	)
	agentHandler := chainHTTP(
		params.GWMux,
		accessLogMiddleware(params.Logger),
		corsMiddleware(params.AllowedOrigins),
	)
	wsHandler := chainHTTP(
		http.HandlerFunc(params.SessionProxyHandler.ProxyWebSocket),
		accessLogMiddleware(params.Logger),
		corsMiddleware(params.AllowedOrigins),
		cradlemw.Auth(params.AuthToken),
	)

	mux := http.NewServeMux()
	mux.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	mux.Handle("POST /api/v1/nodes/register", agentHandler)
	mux.Handle("POST /api/v1/nodes/heartbeat", agentHandler)
	mux.Handle("GET /api/v1/nodes/{name}/sessions/{id}/ws", wsHandler)
	mux.Handle("/", apiHandler)

	return mux
}

// NewHTTPServer 构造完整的 http.Server，供 cmd 直接使用。
func NewHTTPServer(params HTTPHandlerParams) *http.Server {
	handler := NewHTTPHandler(params)
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
