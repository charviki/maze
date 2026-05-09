package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndValidateAccessToken(t *testing.T) {
	secret := "test-secret-key"
	subjectKey := "user:admin"
	ttl := 15 * time.Minute

	token, err := GenerateAccessToken(secret, DefaultIssuer, subjectKey, ttl)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}

	claims, err := ValidateAccessToken(secret, DefaultIssuer, token)
	if err != nil {
		t.Fatalf("ValidateAccessToken failed: %v", err)
	}
	if claims.Subject != subjectKey {
		t.Errorf("expected subject %q, got %q", subjectKey, claims.Subject)
	}
}

func TestValidateAccessToken_Expired(t *testing.T) {
	secret := "test-secret-key"
	subjectKey := "user:admin"
	ttl := -1 * time.Second

	token, _ := GenerateAccessToken(secret, DefaultIssuer, subjectKey, ttl)
	_, err := ValidateAccessToken(secret, DefaultIssuer, token)
	if err != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestValidateAccessToken_InvalidSignature(t *testing.T) {
	secret := "test-secret-key"
	wrongSecret := "wrong-secret"
	subjectKey := "user:admin"
	ttl := 15 * time.Minute

	token, _ := GenerateAccessToken(secret, DefaultIssuer, subjectKey, ttl)
	_, err := ValidateAccessToken(wrongSecret, DefaultIssuer, token)
	if err != ErrTokenInvalid {
		t.Errorf("expected ErrTokenInvalid, got %v", err)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token1, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}
	token2, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}
	if token1 == token2 {
		t.Error("two refresh tokens should not be equal")
	}
	if len(token1) != 64 {
		t.Errorf("expected 64-char hex string, got %d chars", len(token1))
	}
}

func TestHashToken(t *testing.T) {
	token := "test-token"
	hash1 := HashToken(token)
	hash2 := HashToken(token)
	if hash1 != hash2 {
		t.Error("same input should produce same hash")
	}
	if hash1 == token {
		t.Error("hash should differ from input")
	}
}

func TestValidateAccessToken_WrongIssuer(t *testing.T) {
	secret := "test-secret-key"
	subjectKey := "user:admin"
	ttl := 15 * time.Minute

	token, _ := GenerateAccessToken(secret, DefaultIssuer, subjectKey, ttl)
	_, err := ValidateAccessToken(secret, "wrong-issuer", token)
	if err != ErrTokenInvalid {
		t.Errorf("expected ErrTokenInvalid for wrong issuer, got %v", err)
	}
}

func TestValidateAccessToken_MissingIssuer(t *testing.T) {
	secret := "test-secret-key"
	subjectKey := "user:admin"
	ttl := 15 * time.Minute

	// 手动构造不含 Issuer 的 token，模拟旧版签发
	now := time.Now()
	jti, _ := generateJTI()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subjectKey,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        jti,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	_, err := ValidateAccessToken(secret, DefaultIssuer, tokenString)
	if err != ErrTokenInvalid {
		t.Errorf("expected ErrTokenInvalid for missing issuer, got %v", err)
	}
}
