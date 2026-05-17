package gatewayutil

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/charviki/maze/fabrication/cradle/errutil"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
)

// rpcStatusBody 错误响应体，与 OpenAPI spec 中 rpcStatus 定义一致。
type rpcStatusBody struct {
	Code    int32             `json:"code"`
	Message string            `json:"message"`
	Reason  string            `json:"reason,omitempty"`
	Details *rpcStatusDetails `json:"details,omitempty"`
}

type rpcStatusDetails struct {
	FieldViolations        []fieldViolationJSON        `json:"fieldViolations,omitempty"`
	PreconditionViolations []preconditionViolationJSON `json:"preconditionViolations,omitempty"`
}

type fieldViolationJSON struct {
	Field       string `json:"field"`
	Description string `json:"description"`
}

type preconditionViolationJSON struct {
	Type        string `json:"type"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
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

	reason := errutil.ReasonFromError(s.Err())
	fieldViolations := errutil.FieldViolationsFromError(s.Err())
	precondViolations := errutil.PreconditionViolationsFromError(s.Err())

	resp := rpcStatusBody{
		Code:    int32(s.Code()), //nolint:gosec // G115: gRPC codes.Code 在 [0,16] 范围内，无溢出风险
		Message: s.Message(),
		Reason:  errutil.ReasonName(reason),
	}
	if len(fieldViolations) > 0 || len(precondViolations) > 0 {
		resp.Details = &rpcStatusDetails{}
		for _, fv := range fieldViolations {
			resp.Details.FieldViolations = append(resp.Details.FieldViolations, fieldViolationJSON{Field: fv.Field, Description: fv.Description})
		}
		for _, pv := range precondViolations {
			resp.Details.PreconditionViolations = append(resp.Details.PreconditionViolations, preconditionViolationJSON{Type: pv.Type, Subject: pv.Subject, Description: pv.Description})
		}
	}
	body, _ := json.Marshal(resp)
	_, _ = w.Write(body)
}
