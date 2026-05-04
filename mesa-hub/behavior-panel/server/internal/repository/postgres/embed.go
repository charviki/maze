package postgres

import "embed"

// MigrationsFS 包含 Goose 迁移 SQL 文件。
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS
