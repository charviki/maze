package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type APIResponse struct {
	Status  string          `json:"status"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message,omitempty"`
}

type Session struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	WindowCount int    `json:"window_count"`
}

type ConfigItem struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CreateSessionRequest struct {
	Name            string       `json:"name"`
	Command         string       `json:"command"`
	WorkingDir      string       `json:"working_dir,omitempty"`
	SessionConfs    []ConfigItem `json:"session_confs"`
	RestoreStrategy string       `json:"restore_strategy,omitempty"`
	TemplateID      string       `json:"template_id,omitempty"`
}

type TerminalOutput struct {
	SessionID string `json:"session_id"`
	Lines     int    `json:"lines"`
	Output    string `json:"output"`
}

type Node struct {
	Name          string `json:"name"`
	Address       string `json:"address"`
	ExternalAddr  string `json:"external_addr"`
	SessionCount  int    `json:"session_count"`
	Status        string `json:"status"`
	RegisteredAt  string `json:"registered_at"`
	LastHeartbeat string `json:"last_heartbeat"`
}

type Template struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Builtin     bool   `json:"builtin"`
}

type NodeConfig struct {
	WorkingDir string            `json:"working_dir"`
	Env        map[string]string `json:"env"`
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func agent1Port() string      { return getEnv("AGENT_1_PORT", "8081") }
func agent2Port() string      { return getEnv("AGENT_2_PORT", "8082") }
func agent3Port() string      { return getEnv("AGENT_3_PORT", "8083") }
func managerPort() string     { return getEnv("MANAGER_PORT", "8090") }
func webPort() string         { return getEnv("WEB_PORT", "10800") }
func agent1Container() string { return getEnv("AGENT_1_CONTAINER", "agent-claude-1") }

func baseURL(port string) string {
	return fmt.Sprintf("http://localhost:%s", port)
}

func apiGet(t *testing.T, port, path string) *APIResponse {
	t.Helper()
	resp, err := http.Get(baseURL(port) + path)
	if err != nil {
		t.Fatalf("GET %s%s failed: %v", port, path, err)
	}
	defer resp.Body.Close()
	return parseResponse(t, resp.Body)
}

func apiPost(t *testing.T, port, path string, body interface{}) *APIResponse {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body failed: %v", err)
	}
	resp, err := http.Post(baseURL(port)+path, "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST %s%s failed: %v", port, path, err)
	}
	defer resp.Body.Close()
	return parseResponse(t, resp.Body)
}

func apiPut(t *testing.T, port, path string, body interface{}) *APIResponse {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body failed: %v", err)
	}
	req, err := http.NewRequest("PUT", baseURL(port)+path, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("create PUT request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT %s%s failed: %v", port, path, err)
	}
	defer resp.Body.Close()
	return parseResponse(t, resp.Body)
}

func apiDelete(t *testing.T, port, path string) *APIResponse {
	t.Helper()
	req, err := http.NewRequest("DELETE", baseURL(port)+path, nil)
	if err != nil {
		t.Fatalf("create DELETE request failed: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE %s%s failed: %v", port, path, err)
	}
	defer resp.Body.Close()
	return parseResponse(t, resp.Body)
}

func parseResponse(t *testing.T, body io.Reader) *APIResponse {
	t.Helper()
	var apiResp APIResponse
	if err := json.NewDecoder(body).Decode(&apiResp); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	return &apiResp
}

func assertOK(t *testing.T, resp *APIResponse, msg string) {
	t.Helper()
	if resp.Status != "ok" {
		t.Fatalf("%s: expected status=ok, got status=%s, body=%s", msg, resp.Status, string(resp.Data))
	}
}

func assertError(t *testing.T, resp *APIResponse, msg string) {
	t.Helper()
	if resp.Status != "error" {
		t.Fatalf("%s: expected status=error, got status=%s", msg, resp.Status)
	}
}

func decodeData[T any](t *testing.T, resp *APIResponse) T {
	t.Helper()
	var result T
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		t.Fatalf("decode data failed: %v, raw=%s", err, string(resp.Data))
	}
	return result
}

func generateID() string {
	return fmt.Sprintf("test-%d-%d", time.Now().Unix(), time.Now().Nanosecond()%10000)
}

func dockerExec(t *testing.T, container string, args ...string) string {
	t.Helper()
	cmd := exec.Command("docker", append([]string{"exec", container}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("docker exec %s %v failed: %v, output: %s", container, args, err, string(out))
	}
	return strings.TrimSpace(string(out))
}

func dockerExecCheckSession(t *testing.T, container, sessionName string) bool {
	t.Helper()
	cmd := exec.Command("docker", "exec", container, "tmux", "has-session", "-t", sessionName)
	err := cmd.Run()
	return err == nil
}

func dockerExecReadState(t *testing.T, container, sessionName string) string {
	t.Helper()
	cmd := exec.Command("docker", "exec", container, "cat", fmt.Sprintf("/home/agent/.session-state/%s.json", sessionName))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func loadClaudeSettings(t *testing.T) json.RawMessage {
	t.Helper()
	path := os.Getenv("CLAUDE_SETTINGS_PATH")
	if path == "" {
		t.Fatal("CLAUDE_SETTINGS_PATH not set")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read settings.json failed: %v", err)
	}
	if !json.Valid(data) {
		t.Fatal("settings.json is not valid JSON")
	}
	return json.RawMessage(data)
}

func createSessionWithSettings(t *testing.T, port, name, command, workingDir, restoreStrategy string, settings json.RawMessage) *APIResponse {
	t.Helper()
	var settingsStr string
	if err := json.Unmarshal(settings, &settingsStr); err != nil {
		settingsStr = string(settings)
	}

	req := CreateSessionRequest{
		Name:            name,
		Command:         command,
		WorkingDir:      workingDir,
		RestoreStrategy: restoreStrategy,
		TemplateID:      "claude",
		SessionConfs: []ConfigItem{
			{
				Type:  "file",
				Key:   ".claude/settings.json",
				Value: settingsStr,
			},
		},
	}
	return apiPost(t, port, "/api/v1/sessions", req)
}
