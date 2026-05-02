//go:build integration

package integration

import (
	"context"
	"testing"

	client "github.com/charviki/maze-cradle/api/gen/http"
)

// TestTerminalGetOutput — Given: 已上线的 Host 和活跃 Session; When: 获取终端输出; Then: 返回输出内容
func TestTerminalGetOutput(t *testing.T) {
	t.Parallel()
	h := newTestHelper(t)
	defer h.cleanup(t)

	nodeName := uniqueName("test-output")
	h.trackHost(nodeName)
	h.createHostAndWait(t, nodeName, []string{"claude"})

	sid := h.createSession(t, nodeName, "output-test-session")

	t.Log("[step] getting terminal output...")
	output, _, err := h.apiClient.SessionServiceAPI.SessionServiceGetOutput(context.Background(), nodeName, sid).Execute()
	if err != nil {
		t.Fatalf("get output failed: %v", err)
	}
	t.Logf("[step] PASS: terminal output length=%d", len(output.GetOutput()))
}

// TestTerminalSendInput — Given: 已上线的 Host 和活跃 Session; When: 发送命令; Then: 请求成功
func TestTerminalSendInput(t *testing.T) {
	t.Parallel()
	h := newTestHelper(t)
	defer h.cleanup(t)

	nodeName := uniqueName("test-input")
	h.trackHost(nodeName)
	h.createHostAndWait(t, nodeName, []string{"claude"})

	sid := h.createSession(t, nodeName, "input-test-session")

	t.Log("[step] sending input...")
	command := "echo hello\n"
	_, _, err := h.apiClient.SessionServiceAPI.SessionServiceSendInput(context.Background(), nodeName, sid).
		Body(client.SessionServiceSendInputBody{Command: &command}).Execute()
	if err != nil {
		t.Fatalf("send input failed: %v", err)
	}
	t.Log("[step] PASS: input sent successfully")
}

// TestTerminalGetEnv — Given: 已上线的 Host 和活跃 Session; When: 查询环境变量; Then: 返回环境变量列表
func TestTerminalGetEnv(t *testing.T) {
	t.Parallel()
	h := newTestHelper(t)
	defer h.cleanup(t)

	nodeName := uniqueName("test-env")
	h.trackHost(nodeName)
	h.createHostAndWait(t, nodeName, []string{"claude"})

	sid := h.createSession(t, nodeName, "env-test-session")

	t.Log("[step] getting environment variables...")
	envResp, _, err := h.apiClient.SessionServiceAPI.SessionServiceGetEnv(context.Background(), nodeName, sid).Execute()
	if err != nil {
		t.Fatalf("get env failed: %v", err)
	}
	env := envResp.GetEnv()
	t.Logf("[step] PASS: found %d env vars", len(env))
}

// TestTerminalSendSignal — Given: 已上线的 Host 和活跃 Session; When: 发送中断信号; Then: 请求成功
func TestTerminalSendSignal(t *testing.T) {
	t.Parallel()
	h := newTestHelper(t)
	defer h.cleanup(t)

	nodeName := uniqueName("test-signal")
	h.trackHost(nodeName)
	h.createHostAndWait(t, nodeName, []string{"claude"})

	sid := h.createSession(t, nodeName, "signal-test-session")

	t.Log("[step] sending SIGINT signal...")
	signal := "SIGINT"
	_, _, err := h.apiClient.SessionServiceAPI.SessionServiceSendSignal(context.Background(), nodeName, sid).
		Body(client.SessionServiceSendSignalBody{Signal: &signal}).Execute()
	if err != nil {
		t.Fatalf("send signal failed: %v", err)
	}
	t.Log("[step] PASS: signal sent successfully")
}
