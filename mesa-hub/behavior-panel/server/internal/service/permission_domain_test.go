package service

import (
	"testing"
)

func TestValidationError(t *testing.T) {
	err := NewValidationError("invalid subject key")
	if err.Error() != "invalid subject key" {
		t.Errorf("error message = %q, want %q", err.Error(), "invalid subject key")
	}

	_, ok := err.(ValidationError)
	if !ok {
		t.Error("NewValidationError should return ValidationError type")
	}
}

func TestPreconditionError(t *testing.T) {
	err := NewPreconditionError("grant already revoked")
	if err.Error() != "grant already revoked" {
		t.Errorf("error message = %q, want %q", err.Error(), "grant already revoked")
	}
}

func TestPermissionApplicationStatus_Constants(t *testing.T) {
	if PermissionApplicationStatusPending != "pending" {
		t.Errorf("Pending = %q, want pending", PermissionApplicationStatusPending)
	}
	if PermissionApplicationStatusApproved != "approved" {
		t.Errorf("Approved = %q, want approved", PermissionApplicationStatusApproved)
	}
	if PermissionApplicationStatusDenied != "denied" {
		t.Errorf("Denied = %q, want denied", PermissionApplicationStatusDenied)
	}
	if PermissionApplicationStatusRevoked != "revoked" {
		t.Errorf("Revoked = %q, want revoked", PermissionApplicationStatusRevoked)
	}
	if PermissionApplicationStatusExpired != "expired" {
		t.Errorf("Expired = %q, want expired", PermissionApplicationStatusExpired)
	}
}

func TestPermissionGrantStatus_Constants(t *testing.T) {
	if PermissionGrantStatusActive != "active" {
		t.Errorf("Active = %q, want active", PermissionGrantStatusActive)
	}
	if PermissionGrantStatusRevoked != "revoked" {
		t.Errorf("Revoked = %q, want revoked", PermissionGrantStatusRevoked)
	}
	if PermissionGrantStatusExpired != "expired" {
		t.Errorf("Expired = %q, want expired", PermissionGrantStatusExpired)
	}
}

func TestSentinelErrors(t *testing.T) {
	if ErrPermissionApplicationNotFound == nil {
		t.Error("ErrPermissionApplicationNotFound should not be nil")
	}
	if ErrPermissionGrantNotFound == nil {
		t.Error("ErrPermissionGrantNotFound should not be nil")
	}
	if ErrPermissionApplicationStateChanged == nil {
		t.Error("ErrPermissionApplicationStateChanged should not be nil")
	}
}
