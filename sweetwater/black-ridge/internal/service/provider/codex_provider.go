package provider

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
)

// CodexProvider 是 OpenAI Codex CLI 的 Provider 实现。
type CodexProvider struct{}

// ID 返回 Provider 唯一标识符 "codex"。
func (p *CodexProvider) ID() string { return "codex" }

// SessionIDPlaceholder 返回空串（Codex 不支持 session ID 注入）。
func (p *CodexProvider) SessionIDPlaceholder() string { return "" }

// RestoreCommandTemplate 返回空串（恢复命令由模板 YAML 定义）。
func (p *CodexProvider) RestoreCommandTemplate() string { return "" }

// BootstrapTask 返回无操作任务，Codex 无信任注入需求。
func (p *CodexProvider) BootstrapTask() Task {
	return Task{Name: "codex-bootstrap", Description: "no-op"}
}

// EntrypointTasks 返回空列表，Codex 不需要容器启动初始化。
func (p *CodexProvider) EntrypointTasks() []Task { return nil }

// HealthCheckTask 检查 codex 二进制是否在 PATH 中可用。
func (p *CodexProvider) HealthCheckTask() Task {
	return Task{
		Name:        "codex-health-check",
		Description: "检查 codex 二进制是否在 PATH 中可用",
		Run: func(_ TaskContext) error {
			_, err := exec.LookPath("codex")
			return err
		},
	}
}

// CompletionHookConfig 生成 Codex hooks.json 配置，注入 SessionStart + Stop hook。
// Codex 的 hooks.json 是独立的 hook 配置文件，无需与现有配置合并：
//   - SessionStart: CLI 启动就绪时 touch ready 信号文件
//   - Stop: CLI 完成回复时 touch done 信号文件
func (p *CodexProvider) CompletionHookConfig(homeDir, sessionName string) *CompletionHookConfig {
	hooksPath := filepath.Join(homeDir, ".codex", "hooks.json")
	signalFile := filepath.Join(os.TempDir(), "step_done_"+sessionName)
	readySignalFile := filepath.Join(os.TempDir(), "step_ready_"+sessionName)

	hooksConfig := map[string]any{
		"hooks": []any{
			map[string]any{
				"event":   "SessionStart",
				"matcher": "",
				"hooks": []any{
					map[string]any{
						"type":    "command",
						"command": "touch " + readySignalFile,
					},
				},
			},
			map[string]any{
				"event":   "Stop",
				"matcher": "",
				"hooks": []any{
					map[string]any{
						"type":    "command",
						"command": "touch " + signalFile,
					},
				},
			},
		},
	}

	content, err := json.MarshalIndent(hooksConfig, "", "  ")
	if err != nil {
		return nil
	}

	return &CompletionHookConfig{
		ConfigPath:      hooksPath,
		ConfigContent:   string(content),
		SignalFile:      signalFile,
		ReadySignalFile: readySignalFile,
	}
}
