package repository

import (
	"context"
	"time"
)

// User 表示 auth.users 表中的一条用户记录。
type User struct {
	ID           int64
	Username     string
	PasswordHash string
	SubjectKey   string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserStore 定义用户持久化的最小能力集。
type UserStore interface {
	CreateUser(ctx context.Context, username, passwordHash, subjectKey string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserBySubjectKey(ctx context.Context, subjectKey string) (*User, error)
}
