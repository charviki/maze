package builder

import (
	"github.com/charviki/maze-cradle/protocol"
)

// DefaultToolRegistry 默认工具注册表，记录所有可选配的工具
var DefaultToolRegistry = map[string]protocol.ToolConfig{
	"claude": {
		ID:          "claude",
		Image:       "maze-deps-claude:latest",
		SourcePath:  "/opt/claude",
		DestPath:    "/opt/claude",
		BinPaths:    []string{"/opt/claude/bin"},
		Description: "Claude Code CLI — Anthropic 的 AI 编程助手",
		Category:    "cli",
	},
	"codex": {
		ID:          "codex",
		Image:       "maze-deps-codex:latest",
		SourcePath:  "/opt/codex",
		DestPath:    "/opt/codex",
		BinPaths:    []string{"/opt/codex/bin"},
		Description: "OpenAI Codex CLI — AI 编程助手",
		Category:    "cli",
	},
	"go": {
		ID:         "go",
		Image:      "maze-deps-go:latest",
		SourcePath: "/opt/go",
		DestPath:   "/opt/go",
		BinPaths:   []string{"/opt/go/bin"},
		EnvVars: map[string]string{
			"GOROOT":  "/opt/go",
			"GOPROXY": "https://goproxy.cn,direct",
			"GOSUMDB": "sum.golang.org",
		},
		Description: "Go 1.24 工具链及常用开发工具",
		Category:    "language",
	},
	"python": {
		ID:          "python",
		Image:       "maze-deps-python:latest",
		SourcePath:  "/opt/python",
		DestPath:    "/opt/python",
		BinPaths:    []string{"/opt/python/bin"},
		Description: "Python 3.12 及常用包",
		Category:    "language",
	},
	"node": {
		ID:          "node",
		Image:       "maze-deps-node:latest",
		SourcePath:  "/opt/node",
		DestPath:    "/opt/node",
		BinPaths:    []string{"/opt/node/bin"},
		Description: "Node.js pnpm store 及前端依赖",
		Category:    "language",
	},
}

// ListAvailableTools 返回所有可用工具列表
func ListAvailableTools() []protocol.ToolConfig {
	tools := make([]protocol.ToolConfig, 0, len(DefaultToolRegistry))
	for _, cfg := range DefaultToolRegistry {
		tools = append(tools, cfg)
	}
	return tools
}

// ValidateTools 验证工具列表是否全部在注册表中
// 返回未知的工具名列表（空表示全部有效）
func ValidateTools(toolIDs []string) []string {
	var unknown []string
	for _, id := range toolIDs {
		if _, ok := DefaultToolRegistry[id]; !ok {
			unknown = append(unknown, id)
		}
	}
	return unknown
}

// build cache test
