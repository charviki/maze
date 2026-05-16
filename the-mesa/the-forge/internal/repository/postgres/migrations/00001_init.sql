-- +goose Up
CREATE TABLE archives (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    icon        TEXT NOT NULL DEFAULT '',
    author      TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE docs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    archive_id    UUID NOT NULL REFERENCES archives(id) ON DELETE CASCADE,
    parent_id     UUID REFERENCES docs(id) ON DELETE SET NULL,
    kind          TEXT NOT NULL DEFAULT 'doc',
    title         TEXT NOT NULL DEFAULT '',
    content       TEXT NOT NULL DEFAULT '',
    summary       TEXT NOT NULL DEFAULT '',
    status        TEXT DEFAULT NULL,
    priority      TEXT DEFAULT NULL,
    assignee      TEXT NOT NULL DEFAULT '',
    tags          JSONB NOT NULL DEFAULT '[]',
    author        TEXT NOT NULL DEFAULT '',
    visibility    TEXT NOT NULL DEFAULT 'public',
    shared_with   JSONB NOT NULL DEFAULT '[]',
    attachments   JSONB NOT NULL DEFAULT '[]',
    search_vector TSVECTOR GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(content, '')), 'B')
    ) STORED,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_docs_archive_id    ON docs(archive_id);
CREATE INDEX idx_docs_parent_id     ON docs(parent_id);
CREATE INDEX idx_docs_status        ON docs(status);
CREATE INDEX idx_docs_search_vector ON docs USING GIN(search_vector);

CREATE TABLE doc_links (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id     UUID NOT NULL REFERENCES docs(id) ON DELETE CASCADE,
    target_id     UUID NOT NULL REFERENCES docs(id) ON DELETE CASCADE,
    relation_type TEXT NOT NULL DEFAULT 'reference',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(source_id, target_id, relation_type)
);

CREATE INDEX idx_doc_links_source_id ON doc_links(source_id);
CREATE INDEX idx_doc_links_target_id ON doc_links(target_id);

-- +goose Down
DROP TABLE IF EXISTS doc_links;
DROP TABLE IF EXISTS docs;
DROP TABLE IF EXISTS archives;
