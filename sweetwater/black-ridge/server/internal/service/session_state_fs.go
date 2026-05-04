package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charviki/maze-cradle/configutil"
)

// SessionStateRepository 隔离 session 状态的文件系统持久化
type SessionStateRepository interface {
	Save(state *SessionState) error
	Load(sessionName string) (*SessionState, error)
	List() ([]SessionState, error)
	Delete(sessionName string) error
}

type fileSessionStateRepository struct {
	stateDir string
}

func newFileSessionStateRepository(stateDir string) *fileSessionStateRepository {
	return &fileSessionStateRepository{stateDir: stateDir}
}

// Save 将会话状态序列化后原子写入文件
func (r *fileSessionStateRepository) Save(state *SessionState) error {
	if err := os.MkdirAll(r.stateDir, 0750); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	data, err := state.ToJSON()
	if err != nil {
		return fmt.Errorf("serialize state: %w", err)
	}
	// 使用原子写入防止写入过程中崩溃导致状态文件损坏
	filePath := filepath.Join(r.stateDir, state.SessionName+".json")
	if err := configutil.AtomicWriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}
	return nil
}

// Load 从文件读取并反序列化指定会话的状态
func (r *fileSessionStateRepository) Load(sessionName string) (*SessionState, error) {
	filePath := filepath.Join(r.stateDir, sessionName+".json")
	data, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, err
	}
	var state SessionState
	if err := state.FromJSON(data); err != nil {
		return nil, fmt.Errorf("parse state file: %w", err)
	}
	return &state, nil
}

// List 扫描状态目录，返回所有已保存的会话状态
func (r *fileSessionStateRepository) List() ([]SessionState, error) {
	entries, err := os.ReadDir(r.stateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SessionState{}, nil
		}
		return nil, fmt.Errorf("read state dir: %w", err)
	}
	var states []SessionState
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		// 跳过非 session state 文件（如 templates.json）
		if entry.Name() == "templates.json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(r.stateDir, entry.Name()))
		if err != nil {
			continue
		}
		var state SessionState
		if err := state.FromJSON(data); err != nil {
			continue
		}
		if state.SessionName == "" {
			continue
		}
		states = append(states, state)
	}
	return states, nil
}

// Delete 移除指定会话的状态文件，文件不存在时不报错
func (r *fileSessionStateRepository) Delete(sessionName string) error {
	filePath := filepath.Join(r.stateDir, sessionName+".json")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete state file: %w", err)
	}
	return nil
}
