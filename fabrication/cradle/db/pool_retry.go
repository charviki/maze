package db

import (
	"context"
	"fmt"
	"time"

	"github.com/charviki/maze/fabrication/cradle/logutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPoolWithRetry 尝试创建数据库连接池，最多重试 maxAttempts 次（每次间隔 1 秒）。
func NewPoolWithRetry(ctx context.Context, cfg PoolConfig, maxAttempts int, logger logutil.Logger) (*pgxpool.Pool, error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pool, err := NewPool(ctx, cfg)
		if err == nil {
			if attempt > 1 {
				logger.Infof("database became ready after %d attempts", attempt)
			}
			return pool, nil
		}
		lastErr = err
		if attempt == maxAttempts {
			break
		}

		logger.Warnf("database not ready (attempt %d/%d): %v", attempt, maxAttempts, err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return nil, fmt.Errorf("database not ready after %d attempts: %w", maxAttempts, lastErr)
}
