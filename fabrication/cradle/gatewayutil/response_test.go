package gatewayutil

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestProtojsonMarshaler_ContentType(t *testing.T) {
	m := protojsonMarshaler{}
	if got := m.ContentType(nil); got != "application/json" {
		t.Errorf("ContentType() = %q, 期望 %q", got, "application/json")
	}
}

func TestProtojsonMarshaler_Marshal_ProtoMessage(t *testing.T) {
	m := protojsonMarshaler{}
	data, err := m.Marshal(&emptypb.Empty{})
	if err != nil {
		t.Fatalf("Marshal() 返回错误: %v", err)
	}

	// 标准 proto JSON：emptypb.Empty{} 序列化为 {}
	if string(data) != "{}" {
		t.Errorf("Marshal() = %q, 期望 %q", string(data), "{}")
	}
}

func TestProtojsonMarshaler_Marshal_NonProtoMessage(t *testing.T) {
	m := protojsonMarshaler{}
	data, err := m.Marshal(map[string]string{"hello": "world"})
	if err != nil {
		t.Fatalf("Marshal() 返回错误: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("解析结果失败: %v", err)
	}
	if result["hello"] != "world" {
		t.Errorf("hello = %q, 期望 %q", result["hello"], "world")
	}
}

func TestHTTPErrorHandler(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantHTTPCode int
		wantGRPCCode codes.Code
		wantMessage  string
	}{
		{
			name:         "NotFound → 404",
			err:          status.Error(codes.NotFound, "resource not found"),
			wantHTTPCode: http.StatusNotFound,
			wantGRPCCode: codes.NotFound,
			wantMessage:  "resource not found",
		},
		{
			name:         "InvalidArgument → 400",
			err:          status.Error(codes.InvalidArgument, "bad request"),
			wantHTTPCode: http.StatusBadRequest,
			wantGRPCCode: codes.InvalidArgument,
			wantMessage:  "bad request",
		},
		{
			name:         "Internal → 500",
			err:          status.Error(codes.Internal, "something broke"),
			wantHTTPCode: http.StatusInternalServerError,
			wantGRPCCode: codes.Internal,
			wantMessage:  "something broke",
		},
		{
			name:         "Unauthenticated → 401",
			err:          status.Error(codes.Unauthenticated, "no auth"),
			wantHTTPCode: http.StatusUnauthorized,
			wantGRPCCode: codes.Unauthenticated,
			wantMessage:  "no auth",
		},
		{
			name:         "PermissionDenied → 403",
			err:          status.Error(codes.PermissionDenied, "forbidden"),
			wantHTTPCode: http.StatusForbidden,
			wantGRPCCode: codes.PermissionDenied,
			wantMessage:  "forbidden",
		},
		{
			name:         "AlreadyExists → 409",
			err:          status.Error(codes.AlreadyExists, "duplicate"),
			wantHTTPCode: http.StatusConflict,
			wantGRPCCode: codes.AlreadyExists,
			wantMessage:  "duplicate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux := runtime.NewServeMux()

			HTTPErrorHandler(
				context.TODO(), //nolint:staticcheck
				mux,
				protojsonMarshaler{},
				rec,
				httptest.NewRequest("GET", "/test", nil),
				tt.err,
			)

			if rec.Code != tt.wantHTTPCode {
				t.Errorf("HTTP 状态码 = %d, 期望 %d", rec.Code, tt.wantHTTPCode)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Content-Type = %q, 期望 %q", contentType, "application/json")
			}

			// 验证 rpcStatus 格式：{"code": int32, "message": "..."}
			var resp rpcStatusBody
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("解析响应体失败: %v", err)
			}
			if resp.Code != int32(tt.wantGRPCCode) {
				t.Errorf("code = %d, 期望 %d", resp.Code, int32(tt.wantGRPCCode))
			}
			if resp.Message != tt.wantMessage {
				t.Errorf("message = %q, 期望 %q", resp.Message, tt.wantMessage)
			}
		})
	}
}

func TestGrpcCodeToHTTP(t *testing.T) {
	tests := []struct {
		name string
		code codes.Code
		want int
	}{
		{"OK → 200", codes.OK, 200},
		{"InvalidArgument → 400", codes.InvalidArgument, 400},
		{"NotFound → 404", codes.NotFound, 404},
		{"AlreadyExists → 409", codes.AlreadyExists, 409},
		{"Internal → 500", codes.Internal, 500},
		{"Unauthenticated → 401", codes.Unauthenticated, 401},
		{"PermissionDenied → 403", codes.PermissionDenied, 403},
		{"Unavailable → 503", codes.Unavailable, 503},
		{"Canceled → 499", codes.Canceled, 499},
		{"DeadlineExceeded → 504", codes.DeadlineExceeded, 504},
		{"ResourceExhausted → 429", codes.ResourceExhausted, 429},
		{"FailedPrecondition → 400", codes.FailedPrecondition, 400},
		{"Unimplemented → 501", codes.Unimplemented, 501},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := grpcCodeToHTTP(tt.code)
			if got != tt.want {
				t.Errorf("grpcCodeToHTTP(%v) = %d, 期望 %d", tt.code, got, tt.want)
			}
		})
	}
}
