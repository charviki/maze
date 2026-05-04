package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

type memoryPermissionRepository struct {
	subjects     map[string]Subject
	applications map[string]PermissionApplication
	grants       map[string]PermissionGrant
	casbinRules  map[int64]CasbinRule
	auditLogs    []AuditLogEntry
	nextAppID    int
	nextGrantID  int
	nextRuleID   int64
	reviewErr    error
	grantErr     error
}

func newMemoryPermissionRepository() *memoryPermissionRepository {
	return &memoryPermissionRepository{
		subjects:     map[string]Subject{},
		applications: map[string]PermissionApplication{},
		grants:       map[string]PermissionGrant{},
		casbinRules:  map[int64]CasbinRule{},
		nextAppID:    1,
		nextGrantID:  1,
		nextRuleID:   1,
	}
}

func (r *memoryPermissionRepository) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func (r *memoryPermissionRepository) UpsertSubject(_ context.Context, subject Subject) (Subject, error) {
	now := time.Now().UTC()
	current, ok := r.subjects[subject.SubjectKey]
	if ok {
		current.SubjectType = subject.SubjectType
		current.DisplayName = subject.DisplayName
		current.IsSystem = subject.IsSystem
		current.UpdatedAt = now
		r.subjects[subject.SubjectKey] = current
		return current, nil
	}
	subject.CreatedAt = now
	subject.UpdatedAt = now
	r.subjects[subject.SubjectKey] = subject
	return subject, nil
}

