package integration

import "testing"

func TestSessionCRUD(t *testing.T) {
	sid := generateID()

	t.Run("Create", func(t *testing.T) {
		resp := apiPost(t, agent1Port(), "/api/v1/sessions", map[string]string{
			"name":    sid,
			"command": "bash",
		})
		assertOK(t, resp, "create session")
		sess := decodeData[Session](t, resp)
		if sess.Name != sid {
			t.Fatalf("expected name=%s, got %s", sid, sess.Name)
		}
		if sess.Status != "running" {
			t.Fatalf("expected status=running, got %s", sess.Status)
		}
	})

	t.Run("List", func(t *testing.T) {
		resp := apiGet(t, agent1Port(), "/api/v1/sessions")
		assertOK(t, resp, "list sessions")
		sessions := decodeData[[]Session](t, resp)
		found := false
		for _, s := range sessions {
			if s.Name == sid {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("session %s not found in list", sid)
		}
	})

	t.Run("Get", func(t *testing.T) {
		resp := apiGet(t, agent1Port(), "/api/v1/sessions/"+sid)
		assertOK(t, resp, "get session")
		sess := decodeData[Session](t, resp)
		if sess.Name != sid {
			t.Fatalf("expected name=%s, got %s", sid, sess.Name)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		resp := apiDelete(t, agent1Port(), "/api/v1/sessions/"+sid)
		assertOK(t, resp, "delete session")

		listResp := apiGet(t, agent1Port(), "/api/v1/sessions")
		sessions := decodeData[[]Session](t, listResp)
		for _, s := range sessions {
			if s.Name == sid {
				t.Fatalf("session %s still exists after delete", sid)
			}
		}
	})
}
