package transport

import (
	"encoding/json"
	"net/http"
	"time"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/charviki/maze/fabrication/cradle/httputil"
	"github.com/charviki/maze/fabrication/cradle/logutil"
	cradlemw "github.com/charviki/maze/fabrication/cradle/middleware"

	"github.com/charviki/maze/the-mesa/the-forge/internal/config"
)

const (
	httpReadTimeout  = 10 * time.Second
	httpWriteTimeout = 30 * time.Second
	httpIdleTimeout  = 120 * time.Second
)

// HTTPHandlerParams 包含创建 HTTP handler 所需的参数。
type HTTPHandlerParams struct {
	Config *config.Config
	GWMux  *gwruntime.ServeMux
	Logger logutil.Logger
}

// NewHTTPHandler 创建 HTTP handler，路由按 director-core 模式编排：
//   - /health 免鉴权（K8s probe）
//   - / 由 grpc-gateway 代理，鉴权在 gRPC UnaryAuthInterceptor 层处理，HTTP 层不加 JWT
func NewHTTPHandler(params HTTPHandlerParams) http.Handler {
	grpcGatewayHandler := chainHTTP(
		params.GWMux,
		accessLogMiddleware(params.Logger),
		corsMiddleware(params.Config.Server.Origins()),
	)

	mux := http.NewServeMux()
	mux.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	mux.Handle("/", grpcGatewayHandler)

	return mux
}

// NewHTTPServer 创建 HTTP server，供 cmd 直接使用。
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
