package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

// PolicyReloader 在授权结果变更后刷新 Casbin 内存策略。
type PolicyReloader func() error

// PermissionService 处理权限申请闭环。
type PermissionService struct {
	store        PermissionStore
	txm          TxManager
	reloadPolicy PolicyReloader
}

// CreatePermissionApplicationInput 创建权限申请的输入。
type CreatePermissionApplicationInput struct {
	SubjectKey      string
	Targets         []PermissionTarget
	Reason          string
	ExpiresAt       *time.Time
	ActorSubjectKey string
}

// ReviewPermissionApplicationInput 审批输入。
type ReviewPermissionApplicationInput struct {
	PermissionApplicationID string
	Approved                bool
	ReviewComment           string
	ReviewerSubjectKey      string
}

// RevokePermissionApplicationInput 撤销输入。
type RevokePermissionApplicationInput struct {
	PermissionApplicationID string
	RevokeReason            string
	ReviewerSubjectKey      string
}

// ListPermissionApplicationsInput 列表输入。
type ListPermissionApplicationsInput struct {
	Status   string
	Page     int32
	PageSize int32
}

// NewPermissionService 创建 PermissionService。
func NewPermissionService(store PermissionStore, txm TxManager, reload PolicyReloader) *PermissionService {
	return &PermissionService{store: store, txm: txm, reloadPolicy: reload}
}

// CreatePermissionApplication 创建权限申请。
func (s *PermissionService) CreatePermissionApplication(ctx context.Context, input CreatePermissionApplicationInput) (PermissionApplication, error) {
	if err := validateSubjectAndTargets(input.SubjectKey, input.Targets); err != nil {
		return PermissionApplication{}, err
	}
	if strings.TrimSpace(input.Reason) == "" {
		return PermissionApplication{}, NewValidationError("reason is required")
	}
	if input.ExpiresAt != nil && !input.ExpiresAt.After(time.Now().UTC()) {
		// 过期时间必须晚于当前时刻；否则审批成功后授权会立刻进入“应失效但 Casbin 尚未刷新”的不一致窗口。
		return PermissionApplication{}, NewValidationError("expires at must be in the future")
	}

	var created PermissionApplication
	err := s.txm.WithinTx(ctx, func(ctx context.Context) error {
		if err := upsertSubject(ctx, s.store, input.SubjectKey, false); err != nil {
			return err
		}

		row, err := s.store.CreatePermissionApplication(ctx, CreatePermissionApplicationParams{
			SubjectKey: input.SubjectKey,
			Targets:    input.Targets,
			Reason:     strings.TrimSpace(input.Reason),
			ExpiresAt:  input.ExpiresAt,
		})
		if err != nil {
			return err
		}
		created = row

		return s.store.InsertAuditLog(ctx, AuditLogEntry{
			Action:                          "permission_request.created",
			ActorSubjectKey:                 input.ActorSubjectKey,
			TargetSubjectKey:                input.SubjectKey,
			PermissionApplicationInternalID: &row.InternalID,
			Details: map[string]any{
				"targets": input.Targets,
				"reason":  input.Reason,
			},
		})
	})
	return created, err
}

// ListPermissionApplications 列出权限申请。
func (s *PermissionService) ListPermissionApplications(ctx context.Context, input ListPermissionApplicationsInput) ([]PermissionApplication, int64, error) {
	status := strings.TrimSpace(strings.ToLower(input.Status))
	if status == "" {
		return s.store.ListPermissionApplications(ctx, ListPermissionApplicationsParams{
			Page:     input.Page,
			PageSize: input.PageSize,
		})
	}

	parsedStatus, err := parseApplicationStatus(status)
	if err != nil {
		return nil, 0, err
	}
	return s.store.ListPermissionApplications(ctx, ListPermissionApplicationsParams{
		Status:   parsedStatus,
		Page:     input.Page,
		PageSize: input.PageSize,
	})
}

// GetPermissionApplication 获取单个权限申请。
func (s *PermissionService) GetPermissionApplication(ctx context.Context, permissionApplicationID string) (PermissionApplication, error) {
	if strings.TrimSpace(permissionApplicationID) == "" {
		return PermissionApplication{}, NewValidationError("permission application id is required")
	}
	return s.store.GetPermissionApplication(ctx, permissionApplicationID)
}

