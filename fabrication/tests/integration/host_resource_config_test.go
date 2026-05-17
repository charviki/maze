//go:build integration

package integration

import (
	"context"
	"sort"
	"testing"
	"time"

	client "github.com/charviki/maze/fabrication/cradle/api/gen/http"
)

// TestHostCreateWithSkills 验证创建 Host 时携带 skills 字段，
// CreateHost 响应和 GetHost 查询都正确返回 skills
func TestHostCreateWithSkills(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	ctx := context.Background()

	skill1Name := uniqueName("skill-a")
	skill2Name := uniqueName("skill-b")
	desc := "test skill"
	for _, name := range []string{skill1Name, skill2Name} {
		n := name
		_, _, err := h.apiClient.SkillServiceAPI.SkillServiceCreateSkill(ctx).Body(client.V1CreateSkillRequest{
			Name:        &n,
			Description: &desc,
		}).Execute()
		if err != nil {
			t.Fatalf("create skill %s failed: %v", name, err)
		}
	}

	hostName := uniqueName("host-skills")
	h.trackHost(hostName)
	nameField := hostName
	body := client.V1CreateHostRequest{
		Name:   &nameField,
		Tools:  []string{"claude"},
		Skills: []string{skill1Name, skill2Name},
	}

	resp, httpResp, err := h.apiClient.HostServiceAPI.HostServiceCreateHost(ctx).Body(body).Execute()
	if err != nil {
		t.Fatalf("create host failed: %v (status=%d)", err, httpResp.StatusCode)
	}

	gotSkills := resp.GetSkills()
	if len(gotSkills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(gotSkills))
	}
	sort.Strings(gotSkills)
	if gotSkills[0] != skill1Name || gotSkills[1] != skill2Name {
		t.Errorf("expected skills=[%s, %s], got=%v", skill1Name, skill2Name, gotSkills)
	}

	h.waitForHostStatus(t, hostName, "online", 3*time.Minute)
	info, _, err := h.apiClient.HostServiceAPI.HostServiceGetHost(ctx, hostName).Execute()
	if err != nil {
		t.Fatalf("get host failed: %v", err)
	}
	infoSkills := info.GetSkills()
	sort.Strings(infoSkills)
	if len(infoSkills) != 2 || infoSkills[0] != skill1Name || infoSkills[1] != skill2Name {
		t.Errorf("HostInfo skills mismatch: expected=[%s, %s], got=%v", skill1Name, skill2Name, infoSkills)
	}
	t.Logf("[step] PASS: host skills persisted correctly in HostInfo")

	cleanupSkill(t, ctx, h, skill1Name)
	cleanupSkill(t, ctx, h, skill2Name)
}

// TestHostCreateWithMCPServers 验证创建 Host 时携带 mcp_servers 字段
func TestHostCreateWithMCPServers(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	ctx := context.Background()

	mcp1Name := uniqueName("mcp-a")
	mcp2Name := uniqueName("mcp-b")
	mcpType := "stdio"
	command := "/usr/bin/test"
	for _, name := range []string{mcp1Name, mcp2Name} {
		n := name
		_, _, err := h.apiClient.MCPServiceAPI.MCPServiceCreateMCPServer(ctx).Body(client.V1CreateMCPServerRequest{
			Name:    &n,
			Type:    &mcpType,
			Command: &command,
		}).Execute()
		if err != nil {
			t.Fatalf("create MCP server %s failed: %v", name, err)
		}
	}

	hostName := uniqueName("host-mcp")
	h.trackHost(hostName)
	nameField := hostName
	body := client.V1CreateHostRequest{
		Name:       &nameField,
		Tools:      []string{"claude"},
		McpServers: []string{mcp1Name, mcp2Name},
	}

	resp, httpResp, err := h.apiClient.HostServiceAPI.HostServiceCreateHost(ctx).Body(body).Execute()
	if err != nil {
		t.Fatalf("create host failed: %v (status=%d)", err, httpResp.StatusCode)
	}

	gotMcp := resp.GetMcpServers()
	if len(gotMcp) != 2 {
		t.Fatalf("expected 2 mcp_servers, got %d", len(gotMcp))
	}
	sort.Strings(gotMcp)
	if gotMcp[0] != mcp1Name || gotMcp[1] != mcp2Name {
		t.Errorf("expected mcp_servers=[%s, %s], got=%v", mcp1Name, mcp2Name, gotMcp)
	}

	h.waitForHostStatus(t, hostName, "online", 3*time.Minute)
	info, _, err := h.apiClient.HostServiceAPI.HostServiceGetHost(ctx, hostName).Execute()
	if err != nil {
		t.Fatalf("get host failed: %v", err)
	}
	infoMcp := info.GetMcpServers()
	sort.Strings(infoMcp)
	if len(infoMcp) != 2 || infoMcp[0] != mcp1Name || infoMcp[1] != mcp2Name {
		t.Errorf("HostInfo mcp_servers mismatch: expected=[%s, %s], got=%v", mcp1Name, mcp2Name, infoMcp)
	}
	t.Logf("[step] PASS: host mcp_servers persisted correctly in HostInfo")

	cleanupMCPServer(t, ctx, h, mcp1Name)
	cleanupMCPServer(t, ctx, h, mcp2Name)
}

