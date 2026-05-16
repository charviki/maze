package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charviki/maze/fabrication/cradle/configutil"
)

// ClaudeProvider 是 Claude Code CLI 的 Provider 实现。
type ClaudeProvider struct{}

// ID 返回 Provider 唯一标识符 "claude"。
func (p *ClaudeProvider) ID() string { return "claude" }

// SessionIDPlaceholder 返回 {session_id}。
func (p *ClaudeProvider) SessionIDPlaceholder() string { return "{session_id}" }

// RestoreCommandTemplate 返回空串（恢复命令由模板 YAML 定义）。
func (p *ClaudeProvider) RestoreCommandTemplate() string { return "" }

// BootstrapTask 返回创建 Session 前的信任注入任务。
func (p *ClaudeProvider) BootstrapTask() Task {
	return Task{
		Name:        "claude-bootstrap-trust",
		Description: "为工作目录注入 Claude Code 信任配置到 ~/.claude.json",
		Run: func(ctx TaskContext) error {
			if ctx.WorkingDir == "" {
				return nil
			}
			home := ctx.HomeDir
			if home == "" {
				home = ResolveHomeDir()
			}
			configPath := filepath.Join(home, ".claude.json")

			config := make(map[string]json.RawMessage)
			data, err := os.ReadFile(filepath.Clean(configPath))
			if err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				_ = json.Unmarshal(data, &config)
			}

			projects := make(map[string]claudeProjectEntry)
			if rawProjects, ok := config["projects"]; ok && len(rawProjects) > 0 {
				_ = json.Unmarshal(rawProjects, &projects)
			}

			entry := projects[ctx.WorkingDir]
			entry.HasTrustDialogAccepted = true
			entry.HasCompletedProjectOnboarding = true
			projects[ctx.WorkingDir] = entry

			projectsJSON, err := json.Marshal(projects)
			if err != nil {
				return err
			}
			config["projects"] = projectsJSON

			updated, err := json.MarshalIndent(config, "", "  ")
			if err != nil {
				return err
			}
			// 0600: 与 EntrypointTasks 中 ~/.claude.json 写入权限保持一致
			return configutil.AtomicWriteFile(configPath, updated, 0600) //nolint:gosec
		},
	}
}

// EntrypointTasks 返回容器启动时需要的初始化任务。
func (p *ClaudeProvider) EntrypointTasks() []Task {
	return []Task{
		{
			Name:        "init-claude-json",
			Description: "初始化 ~/.claude.json 默认值并确保 onboarding 完成",
			Run: func(ctx TaskContext) error {
				if ctx.HomeDir == "" {
					return nil
				}
				configPath := filepath.Join(ctx.HomeDir, ".claude.json")

				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					defaultCfg := map[string]any{
						"hasCompletedOnboarding":      true,
						"firstStartTime":              "",
						"opusProMigrationComplete":    true,
						"sonnet1m45MigrationComplete": true,
						"migrationVersion":            11,
						"projects": map[string]any{
							ctx.HomeDir: map[string]any{"allowedTools": []any{}},
						},
					}
					data, err := json.MarshalIndent(defaultCfg, "", "  ")
					if err != nil {
						return err
					}
					// 0600: 仅用户可读写，防止其他用户读取 onboarding 状态
					if err := os.WriteFile(configPath, data, 0600); err != nil {
						return err
					}
				}

				//nolint:gosec // configPath 由 HomeDir 拼接，非外部输入
				raw, err := os.ReadFile(configPath)
				if err != nil {
					return err
				}
				var cfg map[string]any
				if err := json.Unmarshal(raw, &cfg); err != nil {
					return err
				}
				cfg["hasCompletedOnboarding"] = true
				updated, err := json.MarshalIndent(cfg, "", "  ")
				if err != nil {
					return err
				}
				// 0600: 仅用户可读写
				return os.WriteFile(configPath, updated, 0600)
			},
		},
		{
			Name:        "init-claude-settings",
			Description: "初始化 ~/.claude/settings.json 默认权限配置",
			Run: func(ctx TaskContext) error {
				if ctx.HomeDir == "" {
					return nil
				}
				settingsDir := filepath.Join(ctx.HomeDir, ".claude")
				settingsPath := filepath.Join(settingsDir, "settings.json")

				// 0750: 用户 rwx，组 rx，其他无权限
				if err := os.MkdirAll(settingsDir, 0750); err != nil {
					return err
				}

				if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
					// 0600: 仅用户可读写
					if err := os.WriteFile(settingsPath, []byte("{}"), 0600); err != nil {
						return err
					}
				}

				//nolint:gosec // settingsPath 由 HomeDir 拼接，非外部输入
				raw, err := os.ReadFile(settingsPath)
				if err != nil {
					return err
				}
				var cfg map[string]any
				if err := json.Unmarshal(raw, &cfg); err != nil {
					return err
				}

				permissions, _ := cfg["permissions"].(map[string]any)
				if permissions == nil {
					permissions = make(map[string]any)
				}
				if _, ok := permissions["allow"]; !ok {
					permissions["allow"] = []string{
						"Bash(*)", "Read(*)", "Write(*)", "Edit(*)",
						"MultiEdit(*)", "WebFetch(*)", "WebSearch(*)",
					}
				}
				if _, ok := permissions["deny"]; !ok {
					permissions["deny"] = []string{}
				}
				permissions["skipDangerousModePermissionPrompt"] = true
				cfg["permissions"] = permissions
				cfg["skipDangerousModePermissionPrompt"] = true
				if _, ok := cfg["theme"]; !ok {
					cfg["theme"] = "dark"
				}

				updated, err := json.MarshalIndent(cfg, "", "  ")
				if err != nil {
					return err
				}
				// 0600: 仅用户可读写
				return os.WriteFile(settingsPath, updated, 0600)
			},
		},
	}
}

// HealthCheckTask 检查 claude 是否可用：二进制在 PATH 中且关键配置文件已初始化且可解析。
// 不仅检查文件存在，还验证 JSON 可读取、可反序列化，覆盖文件损坏/不可读等场景。
// 如果初始化失败导致文件缺失或损坏，HealthCheck 会失败，
// ListAvailable() 不会包含 claude，避免控制面宣告可用但运行时不可用的状态分裂。
func (p *ClaudeProvider) HealthCheckTask() Task {
	return Task{
		Name:        "claude-health-check",
		Description: "检查 claude 二进制和关键配置文件可读可解析",
		Run: func(_ TaskContext) error {
			if _, err := exec.LookPath("claude"); err != nil {
				return err
			}
			home := ResolveHomeDir()
			// ~/.claude.json 必须存在且可解析为 JSON 对象
			configPath := filepath.Join(home, ".claude.json")
			if err := validateJSONFile(configPath); err != nil {
				return fmt.Errorf("claude config invalid: %w", err)
			}
			// ~/.claude/settings.json 必须存在且可解析为 JSON 对象
			settingsPath := filepath.Join(home, ".claude", "settings.json")
			if err := validateJSONFile(settingsPath); err != nil {
				return fmt.Errorf("claude settings invalid: %w", err)
			}
			return nil
		},
	}
}

// validateJSONFile 验证文件存在、可读取、可解析为 JSON 对象。
func validateJSONFile(path string) error {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	var target map[string]any
	if err := json.Unmarshal(data, &target); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	if target == nil {
		return fmt.Errorf("parse %s: expected JSON object, got null", path)
	}
	return nil
}

type claudeProjectEntry struct {
	HasTrustDialogAccepted        bool `json:"hasTrustDialogAccepted"`
	HasCompletedProjectOnboarding bool `json:"hasCompletedProjectOnboarding"`
}
