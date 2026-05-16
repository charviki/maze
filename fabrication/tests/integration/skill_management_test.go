//go:build integration

package integration

import (
	"context"
	"testing"

	client "github.com/charviki/maze/fabrication/cradle/api/gen/http"
)

func TestSkillCRUD(t *testing.T) {
	t.Parallel()

	h := newTestHelper(t)
	defer h.cleanup(t)

	ctx := context.Background()

	// Create
	t.Log("[step] creating skill...")
	name := uniqueName("skill")
	desc := "test skill description"
	cfg := map[string]string{"key1": "value1", "key2": "value2"}
	createBody := client.V1CreateSkillRequest{
		Name:        &name,
		Description: &desc,
		Config:      &cfg,
	}
	created, httpResp, err := h.apiClient.SkillServiceAPI.SkillServiceCreateSkill(ctx).Body(createBody).Execute()
	if err != nil {
		t.Fatalf("create skill failed: %v (status=%d)", err, httpResp.StatusCode)
	}
	if created.GetName() != name {
		t.Errorf("expected name=%s, got=%s", name, created.GetName())
	}
	if created.GetDescription() != desc {
		t.Errorf("expected description=%s, got=%s", desc, created.GetDescription())
	}
	t.Logf("[step] created skill: name=%s", created.GetName())

	// List
	t.Log("[step] listing skills...")
	listResp, _, err := h.apiClient.SkillServiceAPI.SkillServiceListSkills(ctx).Execute()
	if err != nil {
		t.Fatalf("list skills failed: %v", err)
	}
	found := false
	for _, s := range listResp.GetSkills() {
		if s.GetName() == name {
			found = true
		}
	}
	if !found {
		t.Errorf("skill %s not found in list", name)
	}

	// Get
	t.Log("[step] getting skill...")
	skill, _, err := h.apiClient.SkillServiceAPI.SkillServiceGetSkill(ctx, name).Execute()
	if err != nil {
		t.Fatalf("get skill failed: %v", err)
	}
	if skill.GetName() != name {
		t.Errorf("expected name=%s, got=%s", name, skill.GetName())
	}

	// Update
	t.Log("[step] updating skill...")
	newDesc := "updated description"
	newCfg := map[string]string{"key3": "value3"}
	updateBody := client.SkillServiceUpdateSkillBody{
		Description: &newDesc,
		Config:      &newCfg,
	}
	updated, _, err := h.apiClient.SkillServiceAPI.SkillServiceUpdateSkill(ctx, name).Body(updateBody).Execute()
	if err != nil {
		t.Fatalf("update skill failed: %v", err)
	}
	if updated.GetDescription() != newDesc {
		t.Errorf("expected updated description=%s, got=%s", newDesc, updated.GetDescription())
	}

	// Delete
	t.Log("[step] deleting skill...")
	_, httpResp, err = h.apiClient.SkillServiceAPI.SkillServiceDeleteSkill(ctx, name).Execute()
	if err != nil {
		t.Fatalf("delete skill failed: %v (status=%d)", err, httpResp.StatusCode)
	}

	// Verify deleted
	_, httpResp, _ = h.apiClient.SkillServiceAPI.SkillServiceGetSkill(ctx, name).Execute()
	if httpResp.StatusCode != 404 {
		t.Errorf("expected 404 after delete, got=%d", httpResp.StatusCode)
	}
}

func TestSkillCreateDuplicate(t *testing.T) {
	t.Parallel()

	h := newTestHelper(t)
	defer h.cleanup(t)

	ctx := context.Background()

	name := uniqueName("skill-dup")
	desc := "dup test"
	createBody := client.V1CreateSkillRequest{
		Name:        &name,
		Description: &desc,
	}

	_, _, err := h.apiClient.SkillServiceAPI.SkillServiceCreateSkill(ctx).Body(createBody).Execute()
	if err != nil {
		t.Fatalf("first create should succeed: %v", err)
	}

	_, httpResp, err := h.apiClient.SkillServiceAPI.SkillServiceCreateSkill(ctx).Body(createBody).Execute()
	if err == nil {
		t.Fatal("expected error when creating duplicate skill")
	}
	if httpResp.StatusCode != 409 {
		t.Errorf("expected 409 for duplicate, got=%d", httpResp.StatusCode)
	}
}
