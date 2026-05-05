package service

import "context"

// HostTxManager 定义 Host/Node 领域的事务边界。
//
// Host 创建和删除都跨越多张表，必须由 service 显式声明"这些写操作要么一起成功，要么一起回滚"，
// 否则 PG 迁移后仍会保留文件存储时代的半提交问题。
type HostTxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
