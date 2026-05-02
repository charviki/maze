package kit

import (
	"net/http"
	"os"
	"time"

	client "github.com/charviki/maze-cradle/api/gen/http"
)

// TestConfig 集成测试运行参数，从环境变量加载。
type TestConfig struct {
	ManagerURL          string
	Env                 string
	Namespace           string
	DataDir             string
	AuthToken           string
	AgentStorageBackend string
}

// LoadTestConfig 从环境变量加载集成测试配置。
func LoadTestConfig() *TestConfig {
	cfg := &TestConfig{
		ManagerURL:          getEnv("MAZE_TEST_MANAGER_URL", "http://localhost:9091"),
		Env:                 getEnv("MAZE_TEST_ENV", "docker"),
		Namespace:           getEnv("MAZE_TEST_NAMESPACE", "maze-test"),
		DataDir:             getEnv("MAZE_TEST_DATA_DIR", os.Getenv("HOME")+"/.maze-test"),
		AuthToken:           getEnv("MAZE_TEST_AUTH_TOKEN", "test-integration-token"),
		AgentStorageBackend: getEnv("MAZE_TEST_AGENT_STORAGE_BACKEND", defaultAgentStorageBackend(getEnv("MAZE_TEST_ENV", "docker"))),
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

// NewTestAPIClient 创建指向 gRPC-gateway (:9091) 的 OpenAPI 生成 client
func NewTestAPIClient(cfg *TestConfig) *client.APIClient {
	config := client.NewConfiguration()
	config.Servers = client.ServerConfigurations{
		{URL: cfg.ManagerURL},
	}
	config.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	if cfg.AuthToken != "" {
		config.AddDefaultHeader("Authorization", "Bearer "+cfg.AuthToken)
	}
	return client.NewAPIClient(config)
}
