package protocol

import "time"

// GitKey Git 密钥（token 不回传，仅返回 token_mask）
type GitKey struct {
	Name      string    `json:"name"`
	Token     string    `json:"-"`
	TokenMask string    `json:"token_mask"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
