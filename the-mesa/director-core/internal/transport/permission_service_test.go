package transport

import (
	"context"
	"testing"
	"time"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/auth"
	"github.com/charviki/maze/the-mesa/director-core/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockPermissionApplicationService struct {
	createPermissionApplication func(context.Context, service.CreatePermissionApplicationInput) (service.PermissionApplication, error)
	listPermissionApplications  func(context.Context, service.ListPermissionApplicationsInput) ([]service.PermissionApplication, int64, error)
	getPermissionApplication    func(context.Context, string) (service.PermissionApplication, error)
	reviewPermissionApplication func(context.Context, service.ReviewPermissionApplicationInput) (service.PermissionApplication, error)
	revokePermissionApplication func(context.Context, service.RevokePermissionApplicationInput) (service.PermissionApplication, error)
	listSubjectPermissions      func(context.Context, string) ([]service.PermissionGrant, error)
}

func (m *mockPermissionApplicationService) CreatePermissionApplication(ctx context.Context, input service.CreatePermissionApplicationInput) (service.PermissionApplication, error) {
	return m.createPermissionApplication(ctx, input)
}

func (m *mockPermissionApplicationService) ListPermissionApplications(ctx context.Context, input service.ListPermissionApplicationsInput) ([]service.PermissionApplication, int64, error) {
	if m.listPermissionApplications == nil {
		return nil, 0, nil
	}
	return m.listPermissionApplications(ctx, input)
}

func (m *mockPermissionApplicationService) GetPermissionApplication(ctx context.Context, permissionApplicationID string) (service.PermissionApplication, error) {
	if m.getPermissionApplication == nil {
		return service.PermissionApplication{}, nil
	}
	return m.getPermissionApplication(ctx, permissionApplicationID)
}

func (m *mockPermissionApplicationService) ReviewPermissionApplication(ctx context.Context, input service.ReviewPermissionApplicationInput) (service.PermissionApplication, error) {
	if m.reviewPermissionApplication == nil {
		return service.PermissionApplication{}, nil
	}
	return m.reviewPermissionApplication(ctx, input)
}

func (m *mockPermissionApplicationService) RevokePermissionApplication(ctx context.Context, input service.RevokePermissionApplicationInput) (service.PermissionApplication, error) {
	if m.revokePermissionApplication == nil {
		return service.PermissionApplication{}, nil
	}
	return m.revokePermissionApplication(ctx, input)
}

func (m *mockPermissionApplicationService) ListSubjectPermissions(ctx context.Context, subjectKey string) ([]service.PermissionGrant, error) {
	return m.listSubjectPermissions(ctx, subjectKey)
}

func TestPermissionServiceServerCreatePermissionApplicationMapsRequest(t *testing.T) {
	expiresAt := time.Now().UTC().Add(10 * time.Minute).Round(time.Second)
	var captured service.CreatePermissionApplicationInput
	server := NewPermissionServiceServer(&mockPermissionApplicationService{
		createPermissionApplication: func(_ context.Context, input service.CreatePermissionApplicationInput) (service.PermissionApplication, error) {
			captured = input
			now := time.Now().UTC()
			return service.PermissionApplication{
				ID:         "pa_1",
				SubjectKey: input.SubjectKey,
				Targets:    input.Targets,
				Reason:     input.Reason,
				Status:     service.PermissionApplicationStatusPending,
				ExpiresAt:  input.ExpiresAt,
				CreatedAt:  now,
				UpdatedAt:  now,
			}, nil
		},
		listSubjectPermissions: func(context.Context, string) ([]service.PermissionGrant, error) { return nil, nil },
	})

	ctx := auth.WithUserInfo(context.Background(), &auth.UserInfo{SubjectKey: "user:admin"})
	resp, err := server.CreatePermissionApplication(ctx, &pb.CreatePermissionApplicationRequest{
		SubjectKey: "host:test",
		Targets: []*pb.PermissionTarget{
			{Resource: "host/*", Action: "read"},
		},
		Reason:    "need access",
		ExpiresAt: timestamppb.New(expiresAt),
	})
	if err != nil {
		t.Fatalf("create permission application: %v", err)
	}

	if captured.ActorSubjectKey != "user:admin" {
		t.Fatalf("actor subject key = %q, want user:admin", captured.ActorSubjectKey)
	}
	if captured.SubjectKey != "host:test" {
		t.Fatalf("subject key = %q, want host:test", captured.SubjectKey)
	}
	if len(captured.Targets) != 1 || captured.Targets[0].Resource != "host/*" || captured.Targets[0].Action != "read" {
		t.Fatalf("targets = %+v, want host/* read", captured.Targets)
	}
	if captured.ExpiresAt == nil || !captured.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("expires_at = %v, want %v", captured.ExpiresAt, expiresAt)
	}
	if resp.GetId() != "pa_1" {
		t.Fatalf("response id = %q, want pa_1", resp.GetId())
	}
}

