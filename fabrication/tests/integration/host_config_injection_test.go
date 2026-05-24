//go:build integration

package integration

import (
	"context"
	"sort"
	"testing"
	"time"

	client "github.com/charviki/maze/fabrication/cradle/api/gen/http"
)

func TestHostCreateWithRules(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	ctx := context.Background()

	rule1Name := uniqueName("rule-a")
	rule2Name := uniqueName("rule-b")
	content := "test rule content"
	for _, name := range []string{rule1Name, rule2Name} {
		n := name
		_, _, err := h.apiClient.RuleServiceAPI.RuleServiceCreateRule(ctx).Body(client.V1CreateRuleRequest{
			Name:    &n,
			Content: &content,
		}).Execute()
		if err != nil {
			t.Fatalf("create rule %s failed: %v", name, err)
		}
	}

	hostName := uniqueName("host-rules")
	h.trackHost(hostName)
	nameField := hostName
	body := client.V1CreateHostRequest{
		Name:  &nameField,
		Tools: []string{"claude"},
		Rules: []string{rule1Name, rule2Name},
	}

	resp, httpResp, err := h.apiClient.HostServiceAPI.HostServiceCreateHost(ctx).Body(body).Execute()
	if err != nil {
		t.Fatalf("create host failed: %v (status=%d)", err, httpResp.StatusCode)
	}

	gotRules := resp.GetRules()
	if len(gotRules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(gotRules))
	}
	sort.Strings(gotRules)
	if gotRules[0] != rule1Name || gotRules[1] != rule2Name {
		t.Errorf("expected rules=[%s, %s], got=%v", rule1Name, rule2Name, gotRules)
	}

	h.waitForHostStatus(t, hostName, "online", 3*time.Minute)
	info, _, err := h.apiClient.HostServiceAPI.HostServiceGetHost(ctx, hostName).Execute()
	if err != nil {
		t.Fatalf("get host failed: %v", err)
	}
	infoRules := info.GetRules()
	sort.Strings(infoRules)
	if len(infoRules) != 2 || infoRules[0] != rule1Name || infoRules[1] != rule2Name {
		t.Errorf("HostInfo rules mismatch: expected=[%s, %s], got=%v", rule1Name, rule2Name, infoRules)
	}
	t.Logf("[step] PASS: host rules persisted correctly in HostInfo")

	cleanupRule(t, ctx, h, rule1Name)
	cleanupRule(t, ctx, h, rule2Name)
}

func TestHostCreateWithGitKeys(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	ctx := context.Background()

	key1Name := uniqueName("gitkey-a")
	key2Name := uniqueName("gitkey-b")
	token := "ghp_test_integration_token_value"
	tokenType := "PERSONAL_ACCESS_TOKEN"
	host := "github.com"
	for _, name := range []string{key1Name, key2Name} {
		n := name
		_, _, err := h.apiClient.GitKeyServiceAPI.GitKeyServiceCreateGitKey(ctx).Body(client.V1CreateGitKeyRequest{
			Name:      &n,
			Token:     &token,
			TokenType: &tokenType,
			Host:      &host,
		}).Execute()
		if err != nil {
			t.Fatalf("create git key %s failed: %v", name, err)
		}
	}

	hostName := uniqueName("host-gitkeys")
	h.trackHost(hostName)
	nameField := hostName
	body := client.V1CreateHostRequest{
		Name:    &nameField,
		Tools:   []string{"claude"},
		GitKeys: []string{key1Name, key2Name},
	}

	resp, httpResp, err := h.apiClient.HostServiceAPI.HostServiceCreateHost(ctx).Body(body).Execute()
	if err != nil {
		t.Fatalf("create host failed: %v (status=%d)", err, httpResp.StatusCode)
	}

	gotKeys := resp.GetGitKeys()
	if len(gotKeys) != 2 {
		t.Fatalf("expected 2 git keys, got %d", len(gotKeys))
	}
	sort.Strings(gotKeys)
	if gotKeys[0] != key1Name || gotKeys[1] != key2Name {
		t.Errorf("expected gitKeys=[%s, %s], got=%v", key1Name, key2Name, gotKeys)
	}

	h.waitForHostStatus(t, hostName, "online", 3*time.Minute)
	info, _, err := h.apiClient.HostServiceAPI.HostServiceGetHost(ctx, hostName).Execute()
	if err != nil {
		t.Fatalf("get host failed: %v", err)
	}
	infoKeys := info.GetGitKeys()
	sort.Strings(infoKeys)
	if len(infoKeys) != 2 || infoKeys[0] != key1Name || infoKeys[1] != key2Name {
		t.Errorf("HostInfo gitKeys mismatch: expected=[%s, %s], got=%v", key1Name, key2Name, infoKeys)
	}
	t.Logf("[step] PASS: host git keys persisted correctly in HostInfo")

	cleanupGitKey(t, ctx, h, key1Name)
	cleanupGitKey(t, ctx, h, key2Name)
}

