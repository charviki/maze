package postgres

import "embed"

// AuthMigrationsFS 包含权限数据库 (maze_auth) 的 Goose 迁移 SQL 文件。
//
//go:embed migrations/00001_*.sql migrations/00002_*.sql
var AuthMigrationsFS embed.FS

// HostMigrationsFS 包含 Host 数据库 (maze_host) 的 Goose 迁移 SQL 文件。
//
//go:embed migrations/00003_*.sql
var HostMigrationsFS embed.FS
