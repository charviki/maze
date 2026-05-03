package httputil

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

// StatusRecorder 包装 http.ResponseWriter 并记录最终状态码。
// 之所以放在共享层，是因为 WebSocket 升级依赖 Hijacker/Flusher 等可选接口，
// 若每个服务各自包装而不透传这些接口，会让升级在中间件链中失效。
type StatusRecorder struct {
	http.ResponseWriter
	status int
}

// NewStatusRecorder 创建默认状态码为 200 的记录器。
// 这样即使 handler 没有显式调用 WriteHeader，访问日志也能反映真实默认状态。
func NewStatusRecorder(w http.ResponseWriter) *StatusRecorder {
	return &StatusRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

// Status 返回当前记录到的状态码。
func (r *StatusRecorder) Status() int {
	return r.status
}

// WriteHeader 记录状态码后再透传到底层 writer。
func (r *StatusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Unwrap 暴露底层 writer，方便标准库能力探测沿包装链继续下钻。
func (r *StatusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// Hijack 透传到底层 writer，确保 WebSocket/HTTP upgrade 不会因为包装器丢失能力。
func (r *StatusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("http hijacker not supported")
	}
	return hijacker.Hijack()
}

// Flush 透传到底层 writer，确保流式响应仍能及时刷新。
func (r *StatusRecorder) Flush() {
	flusher, ok := r.ResponseWriter.(http.Flusher)
	if !ok {
		return
	}
	flusher.Flush()
}

// Push 透传到底层 writer，保持对 HTTP/2 server push 的兼容。
func (r *StatusRecorder) Push(target string, opts *http.PushOptions) error {
	pusher, ok := r.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return pusher.Push(target, opts)
}
