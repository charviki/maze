package auth

import (
	"encoding/json"
	"net/http"

	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	"github.com/charviki/maze/fabrication/cradle/errutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorReason 是认证失败面向机器消费的稳定错误原因。
//
// Deprecated: use pb.ErrorReason
type ErrorReason string

const (
	// ErrorReasonTokenMissing 表示请求缺失访问令牌。
	//
	// Deprecated: use pb.ErrorReason_ERROR_REASON_TOKEN_MISSING
	ErrorReasonTokenMissing ErrorReason = "TOKEN_MISSING"
	// ErrorReasonTokenInvalid 表示访问令牌格式非法、签名错误或 Bearer 方案错误。
	//
	// Deprecated: use pb.ErrorReason_ERROR_REASON_TOKEN_INVALID
	ErrorReasonTokenInvalid ErrorReason = "TOKEN_INVALID"
	// ErrorReasonTokenExpired 表示访问令牌已过期。
	//
	// Deprecated: use pb.ErrorReason_ERROR_REASON_TOKEN_EXPIRED
	ErrorReasonTokenExpired ErrorReason = "TOKEN_EXPIRED"
)

// ErrorResponse 是 HTTP 与 grpc-gateway 共享的结构化认证错误载荷。
type ErrorResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Reason  string `json:"reason,omitempty"`
}

// MissingTokenError 构造"缺失访问令牌"的 gRPC 状态错误。
func MissingTokenError(message string) error {
	return errutil.NewError(codes.Unauthenticated, pb.ErrorReason_ERROR_REASON_TOKEN_MISSING, message)
}

// InvalidTokenError 构造"访问令牌非法"的 gRPC 状态错误。
func InvalidTokenError(message string) error {
	return errutil.NewError(codes.Unauthenticated, pb.ErrorReason_ERROR_REASON_TOKEN_INVALID, message)
}

// ExpiredTokenError 构造"访问令牌过期"的 gRPC 状态错误。
func ExpiredTokenError(message string) error {
	return errutil.NewError(codes.Unauthenticated, pb.ErrorReason_ERROR_REASON_TOKEN_EXPIRED, message)
}

// ErrorResponseFromError 将任意错误转换为统一的结构化错误响应。
func ErrorResponseFromError(err error) ErrorResponse {
	st := status.Convert(err)
	reason := errutil.ReasonFromError(st.Err())
	return ErrorResponse{
		Code:    int32(st.Code()), //nolint:gosec // G115: gRPC codes.Code 在 [0,16] 范围内，无溢出风险
		Message: st.Message(),
		Reason:  errutil.ReasonName(reason),
	}
}

// ErrorReasonFromError 提取错误中的结构化认证原因；非认证错误返回空字符串。
func ErrorReasonFromError(err error) ErrorReason {
	r := errutil.ReasonFromError(err)
	switch r {
	case pb.ErrorReason_ERROR_REASON_TOKEN_MISSING:
		return ErrorReasonTokenMissing
	case pb.ErrorReason_ERROR_REASON_TOKEN_INVALID:
		return ErrorReasonTokenInvalid
	case pb.ErrorReason_ERROR_REASON_TOKEN_EXPIRED:
		return ErrorReasonTokenExpired
	default:
		return ""
	}
}

// WriteHTTPError 按统一契约写出结构化认证错误。
func WriteHTTPError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if encodeErr := json.NewEncoder(w).Encode(ErrorResponseFromError(err)); encodeErr != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
