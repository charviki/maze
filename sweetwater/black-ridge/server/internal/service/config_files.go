package service

import (
	//nolint:gosec // md5 used for file content fingerprinting, not security
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charviki/sweetwater-black-ridge/internal/model"
)

const configConflictMessage = "配置已变更，请重新加载后再修改"

type configTarget struct {
	displayPath string
	diskPath    string
}

// ConfigConflictError 表示保存时命中了乐观并发冲突。
// 冲突按单文件粒度返回，前端可直接提示用户重新加载真实文件后再编辑。
type ConfigConflictError struct {
	Conflicts []model.ConfigConflict
}

func (e *ConfigConflictError) Error() string {
	return configConflictMessage
}

// ConfigFileService 负责固定路径配置文件的真实读写与 hash 冲突检测。
type ConfigFileService struct{}

// NewConfigFileService 创建 ConfigFileService
func NewConfigFileService() *ConfigFileService {
	return &ConfigFileService{}
}

// ReadGlobalFiles 读取模板声明的全局固定文件集合。
func (s *ConfigFileService) ReadGlobalFiles(defs []model.ConfigFile) ([]model.ConfigFileSnapshot, error) {
	targets, _, err := buildGlobalTargets(defs)
	if err != nil {
		return nil, err
	}
	return s.readTargets(targets)
}

// SaveGlobalFiles 直接写回真实全局文件，并在写入前校验 base_hash。
func (s *ConfigFileService) SaveGlobalFiles(defs []model.ConfigFile, updates []model.ConfigFileUpdate) ([]model.ConfigFileSnapshot, error) {
	targets, targetMap, err := buildGlobalTargets(defs)
	if err != nil {
		return nil, err
	}
	if err := s.validateAndWriteTargets(targets, targetMap, updates); err != nil {
		return nil, err
	}
	return s.readTargets(targets)
}

// ReadProjectFiles 读取 session 工作目录下的固定项目级文件集合。
func (s *ConfigFileService) ReadProjectFiles(workingDir string, defs []model.FileDef) ([]model.ConfigFileSnapshot, error) {
	targets, _, err := buildProjectTargets(workingDir, defs)
	if err != nil {
		return nil, err
	}
	return s.readTargets(targets)
}

// SaveProjectFiles 直接写回 session 工作目录中的真实项目级文件。
func (s *ConfigFileService) SaveProjectFiles(workingDir string, defs []model.FileDef, updates []model.ConfigFileUpdate) ([]model.ConfigFileSnapshot, error) {
	targets, targetMap, err := buildProjectTargets(workingDir, defs)
	if err != nil {
		return nil, err
	}
	if err := s.validateAndWriteTargets(targets, targetMap, updates); err != nil {
		return nil, err
	}
	return s.readTargets(targets)
}

func (s *ConfigFileService) readTargets(targets []configTarget) ([]model.ConfigFileSnapshot, error) {
	files := make([]model.ConfigFileSnapshot, 0, len(targets))
	for _, target := range targets {
		snapshot, err := readConfigSnapshot(target.displayPath, target.diskPath)
		if err != nil {
			return nil, err
		}
		files = append(files, snapshot)
	}
	return files, nil
}

func (s *ConfigFileService) validateAndWriteTargets(_ []configTarget, targetMap map[string]configTarget, updates []model.ConfigFileUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(updates))
	conflicts := make([]model.ConfigConflict, 0)

	// 先统一做 hash 校验，再写入文件，避免半成功半失败把固定文件集合写成不一致状态。
	for _, update := range updates {
		target, ok := targetMap[update.Path]
		if !ok {
			return fmt.Errorf("path %s is not editable", update.Path)
		}
		if _, duplicated := seen[update.Path]; duplicated {
			return fmt.Errorf("duplicate file update for path %s", update.Path)
		}
		seen[update.Path] = struct{}{}

		current, err := readConfigSnapshot(target.displayPath, target.diskPath)
		if err != nil {
			return err
		}
		if current.Hash != update.BaseHash {
			conflicts = append(conflicts, model.ConfigConflict{
				Path:        update.Path,
				CurrentHash: current.Hash,
			})
		}
	}

	if len(conflicts) > 0 {
		return &ConfigConflictError{Conflicts: conflicts}
	}

	for _, update := range updates {
		target := targetMap[update.Path]
		if err := writeConfigFile(target.diskPath, update.Content); err != nil {
			return err
		}
	}

	return nil
}