// TestHostCreateWithAllResources 验证同时携带 tools + skills + mcp_servers 的完整场景
func TestHostCreateWithAllResources(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	ctx := context.Background()

	skillName := uniqueName("skill-full")
	desc := "full test"
	sName := skillName
	_, _, err := h.apiClient.SkillServiceAPI.SkillServiceCreateSkill(ctx).Body(client.V1CreateSkillRequest{
		Name:        &sName,
		Description: &desc,
	}).Execute()
	if err != nil {
		t.Fatalf("create skill failed: %v", err)
	}

	mcpName := uniqueName("mcp-full")
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

	hostName := uniqueName("host-full")
	h.trackHost(hostName)
	nameField := hostName
	body := client.V1CreateHostRequest{
		Name:       &nameField,
		Tools:      []string{"claude"},
		Skills:     []string{skillName},
		McpServers: []string{mcpName},
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
	t.Logf("[step] PASS: host with all resources persisted correctly")

	cleanupSkill(t, ctx, h, skillName)
	cleanupMCPServer(t, ctx, h, mcpName)
}

// TestHostCreateWithEmptyResources 验证不传 skills/mcp_servers 时默认为空数组
func TestHostCreateWithEmptyResources(t *testing.T) {
	h := newTestHelper(t)
	defer h.cleanup(t)

	ctx := context.Background()

	hostName := uniqueName("host-empty-res")
	h.trackHost(hostName)
	nameField := hostName
	body := client.V1CreateHostRequest{
		Name:  &nameField,
		Tools: []string{"claude"},
	}

	resp, httpResp, err := h.apiClient.HostServiceAPI.HostServiceCreateHost(ctx).Body(body).Execute()
	if err != nil {
		t.Fatalf("create host failed: %v (status=%d)", err, httpResp.StatusCode)
	}

	if skills := resp.GetSkills(); len(skills) != 0 {
		t.Errorf("expected empty skills, got=%v", skills)
	}
	if mcp := resp.GetMcpServers(); len(mcp) != 0 {
		t.Errorf("expected empty mcp_servers, got=%v", mcp)
	}

	h.waitForHostStatus(t, hostName, "online", 3*time.Minute)
	info, _, err := h.apiClient.HostServiceAPI.HostServiceGetHost(ctx, hostName).Execute()
	if err != nil {
		t.Fatalf("get host failed: %v", err)
	}
	if skills := info.GetSkills(); len(skills) != 0 {
		t.Errorf("HostInfo expected empty skills, got=%v", skills)
	}
	if mcp := info.GetMcpServers(); len(mcp) != 0 {
		t.Errorf("HostInfo expected empty mcp_servers, got=%v", mcp)
	}
	t.Logf("[step] PASS: host with empty resources works correctly")
}

func cleanupSkill(t *testing.T, ctx context.Context, h *testHelper, name string) {
	t.Helper()
	_, _, err := h.apiClient.SkillServiceAPI.SkillServiceDeleteSkill(ctx, name).Execute()
	if err != nil {
		t.Logf("[cleanup] delete skill %q failed: %v", name, err)
	}
}

func cleanupMCPServer(t *testing.T, ctx context.Context, h *testHelper, name string) {
	t.Helper()
	_, _, err := h.apiClient.MCPServiceAPI.MCPServiceDeleteMCPServer(ctx, name).Execute()
	if err != nil {
		t.Logf("[cleanup] delete mcp_server %q failed: %v", name, err)
	}
}
