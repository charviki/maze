package errutil

import (
	"testing"

	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewError(t *testing.T) {
	err := NewError(codes.NotFound, pb.ErrorReason_ERROR_REASON_RESOURCE_NOT_FOUND, "resource gone")
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
	if st.Message() != "resource gone" {
		t.Errorf("message = %q, want %q", st.Message(), "resource gone")
	}
	var foundInfo *errdetails.ErrorInfo
	for _, d := range st.Details() {
		if info, ok := d.(*errdetails.ErrorInfo); ok {
			foundInfo = info
		}
	}
	if foundInfo == nil {
		t.Fatal("no ErrorInfo detail found")
	}
	if foundInfo.GetReason() != "ERROR_REASON_RESOURCE_NOT_FOUND" {
		t.Errorf("reason = %q, want %q", foundInfo.GetReason(), "ERROR_REASON_RESOURCE_NOT_FOUND")
	}
	if foundInfo.GetDomain() != "maze.v1" {
		t.Errorf("domain = %q, want %q", foundInfo.GetDomain(), "maze.v1")
	}
}

func TestReasonFromError_NewError(t *testing.T) {
	err := NewError(codes.Unauthenticated, pb.ErrorReason_ERROR_REASON_TOKEN_EXPIRED, "expired")
	reason := ReasonFromError(err)
	if reason != pb.ErrorReason_ERROR_REASON_TOKEN_EXPIRED {
		t.Errorf("reason = %v, want TOKEN_EXPIRED", reason)
	}
}

func TestReasonFromError_PlainStatus(t *testing.T) {
	err := status.Error(codes.Internal, "boom")
	reason := ReasonFromError(err)
	if reason != pb.ErrorReason_ERROR_REASON_UNSPECIFIED {
		t.Errorf("reason = %v, want UNSPECIFIED", reason)
	}
}

func TestNewValidationError(t *testing.T) {
	violations := []FieldViolation{
		{Field: "name", Description: "required"},
		{Field: "email", Description: "invalid format"},
	}
	err := NewValidationError(codes.InvalidArgument, pb.ErrorReason_ERROR_REASON_VALIDATION_FAILED, "bad input", violations)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
	var foundBR *errdetails.BadRequest
	var foundInfo *errdetails.ErrorInfo
	for _, d := range st.Details() {
		if br, ok := d.(*errdetails.BadRequest); ok {
			foundBR = br
		}
		if info, ok := d.(*errdetails.ErrorInfo); ok {
			foundInfo = info
		}
	}
	if foundInfo == nil {
		t.Fatal("no ErrorInfo detail found")
	}
	if foundBR == nil {
		t.Fatal("no BadRequest detail found")
	}
	fvs := foundBR.GetFieldViolations()
	if len(fvs) != 2 {
		t.Fatalf("field violations count = %d, want 2", len(fvs))
	}
	if fvs[0].GetField() != "name" || fvs[0].GetDescription() != "required" {
		t.Errorf("first violation = %+v, want field=name desc=required", fvs[0])
	}
	if fvs[1].GetField() != "email" || fvs[1].GetDescription() != "invalid format" {
		t.Errorf("second violation = %+v, want field=email desc=invalid format", fvs[1])
	}
}

func TestFieldViolationsFromError_ValidationError(t *testing.T) {
	violations := []FieldViolation{
		{Field: "name", Description: "required"},
	}
	err := NewValidationError(codes.InvalidArgument, pb.ErrorReason_ERROR_REASON_VALIDATION_FAILED, "bad", violations)
	extracted := FieldViolationsFromError(err)
	if len(extracted) != 1 {
		t.Fatalf("violations count = %d, want 1", len(extracted))
	}
	if extracted[0].Field != "name" || extracted[0].Description != "required" {
		t.Errorf("violation = %+v, want field=name desc=required", extracted[0])
	}
}

func TestFieldViolationsFromError_PlainError(t *testing.T) {
	err := status.Error(codes.Internal, "boom")
	violations := FieldViolationsFromError(err)
	if violations != nil {
		t.Errorf("violations = %v, want nil", violations)
	}
}

func TestNewPreconditionError(t *testing.T) {
	violations := []PreconditionViolation{
		{Type: "CONFIG_CONFLICT", Subject: "/etc/app.yaml", Description: "hash mismatch"},
	}
	err := NewPreconditionError(codes.FailedPrecondition, pb.ErrorReason_ERROR_REASON_CONFIG_CONFLICT, "conflict", violations)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("code = %v, want FailedPrecondition", st.Code())
	}
	var foundPF *errdetails.PreconditionFailure
	var foundInfo *errdetails.ErrorInfo
	for _, d := range st.Details() {
		if pf, ok := d.(*errdetails.PreconditionFailure); ok {
			foundPF = pf
		}
		if info, ok := d.(*errdetails.ErrorInfo); ok {
			foundInfo = info
		}
	}
	if foundInfo == nil {
		t.Fatal("no ErrorInfo detail found")
	}
	if foundPF == nil {
		t.Fatal("no PreconditionFailure detail found")
	}
	pvs := foundPF.GetViolations()
	if len(pvs) != 1 {
		t.Fatalf("violations count = %d, want 1", len(pvs))
	}
	if pvs[0].GetType() != "CONFIG_CONFLICT" || pvs[0].GetSubject() != "/etc/app.yaml" || pvs[0].GetDescription() != "hash mismatch" {
		t.Errorf("violation = %+v, unexpected", pvs[0])
	}
}

func TestPreconditionViolationsFromError_PreconditionError(t *testing.T) {
	violations := []PreconditionViolation{
		{Type: "CONFIG_CONFLICT", Subject: "/etc/app.yaml", Description: "hash mismatch"},
	}
	err := NewPreconditionError(codes.FailedPrecondition, pb.ErrorReason_ERROR_REASON_CONFIG_CONFLICT, "conflict", violations)
	extracted := PreconditionViolationsFromError(err)
	if len(extracted) != 1 {
		t.Fatalf("violations count = %d, want 1", len(extracted))
	}
	if extracted[0].Type != "CONFIG_CONFLICT" || extracted[0].Subject != "/etc/app.yaml" || extracted[0].Description != "hash mismatch" {
		t.Errorf("violation = %+v, unexpected", extracted[0])
	}
}

func TestPreconditionViolationsFromError_PlainError(t *testing.T) {
	err := status.Error(codes.Internal, "boom")
	violations := PreconditionViolationsFromError(err)
	if violations != nil {
		t.Errorf("violations = %v, want nil", violations)
	}
}

func TestReasonName(t *testing.T) {
	tests := []struct {
		reason pb.ErrorReason
		want   string
	}{
		{pb.ErrorReason_ERROR_REASON_UNSPECIFIED, ""},
		{pb.ErrorReason_ERROR_REASON_TOKEN_EXPIRED, "TOKEN_EXPIRED"},
		{pb.ErrorReason_ERROR_REASON_RESOURCE_NOT_FOUND, "RESOURCE_NOT_FOUND"},
		{pb.ErrorReason_ERROR_REASON_CONFIG_CONFLICT, "CONFIG_CONFLICT"},
	}
	for _, tt := range tests {
		got := ReasonName(tt.reason)
		if got != tt.want {
			t.Errorf("ReasonName(%v) = %q, want %q", tt.reason, got, tt.want)
		}
	}
}
