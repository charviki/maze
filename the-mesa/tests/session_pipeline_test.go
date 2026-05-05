package integration

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSessionPipeline(t *testing.T) {
	settings := loadClaudeSettings(t)
	claudeSid := generateID()
	bashSid := generateID()

	t.Run("CreateClaudeSession_AutoWithSettings", func(t *testing.T) {
		resp := createSessionWithSettings(t, agent1Port(), claudeSid,
			"claude --dangerously-skip-permissions",
			"/home/agent/"+claudeSid,
			"auto",
			settings,
		)
		assertOK(t, resp, "create claude session with settings")
		sess := decodeData[Session](t, resp)
		if sess.Name != claudeSid {
			t.Fatalf("expected name=%s, got %s", claudeSid, sess.Name)
		}

		time.Sleep(3 * time.Second)

		if !dockerExecCheckSession(t, agent1Container(), claudeSid) {
			t.Fatalf("tmux session %s not found", claudeSid)
		}

		state := dockerExecReadState(t, agent1Container(), claudeSid)
		if state == "" {
			t.Fatal("pipeline state file not found")
		}
		var stateMap struct {
			RestoreStrategy string `json:"restore_strategy"`
		}
		if err := json.Unmarshal([]byte(state), &stateMap); err != nil {
			t.Fatalf("parse state file failed: %v", err)
		}
		if stateMap.RestoreStrategy != "auto" {
			t.Fatalf("expected restore_strategy=auto, got %s", stateMap.RestoreStrategy)
		}

		settingsInContainer := dockerExec(t, agent1Container(),
			"cat", "/home/agent/"+claudeSid+"/.claude/settings.json")

		var settingsMap map[string]interface{}
		if err := json.Unmarshal([]byte(settingsInContainer), &settingsMap); err != nil {
			t.Fatalf("parse container settings.json failed: %v", err)
		}
		envMap, ok := settingsMap["env"].(map[string]interface{})
		if !ok {
			t.Fatal("settings.json has no env section")
		}
		if _, hasToken := envMap["ANTHROPIC_AUTH_TOKEN"]; !hasToken {
			t.Fatal("settings.json missing ANTHROPIC_AUTH_TOKEN in env")
		}
	})

	t.Run("CreateBashSession_Manual", func(t *testing.T) {
		resp := apiPost(t, agent1Port(), "/api/v1/sessions", CreateSessionRequest{
			Name:            bashSid,
			Command:         "bash",
			WorkingDir:      "/home/agent",
			RestoreStrategy: "manual",
		})
		assertOK(t, resp, "create bash session")
		time.Sleep(2 * time.Second)

		state := dockerExecReadState(t, agent1Container(), bashSid)
		if state == "" {
			t.Fatal("pipeline state file not found")
		}
		var stateMap struct {
			RestoreStrategy string `json:"restore_strategy"`
		}
		if err := json.Unmarshal([]byte(state), &stateMap); err != nil {
			t.Fatalf("parse state file failed: %v", err)
		}
		if stateMap.RestoreStrategy != "manual" {
			t.Fatalf("expected restore_strategy=manual, got %s", stateMap.RestoreStrategy)
		}
	})

	t.Run("ManualSave", func(t *testing.T) {
		resp := apiPost(t, agent1Port(), "/api/v1/sessions/save", map[string]string{})
		assertOK(t, resp, "manual save")
		var data struct {
			SavedAt string `json:"saved_at"`
		}
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			t.Fatalf("parse save response failed: %v", err)
		}
		if data.SavedAt == "" {
			t.Fatal("saved_at is empty")
		}
	})

	t.Run("KillSession_Cleanup", func(t *testing.T) {
		resp := apiDelete(t, agent1Port(), "/api/v1/sessions/"+claudeSid)
		assertOK(t, resp, "kill claude session")

		if dockerExecCheckSession(t, agent1Container(), claudeSid) {
			t.Fatalf("tmux session %s still exists after kill", claudeSid)
		}

		state := dockerExecReadState(t, agent1Container(), claudeSid)
		if state != "" {
			t.Fatalf("state file still exists after kill")
		}

		resp = apiDelete(t, agent1Port(), "/api/v1/sessions/"+bashSid)
		assertOK(t, resp, "kill bash session")

		if dockerExecCheckSession(t, agent1Container(), bashSid) {
			t.Fatalf("tmux session %s still exists after kill", bashSid)
		}
	})
}
