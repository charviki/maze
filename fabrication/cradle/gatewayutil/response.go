package gatewayutil

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/charviki/maze/fabrication/cradle/auth"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
)

// rpcStatusBody 错误响应体，与 OpenAPI spec 中 rpcStatus 定义一致。
type rpcStatusBody struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Reason  string `json:"reason,omitempty"`
}

// HTTPErrorHandler 自定义 gRPC 错误 → HTTP 响应处理器。
// 输出 rpcStatus 格式（{"code": int32, "message": "..."}），
// 与 OpenAPI spec 的 default 响应定义一致。
func HTTPErrorHandler(
	_ context.Context,
	_ *runtime.ServeMux,
	marshaler runtime.Marshaler,
	w http.ResponseWriter,
	_ *http.Request,
	err error,
) {
	s, ok := status.FromError(err)
	if !ok {
		s = status.Convert(err)
	}

	httpCode := grpcCodeToHTTP(s.Code())

	w.Header().Set("Content-Type", marshaler.ContentType(nil))
	w.WriteHeader(httpCode)

	base := auth.ErrorResponseFromError(s.Err())
	resp := rpcStatusBody{
		Code:    base.Code,
		Message: base.Message,
		Reason:  base.Reason,
	}
	body, _ := json.Marshal(resp)
	_, _ = w.Write(body)
}
