//go:build integration

package integration

import (
	"context"
	"testing"

	client "github.com/charviki/maze/fabrication/cradle/api/gen/http"
)

func TestGitKeyCRUD(t *testing.T) {
	t.Parallel()

	h := newTestHelper(t)
	defer h.cleanup(t)

	ctx := context.Background()

	// Create
	t.Log("[step] creating git key...")
	name := uniqueName("gitkey")
	token := "ghp_abcdefghijklmnopqrstuvwxyz1234"
	createBody := client.V1CreateGitKeyRequest{
		Name:  &name,
		Token: &token,
	}
	created, httpResp, err := h.apiClient.GitKeyServiceAPI.GitKeyServiceCreateGitKey(ctx).Body(createBody).Execute()
	if err != nil {
		t.Fatalf("create git key failed: %v (status=%d)", err, httpResp.StatusCode)
	}
	if created.GetName() != name {
		t.Errorf("expected name=%s, got=%s", name, created.GetName())
	}
	if created.GetTokenMask() == "" {
		t.Error("expected token_mask to be non-empty")
	}
	if created.GetTokenMask() == token {
		t.Error("token_mask should not reveal the full token")
	}
	t.Logf("[step] created git key: name=%s token_mask=%s", created.GetName(), created.GetTokenMask())

	// List
	t.Log("[step] listing git keys...")
	listResp, _, err := h.apiClient.GitKeyServiceAPI.GitKeyServiceListGitKeys(ctx).Execute()
	if err != nil {
		t.Fatalf("list git keys failed: %v", err)
	}
	found := false
	for _, k := range listResp.GetGitKeys() {
		if k.GetName() == name {
			found = true
		}
	}
	if !found {
		t.Errorf("git key %s not found in list", name)
	}

	// Get
	t.Log("[step] getting git key...")
	key, _, err := h.apiClient.GitKeyServiceAPI.GitKeyServiceGetGitKey(ctx, name).Execute()
	if err != nil {
		t.Fatalf("get git key failed: %v", err)
	}
	if key.GetName() != name {
		t.Errorf("expected name=%s, got=%s", name, key.GetName())
	}

	// Verify token_mask consistency across Get and List
	if key.GetTokenMask() != created.GetTokenMask() {
		t.Errorf("token_mask mismatch: create returned %q, get returned %q", created.GetTokenMask(), key.GetTokenMask())
	}

	// Delete
	t.Log("[step] deleting git key...")
	_, httpResp, err = h.apiClient.GitKeyServiceAPI.GitKeyServiceDeleteGitKey(ctx, name).Execute()
	if err != nil {
		t.Fatalf("delete git key failed: %v (status=%d)", err, httpResp.StatusCode)
	}

	// Verify deleted
	_, httpResp, _ = h.apiClient.GitKeyServiceAPI.GitKeyServiceGetGitKey(ctx, name).Execute()
	if httpResp.StatusCode != 404 {
		t.Errorf("expected 404 after delete, got=%d", httpResp.StatusCode)
	}
}

func TestGitKeyDuplicateName(t *testing.T) {
	t.Parallel()

	h := newTestHelper(t)
	defer h.cleanup(t)

	ctx := context.Background()

	name := uniqueName("gitkey-dup")
	token := "ghp_duplicate_test_token_value"

	// Create first key
	createBody := client.V1CreateGitKeyRequest{
		Name:  &name,
		Token: &token,
	}
	_, _, err := h.apiClient.GitKeyServiceAPI.GitKeyServiceCreateGitKey(ctx).Body(createBody).Execute()
	if err != nil {
		t.Fatalf("first create should succeed: %v", err)
	}

	// Attempt to create with same name should fail
	_, httpResp, err := h.apiClient.GitKeyServiceAPI.GitKeyServiceCreateGitKey(ctx).Body(createBody).Execute()
	if err == nil {
		t.Fatal("expected error when creating duplicate git key")
	}
	if httpResp.StatusCode != 409 {
		t.Errorf("expected 409 for duplicate, got=%d", httpResp.StatusCode)
	}
}
