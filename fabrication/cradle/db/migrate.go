package db

import (
	"database/sql"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

var (
	openDBFromPool = stdlib.OpenDBFromPool
	gooseSetBaseFS = goose.SetBaseFS
	gooseSetDialect = goose.SetDialect
	gooseUp = goose.Up
	gooseDown = goose.Down
	gooseStatus = goose.Status
)

// RunMigrations 执行 Goose 数据库迁移。
// migrationsFS 由调用方传入（Cradle 不 embed 任何 SQL）。
func RunMigrations(pool *pgxpool.Pool, migrationsFS fs.FS) error {
	db := openDBFromPool(pool)
	defer func() { _ = db.Close() }()

	gooseSetBaseFS(migrationsFS)
	if err := gooseSetDialect("postgres"); err != nil {
		return err
	}
	return gooseUp(db, ".")
}

// RollbackMigration 回滚最近一次迁移。
// 仅用于开发/测试环境。
func RollbackMigration(pool *pgxpool.Pool, migrationsFS fs.FS) error {
	db := openDBFromPool(pool)
	defer func() { _ = db.Close() }()

	gooseSetBaseFS(migrationsFS)
	if err := gooseSetDialect("postgres"); err != nil {
		return err
	}
	return gooseDown(db, ".")
}

// MigrationStatus 返回当前迁移状态。
func MigrationStatus(pool *pgxpool.Pool, migrationsFS fs.FS) error {
	db := openDBFromPool(pool)
	defer func() { _ = db.Close() }()

	gooseSetBaseFS(migrationsFS)
	if err := gooseSetDialect("postgres"); err != nil {
		return err
	}
	return gooseStatus(db, ".")
}

// OpenDBFromPool 将 pgxpool.Pool 转换为 database/sql.DB（Goose 需要）。
func OpenDBFromPool(pool *pgxpool.Pool) *sql.DB {
	return openDBFromPool(pool)
}
