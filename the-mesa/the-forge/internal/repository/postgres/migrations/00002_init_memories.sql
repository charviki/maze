-- +goose Up
CREATE TABLE memories (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    archive_id    UUID NOT NULL REFERENCES archives(id) ON DELETE CASCADE,
    parent_id     UUID REFERENCES memories(id) ON DELETE SET NULL,
    kind          TEXT NOT NULL DEFAULT 'doc',
    title         TEXT NOT NULL DEFAULT '',
    content       TEXT NOT NULL DEFAULT '',
    summary       TEXT NOT NULL DEFAULT '',
    type          TEXT NOT NULL DEFAULT 'shared',
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

CREATE INDEX idx_memories_archive_id ON memories(archive_id);
CREATE INDEX idx_memories_parent_id ON memories(parent_id);
CREATE INDEX idx_memories_search_vector ON memories USING GIN(search_vector);

CREATE TABLE document_links (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id     UUID NOT NULL REFERENCES memories(id) ON DELETE CASCADE,
    target_id     UUID NOT NULL REFERENCES memories(id) ON DELETE CASCADE,
    relation_type TEXT NOT NULL DEFAULT 'reference',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(source_id, target_id, relation_type)
);

CREATE INDEX idx_document_links_source_id ON document_links(source_id);
CREATE INDEX idx_document_links_target_id ON document_links(target_id);

-- +goose Down
DROP TABLE IF EXISTS document_links;
DROP TABLE IF EXISTS memories;
