package postgres

import "embed"

// MigrationsFS 包含 The Forge 数据库的 Goose 迁移 SQL 文件。
//
//go:embed migrations/00001_*.sql migrations/00002_*.sql migrations/00003_*.sql migrations/00004_*.sql
var MigrationsFS embed.FS
