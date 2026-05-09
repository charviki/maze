package kit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	client "github.com/charviki/maze-cradle/api/gen/http"
)

// TestConfig 集成测试运行参数，从环境变量加载。
type TestConfig struct {
	DirectorCoreURL     string
	Env                 string
	Namespace           string
	DataDir             string
	JWTSecret          string
	AdminUsername      string
	AdminPassword      string
	AgentStorageBackend string
	EnableHostPool      bool
	PoolClaudeSize      int
	PoolGoSize          int
	StreamEvents        bool
}

// LoadTestConfig 从环境变量加载集成测试配置。
func LoadTestConfig() *TestConfig {
	cfg := &TestConfig{
		DirectorCoreURL:     getEnv("MAZE_TEST_DIRECTOR_CORE_URL", "http://localhost:9090"),
		Env:                 getEnv("MAZE_TEST_ENV", "docker"),
		Namespace:           getEnv("MAZE_TEST_NAMESPACE", "maze-test"),
		DataDir:             getEnv("MAZE_TEST_DATA_DIR", os.Getenv("HOME")+"/.maze-test"),
		JWTSecret:          getEnv("MAZE_TEST_JWT_SECRET", ""),
		AdminUsername:      getEnv("MAZE_TEST_ADMIN_USERNAME", "admin"),
		AdminPassword:      getEnv("MAZE_TEST_ADMIN_PASSWORD", "admin"),
		AgentStorageBackend: getEnv("MAZE_TEST_AGENT_STORAGE_BACKEND", defaultAgentStorageBackend(getEnv("MAZE_TEST_ENV", "docker"))),
		EnableHostPool:      getEnvBool("MAZE_TEST_ENABLE_HOST_POOL", false),
		PoolClaudeSize:      getEnvInt("MAZE_TEST_POOL_CLAUDE_SIZE", 4),
		PoolGoSize:          getEnvInt("MAZE_TEST_POOL_GO_SIZE", 1),
		StreamEvents:        getEnvBool("MAZE_TEST_STREAM_EVENTS", true),
	}
	return cfg
}

func defaultAgentStorageBackend(env string) string {
	switch env {
	case "docker":
		return "bind"
	case "kubernetes":
		return "hostpath"
	default:
		return "unknown"
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

// loginResponse 对应 grpc-gateway 返回的 LoginResponse JSON。
// proto JSON 会把 int64 序列化为字符串，因此 expiresIn 用 string 接收。
type loginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
}

// LoginAdmin 通过 /api/v1/auth/login 获取 JWT token pair。
func LoginAdmin(ctx context.Context, cfg *TestConfig) (*loginResponse, error) {
	body, _ := json.Marshal(map[string]string{
		"username": cfg.AdminUsername,
		"password": cfg.AdminPassword,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.DirectorCoreURL+"/api/v1/auth/login", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read login response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed (status=%d): %s", resp.StatusCode, string(raw))
	}

	var result loginResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode login response: %w", err)
	}
	return &result, nil
}

// NewTestAPIClient 先登录获取 JWT，再创建带 Authorization header 的 OpenAPI client。
func NewTestAPIClient(cfg *TestConfig) (*client.APIClient, error) {
	loginResult, err := LoginAdmin(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("login for test client: %w", err)
	}

	config := client.NewConfiguration()
	config.Servers = client.ServerConfigurations{
		{URL: cfg.DirectorCoreURL},
	}
	config.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	config.AddDefaultHeader("Authorization", "Bearer "+loginResult.AccessToken)
	return client.NewAPIClient(config), nil
}
