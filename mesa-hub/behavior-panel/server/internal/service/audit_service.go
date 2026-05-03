package service

import (
	"context"

	"github.com/charviki/maze-cradle/protocol"
)

// AuditService 审计日志业务逻辑（Manager 本地），供 HTTP handler 和 gRPC handler 共用
type AuditService struct {
	auditLog AuditLogPaginator
}

// AuditLogPaginator 审计日志分页接口
type AuditLogPaginator interface {
	Log(entry protocol.AuditLogEntry)
	List() []protocol.AuditLogEntry
	ListPage(page, pageSize int) (logs []protocol.AuditLogEntry, total int)
}

// NewAuditService 创建 AuditService
func NewAuditService(auditLog AuditLogPaginator) *AuditService {
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
		logs := s.auditLog.List()
		return &AuditLogsResult{
			Logs:     logs,
			Total:    len(logs),
			Page:     1,
			PageSize: len(logs),
		}, nil
	}

	logs, total := s.auditLog.ListPage(page, pageSize)
	return &AuditLogsResult{
		Logs:     logs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
