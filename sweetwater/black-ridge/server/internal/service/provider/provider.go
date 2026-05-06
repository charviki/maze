package provider

import (
	"fmt"
	"os"
	"sort"

	"github.com/charviki/maze-cradle/logutil"
)

// Provider 定义 AI CLI 工具与 Agent 引擎的集成协议。
type Provider interface {
	ID() string
	SessionIDPlaceholder() string
	RestoreCommandTemplate() string
	BootstrapTask() Task
	EntrypointTasks() []Task
	HealthCheckTask() Task
}

// Task 描述一个 Provider 需要执行的操作。
type Task struct {
	Name        string
	Description string
	Run         func(ctx TaskContext) error
}

// TaskContext 为 Task.Run 提供运行时上下文。
type TaskContext struct {
	HomeDir    string
	WorkingDir string
}

// ResolveHomeDir 返回 Provider 应使用的主目录。
// 优先读取 AGENT_HOME 环境变量（与 entrypoint.sh 保持一致），
// 未设置时降级到 os.UserHomeDir()，确保 Go 代码和 shell 脚本写入同一目录。
func ResolveHomeDir() string {
	if agentHome := os.Getenv("AGENT_HOME"); agentHome != "" {
		return agentHome
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "/home/agent"
	}
	return home
}

// Registry 管理所有已注册的 Provider。
type Registry struct {
	providers map[string]Provider
}

// NewRegistry 创建空的 Provider 注册表。
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register 注册一个 Provider。
func (r *Registry) Register(p Provider) {
	r.providers[p.ID()] = p
}

// Get 按 ID 查找 Provider。
func (r *Registry) Get(id string) Provider {
	return r.providers[id]
}

// All 返回所有已注册 Provider。
func (r *Registry) All() []Provider {
	result := make([]Provider, 0, len(r.providers))
	for _, p := range r.providers {
		result = append(result, p)
	}
	return result
}

// ListAvailable 返回所有通过 HealthCheck 的 Provider ID 列表。
func (r *Registry) ListAvailable() []string {
	var available []string
	for _, p := range r.providers {
		task := p.HealthCheckTask()
		if task.Run == nil {
			available = append(available, p.ID())
			continue
		}
		if err := task.Run(TaskContext{}); err == nil {
			available = append(available, p.ID())
		}
	}
	sort.Strings(available)
	return available
}

// RunTask 统一执行 Provider 任务，处理 nil Run 和错误日志。
func RunTask(logger logutil.Logger, task Task, ctx TaskContext) error {
	if task.Run == nil {
		return nil
	}
	if err := task.Run(ctx); err != nil {
		if logger != nil {
			logger.Warnf("[provider] task %q failed: %v", task.Name, err)
		}
		return fmt.Errorf("provider task %q: %w", task.Name, err)
	}
	return nil
}

// RunEntrypointTasks 执行所有 Provider 的容器启动初始化任务。
// 关键任务失败记录 error 级别日志，不阻塞进程启动；
// 失败的 Provider 在后续 HealthCheck 时仍会被 ListAvailable 过滤。
func RunEntrypointTasks(logger logutil.Logger, registry *Registry, homeDir string) {
	for _, p := range registry.All() {
		for _, task := range p.EntrypointTasks() {
			if err := RunTask(logger, task, TaskContext{HomeDir: homeDir}); err != nil {
				// EntrypointTask 失败记录 error 级别，便于排查初始化问题
				if logger != nil {
					logger.Errorf("[provider] provider %q entrypoint task %q failed (provider may be degraded): %v", p.ID(), task.Name, err)
				}
			}
		}
	}
}