// ReviewPermissionApplication 审批权限申请。
func (s *PermissionService) ReviewPermissionApplication(ctx context.Context, input ReviewPermissionApplicationInput) (PermissionApplication, error) {
	if strings.TrimSpace(input.PermissionApplicationID) == "" {
		return PermissionApplication{}, NewValidationError("permission application id is required")
	}
	if strings.TrimSpace(input.ReviewerSubjectKey) == "" {
		return PermissionApplication{}, NewValidationError("reviewer subject key is required")
	}

	var updated PermissionApplication
	err := s.txm.WithinTx(ctx, func(ctx context.Context) error {
		current, err := s.store.GetPermissionApplication(ctx, input.PermissionApplicationID)
		if err != nil {
			return err
		}
		if current.Status != PermissionApplicationStatusPending {
			return NewPreconditionError(fmt.Sprintf("permission request %s is already %s", input.PermissionApplicationID, current.Status))
		}

		if err := upsertSubject(ctx, s.store, input.ReviewerSubjectKey, true); err != nil {
			return err
		}

		if input.Approved {
			updated, err = s.store.ReviewPermissionApplication(ctx, ReviewPermissionApplicationParams{
				InternalID:     current.InternalID,
				ExpectedStatus: PermissionApplicationStatusPending,
				Status:         PermissionApplicationStatusApproved,
				ReviewedBy:     input.ReviewerSubjectKey,
				ReviewComment:  input.ReviewComment,
			})
			if err != nil {
				if errors.Is(err, ErrPermissionApplicationStateChanged) {
					return NewPreconditionError(fmt.Sprintf("permission request %s is no longer pending", input.PermissionApplicationID))
				}
				return err
			}

			for _, target := range updated.Targets {
				grant, err := s.store.CreatePermissionGrant(ctx, CreatePermissionGrantParams{
					SubjectKey:                            updated.SubjectKey,
					Resource:                              target.Resource,
					Action:                                target.Action,
					SourcePermissionApplicationInternalID: updated.InternalID,
					ExpiresAt:                             updated.ExpiresAt,
				})
				if err != nil {
					if errors.Is(err, ErrPermissionApplicationStateChanged) {
						return NewPreconditionError(fmt.Sprintf("permission request %s is no longer pending", input.PermissionApplicationID))
					}
					return err
				}

				// grant 与 Casbin rule 复用同一个事务上下文；任一步失败都会由 TxManager 统一回滚，避免孤儿规则残留。
				ruleID, err := s.store.InsertCasbinRule(ctx, CasbinRule{
					Ptype: "p",
					V0:    updated.SubjectKey,
					V1:    target.Resource,
					V2:    target.Action,
				})
				if err != nil {
					return err
				}

				if err := s.store.AttachGrantCasbinRule(ctx, grant.InternalID, ruleID); err != nil {
					return err
				}
			}

			return s.store.InsertAuditLog(ctx, AuditLogEntry{
				Action:                          "permission_request.approved",
				ActorSubjectKey:                 input.ReviewerSubjectKey,
				TargetSubjectKey:                updated.SubjectKey,
				PermissionApplicationInternalID: &updated.InternalID,
				Details: map[string]any{
					"review_comment": input.ReviewComment,
				},
			})
		}

		// 拒绝必须留下审批意见，方便后续审计和申请方补充材料。
		if strings.TrimSpace(input.ReviewComment) == "" {
			return NewValidationError("review comment is required when denying a request")
		}

		updated, err = s.store.ReviewPermissionApplication(ctx, ReviewPermissionApplicationParams{
			InternalID:     current.InternalID,
			ExpectedStatus: PermissionApplicationStatusPending,
			Status:         PermissionApplicationStatusDenied,
			ReviewedBy:     input.ReviewerSubjectKey,
			ReviewComment:  input.ReviewComment,
		})
		if err != nil {
			if errors.Is(err, ErrPermissionApplicationStateChanged) {
				return NewPreconditionError(fmt.Sprintf("permission request %s is no longer pending", input.PermissionApplicationID))
			}
			return err
		}

		return s.store.InsertAuditLog(ctx, AuditLogEntry{
			Action:                          "permission_request.denied",
			ActorSubjectKey:                 input.ReviewerSubjectKey,
			TargetSubjectKey:                updated.SubjectKey,
			PermissionApplicationInternalID: &updated.InternalID,
			Details: map[string]any{
				"review_comment": input.ReviewComment,
			},
		})
	})
	if err != nil {
		return PermissionApplication{}, err
	}
	if input.Approved {
		s.reloadAfterCommit("review permission application " + updated.ID)
	}
	return updated, nil
}

