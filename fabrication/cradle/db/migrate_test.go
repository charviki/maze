package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type noopConnector struct{}

func (noopConnector) Connect(context.Context) (driver.Conn, error) { return noopConn{}, nil }

func (noopConnector) Driver() driver.Driver { return noopDriver{} }

type noopDriver struct{}

func (noopDriver) Open(string) (driver.Conn, error) { return noopConn{}, nil }

type noopConn struct{}

func (noopConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("not implemented") }

func (noopConn) Close() error { return nil }

func (noopConn) Begin() (driver.Tx, error) { return nil, errors.New("not implemented") }

func stubSQLDB() *sql.DB {
	return sql.OpenDB(noopConnector{})
}

func restoreMigrationHooks() {
	openDBFromPool = stdlibOpenDBFromPool
	gooseSetBaseFS = gooseSetBaseFSImpl
	gooseSetDialect = gooseSetDialectImpl
	gooseUp = gooseUpImpl
	gooseDown = gooseDownImpl
	gooseStatus = gooseStatusImpl
}

var (
	stdlibOpenDBFromPool = openDBFromPool
	gooseSetBaseFSImpl = gooseSetBaseFS
	gooseSetDialectImpl = gooseSetDialect
	gooseUpImpl = gooseUp
	gooseDownImpl = gooseDown
	gooseStatusImpl = gooseStatus
)

func TestRunMigrations_Success(t *testing.T) {
	t.Cleanup(restoreMigrationHooks)

	migrationsFS := fstest.MapFS{
		"0001_init.sql": &fstest.MapFile{Data: []byte("-- test")},
	}
	db := stubSQLDB()
	var gotFS fs.FS
	openDBFromPool = func(*pgxpool.Pool, ...stdlib.OptionOpenDB) *sql.DB { return db }
	gooseSetBaseFS = func(fsys fs.FS) { gotFS = fsys }
	gooseSetDialect = func(dialect string) error {
		if dialect != "postgres" {
			t.Fatalf("dialect = %q, want postgres", dialect)
		}
		return nil
	}
	upCalled := false
	gooseUp = func(gotDB *sql.DB, dir string, _ ...goose.OptionsFunc) error {
		upCalled = true
		if gotDB != db {
			t.Fatal("RunMigrations 未把转换后的 DB 传给 goose.Up")
		}
		if dir != "." {
			t.Fatalf("dir = %q, want .", dir)
		}
		return nil
	}

	if err := RunMigrations(nil, migrationsFS); err != nil {
		t.Fatalf("RunMigrations 返回错误: %v", err)
	}
	if _, err := fs.ReadFile(gotFS, "0001_init.sql"); err != nil {
		t.Fatalf("RunMigrations 未正确设置 migrations FS: %v", err)
	}
	if !upCalled {
		t.Fatal("RunMigrations 成功路径应调用 goose.Up")
	}
}

func TestRunMigrations_SetDialectError(t *testing.T) {
	t.Cleanup(restoreMigrationHooks)

	openDBFromPool = func(*pgxpool.Pool, ...stdlib.OptionOpenDB) *sql.DB { return stubSQLDB() }
	gooseSetBaseFS = func(fs.FS) {}
	gooseSetDialect = func(string) error { return errors.New("dialect failed") }

	err := RunMigrations(nil, fstest.MapFS{})
	if err == nil || !strings.Contains(err.Error(), "dialect failed") {
		t.Fatalf("err = %v, want dialect failure", err)
	}
}

func TestRunMigrations_UpError(t *testing.T) {
	t.Cleanup(restoreMigrationHooks)

	openDBFromPool = func(*pgxpool.Pool, ...stdlib.OptionOpenDB) *sql.DB { return stubSQLDB() }
	gooseSetBaseFS = func(fs.FS) {}
	gooseSetDialect = func(string) error { return nil }
	gooseUp = func(*sql.DB, string, ...goose.OptionsFunc) error { return errors.New("up failed") }

	err := RunMigrations(nil, fstest.MapFS{})
	if err == nil || !strings.Contains(err.Error(), "up failed") {
		t.Fatalf("err = %v, want goose.Up failure", err)
	}
}

func TestRollbackMigration_UsesGooseDown(t *testing.T) {
	t.Cleanup(restoreMigrationHooks)

	db := stubSQLDB()
	openDBFromPool = func(*pgxpool.Pool, ...stdlib.OptionOpenDB) *sql.DB { return db }
	gooseSetBaseFS = func(fs.FS) {}
	gooseSetDialect = func(string) error { return nil }
	downCalled := false
	gooseDown = func(gotDB *sql.DB, dir string, _ ...goose.OptionsFunc) error {
		downCalled = true
		if gotDB != db {
			t.Fatal("RollbackMigration 未把转换后的 DB 传给 goose.Down")
		}
		if dir != "." {
			t.Fatalf("dir = %q, want .", dir)
		}
		return nil
	}

	if err := RollbackMigration(nil, fstest.MapFS{}); err != nil {
		t.Fatalf("RollbackMigration 返回错误: %v", err)
	}
	if !downCalled {
		t.Fatal("RollbackMigration 应调用 goose.Down")
	}
}

func TestMigrationStatus_UsesGooseStatus(t *testing.T) {
	t.Cleanup(restoreMigrationHooks)

	db := stubSQLDB()
	openDBFromPool = func(*pgxpool.Pool, ...stdlib.OptionOpenDB) *sql.DB { return db }
	gooseSetBaseFS = func(fs.FS) {}
	gooseSetDialect = func(string) error { return nil }
	statusCalled := false
	gooseStatus = func(gotDB *sql.DB, dir string, _ ...goose.OptionsFunc) error {
		statusCalled = true
		if gotDB != db {
			t.Fatal("MigrationStatus 未把转换后的 DB 传给 goose.Status")
		}
		if dir != "." {
			t.Fatalf("dir = %q, want .", dir)
		}
		return nil
	}

	if err := MigrationStatus(nil, fstest.MapFS{}); err != nil {
		t.Fatalf("MigrationStatus 返回错误: %v", err)
	}
	if !statusCalled {
		t.Fatal("MigrationStatus 应调用 goose.Status")
	}
}

func TestOpenDBFromPool_DelegatesToStdlibHelper(t *testing.T) {
	t.Cleanup(restoreMigrationHooks)

	db := stubSQLDB()
	pool := new(pgxpool.Pool)
	openDBFromPool = func(gotPool *pgxpool.Pool, _ ...stdlib.OptionOpenDB) *sql.DB {
		if gotPool != pool {
			t.Fatal("OpenDBFromPool 未透传 pool")
		}
		return db
	}

	if got := OpenDBFromPool(pool); got != db {
		t.Fatal("OpenDBFromPool 应返回 helper 的结果")
	}
}

func TestNoopConnector_IsClosable(t *testing.T) {
	db := stubSQLDB()
	if err := db.Close(); err != nil {
		t.Fatalf("stub SQL DB 应可安全关闭: %v", err)
	}
}
