package service

import (
	"context"
)

// AuthTxManager 定义权限业务所需的事务边界能力。
//
// 事务由 service 发起，具体存储实现通过 context 识别当前执行器，
// 这样可以避免把 tx repository 显式传回业务层。
type AuthTxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// PermissionStore 定义权限闭环所需的最小持久化能力。
//
// 接口 owner 放在 service 边界，避免 PostgreSQL 实现细节反向塑造业务接口。
type PermissionStore interface {
	UpsertSubject(ctx context.Context, subject Subject) (Subject, error)
	CreatePermissionApplication(ctx context.Context, params CreatePermissionApplicationParams) (PermissionApplication, error)
	ListPermissionApplications(ctx context.Context, params ListPermissionApplicationsParams) ([]PermissionApplication, int64, error)
	GetPermissionApplication(ctx context.Context, publicID string) (PermissionApplication, error)
	ReviewPermissionApplication(ctx context.Context, params ReviewPermissionApplicationParams) (PermissionApplication, error)
	UpdatePermissionApplicationStatus(ctx context.Context, params UpdatePermissionApplicationStatusParams) (PermissionApplication, error)
	CreatePermissionGrant(ctx context.Context, params CreatePermissionGrantParams) (PermissionGrant, error)
	AttachGrantCasbinRule(ctx context.Context, grantInternalID int64, casbinRuleID int64) error
	ListSubjectPermissionGrants(ctx context.Context, subjectKey string) ([]PermissionGrant, error)
	RevokeActivePermissionGrantsByApplication(ctx context.Context, params RevokePermissionGrantsParams) ([]PermissionGrant, error)
	ListExpiredPermissionGrants(ctx context.Context) ([]PermissionGrant, error)
	ExpirePermissionGrant(ctx context.Context, grantInternalID int64) error
	InsertAuditLog(ctx context.Context, entry AuditLogEntry) error
	InsertCasbinRule(ctx context.Context, rule CasbinRule) (int64, error)
	DeleteCasbinRule(ctx context.Context, id int64) error
	ListAllCasbinRules(ctx context.Context) ([]CasbinRule, error)
	EnsureAdminBootstrap(ctx context.Context, subjectKey, displayName string) error
}