func TestPermissionServiceServerCreatePermissionApplicationMapsValidationError(t *testing.T) {
	server := NewPermissionServiceServer(&mockPermissionApplicationService{
		createPermissionApplication: func(context.Context, service.CreatePermissionApplicationInput) (service.PermissionApplication, error) {
			return service.PermissionApplication{}, service.NewValidationError("reason is required")
		},
		listSubjectPermissions: func(context.Context, string) ([]service.PermissionGrant, error) { return nil, nil },
	})

	_, err := server.CreatePermissionApplication(context.Background(), &pb.CreatePermissionApplicationRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("grpc code = %s, want %s", status.Code(err), codes.InvalidArgument)
	}
}

func TestPermissionServiceServerListPermissionApplicationsMapsRequestAndResponse(t *testing.T) {
	var captured service.ListPermissionApplicationsInput
	server := NewPermissionServiceServer(&mockPermissionApplicationService{
		listPermissionApplications: func(_ context.Context, input service.ListPermissionApplicationsInput) ([]service.PermissionApplication, int64, error) {
			captured = input
			return []service.PermissionApplication{
				permissionApplicationFixture("pa_1"),
				permissionApplicationFixture("pa_2"),
			}, 2, nil
		},
	})

	resp, err := server.ListPermissionApplications(context.Background(), &pb.ListPermissionApplicationsRequest{
		Status:   "approved",
		Page:     3,
		PageSize: 50,
	})
	if err != nil {
		t.Fatalf("list permission applications: %v", err)
	}
	if captured.Status != "approved" || captured.Page != 3 || captured.PageSize != 50 {
		t.Fatalf("captured request = %+v, want status/page/page_size mapped", captured)
	}
	if resp.GetTotal() != 2 {
		t.Fatalf("total = %d, want 2", resp.GetTotal())
	}
	if len(resp.GetApplications()) != 2 {
		t.Fatalf("application count = %d, want 2", len(resp.GetApplications()))
	}
	if resp.GetApplications()[0].GetId() != "pa_1" || resp.GetApplications()[1].GetId() != "pa_2" {
		t.Fatalf("applications = %+v, want mapped ids", resp.GetApplications())
	}
}

func TestPermissionServiceServerGetPermissionApplicationRequiresID(t *testing.T) {
	server := NewPermissionServiceServer(&mockPermissionApplicationService{})

	_, err := server.GetPermissionApplication(context.Background(), &pb.GetPermissionApplicationRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("grpc code = %s, want %s", status.Code(err), codes.InvalidArgument)
	}
}

func TestPermissionServiceServerGetPermissionApplicationMapsNotFound(t *testing.T) {
	server := NewPermissionServiceServer(&mockPermissionApplicationService{
		getPermissionApplication: func(context.Context, string) (service.PermissionApplication, error) {
			return service.PermissionApplication{}, service.ErrPermissionApplicationNotFound
		},
	})

	_, err := server.GetPermissionApplication(context.Background(), &pb.GetPermissionApplicationRequest{
		PermissionApplicationId: "pa_missing",
	})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("grpc code = %s, want %s", status.Code(err), codes.NotFound)
	}
}

func TestPermissionServiceServerGetPermissionApplicationPreservesExistingStatusError(t *testing.T) {
	server := NewPermissionServiceServer(&mockPermissionApplicationService{
		getPermissionApplication: func(context.Context, string) (service.PermissionApplication, error) {
			return service.PermissionApplication{}, status.Error(codes.PermissionDenied, "denied")
		},
	})

	_, err := server.GetPermissionApplication(context.Background(), &pb.GetPermissionApplicationRequest{
		PermissionApplicationId: "pa_1",
	})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("grpc code = %s, want %s", status.Code(err), codes.PermissionDenied)
	}
}

