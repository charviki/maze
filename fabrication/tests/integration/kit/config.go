package kit

import (
	"os"
)

// TestConfig 从环境变量读取集成测试配置，确保与生产环境隔离
type TestConfig struct {
	// ManagerURL 是 Manager API 的地址
	ManagerURL string
	// Env 是测试环境类型：docker 或 kubernetes
	Env string
	// Namespace 是 K8s 测试 namespace（仅 K8s 环境使用）
	Namespace string
	// DataDir 是测试数据根目录
	DataDir string
	// AuthToken 是 API 鉴权 token
	AuthToken string
}

// LoadTestConfig 从环境变量加载测试配置
func LoadTestConfig() *TestConfig {
	cfg := &TestConfig{
		ManagerURL: getEnv("MAZE_TEST_MANAGER_URL", "http://localhost:9090"),
		Env:        getEnv("MAZE_TEST_ENV", "docker"),
		Namespace:  getEnv("MAZE_TEST_NAMESPACE", "maze-test"),
		DataDir:    getEnv("MAZE_TEST_DATA_DIR", os.Getenv("HOME")+"/.maze-test"),
		AuthToken:  getEnv("MAZE_TEST_AUTH_TOKEN", "test-integration-token"),
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
