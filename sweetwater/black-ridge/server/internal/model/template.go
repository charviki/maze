package model

import (
	"embed"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/charviki/maze-cradle/configutil"
	"github.com/charviki/maze-cradle/logutil"
	"gopkg.in/yaml.v3"
)

//go:embed templates/*.yaml
var builtinTemplatesFS embed.FS

// SessionTemplate 会话模板，定义创建 Agent 会话时需要的命令、配置和用户输入字段
type SessionTemplate struct {
	ID                 string                   `json:"id"`
	Name               string                   `json:"name"`
	Command            string                   `json:"command"`
	RestoreCommand     string                   `json:"restore_command"`
	SessionFilePattern string                   `json:"session_file_pattern"`
	Description        string                   `json:"description"`
	Icon               string                   `json:"icon"`
	Builtin            bool                     `json:"builtin"`
	Defaults           configutil.ConfigLayer   `json:"defaults"`
	SessionSchema      configutil.SessionSchema `json:"session_schema"`
}

// TemplateStore 模板持久化存储。内置模板在每次启动时会被无条件覆盖
type TemplateStore struct {
	mu        sync.RWMutex
	templates map[string]*SessionTemplate
	path      string
	logger    logutil.Logger
}

// NewTemplateStore 创建 TemplateStore，加载已有数据后注入内置模板
func NewTemplateStore(filePath string, logger logutil.Logger) *TemplateStore {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		logger.Warnf("[template] create dir %s: %v", dir, err)
	}
	s := &TemplateStore{
		templates: make(map[string]*SessionTemplate),
		path:      filePath,
		logger:    logger,
	}
	s.load()
	s.ensureBuiltins()
	return s
}

// 从 JSON 文件加载模板数据。文件不存在时为首次启动，解析失败时记录错误日志
func (s *TemplateStore) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		s.logger.Infof("[template] file not found, starting fresh: %s", s.path)
		return
	}
	var templates map[string]*SessionTemplate
	if err := json.Unmarshal(data, &templates); err != nil {
		s.logger.Errorf("[template] parse file %s failed: %v", s.path, err)
		return
	}
	s.mu.Lock()
	s.templates = templates
	s.mu.Unlock()
}

// 持久化模板数据到 JSON 文件（原子写入）
func (s *TemplateStore) save() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.templates, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}
	return configutil.AtomicWriteFile(s.path, data, 0644)
}

// --- YAML 解析中间结构体 ---
// SessionTemplate 及 cradle 共享类型只有 json tag，没有 yaml tag。
// yaml.v3 默认使用字段名小写（如 RestoreCommand → restorecommand），
// 与 YAML 文件中的 snake_case 键名不匹配，因此需要中间结构体桥接。

type yamlConfigFile struct {
	Path    string `yaml:"path"`
	Content string `yaml:"content"`
}

type yamlConfigLayer struct {
	Env   map[string]string `yaml:"env"`
	Files []yamlConfigFile  `yaml:"files"`
}

type yamlEnvDef struct {
	Key         string `yaml:"key"`
	Label       string `yaml:"label"`
	Required    bool   `yaml:"required"`
	Placeholder string `yaml:"placeholder"`
	Sensitive   bool   `yaml:"sensitive"`
}

type yamlFileDef struct {
	Path           string `yaml:"path"`
	Label          string `yaml:"label"`
	Required       bool   `yaml:"required"`
	DefaultContent string `yaml:"default_content"`
}

type yamlSessionSchema struct {
	EnvDefs  []yamlEnvDef  `yaml:"env_defs"`
	FileDefs []yamlFileDef `yaml:"file_defs"`
}

type yamlTemplate struct {
	ID                 string            `yaml:"id"`
	Name               string            `yaml:"name"`
	Command            string            `yaml:"command"`
	RestoreCommand     string            `yaml:"restore_command"`
	SessionFilePattern string            `yaml:"session_file_pattern"`
	Description        string            `yaml:"description"`
	Icon               string            `yaml:"icon"`
	Builtin            bool              `yaml:"builtin"`
	Defaults           yamlConfigLayer   `yaml:"defaults"`
	SessionSchema      yamlSessionSchema `yaml:"session_schema"`
}

