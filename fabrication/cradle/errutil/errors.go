package errutil

import (
	"strings"

	pb "github.com/charviki/maze/fabrication/cradle/api/gen/maze/v1"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FieldViolation 描述单个字段的校验失败。
type FieldViolation struct {
	Field       string
	Description string
}

// PreconditionViolation 描述单个前置条件的违反。
type PreconditionViolation struct {
	Type        string
	Subject     string
	Description string
}

// NewError constructs a gRPC status error with an ErrorInfo detail.
func NewError(code codes.Code, reason pb.ErrorReason, msg string) error {
	st := status.New(code, msg)
	withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
		Reason: reason.String(),
		Domain: "maze.v1",
	})
	if err != nil {
		return st.Err()
	}
	return withDetails.Err()
}

// NewValidationError constructs a gRPC status error with ErrorInfo and BadRequest details.
func NewValidationError(code codes.Code, reason pb.ErrorReason, msg string, violations []FieldViolation) error {
	br := &errdetails.BadRequest{}
	for _, v := range violations {
		br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
			Field:       v.Field,
			Description: v.Description,
		})
	}
	st := status.New(code, msg)
	withDetails, err := st.WithDetails(
		&errdetails.ErrorInfo{Reason: reason.String(), Domain: "maze.v1"},
		br,
	)
	if err != nil {
		return st.Err()
	}
	return withDetails.Err()
}

// NewPreconditionError constructs a gRPC status error with ErrorInfo and PreconditionFailure details.
func NewPreconditionError(code codes.Code, reason pb.ErrorReason, msg string, violations []PreconditionViolation) error {
	pf := &errdetails.PreconditionFailure{}
	for _, v := range violations {
		pf.Violations = append(pf.Violations, &errdetails.PreconditionFailure_Violation{
			Type:        v.Type,
			Subject:     v.Subject,
			Description: v.Description,
		})
	}
	st := status.New(code, msg)
	withDetails, err := st.WithDetails(
		&errdetails.ErrorInfo{Reason: reason.String(), Domain: "maze.v1"},
		pf,
	)
	if err != nil {
		return st.Err()
	}
	return withDetails.Err()
}

// ReasonFromError extracts the ErrorReason from a gRPC status error.
func ReasonFromError(err error) pb.ErrorReason {
	st, ok := status.FromError(err)
	if !ok {
		return pb.ErrorReason_ERROR_REASON_UNSPECIFIED
	}
	for _, detail := range st.Details() {
		info, ok := detail.(*errdetails.ErrorInfo)
		if !ok {
			continue
		}
		reason := info.GetReason()
		if v, ok := pb.ErrorReason_value[reason]; ok {
			return pb.ErrorReason(v)
		}
	}
	return pb.ErrorReason_ERROR_REASON_UNSPECIFIED
}

// FieldViolationsFromError extracts BadRequest field violations from a gRPC status error.
func FieldViolationsFromError(err error) []FieldViolation {
	st, ok := status.FromError(err)
	if !ok {
		return nil
	}
	for _, detail := range st.Details() {
		br, ok := detail.(*errdetails.BadRequest)
		if !ok {
			continue
		}
		var result []FieldViolation
		for _, fv := range br.GetFieldViolations() {
			result = append(result, FieldViolation{
				Field:       fv.GetField(),
				Description: fv.GetDescription(),
			})
		}
		return result
	}
	return nil
}

// PreconditionViolationsFromError extracts PreconditionFailure violations from a gRPC status error.
func PreconditionViolationsFromError(err error) []PreconditionViolation {
	st, ok := status.FromError(err)
	if !ok {
		return nil
	}
	for _, detail := range st.Details() {
		pf, ok := detail.(*errdetails.PreconditionFailure)
		if !ok {
			continue
		}
		var result []PreconditionViolation
		for _, v := range pf.GetViolations() {
			result = append(result, PreconditionViolation{
				Type:        v.GetType(),
				Subject:     v.GetSubject(),
				Description: v.GetDescription(),
			})
		}
		return result
	}
	return nil
}

// ReasonName converts an ErrorReason enum value to its short name string
// (without the ERROR_REASON_ prefix). UNSPECIFIED returns an empty string.
func ReasonName(r pb.ErrorReason) string {
	if r == pb.ErrorReason_ERROR_REASON_UNSPECIFIED {
		return ""
	}
	return strings.TrimPrefix(r.String(), "ERROR_REASON_")
}
