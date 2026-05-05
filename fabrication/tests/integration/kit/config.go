package kit

import (
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
	AuthToken           string
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
		AuthToken:           getEnv("MAZE_TEST_AUTH_TOKEN", "test-integration-token"),
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

// NewTestAPIClient 创建指向 Director Core HTTP 端口的 OpenAPI 生成 client。
// 服务端 grpc-gateway 已返回标准 proto JSON，与 OpenAPI spec 完全一致，无需额外解包。
func NewTestAPIClient(cfg *TestConfig) *client.APIClient {
	config := client.NewConfiguration()
	config.Servers = client.ServerConfigurations{
		{URL: cfg.DirectorCoreURL},
	}
	config.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	if cfg.AuthToken != "" {
		config.AddDefaultHeader("Authorization", "Bearer "+cfg.AuthToken)
	}
	return client.NewAPIClient(config)
}
