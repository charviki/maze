//go:build integration

package integration

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/charviki/maze-integration-tests/kit"
)

type suiteState struct {
	cfg  *kit.TestConfig
	pool *hostPool
}

var suite suiteState

func TestMain(m *testing.M) {
	suite.cfg = kit.LoadTestConfig()

	if suite.cfg.EnableHostPool {
		env := kit.NewTestEnv(suite.cfg)
		if err := env.WaitForManager(15 * time.Second); err != nil {
			// 共享池开启意味着本轮测试依赖“固定容量 + 受控并发”语义；
			// 这里若继续静默降级，会重新退回每个测试各自建 Host 的旧模式。
			fmt.Fprintf(os.Stderr, "host pool enabled but manager not ready: %v\n", err)
			os.Exit(1)
		}

		pool, err := newHostPool(suite.cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "init host pool: %v\n", err)
			os.Exit(1)
		}
		if err := pool.Warmup(); err != nil {
			_ = pool.Cleanup()
			fmt.Fprintf(os.Stderr, "warmup host pool: %v\n", err)
			os.Exit(1)
		}
		suite.pool = pool
	}

	code := m.Run()

	if suite.pool != nil {
		if err := suite.pool.Cleanup(); err != nil {
			fmt.Fprintf(os.Stderr, "cleanup host pool: %v\n", err)
			if code == 0 {
				code = 1
			}
		}
	}

	os.Exit(code)
}
