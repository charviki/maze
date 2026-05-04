package transport

import (
	"context"
	"errors"
	"strings"
	"time"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/auth"
	"github.com/charviki/mesa-hub-behavior-panel/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PermissionApplicationService 抽象 transport 层依赖的权限业务能力，便于单测替换。
type PermissionApplicationService interface {
	CreatePermissionApplication(context.Context, service.CreatePermissionApplicationInput) (service.PermissionApplication, error)
	ListPermissionApplications(context.Context, service.ListPermissionApplicationsInput) ([]service.PermissionApplication, int64, error)
	GetPermissionApplication(context.Context, string) (service.PermissionApplication, error)
	ReviewPermissionApplication(context.Context, service.ReviewPermissionApplicationInput) (service.PermissionApplication, error)
	RevokePermissionApplication(context.Context, service.RevokePermissionApplicationInput) (service.PermissionApplication, error)
	ListSubjectPermissions(context.Context, string) ([]service.PermissionGrant, error)
}

// PermissionServiceServer 实现 PermissionService gRPC 接口。
type PermissionServiceServer struct {
	pb.UnimplementedPermissionServiceServer
	permissionService PermissionApplicationService
}

// NewPermissionServiceServer 创建 PermissionServiceServer。
func NewPermissionServiceServer(permissionService PermissionApplicationService) *PermissionServiceServer {
	return &PermissionServiceServer{permissionService: permissionService}
}

// CreatePermissionApplication 创建权限申请单。
func (s *PermissionServiceServer) CreatePermissionApplication(ctx context.Context, req *pb.CreatePermissionApplicationRequest) (*pb.PermissionApplication, error) {
	var expiresAtPtr *time.Time
	if req.GetExpiresAt() != nil {
		expiresAt := req.GetExpiresAt().AsTime().UTC()
		expiresAtPtr = &expiresAt
	}

	targets := make([]service.PermissionTarget, 0, len(req.GetTargets()))
	for _, target := range req.GetTargets() {
		targets = append(targets, service.PermissionTarget{
			Resource: target.GetResource(),
			Action:   target.GetAction(),
		})
	}

	application, err := s.permissionService.CreatePermissionApplication(ctx, service.CreatePermissionApplicationInput{
		SubjectKey:      req.GetSubjectKey(),
		Targets:         targets,
		Reason:          req.GetReason(),
		ExpiresAt:       expiresAtPtr,
		ActorSubjectKey: subjectKeyFromContext(ctx),
	})
	if err != nil {
		return nil, toStatusError(err)
	}
	return toPBPermissionApplication(application), nil
}

// ListPermissionApplications 列出权限申请单。
func (s *PermissionServiceServer) ListPermissionApplications(ctx context.Context, req *pb.ListPermissionApplicationsRequest) (*pb.ListPermissionApplicationsResponse, error) {
	applications, total, err := s.permissionService.ListPermissionApplications(ctx, service.ListPermissionApplicationsInput{
		Status:   req.GetStatus(),
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		return nil, toStatusError(err)
	}

	items := make([]*pb.PermissionApplication, 0, len(applications))
	for _, application := range applications {
		items = append(items, toPBPermissionApplication(application))
	}
	return &pb.ListPermissionApplicationsResponse{
		Applications: items,
		Total:        safeCountInt32(total),
	}, nil
}

// GetPermissionApplication 获取单个权限申请单。
func (s *PermissionServiceServer) GetPermissionApplication(ctx context.Context, req *pb.GetPermissionApplicationRequest) (*pb.PermissionApplication, error) {
	permissionApplicationID := strings.TrimSpace(req.GetPermissionApplicationId())
	if permissionApplicationID == "" {
		return nil, status.Error(codes.InvalidArgument, "permission application id is required")
	}

	application, err := s.permissionService.GetPermissionApplication(ctx, permissionApplicationID)
	if err != nil {
		return nil, toStatusError(err)
	}
	return toPBPermissionApplication(application), nil
}

// ReviewPermissionApplication 审批权限申请单。
func (s *PermissionServiceServer) ReviewPermissionApplication(ctx context.Context, req *pb.ReviewPermissionApplicationRequest) (*pb.PermissionApplication, error) {
	permissionApplicationID := strings.TrimSpace(req.GetPermissionApplicationId())
	if permissionApplicationID == "" {
		return nil, status.Error(codes.InvalidArgument, "permission application id is required")
	}

	application, err := s.permissionService.ReviewPermissionApplication(ctx, service.ReviewPermissionApplicationInput{
		PermissionApplicationID: permissionApplicationID,
		Approved:                req.GetApproved(),
		ReviewComment:           req.GetReviewComment(),
		ReviewerSubjectKey:      subjectKeyFromContext(ctx),
	})
	if err != nil {
		return nil, toStatusError(err)
	}
	return toPBPermissionApplication(application), nil
}

// RevokePermissionApplication 撤销权限申请单对应的授权结果。
func (s *PermissionServiceServer) RevokePermissionApplication(ctx context.Context, req *pb.RevokePermissionApplicationRequest) (*pb.PermissionApplication, error) {
	permissionApplicationID := strings.TrimSpace(req.GetPermissionApplicationId())
	if permissionApplicationID == "" {
		return nil, status.Error(codes.InvalidArgument, "permission application id is required")
	}

	application, err := s.permissionService.RevokePermissionApplication(ctx, service.RevokePermissionApplicationInput{
		PermissionApplicationID: permissionApplicationID,
		RevokeReason:            req.GetRevokeReason(),
		ReviewerSubjectKey:      subjectKeyFromContext(ctx),
	})
	if err != nil {
		return nil, toStatusError(err)
	}
	return toPBPermissionApplication(application), nil
}

// ListSubjectPermissions 列出主体当前生效的授权结果。
func (s *PermissionServiceServer) ListSubjectPermissions(ctx context.Context, req *pb.ListSubjectPermissionsRequest) (*pb.ListSubjectPermissionsResponse, error) {
	grants, err := s.permissionService.ListSubjectPermissions(ctx, req.GetSubjectKey())
	if err != nil {
		return nil, toStatusError(err)
	}

	items := make([]*pb.SubjectPermissionGrant, 0, len(grants))
	for _, grant := range grants {
		items = append(items, &pb.SubjectPermissionGrant{
			GrantId:                       grant.ID,
			SubjectKey:                    grant.SubjectKey,
			Resource:                      grant.Resource,
			Action:                        grant.Action,
			SourcePermissionApplicationId: grant.SourcePermissionApplicationID,
			Status:                        string(grant.Status),
			ExpiresAt:                     timestampPtr(grant.ExpiresAt),
			CreatedAt:                     timestamppb.New(grant.CreatedAt.UTC()),
			UpdatedAt:                     timestamppb.New(grant.UpdatedAt.UTC()),
		})
	}

	return &pb.ListSubjectPermissionsResponse{
		SubjectKey: req.GetSubjectKey(),
		Grants:     items,
	}, nil
}

func toPBPermissionApplication(application service.PermissionApplication) *pb.PermissionApplication {
	targets := make([]*pb.PermissionTarget, 0, len(application.Targets))
	for _, target := range application.Targets {
		targets = append(targets, &pb.PermissionTarget{
			Resource: target.Resource,
			Action:   target.Action,
		})
	}

	return &pb.PermissionApplication{
		Id:            application.ID,
		SubjectKey:    application.SubjectKey,
		Targets:       targets,
		Reason:        application.Reason,
		Status:        toPBApplicationStatus(application.Status),
		ReviewedBy:    application.ReviewedBy,
		ReviewComment: application.ReviewComment,
		ExpiresAt:     timestampPtr(application.ExpiresAt),
		CreatedAt:     timestamppb.New(application.CreatedAt.UTC()),
		UpdatedAt:     timestamppb.New(application.UpdatedAt.UTC()),
	}
}

func toPBApplicationStatus(statusValue service.PermissionApplicationStatus) pb.PermissionApplicationStatus {
	switch statusValue {
	case service.PermissionApplicationStatusPending:
		return pb.PermissionApplicationStatus_PERMISSION_APPLICATION_STATUS_PENDING
	case service.PermissionApplicationStatusApproved:
		return pb.PermissionApplicationStatus_PERMISSION_APPLICATION_STATUS_APPROVED
	case service.PermissionApplicationStatusDenied:
		return pb.PermissionApplicationStatus_PERMISSION_APPLICATION_STATUS_DENIED
	case service.PermissionApplicationStatusRevoked:
		return pb.PermissionApplicationStatus_PERMISSION_APPLICATION_STATUS_REVOKED
	case service.PermissionApplicationStatusExpired:
		return pb.PermissionApplicationStatus_PERMISSION_APPLICATION_STATUS_EXPIRED
	default:
		return pb.PermissionApplicationStatus_PERMISSION_APPLICATION_STATUS_UNSPECIFIED
	}
}

func timestampPtr(value *time.Time) *timestamppb.Timestamp {
	if value == nil {
		return nil
	}
	return timestamppb.New(value.UTC())
}

func subjectKeyFromContext(ctx context.Context) string {
	userInfo := auth.GetUserInfo(ctx)
	if userInfo == nil {
		return ""
	}
	return userInfo.SubjectKey
}

func toStatusError(err error) error {
	switch {
	case err == nil:
		return nil
	case status.Code(err) != codes.Unknown:
		return err
	case errors.Is(err, service.ErrPermissionApplicationNotFound), errors.Is(err, service.ErrPermissionGrantNotFound):
		return status.Error(codes.NotFound, err.Error())
	case isValidationError(err):
		return status.Error(codes.InvalidArgument, err.Error())
	case isPreconditionError(err):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func isValidationError(err error) bool {
	var statusErr interface{ GRPCStatus() *status.Status }
	if errors.As(err, &statusErr) {
		return false
	}
	var validationErr service.ValidationError
	return errors.As(err, &validationErr)
}

func isPreconditionError(err error) bool {
	var statusErr interface{ GRPCStatus() *status.Status }
	if errors.As(err, &statusErr) {
		return false
	}
	var preconditionErr service.PreconditionError
	return errors.As(err, &preconditionErr)
}

func safeCountInt32(n int64) int32 {
	const maxInt32 = int64(^uint32(0) >> 1)
	const minInt32 = -maxInt32 - 1
	switch {
	case n > maxInt32:
		return int32(maxInt32)
	case n < minInt32:
		return int32(minInt32)
	default:
		return int32(n)
	}
}
