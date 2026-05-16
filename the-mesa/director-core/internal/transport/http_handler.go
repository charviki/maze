package transport

import (
	"encoding/json"
	"net/http"
	"time"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/charviki/maze/fabrication/cradle/httputil"
	"github.com/charviki/maze/fabrication/cradle/logutil"
	cradlemw "github.com/charviki/maze/fabrication/cradle/middleware"
	"github.com/charviki/maze/the-mesa/director-core/internal/config"
)

const (
	httpReadTimeout  = 10 * time.Second
	httpWriteTimeout = 30 * time.Second
	httpIdleTimeout  = 120 * time.Second
)

// HTTPHandlerParams 包含构造 director-core HTTP 入口所需的全部依赖。
type HTTPHandlerParams struct {
	Config              *config.Config
	Logger              logutil.Logger
	GWMux               *gwruntime.ServeMux
	SessionProxyHandler *SessionProxyHandler
	JWTSecret           string
	AllowedOrigins      []string
}

// NewHTTPHandler 构造完整的 HTTP 入口，包含路由注册和 middleware 编排。
// gRPC gateway 统一走 access log + CORS，鉴权由 gRPC interceptor 层处理。
// WebSocket 不走 gRPC，需独立的 HTTP 层 JWT 校验。
func NewHTTPHandler(params HTTPHandlerParams) http.Handler {
	grpcGatewayHandler := httputil.ChainHTTP(
		params.GWMux,
		httputil.AccessLogMiddleware(params.Logger),
		httputil.CORSMiddleware(params.AllowedOrigins),
	)
	wsHandler := httputil.ChainHTTP(
		http.HandlerFunc(params.SessionProxyHandler.ProxyWebSocket),
		httputil.AccessLogMiddleware(params.Logger),
		httputil.CORSMiddleware(params.AllowedOrigins),
		cradlemw.Auth(params.JWTSecret),
	)

	mux := http.NewServeMux()
	mux.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	mux.Handle("GET /api/v1/nodes/{name}/sessions/{id}/ws", wsHandler)
	mux.Handle("/", grpcGatewayHandler)

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
