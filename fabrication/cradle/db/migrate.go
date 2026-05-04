package db

import (
	"database/sql"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// RunMigrations 执行 Goose 数据库迁移。
// migrationsFS 由调用方传入（Cradle 不 embed 任何 SQL）。
func RunMigrations(pool *pgxpool.Pool, migrationsFS fs.FS) error {
	db := stdlib.OpenDBFromPool(pool)
	defer func() { _ = db.Close() }()

	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, ".")
}

// RollbackMigration 回滚最近一次迁移。
// 仅用于开发/测试环境。
func RollbackMigration(pool *pgxpool.Pool, migrationsFS fs.FS) error {
	db := stdlib.OpenDBFromPool(pool)
	defer func() { _ = db.Close() }()

	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Down(db, ".")
}

// MigrationStatus 返回当前迁移状态。
func MigrationStatus(pool *pgxpool.Pool, migrationsFS fs.FS) error {
	db := stdlib.OpenDBFromPool(pool)
	defer func() { _ = db.Close() }()

	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Status(db, ".")
}

// OpenDBFromPool 将 pgxpool.Pool 转换为 database/sql.DB（Goose 需要）。
func OpenDBFromPool(pool *pgxpool.Pool) *sql.DB {
	return stdlib.OpenDBFromPool(pool)
}
