package configutil

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SearchConfigPaths 搜索配置文件，按优先级返回第一个找到的路径。
// 搜索顺序：当前工作目录 → 可执行文件所在目录 → 可执行文件上级目录。
func SearchConfigPaths(filename string) (string, error) {
	searchPaths := []string{filename}

	if wd, err := os.Getwd(); err == nil {
		searchPaths = append(searchPaths, filepath.Join(wd, filename))
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		searchPaths = append(searchPaths, filepath.Join(exeDir, filename), filepath.Join(exeDir, "..", filename))
	}

	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("config file %s not found, searched: %v", filename, searchPaths)
}

// LoadYAML 从指定路径加载 YAML 配置文件并反序列化到目标结构体
func LoadYAML(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}
	return nil
}

// LoadFromExe 搜索并加载配置文件，填充到 target 结构体。
// 使用场景：从可执行文件所在目录及当前工作目录搜索配置文件。
func LoadFromExe(target interface{}, filename ...string) (string, error) {
	name := "config.yaml"
	if len(filename) > 0 {
		name = filename[0]
	}

	path, err := SearchConfigPaths(name)
	if err != nil {
		return "", err
	}

	return path, LoadYAML(path, target)
}
