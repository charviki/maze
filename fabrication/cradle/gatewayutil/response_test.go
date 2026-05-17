package gatewayutil

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charviki/maze/fabrication/cradle/auth"
	"github.com/charviki/maze/fabrication/cradle/errutil"
	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestJSONPb_ContentType(t *testing.T) {
	m := &runtime.JSONPb{}
	if got := m.ContentType(nil); got != "application/json" {
		t.Errorf("ContentType() = %q, 期望 %q", got, "application/json")
	}
}

func TestJSONPb_Marshal_ProtoMessage(t *testing.T) {
	m := &runtime.JSONPb{MarshalOptions: protojson.MarshalOptions{EmitUnpopulated: true}}
	data, err := m.Marshal(&emptypb.Empty{})
	if err != nil {
		t.Fatalf("Marshal() 返回错误: %v", err)
	}
	if string(data) != "{}" {
		t.Errorf("Marshal() = %q, 期望 %q", string(data), "{}")
	}
}

func TestJSONPb_Marshal_NonProtoMessage(t *testing.T) {
	m := &runtime.JSONPb{MarshalOptions: protojson.MarshalOptions{EmitUnpopulated: true}}
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
		wantReason   string
	}{
		{
			name:         "NotFound → 404",
			err:          status.Error(codes.NotFound, "resource not found"),
			wantHTTPCode: http.StatusNotFound,
			wantGRPCCode: codes.NotFound,
			wantMessage:  "resource not found",
			wantReason:   "",
		},
		{
			name:         "InvalidArgument → 400",
			err:          status.Error(codes.InvalidArgument, "bad request"),
			wantHTTPCode: http.StatusBadRequest,
			wantGRPCCode: codes.InvalidArgument,
			wantMessage:  "bad request",
			wantReason:   "",
		},
		{
			name:         "Internal → 500",
			err:          status.Error(codes.Internal, "something broke"),
			wantHTTPCode: http.StatusInternalServerError,
			wantGRPCCode: codes.Internal,
			wantMessage:  "something broke",
			wantReason:   "",
		},
		{
			name:         "Unauthenticated With Reason → 401",
			err:          auth.ExpiredTokenError("no auth"),
			wantHTTPCode: http.StatusUnauthorized,
			wantGRPCCode: codes.Unauthenticated,
			wantMessage:  "no auth",
			wantReason:   string(auth.ErrorReasonTokenExpired),
		},
		{
			name:         "PermissionDenied → 403",
			err:          status.Error(codes.PermissionDenied, "forbidden"),
			wantHTTPCode: http.StatusForbidden,
			wantGRPCCode: codes.PermissionDenied,
			wantMessage:  "forbidden",
			wantReason:   "",
		},
		{
			name:         "AlreadyExists → 409",
			err:          status.Error(codes.AlreadyExists, "duplicate"),
			wantHTTPCode: http.StatusConflict,
			wantGRPCCode: codes.AlreadyExists,
			wantMessage:  "duplicate",
			wantReason:   "",
		},
		{
			name:         "errutil.NewError ARCHIVE_NOT_FOUND → 404",
			err:          errutil.NewError(codes.NotFound, pb.ErrorReason_ERROR_REASON_ARCHIVE_NOT_FOUND, "archive gone"),
			wantHTTPCode: http.StatusNotFound,
			wantGRPCCode: codes.NotFound,
			wantMessage:  "archive gone",
			wantReason:   "ARCHIVE_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux := runtime.NewServeMux()

			HTTPErrorHandler(
				context.TODO(), //nolint:staticcheck
				mux,
				&runtime.JSONPb{},
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
			if resp.Reason != tt.wantReason {
				t.Errorf("reason = %q, 期望 %q", resp.Reason, tt.wantReason)
			}
		})
	}
}

func TestHTTPErrorHandler_ValidationError(t *testing.T) {
	err := errutil.NewValidationError(
		codes.InvalidArgument,
		pb.ErrorReason_ERROR_REASON_VALIDATION_FAILED,
		"bad input",
		[]errutil.FieldViolation{
			{Field: "name", Description: "required"},
			{Field: "email", Description: "invalid format"},
		},
	)

	rec := httptest.NewRecorder()
	mux := runtime.NewServeMux()
	HTTPErrorHandler(
		context.TODO(), //nolint:staticcheck
		mux,
		&runtime.JSONPb{},
		rec,
		httptest.NewRequest("POST", "/test", nil),
		err,
	)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("HTTP 状态码 = %d, 期望 %d", rec.Code, http.StatusBadRequest)
	}

	var resp rpcStatusBody
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}
	if resp.Code != int32(codes.InvalidArgument) {
		t.Errorf("code = %d, 期望 %d", resp.Code, int32(codes.InvalidArgument))
	}
	if resp.Message != "bad input" {
		t.Errorf("message = %q, 期望 %q", resp.Message, "bad input")
	}
	if resp.Reason != "VALIDATION_FAILED" {
		t.Errorf("reason = %q, 期望 %q", resp.Reason, "VALIDATION_FAILED")
	}
	if resp.Details == nil {
		t.Fatal("details 不应为 nil")
	}
	if len(resp.Details.FieldViolations) != 2 {
		t.Fatalf("fieldViolations 数量 = %d, 期望 2", len(resp.Details.FieldViolations))
	}
	if resp.Details.FieldViolations[0].Field != "name" || resp.Details.FieldViolations[0].Description != "required" {
		t.Errorf("第一个 violation = %+v, 期望 field=name desc=required", resp.Details.FieldViolations[0])
	}
	if len(resp.Details.PreconditionViolations) != 0 {
		t.Errorf("preconditionViolations 应为空, 实际有 %d 个", len(resp.Details.PreconditionViolations))
	}
}

func TestHTTPErrorHandler_PreconditionError(t *testing.T) {
	err := errutil.NewPreconditionError(
		codes.FailedPrecondition,
		pb.ErrorReason_ERROR_REASON_CONFIG_CONFLICT,
		"conflict",
		[]errutil.PreconditionViolation{
			{Type: "CONFIG_CONFLICT", Subject: "/etc/app.yaml", Description: "hash mismatch"},
		},
	)

	rec := httptest.NewRecorder()
	mux := runtime.NewServeMux()
	HTTPErrorHandler(
		context.TODO(), //nolint:staticcheck
		mux,
		&runtime.JSONPb{},
		rec,
		httptest.NewRequest("PUT", "/test", nil),
		err,
	)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("HTTP 状态码 = %d, 期望 %d", rec.Code, http.StatusBadRequest)
	}

	var resp rpcStatusBody
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}
	if resp.Code != int32(codes.FailedPrecondition) {
		t.Errorf("code = %d, 期望 %d", resp.Code, int32(codes.FailedPrecondition))
	}
	if resp.Reason != "CONFIG_CONFLICT" {
		t.Errorf("reason = %q, 期望 %q", resp.Reason, "CONFIG_CONFLICT")
	}
	if resp.Details == nil {
		t.Fatal("details 不应为 nil")
	}
	if len(resp.Details.PreconditionViolations) != 1 {
		t.Fatalf("preconditionViolations 数量 = %d, 期望 1", len(resp.Details.PreconditionViolations))
	}
	pv := resp.Details.PreconditionViolations[0]
	if pv.Type != "CONFIG_CONFLICT" || pv.Subject != "/etc/app.yaml" || pv.Description != "hash mismatch" {
		t.Errorf("violation = %+v, unexpected", pv)
	}
	if len(resp.Details.FieldViolations) != 0 {
		t.Errorf("fieldViolations 应为空, 实际有 %d 个", len(resp.Details.FieldViolations))
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
