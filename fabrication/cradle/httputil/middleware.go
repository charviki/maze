package httputil

import (
	"net/http"
	"time"

	"github.com/charviki/maze/fabrication/cradle/logutil"
)

// CORSMiddleware 返回 CORS 中间件，origins 为空时允许所有来源。
func CORSMiddleware(origins []string) func(http.Handler) http.Handler {
	if len(origins) == 0 {
		return CORS()
	}
	return CORSWithOrigins(origins)
}

// ChainHTTP 将多个中间件按从外到内的顺序应用到 handler。
func ChainHTTP(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	wrapped := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}
	return wrapped
}

// AccessLogMiddleware 返回 HTTP 访问日志中间件，记录方法、路径、状态码和耗时。
func AccessLogMiddleware(logger logutil.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := NewStatusRecorder(w)
			startedAt := time.Now()
			next.ServeHTTP(recorder, r)
			if logger != nil {
				logger.Infof("[http] %s %s status=%d duration=%s", r.Method, r.URL.Path, recorder.Status(), time.Since(startedAt))
			}
		})
	}
}
