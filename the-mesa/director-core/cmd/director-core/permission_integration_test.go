package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pb "github.com/charviki/maze-cradle/api/gen/maze/v1"
	"github.com/charviki/maze-cradle/gatewayutil"
	"github.com/charviki/maze/the-mesa/director-core/internal/service"
	"github.com/charviki/maze/the-mesa/director-core/internal/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type workflowPermissionService struct {
	applications map[string]service.PermissionApplication
	grants       map[string][]service.PermissionGrant
	nextAppID    int
	nextGrantID  int
}

func newWorkflowPermissionService() *workflowPermissionService {
	return &workflowPermissionService{
		applications: map[string]service.PermissionApplication{},
		grants:       map[string][]service.PermissionGrant{},
		nextAppID:    1,
		nextGrantID:  1,
	}
}

func (s *workflowPermissionService) CreatePermissionApplication(_ context.Context, input service.CreatePermissionApplicationInput) (service.PermissionApplication, error) {
	// 这里保留最小业务校验，让 HTTP 集成测试能覆盖 gateway 对 InvalidArgument/NotFound 的映射。
	if input.SubjectKey == "" {
		return service.PermissionApplication{}, service.NewValidationError("subject key is required")
	}
	if len(input.Targets) == 0 {
		return service.PermissionApplication{}, service.NewValidationError("at least one permission target is required")
	}
	if input.Reason == "" {
		return service.PermissionApplication{}, service.NewValidationError("reason is required")
	}

	now := time.Now().UTC()
	id := "pa_" + string(rune('0'+s.nextAppID))
	s.nextAppID++
	application := service.PermissionApplication{
		InternalID: int64(s.nextAppID),
		ID:         id,
		SubjectKey: input.SubjectKey,
		Targets:    append([]service.PermissionTarget(nil), input.Targets...),
		Reason:     input.Reason,
		Status:     service.PermissionApplicationStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	s.applications[id] = application
	return application, nil
}

func (s *workflowPermissionService) ListPermissionApplications(_ context.Context, _ service.ListPermissionApplicationsInput) ([]service.PermissionApplication, int64, error) {
	items := make([]service.PermissionApplication, 0, len(s.applications))
	for _, application := range s.applications {
		items = append(items, application)
	}
	return items, int64(len(items)), nil
}

func (s *workflowPermissionService) GetPermissionApplication(_ context.Context, id string) (service.PermissionApplication, error) {
	application, ok := s.applications[id]
	if !ok {
		return service.PermissionApplication{}, service.ErrPermissionApplicationNotFound
	}
	return application, nil
}

func (s *workflowPermissionService) ReviewPermissionApplication(_ context.Context, input service.ReviewPermissionApplicationInput) (service.PermissionApplication, error) {
	application := s.applications[input.PermissionApplicationID]
	application.ReviewComment = input.ReviewComment
	application.ReviewedBy = input.ReviewerSubjectKey
	application.UpdatedAt = time.Now().UTC()
	if input.Approved {
		application.Status = service.PermissionApplicationStatusApproved
		grants := make([]service.PermissionGrant, 0, len(application.Targets))
		for _, target := range application.Targets {
			grant := service.PermissionGrant{
				InternalID:                            int64(s.nextGrantID),
				ID:                                    "pg_" + string(rune('0'+s.nextGrantID)),
				SubjectKey:                            application.SubjectKey,
				Resource:                              target.Resource,
				Action:                                target.Action,
				SourcePermissionApplicationInternalID: application.InternalID,
				SourcePermissionApplicationID:         application.ID,
				Status:                                service.PermissionGrantStatusActive,
				CreatedAt:                             time.Now().UTC(),
				UpdatedAt:                             time.Now().UTC(),
			}
			s.nextGrantID++
			grants = append(grants, grant)
		}
		s.grants[application.SubjectKey] = grants
	} else {
		application.Status = service.PermissionApplicationStatusDenied
	}
	s.applications[input.PermissionApplicationID] = application
	return application, nil
}

func (s *workflowPermissionService) RevokePermissionApplication(_ context.Context, input service.RevokePermissionApplicationInput) (service.PermissionApplication, error) {
	application := s.applications[input.PermissionApplicationID]
	application.Status = service.PermissionApplicationStatusRevoked
	application.ReviewComment = input.RevokeReason
	application.ReviewedBy = input.ReviewerSubjectKey
	application.UpdatedAt = time.Now().UTC()
	s.applications[input.PermissionApplicationID] = application
	delete(s.grants, application.SubjectKey)
	return application, nil
}

func (s *workflowPermissionService) ListSubjectPermissions(_ context.Context, subjectKey string) ([]service.PermissionGrant, error) {
	return s.grants[subjectKey], nil
}

func TestPermissionServiceHTTPWorkflow(t *testing.T) {
	grpcListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen grpc: %v", err)
	}
	t.Cleanup(func() { _ = grpcListener.Close() })

	grpcServer := grpc.NewServer()
	permissionService := newWorkflowPermissionService()
	pb.RegisterPermissionServiceServer(grpcServer, transport.NewPermissionServiceServer(permissionService))
	go func() {
		_ = grpcServer.Serve(grpcListener)
	}()
	t.Cleanup(grpcServer.Stop)

	gwmux := gatewayutil.NewServeMux()
	if err := pb.RegisterPermissionServiceHandlerFromEndpoint(context.Background(), gwmux, grpcListener.Addr().String(), []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}); err != nil {
		t.Fatalf("register permission gateway: %v", err)
	}

	httpServer := httptest.NewServer(gwmux)
	t.Cleanup(httpServer.Close)

	application := postJSON[map[string]any](t, httpServer.Client(), httpServer.URL+"/api/v1/permission-applications", map[string]any{
		"subjectKey": "host:test",
		"targets": []map[string]string{
			{"resource": "host/*", "action": "read"},
		},
		"reason": "need access",
	})
	id, _ := application["id"].(string)
	if id == "" {
		t.Fatalf("created application id is empty: %+v", application)
	}

	postJSON[map[string]any](t, httpServer.Client(), httpServer.URL+"/api/v1/permission-applications/"+id+":review", map[string]any{
		"approved":      true,
		"reviewComment": "approved",
	})

	grantsAfterReview := getJSON[map[string]any](t, httpServer.Client(), httpServer.URL+"/api/v1/subjects/host:test/permissions")
	if got := len(anySlice(grantsAfterReview["grants"])); got != 1 {
		t.Fatalf("grant count after review = %d, want 1", got)
	}

	postJSON[map[string]any](t, httpServer.Client(), httpServer.URL+"/api/v1/permission-applications/"+id+":revoke", map[string]any{
		"revokeReason": "cleanup",
	})

	grantsAfterRevoke := getJSON[map[string]any](t, httpServer.Client(), httpServer.URL+"/api/v1/subjects/host:test/permissions")
	if got := len(anySlice(grantsAfterRevoke["grants"])); got != 0 {
		t.Fatalf("grant count after revoke = %d, want 0", got)
	}
}

