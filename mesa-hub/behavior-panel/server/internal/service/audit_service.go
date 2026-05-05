package service

import (
	"context"
	"fmt"

	"github.com/charviki/maze-cradle/protocol"
)

// AuditService 审计日志业务逻辑（Manager 本地），供 HTTP handler 和 gRPC handler 共用
type AuditService struct {
	auditLog AuditLogRepository
}

// AuditLogWriter 定义审计日志写入边界。
type AuditLogWriter interface {
	Log(ctx context.Context, entry protocol.AuditLogEntry) error
}

// AuditLogRepository 定义审计日志查询与写入边界。
type AuditLogRepository interface {
	AuditLogWriter
	List(ctx context.Context) ([]protocol.AuditLogEntry, error)
	ListPage(ctx context.Context, page, pageSize int) (logs []protocol.AuditLogEntry, total int, err error)
}

// NewAuditService 创建 AuditService
func NewAuditService(auditLog AuditLogRepository) *AuditService {
	return &AuditService{
		auditLog: auditLog,
	}
}

// AuditLogsResult 审计日志查询结果
type AuditLogsResult struct {
	Logs     []protocol.AuditLogEntry
	Total    int
	Page     int
	PageSize int
}

// GetAuditLogs 获取审计日志（支持分页）
// 当 page<=0 时返回全部日志
func (s *AuditService) GetAuditLogs(ctx context.Context, page, pageSize int) (*AuditLogsResult, error) {
	if page <= 0 {
		logs, err := s.auditLog.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("list audit logs: %w", err)
		}
		return &AuditLogsResult{
			Logs:     logs,
			Total:    len(logs),
			Page:     1,
			PageSize: len(logs),
		}, nil
	}

	logs, total, err := s.auditLog.ListPage(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("list audit logs page: %w", err)
	}
	return &AuditLogsResult{
		Logs:     logs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
