package gatewayutil

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
)

// rpcStatusBody 错误响应体，与 OpenAPI spec 中 rpcStatus 定义一致。
type rpcStatusBody struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
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

	resp := rpcStatusBody{
		Code:    int32(s.Code()), //nolint:gosec // G115: gRPC codes.Code 在 [0,16] 范围内，无溢出风险
		Message: s.Message(),
	}
	body, _ := json.Marshal(resp)
	_, _ = w.Write(body)
}
