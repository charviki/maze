package gatewayutil

import (
	"context"
	"net/http"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// NewServeMux 创建预配置的 grpc-gateway ServeMux。
// 统一配置：proto JSON Marshaler（EmitUnpopulated）、rpcStatus 错误处理器、ForwardResponseOption。
func NewServeMux() *runtime.ServeMux {
	return runtime.NewServeMux(
		// 使用 runtime.JSONPb 而非默认的 runtime.JSONBuiltin。
		// 原因：默认 JSONBuiltin 内部用 json.Marshal，proto 生成的 Go struct 的 json tag
		// 带有 omitempty，会省略零值字段（空数组、空字符串、数字 0），导致前端收到
		// undefined 引发运行时错误。runtime.JSONPb 使用 protojson.Marshal，配合
		// EmitUnpopulated: true 输出所有字段（包括零值），确保前端始终收到完整结构。
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{EmitUnpopulated: true},
		}),

		// CreateHost 等异步操作应返回 202 Accepted 而非默认的 200 OK。
		// 通过 ForwardResponseOption 在响应写出前检查 gRPC metadata 中携带的
		// x-http-status key 来覆盖 HTTP 状态码。
		runtime.WithForwardResponseOption(setStatusCodeFromMetadata),

		// 错误响应输出 rpcStatus 格式 {"code": int32, "message": "..."}
		runtime.WithErrorHandler(HTTPErrorHandler),
	)
}

// setStatusCodeFromMetadata 检查 grpc-gateway ServerMetadata 中的 x-http-status key，
// 如果存在则覆盖 HTTP 响应状态码。用于 CreateHost 等 RPC 需要返回非 200 场景。
//
// gRPC handler 端设置方式:
//
//	runtime.SetHeader(ctx, metadata.Pairs("x-http-status", "202"))
func setStatusCodeFromMetadata(ctx context.Context, w http.ResponseWriter, _ proto.Message) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}

	values := md.HeaderMD.Get("x-http-status")
	if len(values) == 0 {
		return nil
	}

	code, err := strconv.Atoi(values[0])
	if err != nil {
		return nil //nolint:nilerr
	}

	w.WriteHeader(code)
	return nil
}
