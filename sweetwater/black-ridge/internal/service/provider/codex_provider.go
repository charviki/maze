package provider

import (
	"os/exec"
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
