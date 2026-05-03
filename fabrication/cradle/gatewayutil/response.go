package gatewayutil

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// protojsonMarshaler grpc-gateway 标准 proto JSON Marshaler。
// 直接输出 protojson 序列化结果（EmitUnpopulated: true），不添加外层包装，
// 与 OpenAPI spec（protoc-gen-openapiv2）定义的响应格式完全一致。
type protojsonMarshaler struct{}

// ContentType 返回 JSON 内容类型。
func (protojsonMarshaler) ContentType(_ interface{}) string {
	return "application/json"
}

// Marshal 将 proto message 直接序列化为 proto JSON（EmitUnpopulated: true）。
func (protojsonMarshaler) Marshal(v interface{}) ([]byte, error) {
	pm, ok := v.(proto.Message)
	if !ok {
		return json.Marshal(v)
	}
	return protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(pm)
}

// NewDecoder 返回将 JSON body 解码为 proto message 的解码器。
func (protojsonMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return protojsonDecoder{dec: json.NewDecoder(r)}
}

// NewEncoder 返回将 proto message 编码为 JSON 并写入 w 的编码器。
func (protojsonMarshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return protojsonEncoder{enc: json.NewEncoder(w)}
}

// Unmarshal 将 JSON bytes 反序列化为 proto message，满足 runtime.Marshaler 接口。
func (protojsonMarshaler) Unmarshal(data []byte, v interface{}) error {
	pm, ok := v.(proto.Message)
	if !ok {
		return json.Unmarshal(data, v)
	}
	return protojson.Unmarshal(data, pm)
}

// protojsonDecoder 将标准 JSON 解码为 proto message。
type protojsonDecoder struct {
	dec *json.Decoder
}

// Decode 将 JSON body 解码为 proto message。
func (d protojsonDecoder) Decode(v interface{}) error {
	pm, ok := v.(proto.Message)
	if !ok {
		return d.dec.Decode(v)
	}
	var raw json.RawMessage
	if err := d.dec.Decode(&raw); err != nil {
		return err
	}
	return protojson.Unmarshal(raw, pm)
}

// protojsonEncoder 将 proto message 编码为 JSON 并写入 io.Writer。
type protojsonEncoder struct {
	enc *json.Encoder
}

// Encode 将 proto message 编码为 JSON 并写入 io.Writer。
func (e protojsonEncoder) Encode(v interface{}) error {
	pm, ok := v.(proto.Message)
	if !ok {
		return e.enc.Encode(v)
	}
	raw, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(pm)
	if err != nil {
		return err
	}
	return e.enc.Encode(json.RawMessage(raw))
}

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

	// gRPC code 范围 [0, 16]，int32 转换安全
	resp := rpcStatusBody{
		Code:    int32(s.Code()), //nolint:gosec // G115: gRPC codes.Code 在 [0,16] 范围内，无溢出风险
		Message: s.Message(),
	}
	body, _ := json.Marshal(resp)
	_, _ = w.Write(body)
}