// RevokePermissionApplication 撤销已批准申请对应的权限。
func (s *PermissionService) RevokePermissionApplication(ctx context.Context, input RevokePermissionApplicationInput) (PermissionApplication, error) {
	if strings.TrimSpace(input.PermissionApplicationID) == "" {
		return PermissionApplication{}, NewValidationError("permission application id is required")
	}
	if strings.TrimSpace(input.ReviewerSubjectKey) == "" {
		return PermissionApplication{}, NewValidationError("reviewer subject key is required")
	}

	var updated PermissionApplication
	err := s.txm.WithinTx(ctx, func(ctx context.Context) error {
		current, err := s.store.GetPermissionApplication(ctx, input.PermissionApplicationID)
		if err != nil {
			return err
		}
		if current.Status != PermissionApplicationStatusApproved {
			return NewPreconditionError(fmt.Sprintf("permission request %s cannot be revoked from status %s", input.PermissionApplicationID, current.Status))
		}

		if err := upsertSubject(ctx, s.store, input.ReviewerSubjectKey, true); err != nil {
			return err
		}

		revokedGrants, err := s.store.RevokeActivePermissionGrantsByApplication(ctx, RevokePermissionGrantsParams{
			SourcePermissionApplicationInternalID: current.InternalID,
			Status:                                PermissionGrantStatusRevoked,
			RevokedBy:                             input.ReviewerSubjectKey,
			RevokedReason:                         input.RevokeReason,
		})
		if err != nil {
			return err
		}
		for _, grant := range revokedGrants {
			if grant.CasbinRuleID != nil {
				if err := s.store.DeleteCasbinRule(ctx, *grant.CasbinRuleID); err != nil {
					return err
				}
			}
		}

		updated, err = s.store.UpdatePermissionApplicationStatus(ctx, UpdatePermissionApplicationStatusParams{
			InternalID:    current.InternalID,
			Status:        PermissionApplicationStatusRevoked,
			ReviewedBy:    input.ReviewerSubjectKey,
			ReviewComment: input.RevokeReason,
		})
		if err != nil {
			return err
		}

		return s.store.InsertAuditLog(ctx, AuditLogEntry{
			Action:                          "permission_request.revoked",
			ActorSubjectKey:                 input.ReviewerSubjectKey,
			TargetSubjectKey:                updated.SubjectKey,
			PermissionApplicationInternalID: &updated.InternalID,
			Details: map[string]any{
				"revoked_grants": len(revokedGrants),
				"revoke_reason":  input.RevokeReason,
			},
		})
	})
	if err != nil {
		return PermissionApplication{}, err
	}
	s.reloadAfterCommit("revoke permission application " + updated.ID)
	return updated, nil
}

// ListSubjectPermissions 返回主体当前授权结果。
func (s *PermissionService) ListSubjectPermissions(ctx context.Context, subjectKey string) ([]PermissionGrant, error) {
	if strings.TrimSpace(subjectKey) == "" {
		return nil, NewValidationError("subject key is required")
	}
	grants, err := s.store.ListSubjectPermissionGrants(ctx, subjectKey)
	if err != nil {
		return nil, err
	}
	return filterVisiblePermissionGrants(grants, time.Now().UTC()), nil
}

