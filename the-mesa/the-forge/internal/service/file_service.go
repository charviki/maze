package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// FileService 管理文件上传下载。
type FileService struct {
	repo FileRepository
}

// NewFileService 创建 FileService。
func NewFileService(repo FileRepository) *FileService {
	return &FileService{repo: repo}
}

// Upload 上传文件，返回生成的 key。
func (s *FileService) Upload(ctx context.Context, name string, data []byte, contentType string) (string, error) {
	key := uuid.New().String()
	if err := s.repo.Upload(ctx, key, data, contentType); err != nil {
		return "", err
	}
	return key, nil
}

// Download 下载文件，返回数据和 content type。
func (s *FileService) Download(ctx context.Context, key string) ([]byte, string, error) {
	if _, err := uuid.Parse(key); err != nil {
		return nil, "", fmt.Errorf("invalid file key: %w", err)
	}
	return s.repo.Download(ctx, key)
}
