package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charviki/maze/the-mesa/the-forge/internal/service"
)

// FileRepository 实现 service.FileRepository，使用本地文件系统。
type FileRepository struct {
	baseDir string
}

// NewFileRepository 创建 FileRepository。
func NewFileRepository(baseDir string) *FileRepository {
	return &FileRepository{baseDir: baseDir}
}

// Upload writes file data and its content type metadata to the local filesystem.
func (r *FileRepository) Upload(ctx context.Context, key string, data []byte, contentType string) error {
	dir := filepath.Join(r.baseDir, key[:2])
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create file dir: %w", err)
	}
	// 写入文件内容
	if err := os.WriteFile(filepath.Join(dir, key), data, 0o600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	// 写入 content type 元数据
	if err := os.WriteFile(filepath.Join(dir, key+".meta"), []byte(contentType), 0o600); err != nil {
		return fmt.Errorf("write meta: %w", err)
	}
	return nil
}

// Download reads file data and its content type from the local filesystem.
func (r *FileRepository) Download(ctx context.Context, key string) ([]byte, string, error) {
	dir := filepath.Join(r.baseDir, key[:2])
	data, err := os.ReadFile(filepath.Join(dir, key)) //nolint:gosec
	if err != nil {
		return nil, "", fmt.Errorf("read file: %w", err)
	}
	contentType := "application/octet-stream"
	if meta, readErr := os.ReadFile(filepath.Join(dir, key+".meta")); readErr == nil { //nolint:gosec
		contentType = string(meta)
	}
	return data, contentType, nil
}

var _ service.FileRepository = (*FileRepository)(nil)
