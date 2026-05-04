package postgres

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	gen "github.com/charviki/mesa-hub-behavior-panel/internal/repository/postgres/sqlc"
	"github.com/charviki/mesa-hub-behavior-panel/internal/service"
)

func TestPermissionApplicationFromRowMapsFields(t *testing.T) {
	reviewedBy := "user:admin"
	reviewComment := "approved"
	publicID := "pa_1"
	expiresAt := time.Date(2026, 5, 4, 12, 0, 0, 0, time.FixedZone("UTC+8", 8*3600))
	reviewedAt := time.Date(2026, 5, 4, 13, 0, 0, 0, time.FixedZone("UTC+8", 8*3600))
	createdAt := time.Date(2026, 5, 4, 14, 0, 0, 0, time.FixedZone("UTC+8", 8*3600))
	updatedAt := time.Date(2026, 5, 4, 15, 0, 0, 0, time.FixedZone("UTC+8", 8*3600))

	application, err := permissionApplicationFromRow(gen.AuthPermissionRequest{
		ID:            7,
		PublicID:      &publicID,
		SubjectKey:    "host:test",
		TargetsJson:   []byte(`[{"resource":"host/*","action":"read"}]`),
		Reason:        "need access",
		Status:        gen.AuthPermissionRequestStatus(service.PermissionApplicationStatusApproved),
		ReviewedBy:    &reviewedBy,
		ReviewComment: &reviewComment,
		ReviewedAt:    tstz(reviewedAt),
		ExpiresAt:     tstz(expiresAt),
		CreatedAt:     tstz(createdAt),
		UpdatedAt:     tstz(updatedAt),
	})
	if err != nil {
		t.Fatalf("permissionApplicationFromRow: %v", err)
	}

	if application.ID != "pa_1" || application.InternalID != 7 {
		t.Fatalf("application ids = %+v, want public/internal ids mapped", application)
	}
	if len(application.Targets) != 1 || application.Targets[0].Resource != "host/*" || application.Targets[0].Action != "read" {
		t.Fatalf("targets = %+v, want decoded target", application.Targets)
	}
	if application.Status != service.PermissionApplicationStatusApproved {
		t.Fatalf("status = %s, want approved", application.Status)
	}
	// Mapper 统一转 UTC，避免 service 和 transport 比较时间时混入数据库会话时区。
	if application.ExpiresAt == nil || !application.ExpiresAt.Equal(expiresAt.UTC()) {
		t.Fatalf("expires_at = %v, want %v", application.ExpiresAt, expiresAt.UTC())
	}
	if application.ReviewedAt == nil || !application.ReviewedAt.Equal(reviewedAt.UTC()) {
		t.Fatalf("reviewed_at = %v, want %v", application.ReviewedAt, reviewedAt.UTC())
	}
	if !application.CreatedAt.Equal(createdAt.UTC()) || !application.UpdatedAt.Equal(updatedAt.UTC()) {
		t.Fatalf("created_at/updated_at = %v/%v, want UTC normalized values", application.CreatedAt, application.UpdatedAt)
	}
}

func TestPermissionApplicationFromRowRejectsInvalidTargetsJSON(t *testing.T) {
	_, err := permissionApplicationFromRow(gen.AuthPermissionRequest{
		TargetsJson: []byte(`{"resource":"host/*"`),
	})
	if err == nil {
		t.Fatal("expected invalid targets json error")
	}
}

func TestPermissionGrantsFromSubjectRowsMapsSourceRequestPublicID(t *testing.T) {
	publicID := "pg_1"
	sourceRequestPublicID := "pa_9"
	casbinRuleID := int64(12)
	expiresAt := time.Date(2026, 5, 4, 16, 0, 0, 0, time.FixedZone("UTC-5", -5*3600))

	grants := permissionGrantsFromSubjectRows([]gen.ListActiveSubjectPermissionGrantsRow{
		{
			AuthPermissionGrant: gen.AuthPermissionGrant{
				ID:              3,
				PublicID:        &publicID,
				SubjectKey:      "host:test",
				Resource:        "host/*",
				Action:          "read",
				SourceRequestID: 9,
				CasbinRuleID:    &casbinRuleID,
				Status:          gen.AuthPermissionGrantStatus(service.PermissionGrantStatusActive),
				ExpiresAt:       tstz(expiresAt),
				CreatedAt:       tstz(expiresAt.Add(-time.Minute)),
				UpdatedAt:       tstz(expiresAt),
			},
			SourceRequestPublicID: &sourceRequestPublicID,
		},
	})

	if len(grants) != 1 {
		t.Fatalf("grant count = %d, want 1", len(grants))
	}
	grant := grants[0]
	if grant.ID != "pg_1" || grant.SourcePermissionApplicationID != "pa_9" {
		t.Fatalf("grant ids = %+v, want mapped public ids", grant)
	}
	if grant.CasbinRuleID == nil || *grant.CasbinRuleID != 12 {
		t.Fatalf("casbin_rule_id = %v, want 12", grant.CasbinRuleID)
	}
	if grant.ExpiresAt == nil || !grant.ExpiresAt.Equal(expiresAt.UTC()) {
		t.Fatalf("expires_at = %v, want %v", grant.ExpiresAt, expiresAt.UTC())
	}
}

func TestNullableStringTrimsWhitespace(t *testing.T) {
	if nullableString("   ") != nil {
		t.Fatal("expected whitespace-only string to map to nil")
	}
	value := nullableString("  host:test  ")
	if value == nil || *value != "host:test" {
		t.Fatalf("nullable string = %v, want trimmed host:test", value)
	}
}

func tstz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}
