package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	gen "github.com/charviki/mesa-hub-behavior-panel/internal/repository/postgres/sqlc"
	"github.com/charviki/mesa-hub-behavior-panel/internal/service"
)

var _ service.PermissionStore = (*PermissionRepository)(nil)
var _ service.AuthTxManager = (*PermissionRepository)(nil)

type authTxContextKey struct{}

// PermissionRepository 是 PostgreSQL/sqlc 驱动的权限仓储实现。
type PermissionRepository struct {
	pool *pgxpool.Pool
}

// NewPermissionRepository 创建 PostgreSQL 权限仓储。
func NewPermissionRepository(pool *pgxpool.Pool) *PermissionRepository {
	return &PermissionRepository{pool: pool}
}

// WithinTx 在 context 中透传事务执行器，避免 service 显式切换 tx store。
func (r *PermissionRepository) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if existing := authTxFromContext(ctx); existing != nil {
		return fn(ctx)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	txCtx := context.WithValue(ctx, authTxContextKey{}, tx)
	if err := fn(txCtx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// UpsertSubject 写入或更新权限主体。
func (r *PermissionRepository) UpsertSubject(ctx context.Context, subject service.Subject) (service.Subject, error) {
	row, err := r.queries(ctx).UpsertSubject(ctx, gen.UpsertSubjectParams{
		SubjectKey:  subject.SubjectKey,
		SubjectType: subject.SubjectType,
		DisplayName: subject.DisplayName,
		IsSystem:    subject.IsSystem,
	})
	if err != nil {
		return service.Subject{}, err
	}
	return subjectFromRow(row), nil
}

// CreatePermissionApplication 创建权限申请单。
func (r *PermissionRepository) CreatePermissionApplication(ctx context.Context, params service.CreatePermissionApplicationParams) (service.PermissionApplication, error) {
	return createPermissionApplication(ctx, r.queries(ctx), params)
}

// ListPermissionApplications 按分页条件列出权限申请单。
func (r *PermissionRepository) ListPermissionApplications(ctx context.Context, params service.ListPermissionApplicationsParams) ([]service.PermissionApplication, int64, error) {
	return listPermissionApplications(ctx, r.queries(ctx), params)
}

// GetPermissionApplication 通过对外 public_id 获取权限申请单。
func (r *PermissionRepository) GetPermissionApplication(ctx context.Context, publicID string) (service.PermissionApplication, error) {
	return getPermissionApplication(ctx, r.queries(ctx), publicID)
}

// ReviewPermissionApplication 审批权限申请单。
func (r *PermissionRepository) ReviewPermissionApplication(ctx context.Context, params service.ReviewPermissionApplicationParams) (service.PermissionApplication, error) {
	return reviewPermissionApplication(ctx, r.queries(ctx), params)
}

// UpdatePermissionApplicationStatus 同步权限申请单状态。
func (r *PermissionRepository) UpdatePermissionApplicationStatus(ctx context.Context, params service.UpdatePermissionApplicationStatusParams) (service.PermissionApplication, error) {
	return updatePermissionApplicationStatus(ctx, r.queries(ctx), params)
}

// CreatePermissionGrant 创建权限授权结果。
func (r *PermissionRepository) CreatePermissionGrant(ctx context.Context, params service.CreatePermissionGrantParams) (service.PermissionGrant, error) {
	return createPermissionGrant(ctx, r.queries(ctx), params)
}

// AttachGrantCasbinRule 绑定授权结果与 Casbin 规则。
func (r *PermissionRepository) AttachGrantCasbinRule(ctx context.Context, grantInternalID int64, casbinRuleID int64) error {
	return attachGrantCasbinRule(ctx, r.queries(ctx), grantInternalID, casbinRuleID)
}

// ListSubjectPermissionGrants 查询主体当前有效授权。
func (r *PermissionRepository) ListSubjectPermissionGrants(ctx context.Context, subjectKey string) ([]service.PermissionGrant, error) {
	rows, err := r.queries(ctx).ListActiveSubjectPermissionGrants(ctx, subjectKey)
	if err != nil {
		return nil, err
	}
	return permissionGrantsFromSubjectRows(rows), nil
}

// RevokeActivePermissionGrantsByApplication 按申请单撤销有效授权。
func (r *PermissionRepository) RevokeActivePermissionGrantsByApplication(ctx context.Context, params service.RevokePermissionGrantsParams) ([]service.PermissionGrant, error) {
	return revokeActivePermissionGrantsByApplication(ctx, r.queries(ctx), params)
}

// ListExpiredPermissionGrants 查询所有已到期但尚未清理的授权。
func (r *PermissionRepository) ListExpiredPermissionGrants(ctx context.Context) ([]service.PermissionGrant, error) {
	rows, err := r.queries(ctx).ListExpiredPermissionGrants(ctx)
	if err != nil {
		return nil, err
	}
	return permissionGrantsFromExpiredRows(rows), nil
}

// ExpirePermissionGrant 将授权结果标记为已过期。
func (r *PermissionRepository) ExpirePermissionGrant(ctx context.Context, grantInternalID int64) error {
	_, err := r.queries(ctx).ExpirePermissionGrant(ctx, grantInternalID)
	return mapNotFoundError(err, service.ErrPermissionGrantNotFound)
}

// InsertAuditLog 写入权限审计日志。
func (r *PermissionRepository) InsertAuditLog(ctx context.Context, entry service.AuditLogEntry) error {
	return insertAuditLog(ctx, r.queries(ctx), entry)
}

// InsertCasbinRule 写入 Casbin 规则。
func (r *PermissionRepository) InsertCasbinRule(ctx context.Context, rule service.CasbinRule) (int64, error) {
	return r.queries(ctx).InsertCasbinRule(ctx, gen.InsertCasbinRuleParams{
		Ptype: rule.Ptype,
		V0:    nullableString(rule.V0),
		V1:    nullableString(rule.V1),
		V2:    nullableString(rule.V2),
	})
}

// DeleteCasbinRule 删除 Casbin 规则。
func (r *PermissionRepository) DeleteCasbinRule(ctx context.Context, id int64) error {
	return r.queries(ctx).DeleteCasbinRule(ctx, id)
}

// ListAllCasbinRules 列出当前所有 Casbin 规则。
func (r *PermissionRepository) ListAllCasbinRules(ctx context.Context) ([]service.CasbinRule, error) {
	rows, err := r.queries(ctx).ListAllCasbinRules(ctx)
	if err != nil {
		return nil, err
	}

	rules := make([]service.CasbinRule, 0, len(rows))
	for _, row := range rows {
		rules = append(rules, service.CasbinRule{
			ID:    row.ID,
			Ptype: row.Ptype,
			V0:    derefString(row.V0),
			V1:    derefString(row.V1),
			V2:    derefString(row.V2),
		})
	}
	return rules, nil
}

// EnsureAdminBootstrap 确保 admin 主体与超级管理员策略存在。
func (r *PermissionRepository) EnsureAdminBootstrap(ctx context.Context, subjectKey, displayName string) error {
	// 这里同时补主体和超级管理员策略，避免权限系统启用后进入半初始化状态。
	return r.WithinTx(ctx, func(ctx context.Context) error {
		if _, err := r.UpsertSubject(ctx, service.Subject{
			SubjectKey:  subjectKey,
			SubjectType: subjectTypeFromKey(subjectKey),
			DisplayName: displayName,
			IsSystem:    true,
		}); err != nil {
			return err
		}

		rules, err := r.ListAllCasbinRules(ctx)
		if err != nil {
			return err
		}
		for _, rule := range rules {
			if rule.Ptype == "p" && rule.V0 == subjectKey && rule.V1 == "*" && rule.V2 == "*" {
				return nil
			}
		}

		_, err = r.InsertCasbinRule(ctx, service.CasbinRule{
			Ptype: "p",
			V0:    subjectKey,
			V1:    "*",
			V2:    "*",
		})
		return err
	})
}

func (r *PermissionRepository) queries(ctx context.Context) *gen.Queries {
	return gen.New(authExecutorFromContext(ctx, r.pool))
}

func authExecutorFromContext(ctx context.Context, pool *pgxpool.Pool) gen.DBTX {
	if tx := authTxFromContext(ctx); tx != nil {
		return tx
	}
	return pool
}

func authTxFromContext(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(authTxContextKey{}).(pgx.Tx)
	return tx
}

func createPermissionApplication(ctx context.Context, q *gen.Queries, params service.CreatePermissionApplicationParams) (service.PermissionApplication, error) {
	targetsJSON, err := json.Marshal(params.Targets)
	if err != nil {
		return service.PermissionApplication{}, err
	}

	row, err := q.InsertPermissionRequest(ctx, gen.InsertPermissionRequestParams{
		SubjectKey:  params.SubjectKey,
		TargetsJson: targetsJSON,
		Reason:      strings.TrimSpace(params.Reason),
		ExpiresAt:   timestamptzFromPtr(params.ExpiresAt),
	})
	if err != nil {
		return service.PermissionApplication{}, err
	}
	return permissionApplicationFromRow(row)
}

func listPermissionApplications(ctx context.Context, q *gen.Queries, params service.ListPermissionApplicationsParams) ([]service.PermissionApplication, int64, error) {
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	if params.Status == "" {
		rows, err := q.ListPermissionRequestsAll(ctx, gen.ListPermissionRequestsAllParams{
			Limit:  pageSize,
			Offset: offset,
		})
		if err != nil {
			return nil, 0, err
		}
		total, err := q.CountPermissionRequestsAll(ctx)
		if err != nil {
			return nil, 0, err
		}
		applications, err := permissionApplicationsFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		return applications, total, nil
	}

	rows, err := q.ListPermissionRequestsByStatus(ctx, gen.ListPermissionRequestsByStatusParams{
		Status: gen.AuthPermissionRequestStatus(params.Status),
		Limit:  pageSize,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := q.CountPermissionRequestsByStatus(ctx, gen.AuthPermissionRequestStatus(params.Status))
	if err != nil {
		return nil, 0, err
	}
	applications, err := permissionApplicationsFromRows(rows)
	if err != nil {
		return nil, 0, err
	}
	return applications, total, nil
}

func getPermissionApplication(ctx context.Context, q *gen.Queries, publicID string) (service.PermissionApplication, error) {
	row, err := q.GetPermissionRequestByPublicID(ctx, nullableString(publicID))
	if err != nil {
		return service.PermissionApplication{}, mapNotFoundError(err, service.ErrPermissionApplicationNotFound)
	}
	return permissionApplicationFromRow(row)
}

func reviewPermissionApplication(ctx context.Context, q *gen.Queries, params service.ReviewPermissionApplicationParams) (service.PermissionApplication, error) {
	row, err := q.ReviewPermissionRequest(ctx, gen.ReviewPermissionRequestParams{
		Status:         gen.AuthPermissionRequestStatus(params.Status),
		ReviewedBy:     nullableString(params.ReviewedBy),
		ReviewComment:  nullableString(params.ReviewComment),
		ID:             params.InternalID,
		ExpectedStatus: gen.AuthPermissionRequestStatus(params.ExpectedStatus),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.PermissionApplication{}, service.ErrPermissionApplicationStateChanged
		}
		return service.PermissionApplication{}, mapNotFoundError(err, service.ErrPermissionApplicationNotFound)
	}
	return permissionApplicationFromRow(row)
}

func updatePermissionApplicationStatus(ctx context.Context, q *gen.Queries, params service.UpdatePermissionApplicationStatusParams) (service.PermissionApplication, error) {
	row, err := q.UpdatePermissionRequestStatus(ctx, gen.UpdatePermissionRequestStatusParams{
		ID:            params.InternalID,
		Status:        gen.AuthPermissionRequestStatus(params.Status),
		ReviewedBy:    nullableString(params.ReviewedBy),
		ReviewComment: nullableString(params.ReviewComment),
	})
	if err != nil {
		return service.PermissionApplication{}, mapNotFoundError(err, service.ErrPermissionApplicationNotFound)
	}
	return permissionApplicationFromRow(row)
}

func createPermissionGrant(ctx context.Context, q *gen.Queries, params service.CreatePermissionGrantParams) (service.PermissionGrant, error) {
	row, err := q.InsertPermissionGrant(ctx, gen.InsertPermissionGrantParams{
		SubjectKey:      params.SubjectKey,
		Resource:        params.Resource,
		Action:          params.Action,
		SourceRequestID: params.SourcePermissionApplicationInternalID,
		ExpiresAt:       timestamptzFromPtr(params.ExpiresAt),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "uq_permission_grants_request_resource_action" {
			return service.PermissionGrant{}, service.ErrPermissionApplicationStateChanged
		}
		return service.PermissionGrant{}, err
	}
	return permissionGrantFromRow(row), nil
}

func attachGrantCasbinRule(ctx context.Context, q *gen.Queries, grantInternalID int64, casbinRuleID int64) error {
	_, err := q.AttachGrantCasbinRule(ctx, gen.AttachGrantCasbinRuleParams{
		ID:           grantInternalID,
		CasbinRuleID: &casbinRuleID,
	})
	return err
}

func revokeActivePermissionGrantsByApplication(ctx context.Context, q *gen.Queries, params service.RevokePermissionGrantsParams) ([]service.PermissionGrant, error) {
	rows, err := q.RevokeActivePermissionGrantsByRequest(ctx, gen.RevokeActivePermissionGrantsByRequestParams{
		SourceRequestID: params.SourcePermissionApplicationInternalID,
		Status:          gen.AuthPermissionGrantStatus(params.Status),
		RevokedBy:       nullableString(params.RevokedBy),
		RevokedReason:   nullableString(params.RevokedReason),
	})
	if err != nil {
		return nil, err
	}
	return permissionGrantsFromRows(rows), nil
}

func insertAuditLog(ctx context.Context, q *gen.Queries, entry service.AuditLogEntry) error {
	payload := entry.Details
	if payload == nil {
		payload = map[string]any{}
	}
	detailsJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return q.InsertAuditLog(ctx, gen.InsertAuditLogParams{
		Action:           entry.Action,
		ActorSubjectKey:  nullableString(entry.ActorSubjectKey),
		TargetSubjectKey: nullableString(entry.TargetSubjectKey),
		RequestID:        entry.PermissionApplicationInternalID,
		DetailsJson:      detailsJSON,
	})
}

func permissionApplicationsFromRows(rows []gen.AuthPermissionRequest) ([]service.PermissionApplication, error) {
	applications := make([]service.PermissionApplication, 0, len(rows))
	for _, row := range rows {
		application, err := permissionApplicationFromRow(row)
		if err != nil {
			return nil, err
		}
		applications = append(applications, application)
	}
	return applications, nil
}

func permissionApplicationFromRow(row gen.AuthPermissionRequest) (service.PermissionApplication, error) {
	var targets []service.PermissionTarget
	if err := json.Unmarshal(row.TargetsJson, &targets); err != nil {
		return service.PermissionApplication{}, err
	}

	return service.PermissionApplication{
		InternalID:    row.ID,
		ID:            derefString(row.PublicID),
		SubjectKey:    row.SubjectKey,
		Targets:       targets,
		Reason:        row.Reason,
		Status:        service.PermissionApplicationStatus(row.Status),
		ReviewedBy:    derefString(row.ReviewedBy),
		ReviewComment: derefString(row.ReviewComment),
		ReviewedAt:    timePtr(row.ReviewedAt),
		ExpiresAt:     timePtr(row.ExpiresAt),
		CreatedAt:     row.CreatedAt.Time.UTC(),
		UpdatedAt:     row.UpdatedAt.Time.UTC(),
	}, nil
}

func permissionGrantsFromRows(rows []gen.AuthPermissionGrant) []service.PermissionGrant {
	grants := make([]service.PermissionGrant, 0, len(rows))
	for _, row := range rows {
		grants = append(grants, permissionGrantFromRow(row))
	}
	return grants
}

func permissionGrantFromRow(row gen.AuthPermissionGrant) service.PermissionGrant {
	return service.PermissionGrant{
		InternalID:                            row.ID,
		ID:                                    derefString(row.PublicID),
		SubjectKey:                            row.SubjectKey,
		Resource:                              row.Resource,
		Action:                                row.Action,
		SourcePermissionApplicationInternalID: row.SourceRequestID,
		CasbinRuleID:                          row.CasbinRuleID,
		Status:                                service.PermissionGrantStatus(row.Status),
		RevokedBy:                             derefString(row.RevokedBy),
		RevokedReason:                         derefString(row.RevokedReason),
		ExpiresAt:                             timePtr(row.ExpiresAt),
		CreatedAt:                             row.CreatedAt.Time.UTC(),
		UpdatedAt:                             row.UpdatedAt.Time.UTC(),
	}
}

func permissionGrantsFromSubjectRows(rows []gen.ListActiveSubjectPermissionGrantsRow) []service.PermissionGrant {
	grants := make([]service.PermissionGrant, 0, len(rows))
	for _, row := range rows {
		grant := permissionGrantFromRow(row.AuthPermissionGrant)
		grant.SourcePermissionApplicationID = derefString(row.SourceRequestPublicID)
		grants = append(grants, grant)
	}
	return grants
}

func permissionGrantsFromExpiredRows(rows []gen.ListExpiredPermissionGrantsRow) []service.PermissionGrant {
	grants := make([]service.PermissionGrant, 0, len(rows))
	for _, row := range rows {
		grant := permissionGrantFromRow(row.AuthPermissionGrant)
		grant.SourcePermissionApplicationID = derefString(row.SourceRequestPublicID)
		grants = append(grants, grant)
	}
	return grants
}

func subjectFromRow(row gen.AuthSubject) service.Subject {
	return service.Subject{
		SubjectKey:  row.SubjectKey,
		SubjectType: row.SubjectType,
		DisplayName: row.DisplayName,
		IsSystem:    row.IsSystem,
		CreatedAt:   row.CreatedAt.Time.UTC(),
		UpdatedAt:   row.UpdatedAt.Time.UTC(),
	}
}

func timestamptzFromPtr(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func timePtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time.UTC()
	return &t
}

func nullableString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func subjectTypeFromKey(subjectKey string) string {
	parts := strings.SplitN(subjectKey, ":", 2)
	if len(parts) == 0 || parts[0] == "" {
		return "subject"
	}
	return parts[0]
}

func mapNotFoundError(err error, target error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return target
	}
	return err
}
