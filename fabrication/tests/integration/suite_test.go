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

	env := kit.NewTestEnv(suite.cfg)
	if err := env.WaitForManager(15 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "manager not ready: %v\n", err)
		os.Exit(1)
	}

	if suite.cfg.EnableHostPool {
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
