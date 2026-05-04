package auth

import "context"

// UserInfo 是授权层消费的稳定主体信息。
// 当前阶段只保留最小字段，避免把未来角色体系提前写死到共享契约里。
type UserInfo struct {
	SubjectKey  string `json:"subject_key"`
	DisplayName string `json:"display_name,omitempty"`
}

// UserInfoExtractor 从上下文中提取稳定主体。
// 各项目自己决定如何把认证结果映射成 SubjectKey。
type UserInfoExtractor func(ctx context.Context) (*UserInfo, error)

type contextKey struct{}

// WithUserInfo 将主体信息注入上下文。
func WithUserInfo(ctx context.Context, user *UserInfo) context.Context {
	return context.WithValue(ctx, contextKey{}, user)
}

// GetUserInfo 从上下文读取主体信息。
func GetUserInfo(ctx context.Context) *UserInfo {
	user, _ := ctx.Value(contextKey{}).(*UserInfo)
	return user
}
