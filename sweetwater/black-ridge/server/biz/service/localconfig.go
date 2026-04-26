package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/charviki/maze-cradle/configutil"
	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/maze-cradle/protocol"
)

// LocalConfigStore Agent 本地记忆配置管理器。
// 维护 Agent 自身的工作目录和环境变量配置，替代原来由 Manager 端管理的 NodeConfigStore。
// 配置持久化到 ~/.maze/config.json，Agent 重启后自动恢复。
type LocalConfigStore struct {
	mu            sync.RWMutex
	path          string
	workspaceRoot string
	config        protocol.LocalAgentConfig
	logger        logutil.Logger
}

// NewLocalConfigStore 创建本地配置管理器并从文件加载已有配置。
// 文件不存在时使用默认配置初始化。
func NewLocalConfigStore(workspaceRoot string, logger logutil.Logger) *LocalConfigStore {
	configDir := filepath.Join(workspaceRoot, ".maze")
	path := filepath.Join(configDir, "config.json")

	store := &LocalConfigStore{
		path:          path,
		workspaceRoot: workspaceRoot,
		logger:        logger,
		config: protocol.LocalAgentConfig{
			WorkingDir: workspaceRoot,
			Env:        make(map[string]string),
		},
	}

	store.load()
	return store
}

// load 从 JSON 文件加载配置。文件不存在视为首次启动，使用默认配置。
func (s *LocalConfigStore) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		s.logger.Infof("[local-config] file not found, using defaults: %s", s.path)
		return
	}
	var cfg protocol.LocalAgentConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		s.logger.Errorf("[local-config] parse file %s failed: %v", s.path, err)
		return
	}
	if cfg.Env != nil {
		s.config.Env = cfg.Env
	}
	// 基础工作目录由服务端配置决定，只读展示，不接受本地文件覆盖。
	s.config.WorkingDir = s.workspaceRoot
}

// save 持久化当前配置到 JSON 文件。
// 确保父目录存在后原子写入。
func (s *LocalConfigStore) save() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}

	return configutil.AtomicWriteFile(s.path, data, 0644)
}

// Get 返回当前配置的只读副本
func (s *LocalConfigStore) Get() protocol.LocalAgentConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 返回副本，避免外部修改
	env := make(map[string]string, len(s.config.Env))
	for k, v := range s.config.Env {
		env[k] = v
	}
	return protocol.LocalAgentConfig{
		WorkingDir: s.workspaceRoot,
		Env:        env,
	}
}

// UpdateEnv 更新环境变量（合并，不删除已有的 key 除非值为空）
func (s *LocalConfigStore) UpdateEnv(env map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range env {
		if v == "" {
			delete(s.config.Env, k)
		} else {
			s.config.Env[k] = v
		}
	}
	return s.save()
}

// SetEnv 设置单个环境变量并持久化
func (s *LocalConfigStore) SetEnv(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config.Env[key] = value
	return s.save()
}
