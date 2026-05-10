package service

import "context"

// FileRepository 定义文件存储的边界。
type FileRepository interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) error
	Download(ctx context.Context, key string) ([]byte, string, error)
}
