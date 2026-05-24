package protocol

import "time"

// GitTokenType 凭证类型枚举
type GitTokenType = string

// Git key type constants
const (
	GitKeyTypePAT    GitTokenType = "PERSONAL_ACCESS_TOKEN"
	GitKeyTypeSSHKey GitTokenType = "SSH_KEY"
)

// GitKey Git 密钥（token 不回传，仅返回 token_mask）
type GitKey struct {
	Name           string       `json:"name"`
	Token          string       `json:"-"`
	TokenMask      string       `json:"token_mask"`
	TokenType      GitTokenType `json:"token_type,omitempty"`
	Host           string       `json:"host,omitempty"`
	DecryptedToken string       `json:"-"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}
