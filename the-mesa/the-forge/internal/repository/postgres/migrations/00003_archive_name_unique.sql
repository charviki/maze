-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS idx_archives_name ON archives(name);

-- +goose Down
DROP INDEX IF EXISTS idx_archives_name;
