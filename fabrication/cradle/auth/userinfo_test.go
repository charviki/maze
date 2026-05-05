package auth

import (
	"context"
	"testing"
)

func TestWithUserInfo(t *testing.T) {
	user := &UserInfo{SubjectKey: "user-1", DisplayName: "Test User"}
	ctx := WithUserInfo(context.Background(), user)

	got := GetUserInfo(ctx)
	if got == nil {
		t.Fatal("GetUserInfo returned nil for injected user")
	}
	if got.SubjectKey != "user-1" {
		t.Errorf("SubjectKey = %q, want %q", got.SubjectKey, "user-1")
	}
	if got.DisplayName != "Test User" {
		t.Errorf("DisplayName = %q, want %q", got.DisplayName, "Test User")
	}
}

func TestGetUserInfo_Nil_WhenNotSet(t *testing.T) {
	got := GetUserInfo(context.Background())
	if got != nil {
		t.Errorf("GetUserInfo should return nil when not set, got %+v", got)
	}
}

func TestGetUserInfo_Nil_WhenWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextKey{}, "not-a-user")
	got := GetUserInfo(ctx)
	if got != nil {
		t.Errorf("GetUserInfo should return nil for wrong type, got %+v", got)
	}
}

func TestUserInfoExtractor(t *testing.T) {
	var extractor UserInfoExtractor = func(ctx context.Context) (*UserInfo, error) {
		return &UserInfo{SubjectKey: "extracted"}, nil
	}

	user, err := extractor(context.Background())
	if err != nil {
		t.Fatalf("extractor returned error: %v", err)
	}
	if user.SubjectKey != "extracted" {
		t.Errorf("SubjectKey = %q, want %q", user.SubjectKey, "extracted")
	}
}

func TestWithUserInfo_Overwrite(t *testing.T) {
	first := &UserInfo{SubjectKey: "first"}
	second := &UserInfo{SubjectKey: "second"}

	ctx := WithUserInfo(context.Background(), first)
	ctx = WithUserInfo(ctx, second)

	got := GetUserInfo(ctx)
	if got.SubjectKey != "second" {
		t.Errorf("SubjectKey = %q, want %q", got.SubjectKey, "second")
	}
}
