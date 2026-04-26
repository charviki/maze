package configutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// AtomicWriteFile 原子写入文件：先写入同目录临时文件，再 rename 覆盖目标文件。
// 防止写入过程中进程崩溃导致文件损坏
func AtomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write temp file %s: %w", tmpPath, err)
	}
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("sync temp file %s: %w", tmpPath, err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpPath, perm); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("chmod temp file %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename %s to %s: %w", tmpPath, path, err)
	}
	return nil
}