func TestPermissionServiceServerReviewPermissionApplicationRequiresID(t *testing.T) {
	server := NewPermissionServiceServer(&mockPermissionApplicationService{})

	_, err := server.ReviewPermissionApplication(context.Background(), &pb.ReviewPermissionApplicationRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("grpc code = %s, want %s", status.Code(err), codes.InvalidArgument)
	}
}

func TestPermissionServiceServerReviewPermissionApplicationMapsID(t *testing.T) {
	var captured service.ReviewPermissionApplicationInput
	server := NewPermissionServiceServer(&mockPermissionApplicationService{
		reviewPermissionApplication: func(_ context.Context, input service.ReviewPermissionApplicationInput) (service.PermissionApplication, error) {
			captured = input
			return permissionApplicationFixture("pa_1"), nil
		},
	})

	ctx := auth.WithUserInfo(context.Background(), &auth.UserInfo{SubjectKey: "user:admin"})
	_, err := server.ReviewPermissionApplication(ctx, &pb.ReviewPermissionApplicationRequest{
		PermissionApplicationId: " pa_1 ",
		Approved:                true,
		ReviewComment:           "approved",
	})
	if err != nil {
		t.Fatalf("review permission application: %v", err)
	}
	if captured.PermissionApplicationID != "pa_1" {
		t.Fatalf("permission application id = %q, want pa_1", captured.PermissionApplicationID)
	}
	if captured.ReviewerSubjectKey != "user:admin" {
		t.Fatalf("reviewer subject key = %q, want user:admin", captured.ReviewerSubjectKey)
	}
}

func TestPermissionServiceServerRevokePermissionApplicationRequiresID(t *testing.T) {
	server := NewPermissionServiceServer(&mockPermissionApplicationService{})

	_, err := server.RevokePermissionApplication(context.Background(), &pb.RevokePermissionApplicationRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("grpc code = %s, want %s", status.Code(err), codes.InvalidArgument)
	}
}

func TestPermissionServiceServerListSubjectPermissionsMapsResponse(t *testing.T) {
	expiresAt := time.Now().UTC().Add(5 * time.Minute).Round(time.Second)
	createdAt := time.Now().UTC().Add(-1 * time.Minute).Round(time.Second)
	updatedAt := time.Now().UTC().Round(time.Second)

	server := NewPermissionServiceServer(&mockPermissionApplicationService{
		createPermissionApplication: func(context.Context, service.CreatePermissionApplicationInput) (service.PermissionApplication, error) {
			return service.PermissionApplication{}, nil
		},
		listSubjectPermissions: func(_ context.Context, subjectKey string) ([]service.PermissionGrant, error) {
			if subjectKey != "host:test" {
				t.Fatalf("subject key = %q, want host:test", subjectKey)
			}
			return []service.PermissionGrant{
				{
					ID:                            "pg_1",
					SubjectKey:                    "host:test",
					Resource:                      "host/*",
					Action:                        "read",
					SourcePermissionApplicationID: "pa_1",
					Status:                        service.PermissionGrantStatusActive,
					ExpiresAt:                     &expiresAt,
					CreatedAt:                     createdAt,
					UpdatedAt:                     updatedAt,
				},
			}, nil
		},
	})

	resp, err := server.ListSubjectPermissions(context.Background(), &pb.ListSubjectPermissionsRequest{SubjectKey: "host:test"})
	if err != nil {
		t.Fatalf("list subject permissions: %v", err)
	}
	if len(resp.GetGrants()) != 1 {
		t.Fatalf("grant count = %d, want 1", len(resp.GetGrants()))
	}
	grant := resp.GetGrants()[0]
	if grant.GetGrantId() != "pg_1" || grant.GetSourcePermissionApplicationId() != "pa_1" {
		t.Fatalf("grant = %+v, want mapped ids", grant)
	}
	if grant.GetExpiresAt().AsTime() != expiresAt || grant.GetCreatedAt().AsTime() != createdAt || grant.GetUpdatedAt().AsTime() != updatedAt {
		t.Fatalf("timestamps not mapped correctly: %+v", grant)
	}
}

func permissionApplicationFixture(id string) service.PermissionApplication {
	now := time.Now().UTC()
	return service.PermissionApplication{
		ID:         id,
		SubjectKey: "host:test",
		Targets: []service.PermissionTarget{
			{Resource: "host/*", Action: "read"},
		},
		Reason:    "need access",
		Status:    service.PermissionApplicationStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