func (r *memoryPermissionRepository) CreatePermissionApplication(_ context.Context, params CreatePermissionApplicationParams) (PermissionApplication, error) {
	now := time.Now().UTC()
	internalID := int64(r.nextAppID)
	id := appID(int(internalID))
	r.nextAppID++

	application := PermissionApplication{
		InternalID: internalID,
		ID:         id,
		SubjectKey: params.SubjectKey,
		Targets:    append([]PermissionTarget(nil), params.Targets...),
		Reason:     params.Reason,
		Status:     PermissionApplicationStatusPending,
		ExpiresAt:  params.ExpiresAt,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	r.applications[id] = application
	return application, nil
}

func (r *memoryPermissionRepository) ListPermissionApplications(_ context.Context, params ListPermissionApplicationsParams) ([]PermissionApplication, int64, error) {
	items := make([]PermissionApplication, 0, len(r.applications))
	for _, application := range r.applications {
		if params.Status != "" && application.Status != params.Status {
			continue
		}
		items = append(items, application)
	}
	return items, int64(len(items)), nil
}

func (r *memoryPermissionRepository) GetPermissionApplication(_ context.Context, id string) (PermissionApplication, error) {
	application, ok := r.applications[id]
	if !ok {
		return PermissionApplication{}, ErrPermissionApplicationNotFound
	}
	return application, nil
}

func (r *memoryPermissionRepository) ReviewPermissionApplication(_ context.Context, params ReviewPermissionApplicationParams) (PermissionApplication, error) {
	if r.reviewErr != nil {
		return PermissionApplication{}, r.reviewErr
	}
	application, key, ok := r.lookupApplication(params.InternalID)
	if !ok {
		return PermissionApplication{}, ErrPermissionApplicationNotFound
	}
	if params.ExpectedStatus != "" && application.Status != params.ExpectedStatus {
		return PermissionApplication{}, ErrPermissionApplicationStateChanged
	}
	now := time.Now().UTC()
	application.Status = params.Status
	application.ReviewedBy = params.ReviewedBy
	application.ReviewComment = params.ReviewComment
	application.ReviewedAt = &now
	application.UpdatedAt = now
	r.applications[key] = application
	return application, nil
}

func (r *memoryPermissionRepository) UpdatePermissionApplicationStatus(_ context.Context, params UpdatePermissionApplicationStatusParams) (PermissionApplication, error) {
	application, key, ok := r.lookupApplication(params.InternalID)
	if !ok {
		return PermissionApplication{}, ErrPermissionApplicationNotFound
	}
	now := time.Now().UTC()
	application.Status = params.Status
	if params.ReviewedBy != "" {
		application.ReviewedBy = params.ReviewedBy
		reviewedAt := now
		application.ReviewedAt = &reviewedAt
	}
	if params.ReviewComment != "" {
		application.ReviewComment = params.ReviewComment
	}
	application.UpdatedAt = now
	r.applications[key] = application
	return application, nil
}

func (r *memoryPermissionRepository) CreatePermissionGrant(_ context.Context, params CreatePermissionGrantParams) (PermissionGrant, error) {
	if r.grantErr != nil {
		return PermissionGrant{}, r.grantErr
	}
	now := time.Now().UTC()
	internalID := int64(r.nextGrantID)
	id := grantID(int(internalID))
	r.nextGrantID++
	grant := PermissionGrant{
		InternalID:                            internalID,
		ID:                                    id,
		SubjectKey:                            params.SubjectKey,
		Resource:                              params.Resource,
		Action:                                params.Action,
		SourcePermissionApplicationInternalID: params.SourcePermissionApplicationInternalID,
		SourcePermissionApplicationID:         appID(int(params.SourcePermissionApplicationInternalID)),
		Status:                                PermissionGrantStatusActive,
		ExpiresAt:                             params.ExpiresAt,
		CreatedAt:                             now,
		UpdatedAt:                             now,
	}
	r.grants[id] = grant
	return grant, nil
}

func (r *memoryPermissionRepository) AttachGrantCasbinRule(_ context.Context, grantInternalID int64, casbinRuleID int64) error {
	grant, key, ok := r.lookupGrant(grantInternalID)
	if !ok {
		return ErrPermissionGrantNotFound
	}
	grant.CasbinRuleID = &casbinRuleID
	grant.UpdatedAt = time.Now().UTC()
	r.grants[key] = grant
	return nil
}

func (r *memoryPermissionRepository) ListSubjectPermissionGrants(_ context.Context, subjectKey string) ([]PermissionGrant, error) {
	var grants []PermissionGrant
	for _, grant := range r.grants {
		if grant.SubjectKey == subjectKey && grant.Status == PermissionGrantStatusActive {
			grants = append(grants, grant)
		}
	}
	return grants, nil
}

func (r *memoryPermissionRepository) RevokeActivePermissionGrantsByApplication(_ context.Context, params RevokePermissionGrantsParams) ([]PermissionGrant, error) {
	var revoked []PermissionGrant
	for id, grant := range r.grants {
		if grant.SourcePermissionApplicationInternalID != params.SourcePermissionApplicationInternalID || grant.Status != PermissionGrantStatusActive {
			continue
		}
		grant.Status = params.Status
		grant.RevokedBy = params.RevokedBy
		grant.RevokedReason = params.RevokedReason
		grant.UpdatedAt = time.Now().UTC()
		r.grants[id] = grant
		revoked = append(revoked, grant)
	}
	return revoked, nil
}

func (r *memoryPermissionRepository) ListExpiredPermissionGrants(_ context.Context) ([]PermissionGrant, error) {
	now := time.Now().UTC()
	var grants []PermissionGrant
	for _, grant := range r.grants {
		if grant.Status == PermissionGrantStatusActive && grant.ExpiresAt != nil && !grant.ExpiresAt.After(now) {
			grants = append(grants, grant)
		}
	}
	return grants, nil
}

func (r *memoryPermissionRepository) ExpirePermissionGrant(_ context.Context, grantInternalID int64) error {
	grant, key, ok := r.lookupGrant(grantInternalID)
	if !ok {
		return ErrPermissionGrantNotFound
	}
	grant.Status = PermissionGrantStatusExpired
	grant.UpdatedAt = time.Now().UTC()
	r.grants[key] = grant
	return nil
}

func (r *memoryPermissionRepository) InsertAuditLog(_ context.Context, entry AuditLogEntry) error {
	r.auditLogs = append(r.auditLogs, entry)
	return nil
}

func (r *memoryPermissionRepository) InsertCasbinRule(_ context.Context, rule CasbinRule) (int64, error) {
	rule.ID = r.nextRuleID
	r.nextRuleID++
	r.casbinRules[rule.ID] = rule
	return rule.ID, nil
}

func (r *memoryPermissionRepository) DeleteCasbinRule(_ context.Context, id int64) error {
	delete(r.casbinRules, id)
	return nil
}

func (r *memoryPermissionRepository) ListAllCasbinRules(_ context.Context) ([]CasbinRule, error) {
	rules := make([]CasbinRule, 0, len(r.casbinRules))
	for _, rule := range r.casbinRules {
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *memoryPermissionRepository) EnsureAdminBootstrap(ctx context.Context, subjectKey, displayName string) error {
	if _, err := r.UpsertSubject(ctx, Subject{
		SubjectKey:  subjectKey,
		SubjectType: "user",
		DisplayName: displayName,
		IsSystem:    true,
	}); err != nil {
		return err
	}
	_, err := r.InsertCasbinRule(ctx, CasbinRule{
		Ptype: "p",
		V0:    subjectKey,
		V1:    "*",
		V2:    "*",
	})
	return err
}

func (r *memoryPermissionRepository) lookupApplication(internalID int64) (PermissionApplication, string, bool) {
	for key, application := range r.applications {
		if application.InternalID == internalID {
			return application, key, true
		}
	}
	return PermissionApplication{}, "", false
}

func (r *memoryPermissionRepository) lookupGrant(internalID int64) (PermissionGrant, string, bool) {
	for key, grant := range r.grants {
		if grant.InternalID == internalID {
			return grant, key, true
		}
	}
	return PermissionGrant{}, "", false
}

func TestPermissionServiceCreatePermissionApplicationRequiresReason(t *testing.T) {
	repo := newMemoryPermissionRepository()
	svc := NewPermissionService(repo, repo, nil)

	_, err := svc.CreatePermissionApplication(context.Background(), CreatePermissionApplicationInput{
		SubjectKey: "host:test",
		Targets: []PermissionTarget{
			{Resource: "host/*", Action: "read"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "reason is required") {
		t.Fatalf("expected reason validation error, got %v", err)
	}
}

func TestPermissionServiceCreatePermissionApplicationRejectsPastExpiry(t *testing.T) {
	repo := newMemoryPermissionRepository()
	svc := NewPermissionService(repo, repo, nil)

	expiresAt := time.Now().UTC().Add(-time.Minute)
	_, err := svc.CreatePermissionApplication(context.Background(), CreatePermissionApplicationInput{
		SubjectKey: "host:test",
		Targets: []PermissionTarget{
			{Resource: "host/*", Action: "read"},
		},
		Reason:    "temporary access",
		ExpiresAt: &expiresAt,
	})
	if err == nil || !strings.Contains(err.Error(), "expires at must be in the future") {
		t.Fatalf("expected expires_at validation error, got %v", err)
	}
}

func TestPermissionServiceCreatePermissionApplicationRejectsDuplicateTargets(t *testing.T) {
	repo := newMemoryPermissionRepository()
	svc := NewPermissionService(repo, repo, nil)

	_, err := svc.CreatePermissionApplication(context.Background(), CreatePermissionApplicationInput{
		SubjectKey: "host:test",
		Targets: []PermissionTarget{
			{Resource: "host/*", Action: "read"},
			{Resource: "host/*", Action: "read"},
		},
		Reason: "need access",
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate permission target") {
		t.Fatalf("expected duplicate target validation error, got %v", err)
	}
}

func TestPermissionServiceListPermissionApplicationsNormalizesStatus(t *testing.T) {
	repo := newMemoryPermissionRepository()
	svc := NewPermissionService(repo, repo, nil)

	pendingApplication, err := svc.CreatePermissionApplication(context.Background(), CreatePermissionApplicationInput{
		SubjectKey: "host:pending",
		Targets: []PermissionTarget{
			{Resource: "host/*", Action: "read"},
		},
		Reason: "need access",
	})
	if err != nil {
		t.Fatalf("create pending permission application: %v", err)
	}
	approvedApplication, err := svc.CreatePermissionApplication(context.Background(), CreatePermissionApplicationInput{
		SubjectKey: "host:approved",
		Targets: []PermissionTarget{
			{Resource: "host/*", Action: "read"},
		},
		Reason: "need access",
	})
	if err != nil {
		t.Fatalf("create approved permission application: %v", err)
	}
	if _, err := svc.ReviewPermissionApplication(context.Background(), ReviewPermissionApplicationInput{
		PermissionApplicationID: approvedApplication.ID,
		Approved:                true,
		ReviewComment:           "approved",
		ReviewerSubjectKey:      "user:admin",
	}); err != nil {
		t.Fatalf("approve permission application: %v", err)
	}

	applications, total, err := svc.ListPermissionApplications(context.Background(), ListPermissionApplicationsInput{
		Status:   "  APPROVED ",
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("list permission applications: %v", err)
	}
	if total != 1 {
		t.Fatalf("total = %d, want 1", total)
	}
	if len(applications) != 1 || applications[0].ID != approvedApplication.ID {
		t.Fatalf("applications = %+v, want only %s", applications, approvedApplication.ID)
	}
	if pendingApplication.Status != PermissionApplicationStatusPending {
		t.Fatalf("pending application status = %s, want pending", pendingApplication.Status)
	}
}

func TestPermissionServiceListPermissionApplicationsRejectsInvalidStatus(t *testing.T) {
	repo := newMemoryPermissionRepository()
	svc := NewPermissionService(repo, repo, nil)

	_, _, err := svc.ListPermissionApplications(context.Background(), ListPermissionApplicationsInput{
		Status: "processing",
	})
	if err == nil || !strings.Contains(err.Error(), "invalid permission request status") {
		t.Fatalf("expected invalid status error, got %v", err)
	}
}

func TestPermissionServiceReviewPermissionApplicationApproveFlow(t *testing.T) {
	repo := newMemoryPermissionRepository()
	reloadCalls := 0
	svc := NewPermissionService(repo, repo, func() error {
		reloadCalls++
		return nil
	})

	application, err := svc.CreatePermissionApplication(context.Background(), CreatePermissionApplicationInput{
		SubjectKey: "host:test",
		Targets: []PermissionTarget{
			{Resource: "host/*", Action: "read"},
			{Resource: "session/*", Action: "create"},
		},
		Reason:          "need access",
		ActorSubjectKey: "user:admin",
	})
	if err != nil {
		t.Fatalf("create permission application: %v", err)
	}

	updated, err := svc.ReviewPermissionApplication(context.Background(), ReviewPermissionApplicationInput{
		PermissionApplicationID: application.ID,
		Approved:                true,
		ReviewComment:           "approved",
		ReviewerSubjectKey:      "user:admin",
	})
	if err != nil {
		t.Fatalf("review permission application: %v", err)
	}

	if updated.Status != PermissionApplicationStatusApproved {
		t.Fatalf("status = %s, want approved", updated.Status)
	}
	if reloadCalls != 1 {
		t.Fatalf("reload calls = %d, want 1", reloadCalls)
	}
	if len(repo.grants) != 2 {
		t.Fatalf("grant count = %d, want 2", len(repo.grants))
	}
	if len(repo.casbinRules) != 2 {
		t.Fatalf("casbin rule count = %d, want 2", len(repo.casbinRules))
	}
}

func TestPermissionServiceReviewPermissionApplicationDenyRequiresComment(t *testing.T) {
	repo := newMemoryPermissionRepository()
	svc := NewPermissionService(repo, repo, nil)

	application, err := svc.CreatePermissionApplication(context.Background(), CreatePermissionApplicationInput{
		SubjectKey: "host:test",
		Targets: []PermissionTarget{
			{Resource: "host/*", Action: "read"},
		},
		Reason: "need access",
	})
	if err != nil {
		t.Fatalf("create permission application: %v", err)
	}

	_, err = svc.ReviewPermissionApplication(context.Background(), ReviewPermissionApplicationInput{
		PermissionApplicationID: application.ID,
		Approved:                false,
		ReviewerSubjectKey:      "user:admin",
	})
	if err == nil || !strings.Contains(err.Error(), "review comment is required") {
		t.Fatalf("expected deny comment validation error, got %v", err)
	}
}

func TestPermissionServiceReviewPermissionApplicationDenyDoesNotReloadPolicy(t *testing.T) {
	repo := newMemoryPermissionRepository()
	reloadCalls := 0
	svc := NewPermissionService(repo, repo, func() error {
		reloadCalls++
		return nil
	})

	application, err := svc.CreatePermissionApplication(context.Background(), CreatePermissionApplicationInput{
		SubjectKey: "host:test",
		Targets: []PermissionTarget{
			{Resource: "host/*", Action: "read"},
		},
		Reason: "need access",
	})
	if err != nil {
		t.Fatalf("create permission application: %v", err)
	}

	updated, err := svc.ReviewPermissionApplication(context.Background(), ReviewPermissionApplicationInput{
		PermissionApplicationID: application.ID,
		Approved:                false,
		ReviewComment:           "not allowed",
		ReviewerSubjectKey:      "user:admin",
	})
	if err != nil {
		t.Fatalf("deny permission application: %v", err)
	}
	if updated.Status != PermissionApplicationStatusDenied {
		t.Fatalf("status = %s, want denied", updated.Status)
	}
	if reloadCalls != 0 {
		t.Fatalf("reload calls = %d, want 0 for deny flow", reloadCalls)
	}
}

func TestPermissionServiceReviewPermissionApplicationRejectsRepeatReview(t *testing.T) {
	repo := newMemoryPermissionRepository()
	svc := NewPermissionService(repo, repo, nil)

	application, err := svc.CreatePermissionApplication(context.Background(), CreatePermissionApplicationInput{
		SubjectKey: "host:test",
		Targets: []PermissionTarget{
			{Resource: "host/*", Action: "read"},
		},
		Reason: "need access",
	})
	if err != nil {
		t.Fatalf("create permission application: %v", err)
	}

	_, err = svc.ReviewPermissionApplication(context.Background(), ReviewPermissionApplicationInput{
		PermissionApplicationID: application.ID,
		Approved:                true,
		ReviewComment:           "approved",
		ReviewerSubjectKey:      "user:admin",
	})
	if err != nil {
		t.Fatalf("first review permission application: %v", err)
	}

	_, err = svc.ReviewPermissionApplication(context.Background(), ReviewPermissionApplicationInput{
		PermissionApplicationID: application.ID,
		Approved:                true,
		ReviewComment:           "approved again",
		ReviewerSubjectKey:      "user:admin",
	})
	if err == nil || !strings.Contains(err.Error(), "already") {
		t.Fatalf("expected repeated review error, got %v", err)
	}
}

func TestPermissionServiceRevokePermissionApplication(t *testing.T) {
	repo := newMemoryPermissionRepository()
	reloadCalls := 0
	svc := NewPermissionService(repo, repo, func() error {
		reloadCalls++
		return nil
	})

	application, err := svc.CreatePermissionApplication(context.Background(), CreatePermissionApplicationInput{
		SubjectKey: "host:test",
		Targets: []PermissionTarget{
			{Resource: "host/*", Action: "read"},
		},
		Reason: "need access",
	})
	if err != nil {
		t.Fatalf("create permission application: %v", err)
	}
	if _, err := svc.ReviewPermissionApplication(context.Background(), ReviewPermissionApplicationInput{
		PermissionApplicationID: application.ID,
		Approved:                true,
		ReviewComment:           "approved",
		ReviewerSubjectKey:      "user:admin",
	}); err != nil {
		t.Fatalf("approve permission application: %v", err)
	}

	updated, err := svc.RevokePermissionApplication(context.Background(), RevokePermissionApplicationInput{
		PermissionApplicationID: application.ID,
		RevokeReason:            "session closed",
		ReviewerSubjectKey:      "user:admin",
	})
	if err != nil {
		t.Fatalf("revoke permission application: %v", err)
	}

	if updated.Status != PermissionApplicationStatusRevoked {
		t.Fatalf("status = %s, want revoked", updated.Status)
	}
	if reloadCalls != 2 {
		t.Fatalf("reload calls = %d, want 2", reloadCalls)
	}
	for _, grant := range repo.grants {
		if grant.Status != PermissionGrantStatusRevoked {
			t.Fatalf("grant status = %s, want revoked", grant.Status)
		}
	}
	if len(repo.casbinRules) != 0 {
		t.Fatalf("casbin rule count = %d, want 0 after revoke", len(repo.casbinRules))
	}
}

func TestPermissionServiceRevokePermissionApplicationRejectsExpiredRequest(t *testing.T) {
	repo := newMemoryPermissionRepository()
	svc := NewPermissionService(repo, repo, nil)

	application, err := svc.CreatePermissionApplication(context.Background(), CreatePermissionApplicationInput{
		SubjectKey: "host:test",
		Targets: []PermissionTarget{
			{Resource: "host/*", Action: "read"},
		},
		Reason: "temporary access",
	})
	if err != nil {
		t.Fatalf("create permission application: %v", err)
	}
	current := repo.applications[application.ID]
	current.Status = PermissionApplicationStatusExpired
	repo.applications[application.ID] = current

	_, err = svc.RevokePermissionApplication(context.Background(), RevokePermissionApplicationInput{
		PermissionApplicationID: application.ID,
		RevokeReason:            "cleanup",
		ReviewerSubjectKey:      "user:admin",
	})
	if err == nil || !strings.Contains(err.Error(), "cannot be revoked") {
		t.Fatalf("expected revoke precondition error, got %v", err)
	}
}

func TestPermissionServiceExpirePermissionGrants(t *testing.T) {
	repo := newMemoryPermissionRepository()
	reloadCalls := 0
	svc := NewPermissionService(repo, repo, func() error {
		reloadCalls++
		return nil
	})

	expiresAt := time.Now().UTC().Add(time.Minute)
	application, err := svc.CreatePermissionApplication(context.Background(), CreatePermissionApplicationInput{
		SubjectKey: "host:test",
		Targets: []PermissionTarget{
			{Resource: "host/*", Action: "read"},
		},
		Reason:    "temporary access",
		ExpiresAt: &expiresAt,
	})
	if err != nil {
		t.Fatalf("create permission application: %v", err)
	}
	if _, err := svc.ReviewPermissionApplication(context.Background(), ReviewPermissionApplicationInput{
		PermissionApplicationID: application.ID,
		Approved:                true,
		ReviewComment:           "approved",
		ReviewerSubjectKey:      "user:admin",
	}); err != nil {
		t.Fatalf("approve permission application: %v", err)
	}
	for id, grant := range repo.grants {
		expiredAt := time.Now().UTC().Add(-time.Minute)
		grant.ExpiresAt = &expiredAt
		repo.grants[id] = grant
	}

	expiredCount, err := svc.ExpirePermissionGrants(context.Background())
	if err != nil {
		t.Fatalf("expire permission grants: %v", err)
	}

	if expiredCount != 1 {
		t.Fatalf("expired count = %d, want 1", expiredCount)
	}
	if reloadCalls != 2 {
		t.Fatalf("reload calls = %d, want 2", reloadCalls)
	}
	for _, grant := range repo.grants {
		if grant.Status != PermissionGrantStatusExpired {
			t.Fatalf("grant status = %s, want expired", grant.Status)
		}
	}
	currentApplication := repo.applications[application.ID]
	if currentApplication.Status != PermissionApplicationStatusExpired {
		t.Fatalf("application status = %s, want expired", currentApplication.Status)
	}
}

func TestPermissionServiceListSubjectPermissionsHidesExpiredGrant(t *testing.T) {
	repo := newMemoryPermissionRepository()
	svc := NewPermissionService(repo, repo, nil)

	application, err := svc.CreatePermissionApplication(context.Background(), CreatePermissionApplicationInput{
		SubjectKey: "host:test",
		Targets: []PermissionTarget{
			{Resource: "host/*", Action: "read"},
		},
		Reason: "temporary access",
	})
	if err != nil {
		t.Fatalf("create permission application: %v", err)
	}
	if _, err := svc.ReviewPermissionApplication(context.Background(), ReviewPermissionApplicationInput{
		PermissionApplicationID: application.ID,
		Approved:                true,
		ReviewComment:           "approved",
		ReviewerSubjectKey:      "user:admin",
	}); err != nil {
		t.Fatalf("approve permission application: %v", err)
	}
	for id, grant := range repo.grants {
		expiredAt := time.Now().UTC().Add(-time.Minute)
		grant.ExpiresAt = &expiredAt
		repo.grants[id] = grant
	}

	grants, err := svc.ListSubjectPermissions(context.Background(), "host:test")
	if err != nil {
		t.Fatalf("list subject permissions: %v", err)
	}
	if len(grants) != 0 {
		t.Fatalf("visible grants = %d, want 0 once expiry is reached", len(grants))
	}
}

func TestPermissionServiceReviewPermissionApplicationDetectsConcurrentStateChange(t *testing.T) {
	repo := newMemoryPermissionRepository()
	svc := NewPermissionService(repo, repo, nil)

	application, err := svc.CreatePermissionApplication(context.Background(), CreatePermissionApplicationInput{
		SubjectKey: "host:test",
		Targets: []PermissionTarget{
			{Resource: "host/*", Action: "read"},
		},
		Reason: "need access",
	})
	if err != nil {
		t.Fatalf("create permission application: %v", err)
	}
	repo.reviewErr = ErrPermissionApplicationStateChanged

	_, err = svc.ReviewPermissionApplication(context.Background(), ReviewPermissionApplicationInput{
		PermissionApplicationID: application.ID,
		Approved:                true,
		ReviewComment:           "approved",
		ReviewerSubjectKey:      "user:admin",
	})
	if err == nil || !strings.Contains(err.Error(), "no longer pending") {
		t.Fatalf("expected concurrent review precondition error, got %v", err)
	}
}

func TestPermissionServiceReviewPermissionApplicationReloadFailureDoesNotReturnMutationError(t *testing.T) {
	repo := newMemoryPermissionRepository()
	reloadCalls := 0
	svc := NewPermissionService(repo, repo, func() error {
		reloadCalls++
		return errors.New("reload failed")
	})

	application, err := svc.CreatePermissionApplication(context.Background(), CreatePermissionApplicationInput{
		SubjectKey: "host:test",
		Targets: []PermissionTarget{
			{Resource: "host/*", Action: "read"},
		},
		Reason: "need access",
	})
	if err != nil {
		t.Fatalf("create permission application: %v", err)
	}

	updated, err := svc.ReviewPermissionApplication(context.Background(), ReviewPermissionApplicationInput{
		PermissionApplicationID: application.ID,
		Approved:                true,
		ReviewComment:           "approved",
		ReviewerSubjectKey:      "user:admin",
	})
	if err != nil {
		t.Fatalf("review should still succeed after commit when reload fails: %v", err)
	}
	if updated.Status != PermissionApplicationStatusApproved {
		t.Fatalf("status = %s, want approved", updated.Status)
	}
	if reloadCalls != 3 {
		t.Fatalf("reload calls = %d, want 3 retries", reloadCalls)
	}
}

func TestPermissionServiceGetPermissionApplicationReturnsStableNotFound(t *testing.T) {
	repo := newMemoryPermissionRepository()
	svc := NewPermissionService(repo, repo, nil)

	_, err := svc.GetPermissionApplication(context.Background(), "pa_missing")
	if !errors.Is(err, ErrPermissionApplicationNotFound) {
		t.Fatalf("error = %v, want ErrPermissionApplicationNotFound", err)
	}
}

func appID(n int) string { return fmt.Sprintf("pa_%d", n) }

func grantID(n int) string { return fmt.Sprintf("pg_%d", n) }