func TestHostCreateWithAllResourcesIncludingRulesAndGitKeys(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	ctx := context.Background()

	// Create skill
	skillName := uniqueName("skill-all")
	desc := "full resource test"
	sName := skillName
	_, _, err := h.apiClient.SkillServiceAPI.SkillServiceCreateSkill(ctx).Body(client.V1CreateSkillRequest{
		Name:        &sName,
		Description: &desc,
	}).Execute()
	if err != nil {
		t.Fatalf("create skill failed: %v", err)
	}

	// Create MCP server
	mcpName := uniqueName("mcp-all")
	mcpType := "sse"
	mcpURL := "http://localhost:3000/mcp"
	mName := mcpName
	_, _, err = h.apiClient.MCPServiceAPI.MCPServiceCreateMCPServer(ctx).Body(client.V1CreateMCPServerRequest{
		Name: &mName,
		Type: &mcpType,
		Url:  &mcpURL,
	}).Execute()
	if err != nil {
		t.Fatalf("create MCP server failed: %v", err)
	}

	// Create rule
	ruleName := uniqueName("rule-all")
	ruleContent := "full resource test rule"
	rName := ruleName
	_, _, err = h.apiClient.RuleServiceAPI.RuleServiceCreateRule(ctx).Body(client.V1CreateRuleRequest{
		Name:    &rName,
		Content: &ruleContent,
	}).Execute()
	if err != nil {
		t.Fatalf("create rule failed: %v", err)
	}

	// Create git key
	gitKeyName := uniqueName("gitkey-all")
	gitToken := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAItestall"
	gitTokenType := "SSH_KEY"
	gitHost := "github.com"
	gName := gitKeyName
	_, _, err = h.apiClient.GitKeyServiceAPI.GitKeyServiceCreateGitKey(ctx).Body(client.V1CreateGitKeyRequest{
		Name:      &gName,
		Token:     &gitToken,
		TokenType: &gitTokenType,
		Host:      &gitHost,
	}).Execute()
	if err != nil {
		t.Fatalf("create git key failed: %v", err)
	}

	// Create host with all 4 resource types
	hostName := uniqueName("host-all-res")
	h.trackHost(hostName)
	nameField := hostName
	body := client.V1CreateHostRequest{
		Name:       &nameField,
		Tools:      []string{"claude"},
		Skills:     []string{skillName},
		McpServers: []string{mcpName},
		Rules:      []string{ruleName},
		GitKeys:    []string{gitKeyName},
	}

	resp, httpResp, err := h.apiClient.HostServiceAPI.HostServiceCreateHost(ctx).Body(body).Execute()
	if err != nil {
		t.Fatalf("create host failed: %v (status=%d)", err, httpResp.StatusCode)
	}

	if gotSkills := resp.GetSkills(); len(gotSkills) != 1 || gotSkills[0] != skillName {
		t.Errorf("expected skills=[%s], got=%v", skillName, gotSkills)
	}
	if gotMcp := resp.GetMcpServers(); len(gotMcp) != 1 || gotMcp[0] != mcpName {
		t.Errorf("expected mcp_servers=[%s], got=%v", mcpName, gotMcp)
	}
	if gotRules := resp.GetRules(); len(gotRules) != 1 || gotRules[0] != ruleName {
		t.Errorf("expected rules=[%s], got=%v", ruleName, gotRules)
	}
	if gotKeys := resp.GetGitKeys(); len(gotKeys) != 1 || gotKeys[0] != gitKeyName {
		t.Errorf("expected gitKeys=[%s], got=%v", gitKeyName, gotKeys)
	}

	h.waitForHostStatus(t, hostName, "online", 3*time.Minute)
	info, _, err := h.apiClient.HostServiceAPI.HostServiceGetHost(ctx, hostName).Execute()
	if err != nil {
		t.Fatalf("get host failed: %v", err)
	}
	if gotSkills := info.GetSkills(); len(gotSkills) != 1 || gotSkills[0] != skillName {
		t.Errorf("HostInfo skills mismatch: expected=[%s], got=%v", skillName, gotSkills)
	}
	if gotMcp := info.GetMcpServers(); len(gotMcp) != 1 || gotMcp[0] != mcpName {
		t.Errorf("HostInfo mcp_servers mismatch: expected=[%s], got=%v", mcpName, gotMcp)
	}
	if gotRules := info.GetRules(); len(gotRules) != 1 || gotRules[0] != ruleName {
		t.Errorf("HostInfo rules mismatch: expected=[%s], got=%v", ruleName, gotRules)
	}
	if gotKeys := info.GetGitKeys(); len(gotKeys) != 1 || gotKeys[0] != gitKeyName {
		t.Errorf("HostInfo gitKeys mismatch: expected=[%s], got=%v", gitKeyName, gotKeys)
	}
	t.Logf("[step] PASS: host with all 4 resource types persisted correctly")

	cleanupSkill(t, ctx, h, skillName)
	cleanupMCPServer(t, ctx, h, mcpName)
	cleanupRule(t, ctx, h, ruleName)
	cleanupGitKey(t, ctx, h, gitKeyName)
}

