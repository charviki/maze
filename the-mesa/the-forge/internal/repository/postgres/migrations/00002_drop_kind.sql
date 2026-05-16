-- +goose Up
ALTER TABLE docs DROP COLUMN kind;

-- +goose Down
ALTER TABLE docs ADD COLUMN kind TEXT NOT NULL DEFAULT 'doc';
