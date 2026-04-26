package integration

import (
	"testing"
)

func TestAgentManager(t *testing.T) {
	t.Run("ListNodes", func(t *testing.T) {
		resp := apiGet(t, managerPort(), "/api/v1/nodes")
		assertOK(t, resp, "list nodes")
		nodes := decodeData[[]Node](t, resp)
		if len(nodes) < 3 {
			t.Fatalf("expected at least 3 nodes, got %d", len(nodes))
		}
		nodeNames := make(map[string]bool)
		for _, n := range nodes {
			nodeNames[n.Name] = true
		}
		for _, expected := range []string{"claude-1", "claude-2", "codex-1"} {
			if !nodeNames[expected] {
				t.Fatalf("node %s not found", expected)
			}
		}
	})

	t.Run("TemplateCRUD", func(t *testing.T) {
		tplID := "test-tpl-" + generateID()

		resp := apiPost(t, managerPort(), "/api/v1/templates", map[string]interface{}{
			"id":      tplID,
			"name":    "Test Template",
			"command": "echo hello",
			"builtin": false,
		})
		assertOK(t, resp, "create template")
		tpl := decodeData[Template](t, resp)
		if tpl.ID != tplID {
			t.Fatalf("expected id=%s, got %s", tplID, tpl.ID)
		}

		listResp := apiGet(t, managerPort(), "/api/v1/templates")
		assertOK(t, listResp, "list templates")
		templates := decodeData[[]Template](t, listResp)
		found := false
		for _, t := range templates {
			if t.ID == tplID {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("template %s not found in list", tplID)
		}

		updateResp := apiPut(t, managerPort(), "/api/v1/templates/"+tplID, map[string]string{
			"name":    "Updated Test",
			"command": "echo updated",
		})
		assertOK(t, updateResp, "update template")

		delResp := apiDelete(t, managerPort(), "/api/v1/templates/"+tplID)
		assertOK(t, delResp, "delete template")
	})


}
