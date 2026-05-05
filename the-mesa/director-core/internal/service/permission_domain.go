package service

import (
	"errors"
	"time"
)

var (
	// ErrPermissionApplicationNotFound 表示权限申请资源不存在。
	ErrPermissionApplicationNotFound = errors.New("permission application not found")
	// ErrPermissionGrantNotFound 表示权限授权资源不存在。
	ErrPermissionGrantNotFound = errors.New("permission grant not found")
	// ErrPermissionApplicationStateChanged 表示申请单在当前事务外已被并发修改。
	ErrPermissionApplicationStateChanged = errors.New("permission application state changed")
)

// ValidationError 表示业务输入不满足权限域约束。
type ValidationError struct {
	message string
}

// Error 返回可直接暴露给调用方的校验错误信息。
func (e ValidationError) Error() string {
	return e.message
}

// NewValidationError 创建权限域输入校验错误。
func NewValidationError(message string) error {
	return ValidationError{message: message}
}

// PreconditionError 表示资源当前状态不允许执行本次操作。
type PreconditionError struct {
	message string
}

// Error 返回业务前置条件错误信息。
func (e PreconditionError) Error() string {
	return e.message
}

// NewPreconditionError 创建权限域状态冲突错误。
func NewPreconditionError(message string) error {
	return PreconditionError{message: message}
}

// PermissionApplicationStatus 是权限申请单状态的内部表示。
type PermissionApplicationStatus string

const (
	// PermissionApplicationStatusPending 表示申请单待审批。
	PermissionApplicationStatusPending PermissionApplicationStatus = "pending"
	// PermissionApplicationStatusApproved 表示申请单已批准。
	PermissionApplicationStatusApproved PermissionApplicationStatus = "approved"
	// PermissionApplicationStatusDenied 表示申请单已拒绝。
	PermissionApplicationStatusDenied PermissionApplicationStatus = "denied"
	// PermissionApplicationStatusRevoked 表示申请单已撤销。
	PermissionApplicationStatusRevoked PermissionApplicationStatus = "revoked"
	// PermissionApplicationStatusExpired 表示申请单对应授权已过期。
	PermissionApplicationStatusExpired PermissionApplicationStatus = "expired"
)

// PermissionGrantStatus 是授权结果状态的内部表示。
type PermissionGrantStatus string

const (
	// PermissionGrantStatusActive 表示授权结果当前生效。
	PermissionGrantStatusActive PermissionGrantStatus = "active"
	// PermissionGrantStatusRevoked 表示授权结果已被人工撤销。
	PermissionGrantStatusRevoked PermissionGrantStatus = "revoked"
	// PermissionGrantStatusExpired 表示授权结果因过期失效。
	PermissionGrantStatusExpired PermissionGrantStatus = "expired"
)

// Subject 描述权限系统中的稳定主体。
type Subject struct {
	SubjectKey  string
	SubjectType string
	DisplayName string
	IsSystem    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PermissionTarget 是权限申请单中的最小资源动作对。
type PermissionTarget struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// PermissionApplication 描述权限申请资源。
type PermissionApplication struct {
	InternalID    int64
	ID            string
	SubjectKey    string
	Targets       []PermissionTarget
	Reason        string
	Status        PermissionApplicationStatus
	ReviewedBy    string
	ReviewComment string
	ReviewedAt    *time.Time
	ExpiresAt     *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// PermissionGrant 描述已经生效的授权结果。
type PermissionGrant struct {
	InternalID                            int64
	ID                                    string
	SubjectKey                            string
	Resource                              string
	Action                                string
	SourcePermissionApplicationInternalID int64
	SourcePermissionApplicationID         string
	CasbinRuleID                          *int64
	Status                                PermissionGrantStatus
	RevokedBy                             string
	RevokedReason                         string
	ExpiresAt                             *time.Time
	CreatedAt                             time.Time
	UpdatedAt                             time.Time
}

// CasbinRule 是运行时策略的持久化表示。
type CasbinRule struct {
	ID    int64
	Ptype string
	V0    string
	V1    string
	V2    string
}

// AuditLogEntry 记录权限闭环关键行为。
type AuditLogEntry struct {
	Action                          string
	ActorSubjectKey                 string
	TargetSubjectKey                string
	PermissionApplicationInternalID *int64
	Details                         map[string]any
}

// CreatePermissionApplicationParams 是创建申请单所需参数。
type CreatePermissionApplicationParams struct {
	SubjectKey string
	Targets    []PermissionTarget
	Reason     string
	ExpiresAt  *time.Time
}

// ListPermissionApplicationsParams 是申请单列表查询参数。
type ListPermissionApplicationsParams struct {
	Status   PermissionApplicationStatus
	Page     int32
	PageSize int32
}

// ReviewPermissionApplicationParams 是审批申请单的写入参数。
type ReviewPermissionApplicationParams struct {
	InternalID     int64
	ExpectedStatus PermissionApplicationStatus
	Status         PermissionApplicationStatus
	ReviewedBy     string
	ReviewComment  string
}

// UpdatePermissionApplicationStatusParams 用于不改变业务载荷的状态同步。
type UpdatePermissionApplicationStatusParams struct {
	InternalID    int64
	Status        PermissionApplicationStatus
	ReviewedBy    string
	ReviewComment string
}

// CreatePermissionGrantParams 是创建授权结果的写入参数。
type CreatePermissionGrantParams struct {
	SubjectKey                            string
	Resource                              string
	Action                                string
	SourcePermissionApplicationInternalID int64
	ExpiresAt                             *time.Time
}

// RevokePermissionGrantsParams 是批量撤销申请单授权结果的写入参数。
type RevokePermissionGrantsParams struct {
	SourcePermissionApplicationInternalID int64
	Status                                PermissionGrantStatus
	RevokedBy                             string
	RevokedReason                         string
}
