package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/charviki/maze/fabrication/cradle/auth"
	"github.com/charviki/maze/the-mesa/director-core/internal/repository"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

// AuthBootstrapStore 定义 BootstrapAdmin 所需的最小持久化能力。
type AuthBootstrapStore interface {
	UpsertSubject(ctx context.Context, subject Subject) (Subject, error)
	InsertCasbinRule(ctx context.Context, rule CasbinRule) (int64, error)
	ListAllCasbinRules(ctx context.Context) ([]CasbinRule, error)
}

var (
	// ErrInvalidCredentials 用户名或密码错误。
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserDisabled 用户已被禁用。
	ErrUserDisabled = errors.New("user is disabled")
	// ErrRefreshTokenNotFound refresh token 不存在或已被并发消费。
	ErrRefreshTokenNotFound = errors.New("refresh token not found or already used")
	// ErrRefreshTokenRevoked refresh token 已被撤销。
	ErrRefreshTokenRevoked = errors.New("refresh token has been revoked")
	// ErrRefreshTokenExpired refresh token 已过期。
	ErrRefreshTokenExpired = errors.New("refresh token has expired")
)

// AuthService 处理用户认证业务逻辑。
type AuthService struct {
	users        repository.UserStore
	credentials  repository.CredentialStore
	bootstrap    AuthBootstrapStore
	txm          AuthTxManager
	jwtSecret    string
	accessTTL    time.Duration
	refreshTTL   time.Duration
	adminUser    string
	adminPass    string
	adminSubject string
}

// NewAuthService 创建 AuthService。
func NewAuthService(
	users repository.UserStore,
	credentials repository.CredentialStore,
	bootstrap AuthBootstrapStore,
	txm AuthTxManager,
	jwtSecret string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
	adminUser string,
	adminPass string,
	adminSubject string,
) *AuthService {
	return &AuthService{
		users:        users,
		credentials:  credentials,
		bootstrap:    bootstrap,
		txm:          txm,
		jwtSecret:    jwtSecret,
		accessTTL:    accessTTL,
		refreshTTL:   refreshTTL,
		adminUser:    adminUser,
		adminPass:    adminPass,
		adminSubject: adminSubject,
	}
}

// LoginResult 包含登录和刷新操作返回的令牌对。
type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// Login 校验用户名密码，签发 access/refresh token 对。
func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	user, err := s.users.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("internal error during user lookup")
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrUserDisabled
	}

	// bcrypt 返回密码不匹配、哈希格式损坏、版本过新等多种错误；
	// 统一返回 ErrInvalidCredentials，避免攻击者通过 500 vs 401 差异探测哈希状态。
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokenPair(ctx, user.SubjectKey)
}

// Refresh 使用 refresh token 换发新的令牌对，旧 token 在同一事务中原子消费，防止并发重放。
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*LoginResult, error) {
	tokenHash := auth.HashToken(refreshToken)
	var result *LoginResult
	err := s.txm.WithinTx(ctx, func(txCtx context.Context) error {
		// 原子消费：UPDATE ... WHERE status='active' RETURNING 确保并发安全。
		consumed, err := s.credentials.ConsumeRefreshCredential(txCtx, tokenHash)
		if err != nil {
			return fmt.Errorf("consume refresh credential: %w", err)
		}
		if consumed == nil {
			return ErrRefreshTokenNotFound
		}
		if consumed.ExpiresAt != nil && consumed.ExpiresAt.Before(time.Now().UTC()) {
			return ErrRefreshTokenExpired
		}

		pair, err := s.generateTokenPair(txCtx, consumed.SubjectKey)
		if err != nil {
			return err
		}
		result = pair
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Logout 仅撤销当前主体自己的 refresh token。
func (s *AuthService) Logout(ctx context.Context, subjectKey, refreshToken string) error {
	tokenHash := auth.HashToken(refreshToken)
	cred, err := s.credentials.GetCredentialByTokenHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("lookup credential: %w", err)
	}
	// 这里必须保持幂等且不泄露 token 是否存在、是否已撤销、是否属于当前主体。
	if cred == nil || cred.Type != repository.CredentialTypeUserRefresh {
		return nil
	}
	if cred.SubjectKey != subjectKey || cred.Status != repository.CredentialStatusActive {
		return nil
	}
	if cred.ExpiresAt != nil && cred.ExpiresAt.Before(time.Now().UTC()) {
		return nil
	}

	return s.credentials.RevokeCredential(ctx, tokenHash)
}

// BootstrapAdmin 在无用户时创建默认管理员，含 subject、用户记录和超级管理员 Casbin 策略。
// 返回 true 表示新创建了管理员，false 表示管理员已存在。
func (s *AuthService) BootstrapAdmin(ctx context.Context) (bool, error) {
	existing, err := s.users.GetUserByUsername(ctx, s.adminUser)
	if err != nil {
		return false, fmt.Errorf("check admin user: %w", err)
	}
	if existing != nil {
		s.adminPass = ""
		return false, nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(s.adminPass), bcrypt.DefaultCost)
	if err != nil {
		return false, fmt.Errorf("hash admin password: %w", err)
	}

	err = s.txm.WithinTx(ctx, func(txCtx context.Context) error {
		subject, err := s.bootstrap.UpsertSubject(txCtx, Subject{
			SubjectKey:  s.adminSubject,
			SubjectType: "user",
			DisplayName: s.adminUser,
			IsSystem:    true,
		})
		if err != nil {
			return fmt.Errorf("upsert admin subject: %w", err)
		}

		if _, err := s.users.CreateUser(txCtx, s.adminUser, string(hash), subject.SubjectKey); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return nil
			}
			return fmt.Errorf("create admin user: %w", err)
		}

		// admin 超级管理员策略幂等写入：只在规则不存在时插入。
		rules, err := s.bootstrap.ListAllCasbinRules(txCtx)
		if err != nil {
			return fmt.Errorf("list casbin rules: %w", err)
		}
		for _, rule := range rules {
			if rule.Ptype == "p" && rule.V0 == s.adminSubject && rule.V1 == "*" && rule.V2 == "*" {
				return nil
			}
		}
		if _, insertErr := s.bootstrap.InsertCasbinRule(txCtx, CasbinRule{
			Ptype: "p",
			V0:    s.adminSubject,
			V1:    "*",
			V2:    "*",
		}); insertErr != nil {
			return fmt.Errorf("insert admin casbin rule: %w", insertErr)
		}

		return nil
	})
	if err != nil {
		return false, err
	}
	// BootstrapAdmin 结束后清零，减少内存中敏感数据暴露窗口。
	// 无论是否新创建了管理员，明文密码都不再需要。
	s.adminPass = ""
	return true, nil
}

func (s *AuthService) generateTokenPair(ctx context.Context, subjectKey string) (*LoginResult, error) {
	accessToken, err := auth.GenerateAccessToken(s.jwtSecret, auth.DefaultIssuer, subjectKey, s.accessTTL)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	refreshTokenHash := auth.HashToken(refreshToken)
	expiresAt := time.Now().UTC().Add(s.refreshTTL)

	if err := s.credentials.StoreCredential(ctx, &repository.Credential{
		Type:       repository.CredentialTypeUserRefresh,
		TokenHash:  refreshTokenHash,
		SubjectKey: subjectKey,
		ExpiresAt:  &expiresAt,
		Status:     repository.CredentialStatusActive,
	}); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}
