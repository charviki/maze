package service

import "errors"

var (
	// ErrArchiveNotFound 表示知识库不存在。
	ErrArchiveNotFound = errors.New("archive not found")
	// ErrDocNotFound 表示文档不存在。
	ErrDocNotFound = errors.New("doc not found")
	// ErrLinkNotFound 表示关联不存在。
	ErrLinkNotFound = errors.New("link not found")
	// ErrAlreadyExists 表示资源已存在。
	ErrAlreadyExists = errors.New("resource already exists")
)

// ValidationError 表示请求参数校验失败。
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Message }
