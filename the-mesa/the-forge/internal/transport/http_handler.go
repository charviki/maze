package transport

import (
	"encoding/json"
	"net/http"
	"time"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/charviki/maze/fabrication/cradle/httputil"
	"github.com/charviki/maze/fabrication/cradle/logutil"

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
	grpcGatewayHandler := httputil.ChainHTTP(
		params.GWMux,
		httputil.AccessLogMiddleware(params.Logger),
		httputil.CORSMiddleware(params.Config.Server.Origins()),
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

