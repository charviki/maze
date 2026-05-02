package service

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/charviki/maze-cradle/configutil"
)

// claudeProjectEntry Claude CLI 对单个项目目录的信任配置
type claudeProjectEntry struct {
	HasTrustDialogAccepted        bool `json:"hasTrustDialogAccepted"`
	HasCompletedProjectOnboarding bool `json:"hasCompletedProjectOnboarding"`
}

// ClaudeTrustBootstrapper 为 Claude CLI 注入工作目录信任配置。
// 修改 ~/.claude.json 中的 projects 字段，标记指定目录的信任状态。
type ClaudeTrustBootstrapper struct{}

// TrustDir 为指定工作目录设置 Claude Code 信任
func (c *ClaudeTrustBootstrapper) TrustDir(workingDir string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(home, ".claude.json")

	config := make(map[string]json.RawMessage)
	data, err := os.ReadFile(filepath.Clean(configPath))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// 文件不存在，使用零值初始化
	} else {
		// 解析失败时降级为空配置，不阻塞流程，但要尽量保留已有顶层字段。
		_ = json.Unmarshal(data, &config)
	}

	projects := make(map[string]claudeProjectEntry)
	if rawProjects, ok := config["projects"]; ok && len(rawProjects) > 0 {
		// projects 解析失败时降级为空，避免因为单个坏字段阻塞 trust 注入。
		_ = json.Unmarshal(rawProjects, &projects)
	}

	entry := projects[workingDir]
	entry.HasTrustDialogAccepted = true
	entry.HasCompletedProjectOnboarding = true
	projects[workingDir] = entry

	projectsJSON, err := json.Marshal(projects)
	if err != nil {
		return err
	}
	config["projects"] = projectsJSON

	updated, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	// 使用原子写入避免并发信任注入时文件损坏
	return configutil.AtomicWriteFile(configPath, updated, 0644)
}
