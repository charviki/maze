package postgres

import "embed"

// MigrationsFS 包含 The Forge 数据库的 Goose 迁移 SQL 文件。
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS
