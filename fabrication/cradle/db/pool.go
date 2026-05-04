package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig 包含 PostgreSQL 连接池配置。
type PoolConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

// ConnString 构建 PostgreSQL 连接字符串。
func (c PoolConfig) ConnString() string {
	port := c.Port
	if port == 0 {
		port = 5432
	}
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		c.User, c.Password, c.Host, port, c.Name)
}

// NewPool 创建 PostgreSQL 连接池。
func NewPool(ctx context.Context, cfg PoolConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.ConnString())
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create db pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return pool, nil
}