func TestGitKeyUpdatePreservesFields(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	ctx := context.Background()

	// Create key with tokenType and host
	name := uniqueName("gitkey-upd")
	token := "ghp_original_token_value"
	tokenType := "PERSONAL_ACCESS_TOKEN"
	host := "gitlab.com"
	_, _, err := h.apiClient.GitKeyServiceAPI.GitKeyServiceCreateGitKey(ctx).Body(client.V1CreateGitKeyRequest{
		Name:      &name,
		Token:     &token,
		TokenType: &tokenType,
		Host:      &host,
	}).Execute()
	if err != nil {
		t.Fatalf("create git key failed: %v", err)
	}

	// Update only the token
	newToken := "ghp_updated_token_value"
	updated, _, err := h.apiClient.GitKeyServiceAPI.GitKeyServiceUpdateGitKey(ctx, name).Body(client.GitKeyServiceUpdateGitKeyBody{
		Token: &newToken,
	}).Execute()
	if err != nil {
		t.Fatalf("update git key failed: %v", err)
	}

	// tokenType and host should be preserved
	if updated.GetTokenType() != tokenType {
		t.Errorf("TokenType = %q, want %q", updated.GetTokenType(), tokenType)
	}
	if updated.GetHost() != host {
		t.Errorf("Host = %q, want %q", updated.GetHost(), host)
	}

	// Get and verify again
	key, _, err := h.apiClient.GitKeyServiceAPI.GitKeyServiceGetGitKey(ctx, name).Execute()
	if err != nil {
		t.Fatalf("get git key failed: %v", err)
	}
	if key.GetTokenType() != tokenType {
		t.Errorf("Get TokenType = %q, want %q", key.GetTokenType(), tokenType)
	}
	if key.GetHost() != host {
		t.Errorf("Get Host = %q, want %q", key.GetHost(), host)
	}
	t.Logf("[step] PASS: git key update preserves tokenType and host")

	cleanupGitKey(t, ctx, h, name)
}

func TestGitKeyUpdateWithHostAndTokenType(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	ctx := context.Background()

	// Create key with initial values
	name := uniqueName("gitkey-upd2")
	token := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIoriginal"
	tokenType := "SSH_KEY"
	host := "gitlab.com"
	_, _, err := h.apiClient.GitKeyServiceAPI.GitKeyServiceCreateGitKey(ctx).Body(client.V1CreateGitKeyRequest{
		Name:      &name,
		Token:     &token,
		TokenType: &tokenType,
		Host:      &host,
	}).Execute()
	if err != nil {
		t.Fatalf("create git key failed: %v", err)
	}

	// Update host and tokenType
	newHost := "github.com"
	newTokenType := "PERSONAL_ACCESS_TOKEN"
	updated, _, err := h.apiClient.GitKeyServiceAPI.GitKeyServiceUpdateGitKey(ctx, name).Body(client.GitKeyServiceUpdateGitKeyBody{
		Host:      &newHost,
		TokenType: &newTokenType,
	}).Execute()
	if err != nil {
		t.Fatalf("update git key failed: %v", err)
	}

	if updated.GetHost() != newHost {
		t.Errorf("Host = %q, want %q", updated.GetHost(), newHost)
	}
	if updated.GetTokenType() != newTokenType {
		t.Errorf("TokenType = %q, want %q", updated.GetTokenType(), newTokenType)
	}

	// Verify via Get
	key, _, err := h.apiClient.GitKeyServiceAPI.GitKeyServiceGetGitKey(ctx, name).Execute()
	if err != nil {
		t.Fatalf("get git key failed: %v", err)
	}
	if key.GetHost() != newHost {
		t.Errorf("Get Host = %q, want %q", key.GetHost(), newHost)
	}
	if key.GetTokenType() != newTokenType {
		t.Errorf("Get TokenType = %q, want %q", key.GetTokenType(), newTokenType)
	}
	t.Logf("[step] PASS: git key update with host and tokenType works correctly")

	cleanupGitKey(t, ctx, h, name)
}

func cleanupRule(t *testing.T, ctx context.Context, h *testHelper, name string) {
	t.Helper()
	_, _, err := h.apiClient.RuleServiceAPI.RuleServiceDeleteRule(ctx, name).Execute()
	if err != nil {
		t.Logf("[cleanup] delete rule %q failed: %v", name, err)
	}
}

func cleanupGitKey(t *testing.T, ctx context.Context, h *testHelper, name string) {
	t.Helper()
	_, _, err := h.apiClient.GitKeyServiceAPI.GitKeyServiceDeleteGitKey(ctx, name).Execute()
	if err != nil {
		t.Logf("[cleanup] delete git key %q failed: %v", name, err)
	}
}
