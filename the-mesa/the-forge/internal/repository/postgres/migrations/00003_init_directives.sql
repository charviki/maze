-- +goose Up
CREATE TABLE directives (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title           TEXT NOT NULL DEFAULT '',
    description     TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'pending',
    priority        TEXT NOT NULL DEFAULT 'normal',
    assignee        TEXT NOT NULL DEFAULT '',
    author          TEXT NOT NULL DEFAULT '',
    require_doc_ids JSONB NOT NULL DEFAULT '[]',
    narrative_id    TEXT NOT NULL DEFAULT '',
    archive_id      UUID REFERENCES archives(id) ON DELETE SET NULL,
    visibility      TEXT NOT NULL DEFAULT 'public',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_directives_archive_id ON directives(archive_id);
CREATE INDEX idx_directives_status ON directives(status);

-- +goose Down
DROP TABLE IF EXISTS directives;
