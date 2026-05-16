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
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE doc_links (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id     UUID NOT NULL REFERENCES docs(id) ON DELETE CASCADE,
    target_id     UUID NOT NULL REFERENCES docs(id) ON DELETE CASCADE,
    relation_type TEXT NOT NULL DEFAULT 'reference',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(source_id, target_id, relation_type)
);
