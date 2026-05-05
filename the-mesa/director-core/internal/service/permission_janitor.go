package service

import (
	"context"
	"log"
	"time"
)

const permissionJanitorInterval = 5 * time.Second

type permissionGrantExpirer interface {
	ExpirePermissionGrants(context.Context) (int, error)
}

// PermissionJanitor 定时扫描过期授权并清理。
type PermissionJanitor struct {
	service permissionGrantExpirer
}

// NewPermissionJanitor 创建 PermissionJanitor。
func NewPermissionJanitor(service permissionGrantExpirer) *PermissionJanitor {
	return &PermissionJanitor{service: service}
}

// Run 启动定时清理任务。
// 过期 grant 不应长时间继续留在 Casbin 内存策略里，因此这里缩短扫描周期，收敛“到期但尚未回收”的窗口。
func (j *PermissionJanitor) Run(ctx context.Context) {
	ticker := time.NewTicker(permissionJanitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := j.cleanup(ctx); err != nil {
				log.Printf("[janitor] cleanup error: %v", err)
			}
		}
	}
}

func (j *PermissionJanitor) cleanup(ctx context.Context) error {
	expiredCount, err := j.service.ExpirePermissionGrants(ctx)
	if err != nil {
		return err
	}
	if expiredCount > 0 {
		log.Printf("[janitor] expired %d permission grants", expiredCount)
	}
	return nil
}
