package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrTokenExpired 表示 JWT 已过期。
var ErrTokenExpired = errors.New("token expired")

// ErrTokenInvalid 表示 JWT 无效。
var ErrTokenInvalid = errors.New("invalid token")

// Claims 封装标准 JWT claims，用于 access token 签发和验证。
// 主体标识通过 RegisteredClaims.Subject 传递，中间件通过 claims.Subject 提取。
type Claims struct {
	jwt.RegisteredClaims
}

// TokenPair 包含访问令牌和刷新令牌。
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// DefaultIssuer 是 JWT 签发方标识，所有签发和校验共用。
const DefaultIssuer = "maze-director-core"

// GenerateAccessToken 使用 HMAC-SHA256 签发访问令牌。
func GenerateAccessToken(secret string, issuer string, subjectKey string, ttl time.Duration) (string, error) {
	now := time.Now()
	jti, err := generateJTI()
	if err != nil {
		return "", err
	}
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   subjectKey,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        jti,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateAccessToken 验证 JWT 签名、有效期和签发方，返回解析后的 Claims。
func ValidateAccessToken(secret string, issuer string, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return []byte(secret), nil
	}, jwt.WithIssuer(issuer))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

// GenerateRefreshToken 生成加密安全的刷新令牌。
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// HashToken 对令牌进行 SHA-256 哈希，用于安全存储。
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func generateJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate JTI: %w", err)
	}
	return hex.EncodeToString(b), nil
}
