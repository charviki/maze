//go:build integration

package integration

import (
	"context"
	"testing"

	client "github.com/charviki/maze/fabrication/cradle/api/gen/http"
)

func TestMCPServerCRUD(t *testing.T) {
	t.Parallel()

	h := newTestHelper(t)
	defer h.cleanup(t)

	ctx := context.Background()

	// Create (stdio type)
	t.Log("[step] creating MCP server (stdio)...")
	name := uniqueName("mcp")
	mcpType := "stdio"
	command := "/usr/bin/mcp-server"
	args := []string{"--port", "8080"}
	env := map[string]string{"API_KEY": "test123"}
	createBody := client.V1CreateMCPServerRequest{
		Name:    &name,
		Type:    &mcpType,
		Command: &command,
		Args:    args,
		Env:     &env,
	}
	created, httpResp, err := h.apiClient.MCPServiceAPI.MCPServiceCreateMCPServer(ctx).Body(createBody).Execute()
	if err != nil {
		t.Fatalf("create MCP server failed: %v (status=%d)", err, httpResp.StatusCode)
	}
	if created.GetName() != name {
		t.Errorf("expected name=%s, got=%s", name, created.GetName())
	}
	if created.GetType() != mcpType {
		t.Errorf("expected type=%s, got=%s", mcpType, created.GetType())
	}
	t.Logf("[step] created MCP server: name=%s type=%s", created.GetName(), created.GetType())

	// List
	t.Log("[step] listing MCP servers...")
	listResp, _, err := h.apiClient.MCPServiceAPI.MCPServiceListMCPServers(ctx).Execute()
	if err != nil {
		t.Fatalf("list MCP servers failed: %v", err)
	}
	found := false
	for _, s := range listResp.GetServers() {
		if s.GetName() == name {
			found = true
		}
	}
	if !found {
		t.Errorf("MCP server %s not found in list", name)
	}

	// Get
	t.Log("[step] getting MCP server...")
	server, _, err := h.apiClient.MCPServiceAPI.MCPServiceGetMCPServer(ctx, name).Execute()
	if err != nil {
		t.Fatalf("get MCP server failed: %v", err)
	}
	if server.GetName() != name {
		t.Errorf("expected name=%s, got=%s", name, server.GetName())
	}

	// Update (change to sse type)
	t.Log("[step] updating MCP server...")
	newType := "sse"
	newURL := "http://localhost:3000/mcp"
	updateBody := client.MCPServiceUpdateMCPServerBody{
		Type: &newType,
		Url:  &newURL,
	}
	updated, _, err := h.apiClient.MCPServiceAPI.MCPServiceUpdateMCPServer(ctx, name).Body(updateBody).Execute()
	if err != nil {
		t.Fatalf("update MCP server failed: %v", err)
	}
	if updated.GetType() != newType {
		t.Errorf("expected updated type=%s, got=%s", newType, updated.GetType())
	}

	// Delete
	t.Log("[step] deleting MCP server...")
	_, httpResp, err = h.apiClient.MCPServiceAPI.MCPServiceDeleteMCPServer(ctx, name).Execute()
	if err != nil {
		t.Fatalf("delete MCP server failed: %v (status=%d)", err, httpResp.StatusCode)
	}

	// Verify deleted
	_, httpResp, _ = h.apiClient.MCPServiceAPI.MCPServiceGetMCPServer(ctx, name).Execute()
	if httpResp.StatusCode != 404 {
		t.Errorf("expected 404 after delete, got=%d", httpResp.StatusCode)
	}
}

func TestMCPServerCreateDuplicate(t *testing.T) {
	t.Parallel()

	h := newTestHelper(t)
	defer h.cleanup(t)

	ctx := context.Background()

	name := uniqueName("mcp-dup")
	mcpType := "stdio"
	command := "/usr/bin/test"
	createBody := client.V1CreateMCPServerRequest{
		Name:    &name,
		Type:    &mcpType,
		Command: &command,
	}

	_, _, err := h.apiClient.MCPServiceAPI.MCPServiceCreateMCPServer(ctx).Body(createBody).Execute()
	if err != nil {
		t.Fatalf("first create should succeed: %v", err)
	}

	_, httpResp, err := h.apiClient.MCPServiceAPI.MCPServiceCreateMCPServer(ctx).Body(createBody).Execute()
	if err == nil {
		t.Fatal("expected error when creating duplicate MCP server")
	}
	if httpResp.StatusCode != 409 {
		t.Errorf("expected 409 for duplicate, got=%d", httpResp.StatusCode)
	}
}
