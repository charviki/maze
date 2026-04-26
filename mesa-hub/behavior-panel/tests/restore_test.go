package integration

import (
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestRestore 验证自动保存、容器重启后 auto 恢复、manual 恢复。
// 注意：此测试会重启 agent-1 容器，建议单独运行：
//   go test -v -run TestRestore -timeout 300s
func TestRestore(t *testing.T) {
	settings := loadClaudeSettings(t)
	claudeSid := generateID()
	bashSid := generateID()

	t.Run("Setup", func(t *testing.T) {
		resp := createSessionWithSettings(t, agent1Port(), claudeSid,
			"claude --dangerously-skip-permissions",
			"/home/agent/"+claudeSid,
			"auto",
			settings,
		)
		assertOK(t, resp, "create claude session for restore test")

		resp = apiPost(t, agent1Port(), "/api/v1/sessions", CreateSessionRequest{
			Name:            bashSid,
			Command:         "bash",
			WorkingDir:      "/home/agent",
			RestoreStrategy: "manual",
		})
		assertOK(t, resp, "create bash session for restore test")

		time.Sleep(3 * time.Second)
	})

	t.Run("AutoSave", func(t *testing.T) {
		beforeTime := dockerExec(t, agent1Container(),
			"stat", "-c", "%Y", "/home/agent/.session-state/"+claudeSid+".json")

		interval := 60
		t.Logf("Waiting %d seconds for autosave...", interval+10)
		time.Sleep(time.Duration(interval+10) * time.Second)

		afterTime := dockerExec(t, agent1Container(),
			"stat", "-c", "%Y", "/home/agent/.session-state/"+claudeSid+".json")

		if beforeTime == afterTime || afterTime == "" {
			t.Fatalf("autosave did not update file mtime (before=%s, after=%s)", beforeTime, afterTime)
		}
		t.Logf("File mtime changed: %s -> %s", beforeTime, afterTime)
	})

	t.Run("AutoRestore_AfterRestart", func(t *testing.T) {
		t.Log("Restarting agent-1 container...")
		cmd := exec.Command("docker", "compose", "-f", "../docker-compose.yml", "restart", "agent-1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("restart agent-1 failed: %v, output: %s", err, string(out))
		}

		// 轮询等待 health check 通过
		t.Log("Waiting for agent-1 to become healthy...")
		healthy := false
		for i := 0; i < 30; i++ {
			time.Sleep(2 * time.Second)
			resp, err := http.Get("http://localhost:" + agent1Port() + "/health")
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				healthy = true
				t.Logf("agent-1 healthy after %d seconds", (i+1)*2)
				break
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
		if !healthy {
			t.Fatal("agent-1 did not become healthy within 60 seconds")
		}

		// 额外等待 entrypoint.sh 完成恢复流程
		time.Sleep(5 * time.Second)

		if !dockerExecCheckSession(t, agent1Container(), claudeSid) {
			t.Fatalf("auto strategy session %s not restored after restart", claudeSid)
		}

		logCmd := exec.Command("docker", "compose", "-f", "../docker-compose.yml", "logs", "agent-1", "--tail", "50")
		logOut, _ := logCmd.CombinedOutput()
		logStr := string(logOut)
		if !strings.Contains(logStr, "[restore] restoring session") {
			t.Fatalf("restore log not found in agent-1 logs")
		}
	})

	t.Run("ManualRestore", func(t *testing.T) {
		resp := apiPost(t, agent1Port(), "/api/v1/sessions/"+bashSid+"/restore", map[string]string{})
		assertOK(t, resp, "restore bash session")
		time.Sleep(3 * time.Second)

		if !dockerExecCheckSession(t, agent1Container(), bashSid) {
			t.Fatalf("manual restore: tmux session %s not found", bashSid)
		}
	})

	t.Run("Cleanup", func(t *testing.T) {
		apiDelete(t, agent1Port(), "/api/v1/sessions/"+claudeSid)
		apiDelete(t, agent1Port(), "/api/v1/sessions/"+bashSid)
	})
}