// 将 YAML 中间结构体转换为 SessionTemplate
func (y *yamlTemplate) toSessionTemplate() *SessionTemplate {
	files := make([]configutil.ConfigFile, len(y.Defaults.Files))
	for i, f := range y.Defaults.Files {
		files[i] = configutil.ConfigFile{Path: f.Path, Content: f.Content}
	}

	envDefs := make([]configutil.EnvDef, len(y.SessionSchema.EnvDefs))
	for i, d := range y.SessionSchema.EnvDefs {
		envDefs[i] = configutil.EnvDef{
			Key: d.Key, Label: d.Label, Required: d.Required,
			Placeholder: d.Placeholder, Sensitive: d.Sensitive,
		}
	}

	fileDefs := make([]configutil.FileDef, len(y.SessionSchema.FileDefs))
	for i, d := range y.SessionSchema.FileDefs {
		fileDefs[i] = configutil.FileDef{
			Path: d.Path, Label: d.Label, Required: d.Required,
			DefaultContent: d.DefaultContent,
		}
	}

	return &SessionTemplate{
		ID: y.ID, Name: y.Name, Command: y.Command,
		RestoreCommand: y.RestoreCommand, SessionFilePattern: y.SessionFilePattern,
		Description: y.Description, Icon: y.Icon, Builtin: y.Builtin,
		Defaults:      configutil.ConfigLayer{Env: y.Defaults.Env, Files: files},
		SessionSchema: configutil.SessionSchema{EnvDefs: envDefs, FileDefs: fileDefs},
	}
}

// loadBuiltinFromYAML 从 embed.FS 读取并解析单个 YAML 模板文件
func loadBuiltinFromYAML(name string) (*SessionTemplate, error) {
	data, err := builtinTemplatesFS.ReadFile("templates/" + name)
	if err != nil {
		return nil, err
	}
	var yt yamlTemplate
	if err := yaml.Unmarshal(data, &yt); err != nil {
		return nil, err
	}
	return yt.toSessionTemplate(), nil
}

// 注入内置模板（Claude Code、Codex、Bash Shell）。
// 优先从 embed FS 中的 YAML 文件加载；加载失败时降级到硬编码模板（向后兼容）
func (s *TemplateStore) ensureBuiltins() {
	builtins, err := s.loadBuiltinsFromFS()
	if err != nil {
		s.logger.Errorf("[template] load builtins from embed FS failed: %v, falling back to hardcoded", err)
		builtins = hardcodedBuiltins()
	}

	s.mu.Lock()
	for _, t := range builtins {
		tpl := t
		s.templates[t.ID] = &tpl
	}
	s.mu.Unlock()
	if err := s.save(); err != nil {
		s.logger.Errorf("[template] save builtins failed: %v", err)
	}
}

// loadBuiltinsFromFS 从 embed FS 中依次加载所有内置模板 YAML 文件
func (s *TemplateStore) loadBuiltinsFromFS() ([]SessionTemplate, error) {
	names := []string{"claude.yaml", "codex.yaml", "bash.yaml"}
	var builtins []SessionTemplate
	for _, name := range names {
		tpl, err := loadBuiltinFromYAML(name)
		if err != nil {
			return nil, err
		}
		builtins = append(builtins, *tpl)
	}
	return builtins, nil
}

