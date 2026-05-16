package provider

// BashProvider 是 bash 模板的 Provider 实现，无需任何特殊初始化。
type BashProvider struct{}

// ID 返回 Provider 唯一标识符 "bash"。
func (p *BashProvider) ID() string { return "bash" }

// SessionIDPlaceholder 返回空串（bash 不支持 session ID 注入）。
func (p *BashProvider) SessionIDPlaceholder() string { return "" }

// RestoreCommandTemplate 返回空串（bash 不支持恢复）。
func (p *BashProvider) RestoreCommandTemplate() string { return "" }

// BootstrapTask 返回无操作任务，bash 不需要前置初始化。
func (p *BashProvider) BootstrapTask() Task {
	return Task{Name: "bash-bootstrap", Description: "no-op"}
}

// EntrypointTasks 返回空列表，bash 不需要容器启动初始化。
func (p *BashProvider) EntrypointTasks() []Task { return nil }

// HealthCheckTask 返回 Run=nil 的任务，bash 始终可用。
func (p *BashProvider) HealthCheckTask() Task {
	return Task{Name: "bash-health-check", Description: "always available"}
}