// ExpirePermissionGrants 清理已过期授权。
func (s *PermissionService) ExpirePermissionGrants(ctx context.Context) (int, error) {
	expired, err := s.store.ListExpiredPermissionGrants(ctx)
	if err != nil {
		return 0, err
	}
	if len(expired) == 0 {
		return 0, nil
	}

	count := 0
	err = s.txm.WithinTx(ctx, func(ctx context.Context) error {
		// 逐条处理是为了同步删除对应 Casbin 规则并回写申请单状态，保证审计和运行时策略一致。
		for _, grant := range expired {
			if err := s.store.ExpirePermissionGrant(ctx, grant.InternalID); err != nil {
				return err
			}
			if grant.CasbinRuleID != nil {
				if err := s.store.DeleteCasbinRule(ctx, *grant.CasbinRuleID); err != nil {
					return err
				}
			}
			if _, err := s.store.UpdatePermissionApplicationStatus(ctx, UpdatePermissionApplicationStatusParams{
				InternalID: grant.SourcePermissionApplicationInternalID,
				Status:     PermissionApplicationStatusExpired,
			}); err != nil {
				return err
			}
			if err := s.store.InsertAuditLog(ctx, AuditLogEntry{
				Action:                          "permission_grant.expired",
				TargetSubjectKey:                grant.SubjectKey,
				PermissionApplicationInternalID: &grant.SourcePermissionApplicationInternalID,
				Details: map[string]any{
					"resource": grant.Resource,
					"action":   grant.Action,
				},
			}); err != nil {
				return err
			}
			count++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	s.reloadAfterCommit("expire permission grants")
	return count, nil
}

func (s *PermissionService) reloadAfterCommit(action string) {
	if s.reloadPolicy == nil {
		return
	}

	// 权限变更一旦提交，接口就不能再把它伪装成失败；这里只做提交后的重试刷新并显式记录告警。
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		err = s.reloadPolicy()
		if err == nil {
			return
		}
		time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
	}
	log.Printf("permission policy reload failed after committed action=%s: %v", action, err)
}

func validateSubjectAndTargets(subjectKey string, targets []PermissionTarget) error {
	if strings.TrimSpace(subjectKey) == "" {
		return NewValidationError("subject key is required")
	}
	if len(targets) == 0 {
		return NewValidationError("at least one permission target is required")
	}
	seen := make(map[string]struct{}, len(targets))
	for _, target := range targets {
		if strings.TrimSpace(target.Resource) == "" || strings.TrimSpace(target.Action) == "" {
			return NewValidationError("permission target resource and action are required")
		}
		key := strings.TrimSpace(target.Resource) + "\x00" + strings.TrimSpace(target.Action)
		if _, ok := seen[key]; ok {
			return NewValidationError("duplicate permission target is not allowed")
		}
		seen[key] = struct{}{}
	}
	return nil
}

func filterVisiblePermissionGrants(grants []PermissionGrant, now time.Time) []PermissionGrant {
	filtered := make([]PermissionGrant, 0, len(grants))
	for _, grant := range grants {
		if grant.Status != PermissionGrantStatusActive {
			continue
		}
		if grant.ExpiresAt != nil && !grant.ExpiresAt.After(now) {
			continue
		}
		filtered = append(filtered, grant)
	}
	return filtered
}

func upsertSubject(ctx context.Context, store PermissionStore, subjectKey string, isSystem bool) error {
	if strings.TrimSpace(subjectKey) == "" {
		return NewValidationError("subject key is required")
	}
	_, err := store.UpsertSubject(ctx, Subject{
		SubjectKey:  subjectKey,
		SubjectType: subjectTypeFromKey(subjectKey),
		DisplayName: displayNameFromKey(subjectKey),
		IsSystem:    isSystem,
	})
	return err
}

func parseApplicationStatus(raw string) (PermissionApplicationStatus, error) {
	switch PermissionApplicationStatus(raw) {
	case PermissionApplicationStatusPending,
		PermissionApplicationStatusApproved,
		PermissionApplicationStatusDenied,
		PermissionApplicationStatusRevoked,
		PermissionApplicationStatusExpired:
		return PermissionApplicationStatus(raw), nil
	default:
		return "", NewValidationError(fmt.Sprintf("invalid permission request status %q", raw))
	}
}

func subjectTypeFromKey(subjectKey string) string {
	parts := strings.SplitN(subjectKey, ":", 2)
	if len(parts) == 0 || parts[0] == "" {
		return "subject"
	}
	return parts[0]
}

func displayNameFromKey(subjectKey string) string {
	parts := strings.SplitN(subjectKey, ":", 2)
	if len(parts) == 2 && parts[1] != "" {
		return parts[1]
	}
	return subjectKey
}