// hardcodedBuiltins 返回硬编码的内置模板（作为 embed FS 加载失败时的降级方案）
func hardcodedBuiltins() []SessionTemplate {
	return []SessionTemplate{
		{
			ID: "claude", Name: "Claude Code", Command: `IS_SANDBOX=1 claude --dangerously-skip-permissions --session-id {session_id}`,
			RestoreCommand:     `IS_SANDBOX=1 claude --dangerously-skip-permissions --resume {session_id}`,
			SessionFilePattern: `~/.claude/projects/{encoded_working_dir}/*.jsonl`,
			Description:        "Anthropic Claude CLI Agent", Icon: "🤖", Builtin: true,
			Defaults: configutil.ConfigLayer{
				Env: map[string]string{},
				Files: []configutil.ConfigFile{
					{Path: "~/.claude.json", Content: "{\n  \"hasCompletedOnboarding\": true,\n  \"firstStartTime\": \"\",\n  \"opusProMigrationComplete\": true,\n  \"sonnet1m45MigrationComplete\": true,\n  \"migrationVersion\": 11,\n  \"projects\": {}\n}\n"},
					{Path: "~/.claude/settings.json", Content: "{\n  \"permissions\": {\n    \"allow\": [\n      \"Bash(*)\",\n      \"Read(*)\",\n      \"Write(*)\",\n      \"Edit(*)\",\n      \"MultiEdit(*)\",\n      \"WebFetch(*)\",\n      \"WebSearch(*)\"\n    ],\n    \"deny\": [],\n    \"skipDangerousModePermissionPrompt\": true\n  },\n  \"skipDangerousModePermissionPrompt\": true,\n  \"theme\": \"dark\"\n}\n"},
					{Path: "~/.claude/CLAUDE.md", Content: "# Global Instructions\n"},
				},
			},
			SessionSchema: configutil.SessionSchema{
				EnvDefs: []configutil.EnvDef{},
				FileDefs: []configutil.FileDef{
					{Path: "CLAUDE.md", Label: "CLAUDE.md（项目记忆）", Required: false, DefaultContent: "# Project Instructions\n"},
					{Path: ".claude/settings.json", Label: "项目级 Settings", Required: false, DefaultContent: "{}"},
				},
			},
		},
		{
			ID: "codex", Name: "Codex", Command: "codex --full-auto",
			Description: "OpenAI Codex Agent", Icon: "⚡", Builtin: true,
			Defaults: configutil.ConfigLayer{
				Env: map[string]string{},
				Files: []configutil.ConfigFile{
					{Path: "~/.codex/config.toml", Content: "# model = \"o3\"\n# approval_policy = \"on-request\"\n"},
					{Path: "~/AGENTS.md", Content: "# Global Instructions\n"},
				},
			},
			SessionSchema: configutil.SessionSchema{
				EnvDefs: []configutil.EnvDef{},
				FileDefs: []configutil.FileDef{
					{Path: "AGENTS.md", Label: "AGENTS.md（项目指令）", Required: false, DefaultContent: "# Project Instructions\n"},
					{Path: ".codex/config.toml", Label: "项目级 Config", Required: false, DefaultContent: "# model = \"o3\"\n# approval_policy = \"on-request\"\n"},
				},
			},
		},
		{
			ID: "bash", Name: "Bash Shell", Command: "",
			Description: "纯 Bash 终端", Icon: "🖥️", Builtin: true,
			Defaults: configutil.ConfigLayer{
				Env:   map[string]string{},
				Files: []configutil.ConfigFile{},
			},
			SessionSchema: configutil.SessionSchema{EnvDefs: []configutil.EnvDef{}, FileDefs: []configutil.FileDef{}},
		},
	}
}

// List 列出所有模板
func (s *TemplateStore) List() []*SessionTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*SessionTemplate, 0, len(s.templates))
	for _, t := range s.templates {
		result = append(result, t)
	}
	return result
}

// Get 获取指定 ID 的模板
func (s *TemplateStore) Get(id string) *SessionTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.templates[id]
}

// Set 保存模板
func (s *TemplateStore) Set(t *SessionTemplate) error {
	s.mu.Lock()
	s.templates[t.ID] = t
	s.mu.Unlock()
	return s.save()
}

// Delete 删除模板。内置模板静默返回 nil（不报错也不删除）
func (s *TemplateStore) Delete(id string) error {
	s.mu.Lock()
	t, exists := s.templates[id]
	if exists && t.Builtin {
		s.mu.Unlock()
		return nil
	}
	delete(s.templates, id)
	s.mu.Unlock()
	return s.save()
}
