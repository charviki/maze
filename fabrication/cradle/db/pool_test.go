package db

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPoolConfig_ConnString(t *testing.T) {
	tests := []struct {
		name     string
		cfg      PoolConfig
		wantHost string
		wantPort string
		wantDB   string
	}{
		{
			name:     "full config",
			cfg:      PoolConfig{Host: "localhost", Port: 5432, Name: "mydb", User: "admin", Password: "secret"},
			wantHost: "localhost",
			wantPort: "5432",
			wantDB:   "mydb",
		},
		{
			name:     "default port",
			cfg:      PoolConfig{Host: "db.example.com", Port: 0, Name: "testdb", User: "user", Password: "pass"},
			wantHost: "db.example.com",
			wantPort: "5432",
			wantDB:   "testdb",
		},
		{
			name:     "non-standard port",
			cfg:      PoolConfig{Host: "127.0.0.1", Port: 15432, Name: "maze", User: "postgres", Password: "postgres"},
			wantHost: "127.0.0.1",
			wantPort: "15432",
			wantDB:   "maze",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connStr := tt.cfg.ConnString()
			if connStr == "" {
				t.Fatal("ConnString returned empty string")
			}
			if !containsAll(connStr, tt.wantHost, tt.wantPort, tt.wantDB) {
				t.Errorf("ConnString = %q, expected to contain host=%q port=%q db=%q", connStr, tt.wantHost, tt.wantPort, tt.wantDB)
			}
			if !containsAll(connStr, "sslmode=disable") {
				t.Errorf("ConnString = %q, expected sslmode=disable", connStr)
			}
		})
	}
}

func TestPoolConfig_ConnString_Format(t *testing.T) {
	cfg := PoolConfig{Host: "pg", Port: 5432, Name: "db", User: "u", Password: "p"}
	got := cfg.ConnString()
	expected := "postgresql://u:p@pg:5432/db?sslmode=disable"
	if got != expected {
		t.Errorf("ConnString = %q, want %q", got, expected)
	}
}

func containsAll(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(s) < len(sub) {
			return false
		}
		found := false
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestNewPool_ParseConfigError(t *testing.T) {
	origParsePoolConfig := parsePoolConfig
	t.Cleanup(func() {
		parsePoolConfig = origParsePoolConfig
	})

	parsePoolConfig = func(connString string) (*pgxpool.Config, error) {
		if !strings.Contains(connString, "db:5432") {
			t.Fatalf("ConnString 未传递给 ParseConfig: %q", connString)
		}
		return nil, errors.New("bad config")
	}

	_, err := NewPool(context.Background(), PoolConfig{
		Host: "db",
		Name: "maze",
		User: "user",
	})
	if err == nil || !strings.Contains(err.Error(), "parse db config: bad config") {
		t.Fatalf("err = %v, want parse db config wrapper", err)
	}
}

func TestNewPool_CreatePoolError(t *testing.T) {
	origParsePoolConfig := parsePoolConfig
	origNewPoolWithConfig := newPoolWithConfig
	t.Cleanup(func() {
		parsePoolConfig = origParsePoolConfig
		newPoolWithConfig = origNewPoolWithConfig
	})

	cfg := &pgxpool.Config{}
	parsePoolConfig = func(string) (*pgxpool.Config, error) {
		return cfg, nil
	}
	newPoolWithConfig = func(ctx context.Context, gotCfg *pgxpool.Config) (*pgxpool.Pool, error) {
		if gotCfg != cfg {
			t.Fatal("NewPool 未把解析后的配置传给 NewWithConfig")
		}
		return nil, errors.New("dial failed")
	}

	_, err := NewPool(context.Background(), PoolConfig{})
	if err == nil || !strings.Contains(err.Error(), "create db pool: dial failed") {
		t.Fatalf("err = %v, want create db pool wrapper", err)
	}
}

func TestNewPool_PingErrorClosesPool(t *testing.T) {
	origParsePoolConfig := parsePoolConfig
	origNewPoolWithConfig := newPoolWithConfig
	origPingPool := pingPool
	origClosePool := closePool
	t.Cleanup(func() {
		parsePoolConfig = origParsePoolConfig
		newPoolWithConfig = origNewPoolWithConfig
		pingPool = origPingPool
		closePool = origClosePool
	})

	parsePoolConfig = func(string) (*pgxpool.Config, error) {
		return &pgxpool.Config{}, nil
	}

	pool := new(pgxpool.Pool)
	newPoolWithConfig = func(context.Context, *pgxpool.Config) (*pgxpool.Pool, error) {
		return pool, nil
	}

	closed := false
	pingPool = func(context.Context, *pgxpool.Pool) error {
		return errors.New("ping failed")
	}
	closePool = func(gotPool *pgxpool.Pool) {
		if gotPool != pool {
			t.Fatal("closePool 收到的 pool 不正确")
		}
		closed = true
	}

	_, err := NewPool(context.Background(), PoolConfig{})
	if err == nil || !strings.Contains(err.Error(), "ping db: ping failed") {
		t.Fatalf("err = %v, want ping db wrapper", err)
	}
	if !closed {
		t.Fatal("ping 失败后应关闭连接池")
	}
}

func TestNewPool_Success(t *testing.T) {
	origParsePoolConfig := parsePoolConfig
	origNewPoolWithConfig := newPoolWithConfig
	origPingPool := pingPool
	t.Cleanup(func() {
		parsePoolConfig = origParsePoolConfig
		newPoolWithConfig = origNewPoolWithConfig
		pingPool = origPingPool
	})

	parsePoolConfig = func(string) (*pgxpool.Config, error) {
		return &pgxpool.Config{}, nil
	}

	pool := new(pgxpool.Pool)
	newPoolWithConfig = func(context.Context, *pgxpool.Config) (*pgxpool.Pool, error) {
		return pool, nil
	}
	pingCalled := false
	pingPool = func(context.Context, *pgxpool.Pool) error {
		pingCalled = true
		return nil
	}

	got, err := NewPool(context.Background(), PoolConfig{})
	if err != nil {
		t.Fatalf("NewPool 返回错误: %v", err)
	}
	if got != pool {
		t.Fatal("NewPool 应返回创建出的 pool")
	}
	if !pingCalled {
		t.Fatal("NewPool 成功路径应执行 Ping")
	}
}
