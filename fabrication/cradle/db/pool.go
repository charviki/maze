package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	parsePoolConfig = pgxpool.ParseConfig
	newPoolWithConfig = pgxpool.NewWithConfig
	pingPool = func(ctx context.Context, pool *pgxpool.Pool) error {
		return pool.Ping(ctx)
	}
	closePool = func(pool *pgxpool.Pool) {
		pool.Close()
	}
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
	poolCfg, err := parsePoolConfig(cfg.ConnString())
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	pool, err := newPoolWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create db pool: %w", err)
	}

	if err := pingPool(ctx, pool); err != nil {
		closePool(pool)
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return pool, nil
}
