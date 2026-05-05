package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"

	"github.com/charviki/maze-cradle/protocol"
	hostgen "github.com/charviki/maze/the-mesa/director-core/internal/repository/postgres/sqlc/host"
)

// HostSpecRepository 是 PostgreSQL/sqlc 驱动的 HostSpec 仓储实现。
type HostSpecRepository struct {
	db hostgen.DBTX
}

// NewHostSpecRepository 创建 PG HostSpec 仓储。
func NewHostSpecRepository(db hostgen.DBTX) *HostSpecRepository {
	return &HostSpecRepository{db: db}
}

func (r *HostSpecRepository) queries(ctx context.Context) *hostgen.Queries {
	return hostgen.New(hostExecutorFromContext(ctx, r.db))
}

// Create 插入一条 HostSpec，ON CONFLICT DO NOTHING；返回是否真正插入。
func (r *HostSpecRepository) Create(ctx context.Context, spec *protocol.HostSpec) (bool, error) {
	tools, err := json.Marshal(spec.Tools)
	if err != nil {
		return false, err
	}
	resources, err := json.Marshal(spec.Resources)
	if err != nil {
		return false, err
	}
	rowsAffected, err := r.queries(ctx).InsertHostSpec(ctx, hostgen.InsertHostSpecParams{
		Name:        spec.Name,
		DisplayName: spec.DisplayName,
		Tools:       tools,
		Resources:   resources,
		AuthToken:   spec.AuthToken,
		Status:      spec.Status,
	})
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

// Get 按名称查询 HostSpec，不存在时返回 nil, nil。
func (r *HostSpecRepository) Get(ctx context.Context, name string) (*protocol.HostSpec, error) {
	row, err := r.queries(ctx).GetHostSpecByName(ctx, name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	spec := hostSpecFromRow(row)
	return &spec, nil
}

// List 返回所有 HostSpec，按名称排序。
func (r *HostSpecRepository) List(ctx context.Context) ([]*protocol.HostSpec, error) {
	rows, err := r.queries(ctx).ListHostSpecs(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*protocol.HostSpec, 0, len(rows))
	for _, row := range rows {
		spec := hostSpecFromRow(row)
		result = append(result, &spec)
	}
	return result, nil
}

// UpdateStatus 更新 HostSpec 的状态和错误信息。
func (r *HostSpecRepository) UpdateStatus(ctx context.Context, name, status, errMsg string) (bool, error) {
	_, err := r.queries(ctx).UpdateHostSpecStatus(ctx, hostgen.UpdateHostSpecStatusParams{
		Name:     name,
		Status:   status,
		ErrorMsg: errMsg,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Delete 按名称删除 HostSpec，返回是否真正删除。
func (r *HostSpecRepository) Delete(ctx context.Context, name string) (bool, error) {
	tag, err := hostExecutorFromContext(ctx, r.db).Exec(ctx, "DELETE FROM host_specs WHERE name = $1", name)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// IncrementRetry 递增 HostSpec 的 retry_count。
func (r *HostSpecRepository) IncrementRetry(ctx context.Context, name string) (bool, error) {
	_, err := r.queries(ctx).IncrementHostSpecRetry(ctx, name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func hostSpecFromRow(row hostgen.HostSpec) protocol.HostSpec {
	var tools []string
	if len(row.Tools) > 0 {
		_ = json.Unmarshal(row.Tools, &tools)
	}
	var resources protocol.ResourceLimits
	if len(row.Resources) > 0 {
		_ = json.Unmarshal(row.Resources, &resources)
	}
	return protocol.HostSpec{
		Name:        row.Name,
		DisplayName: row.DisplayName,
		Tools:       tools,
		Resources:   resources,
		AuthToken:   row.AuthToken,
		Status:      row.Status,
		ErrorMsg:    row.ErrorMsg,
		RetryCount:  int(row.RetryCount),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}
