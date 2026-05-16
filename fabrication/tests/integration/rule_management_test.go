//go:build integration

package integration

import (
	"context"
	"testing"

	client "github.com/charviki/maze/fabrication/cradle/api/gen/http"
)

func TestRuleCRUD(t *testing.T) {
	t.Parallel()

	h := newTestHelper(t)
	defer h.cleanup(t)

	ctx := context.Background()

	// Create
	t.Log("[step] creating rule...")
	name := uniqueName("rule")
	content := "# Rule Content\n\nThis is a test rule."
	createBody := client.V1CreateRuleRequest{
		Name:    &name,
		Content: &content,
	}
	created, httpResp, err := h.apiClient.RuleServiceAPI.RuleServiceCreateRule(ctx).Body(createBody).Execute()
	if err != nil {
		t.Fatalf("create rule failed: %v (status=%d)", err, httpResp.StatusCode)
	}
	if created.GetName() != name {
		t.Errorf("expected name=%s, got=%s", name, created.GetName())
	}
	t.Logf("[step] created rule: name=%s", created.GetName())

	// List
	t.Log("[step] listing rules...")
	listResp, _, err := h.apiClient.RuleServiceAPI.RuleServiceListRules(ctx).Execute()
	if err != nil {
		t.Fatalf("list rules failed: %v", err)
	}
	found := false
	for _, r := range listResp.GetRules() {
		if r.GetName() == name {
			found = true
		}
	}
	if !found {
		t.Errorf("rule %s not found in list", name)
	}

	// Get
	t.Log("[step] getting rule...")
	rule, _, err := h.apiClient.RuleServiceAPI.RuleServiceGetRule(ctx, name).Execute()
	if err != nil {
		t.Fatalf("get rule failed: %v", err)
	}
	if rule.GetName() != name {
		t.Errorf("expected name=%s, got=%s", name, rule.GetName())
	}

	// Update
	t.Log("[step] updating rule...")
	newContent := "# Updated Rule\n\nUpdated content."
	updateBody := client.RuleServiceUpdateRuleBody{
		Content: &newContent,
	}
	updated, _, err := h.apiClient.RuleServiceAPI.RuleServiceUpdateRule(ctx, name).Body(updateBody).Execute()
	if err != nil {
		t.Fatalf("update rule failed: %v", err)
	}
	if updated.GetContent() != newContent {
		t.Errorf("expected updated content=%s, got=%s", newContent, updated.GetContent())
	}

	// Delete
	t.Log("[step] deleting rule...")
	_, httpResp, err = h.apiClient.RuleServiceAPI.RuleServiceDeleteRule(ctx, name).Execute()
	if err != nil {
		t.Fatalf("delete rule failed: %v (status=%d)", err, httpResp.StatusCode)
	}

	// Verify deleted
	_, httpResp, _ = h.apiClient.RuleServiceAPI.RuleServiceGetRule(ctx, name).Execute()
	if httpResp.StatusCode != 404 {
		t.Errorf("expected 404 after delete, got=%d", httpResp.StatusCode)
	}
}

func TestRuleCreateDuplicate(t *testing.T) {
	t.Parallel()

	h := newTestHelper(t)
	defer h.cleanup(t)

	ctx := context.Background()

	name := uniqueName("rule-dup")
	content := "test rule content"
	createBody := client.V1CreateRuleRequest{
		Name:    &name,
		Content: &content,
	}

	_, _, err := h.apiClient.RuleServiceAPI.RuleServiceCreateRule(ctx).Body(createBody).Execute()
	if err != nil {
		t.Fatalf("first create should succeed: %v", err)
	}

	_, httpResp, err := h.apiClient.RuleServiceAPI.RuleServiceCreateRule(ctx).Body(createBody).Execute()
	if err == nil {
		t.Fatal("expected error when creating duplicate rule")
	}
	if httpResp.StatusCode != 409 {
		t.Errorf("expected 409 for duplicate, got=%d", httpResp.StatusCode)
	}
}
