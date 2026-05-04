package service

import (
	"context"
	"errors"
	"testing"
)

type expireTrackingPermissionService struct {
	expiredCount int
	err          error
	calls        int
}

func (s *expireTrackingPermissionService) ExpirePermissionGrants(context.Context) (int, error) {
	s.calls++
	return s.expiredCount, s.err
}

func TestPermissionJanitorCleanupCallsExpirePermissionGrants(t *testing.T) {
	tracker := &expireTrackingPermissionService{expiredCount: 2}
	janitor := NewPermissionJanitor(tracker)

	if err := janitor.cleanup(context.Background()); err != nil {
		t.Fatalf("cleanup returned error: %v", err)
	}
	if tracker.calls != 1 {
		t.Fatalf("expire calls = %d, want 1", tracker.calls)
	}
}

func TestPermissionJanitorCleanupReturnsServiceError(t *testing.T) {
	expectedErr := errors.New("expire failed")
	tracker := &expireTrackingPermissionService{err: expectedErr}
	janitor := NewPermissionJanitor(tracker)

	err := janitor.cleanup(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("cleanup error = %v, want %v", err, expectedErr)
	}
}
