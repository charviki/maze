package integration

import (
	"testing"
	"time"
)

func TestTerminalOps(t *testing.T) {
	sid := generateID()

	resp := apiPost(t, agent1Port(), "/api/v1/sessions", map[string]string{
		"name":    sid,
		"command": "bash",
	})
	assertOK(t, resp, "create session for terminal ops")
	time.Sleep(2 * time.Second)

	t.Run("SendInput", func(t *testing.T) {
		resp := apiPost(t, agent1Port(), "/api/v1/sessions/"+sid+"/input", map[string]string{
			"command": "echo hello",
		})
		assertOK(t, resp, "send input")
		time.Sleep(1 * time.Second)
	})

	t.Run("GetOutput", func(t *testing.T) {
		resp := apiGet(t, agent1Port(), "/api/v1/sessions/"+sid+"/output")
		assertOK(t, resp, "get output")
		output := decodeData[TerminalOutput](t, resp)
		if output.Output == "" {
			t.Fatal("output is empty")
		}
	})

	t.Run("SendSignal", func(t *testing.T) {
		resp := apiPost(t, agent1Port(), "/api/v1/sessions/"+sid+"/signal", map[string]string{
			"signal": "SIGINT",
		})
		assertOK(t, resp, "send signal")
	})

	t.Run("GetEnv", func(t *testing.T) {
		resp := apiGet(t, agent1Port(), "/api/v1/sessions/"+sid+"/env")
		assertOK(t, resp, "get env")
		decodeData[map[string]string](t, resp)
	})

	apiDelete(t, agent1Port(), "/api/v1/sessions/"+sid)
}