func TestPermissionServiceHTTPCreateValidationError(t *testing.T) {
	httpServer := newPermissionGatewayTestServer(t, newWorkflowPermissionService())

	resp, err := httpServer.Client().Post(httpServer.URL+"/api/v1/permission-applications", "application/json", bytes.NewReader([]byte(`{}`)))
	if err != nil {
		t.Fatalf("post create permission application: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want %d, body = %s", resp.StatusCode, http.StatusBadRequest, string(raw))
	}
}

func TestPermissionServiceHTTPGetNotFound(t *testing.T) {
	httpServer := newPermissionGatewayTestServer(t, newWorkflowPermissionService())

	resp, err := httpServer.Client().Get(httpServer.URL + "/api/v1/permission-applications/pa_missing")
	if err != nil {
		t.Fatalf("get permission application: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want %d, body = %s", resp.StatusCode, http.StatusNotFound, string(raw))
	}
}

func newPermissionGatewayTestServer(t *testing.T, permissionService *workflowPermissionService) *httptest.Server {
	t.Helper()

	grpcListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen grpc: %v", err)
	}
	t.Cleanup(func() { _ = grpcListener.Close() })

	grpcServer := grpc.NewServer()
	pb.RegisterPermissionServiceServer(grpcServer, transport.NewPermissionServiceServer(permissionService))
	go func() {
		_ = grpcServer.Serve(grpcListener)
	}()
	t.Cleanup(grpcServer.Stop)

	gwmux := gatewayutil.NewServeMux()
	if err := pb.RegisterPermissionServiceHandlerFromEndpoint(context.Background(), gwmux, grpcListener.Addr().String(), []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}); err != nil {
		t.Fatalf("register permission gateway: %v", err)
	}

	httpServer := httptest.NewServer(gwmux)
	t.Cleanup(httpServer.Close)
	return httpServer
}

func postJSON[T any](t *testing.T, client *http.Client, url string, payload any) T {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("post %s status = %d, body = %s", url, resp.StatusCode, string(raw))
	}
	return decodeJSON[T](t, resp.Body)
}

func getJSON[T any](t *testing.T, client *http.Client, url string) T {
	t.Helper()

	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("get %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("get %s status = %d, body = %s", url, resp.StatusCode, string(raw))
	}
	return decodeJSON[T](t, resp.Body)
}

func decodeJSON[T any](t *testing.T, body io.Reader) T {
	t.Helper()

	var result T
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	return result
}

func anySlice(value any) []any {
	items, _ := value.([]any)
	return items
}
