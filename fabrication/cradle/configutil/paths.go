package configutil

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandHomePath 将 ~/ 开头的路径展开为用户 home 目录。
// Docker volume、持久化目录等场景依赖绝对路径，因此需要在配置层统一展开。
func ExpandHomePath(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}