func buildGlobalTargets(defs []model.ConfigFile) ([]configTarget, map[string]configTarget, error) {
	targets := make([]configTarget, 0, len(defs))
	targetMap := make(map[string]configTarget, len(defs))

	for _, def := range defs {
		displayPath := strings.TrimSpace(def.Path)
		if displayPath == "" {
			return nil, nil, errors.New("template global file path is required")
		}

		diskPath, err := expandHomePath(displayPath)
		if err != nil {
			return nil, nil, fmt.Errorf("expand global path %s: %w", displayPath, err)
		}
		if !filepath.IsAbs(diskPath) {
			return nil, nil, fmt.Errorf("global path %s must resolve to an absolute path", displayPath)
		}

		target := configTarget{displayPath: displayPath, diskPath: filepath.Clean(diskPath)}
		if _, duplicated := targetMap[target.displayPath]; duplicated {
			return nil, nil, fmt.Errorf("duplicate global path %s", target.displayPath)
		}
		targets = append(targets, target)
		targetMap[target.displayPath] = target
	}

	return targets, targetMap, nil
}

func buildProjectTargets(workingDir string, defs []model.FileDef) ([]configTarget, map[string]configTarget, error) {
	if strings.TrimSpace(workingDir) == "" {
		return nil, nil, errors.New("working_dir is required")
	}

	root := filepath.Clean(workingDir)
	targets := make([]configTarget, 0, len(defs))
	targetMap := make(map[string]configTarget, len(defs))

	for _, def := range defs {
		displayPath := strings.TrimSpace(def.Path)
		if displayPath == "" {
			return nil, nil, errors.New("template project file path is required")
		}
		if filepath.IsAbs(displayPath) {
			return nil, nil, fmt.Errorf("project path %s must be relative", displayPath)
		}

		cleaned := filepath.Clean(displayPath)
		if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
			return nil, nil, fmt.Errorf("project path %s escapes working directory", displayPath)
		}

		target := configTarget{
			displayPath: cleaned,
			diskPath:    filepath.Join(root, cleaned),
		}
		if _, duplicated := targetMap[target.displayPath]; duplicated {
			return nil, nil, fmt.Errorf("duplicate project path %s", target.displayPath)
		}
		targets = append(targets, target)
		targetMap[target.displayPath] = target
	}

	return targets, targetMap, nil
}

func readConfigSnapshot(displayPath string, diskPath string) (model.ConfigFileSnapshot, error) {
	data, err := os.ReadFile(filepath.Clean(diskPath))
	if err != nil {
		if os.IsNotExist(err) {
			return model.ConfigFileSnapshot{
				Path:    displayPath,
				Content: "",
				Exists:  false,
				Hash:    contentHash(""),
			}, nil
		}
		return model.ConfigFileSnapshot{}, fmt.Errorf("read config file %s: %w", displayPath, err)
	}

	content := string(data)
	return model.ConfigFileSnapshot{
		Path:    displayPath,
		Content: content,
		Exists:  true,
		Hash:    contentHash(content),
	}, nil
}

func writeConfigFile(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return fmt.Errorf("create config parent dir for %s: %w", path, err)
	}
	if err := model.AtomicWriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write config file %s: %w", path, err)
	}
	return nil
}

func contentHash(content string) string {
	if content == "" {
		return "md5:empty"
	}
	//nolint:gosec
	sum := md5.Sum([]byte(content))
	return "md5:" + hex.EncodeToString(sum[:])
}

func expandHomePath(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") {
		return filepath.Clean(path), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, path[2:]), nil
}
