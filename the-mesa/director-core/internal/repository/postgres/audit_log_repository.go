package postgres

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/charviki/maze-cradle/protocol"
	hostgen "github.com/charviki/maze/the-mesa/director-core/internal/repository/postgres/sqlc/host"
)

// AuditLogRepository 是 PostgreSQL/sqlc 驱动的审计日志仓储实现。
type AuditLogRepository struct {
	db hostgen.DBTX
}

// NewAuditLogRepository 创建 PG 审计日志仓储。
func NewAuditLogRepository(db hostgen.DBTX) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) queries(ctx context.Context) *hostgen.Queries {
	return hostgen.New(hostExecutorFromContext(ctx, r.db))
}

// Log 写入一条审计日志条目。
func (r *AuditLogRepository) Log(ctx context.Context, entry protocol.AuditLogEntry) error {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	if entry.ID == "" {
		entry.ID = generateAuditIDPG()
	}

	var ts pgtype.Timestamptz
	if err := ts.Scan(entry.Timestamp); err != nil {
		ts = pgtype.Timestamptz{Time: entry.Timestamp, Valid: true}
	}

	return r.queries(ctx).InsertAuditEntry(ctx, hostgen.InsertAuditEntryParams{
		AuditID:        entry.ID,
		Timestamp:      ts,
		Operator:       entry.Operator,
		TargetNode:     entry.TargetNode,
		Action:         entry.Action,
		PayloadSummary: entry.PayloadSummary,
		Result:         entry.Result,
		StatusCode:     int32(entry.StatusCode), //nolint:gosec // HTTP 状态码范围 100-599，不会溢出 int32
	})
}

// List 返回全部审计日志，按时间倒序。
func (r *AuditLogRepository) List(ctx context.Context) ([]protocol.AuditLogEntry, error) {
	rows, err := r.queries(ctx).ListAuditEntries(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]protocol.AuditLogEntry, 0, len(rows))
	for _, row := range rows {
		result = append(result, auditEntryFromRow(row))
	}
	return result, nil
}

// ListPage 分页查询审计日志，返回条目和总数。
func (r *AuditLogRepository) ListPage(ctx context.Context, page, pageSize int) ([]protocol.AuditLogEntry, int, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	total, err := r.queries(ctx).CountAuditEntries(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries(ctx).ListAuditEntriesPage(ctx, hostgen.ListAuditEntriesPageParams{
		Limit:  int32(pageSize), //nolint:gosec // 分页大小由调用方控制，合理范围内不会溢出
		Offset: int32(offset),   //nolint:gosec // 偏移量由分页参数推导，合理范围内不会溢出
	})
	if err != nil {
		return nil, 0, err
	}
	result := make([]protocol.AuditLogEntry, 0, len(rows))
	for _, row := range rows {
		result = append(result, auditEntryFromRow(row))
	}
	return result, int(total), nil
}

// Query 按节点和操作模糊查询审计日志。
func (r *AuditLogRepository) Query(ctx context.Context, node, action string) ([]protocol.AuditLogEntry, error) {
	rows, err := r.queries(ctx).QueryAuditEntries(ctx, hostgen.QueryAuditEntriesParams{
		Column1: node,
		Column2: action,
	})
	if err != nil {
		return nil, err
	}
	result := make([]protocol.AuditLogEntry, 0, len(rows))
	for _, row := range rows {
		result = append(result, auditEntryFromRow(row))
	}
	return result, nil
}

// Close 释放资源（PG 实现为 no-op，连接由 pool 管理）。
func (r *AuditLogRepository) Close() {}

func auditEntryFromRow(row hostgen.AuditEntry) protocol.AuditLogEntry {
	return protocol.AuditLogEntry{
		ID:             row.AuditID,
		Timestamp:      row.Timestamp.Time,
		Operator:       row.Operator,
		TargetNode:     row.TargetNode,
		Action:         row.Action,
		PayloadSummary: row.PayloadSummary,
		Result:         row.Result,
		StatusCode:     int(row.StatusCode),
	}
}

func generateAuditIDPG() string {
	b := make([]byte, 16)
	n, err := rand.Read(b)
	if err != nil || n != len(b) {
		binary.BigEndian.PutUint64(b, uint64(time.Now().UnixNano()))
	}
	return "audit-" + hex.EncodeToString(b)
}
