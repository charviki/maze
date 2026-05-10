-- schema.sql 声明 sqlc 需要的类型和表结构（从 migrations 精简提取）。

CREATE TABLE archives (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    icon        TEXT NOT NULL DEFAULT '',
    author      TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

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
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE document_links (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id     UUID NOT NULL REFERENCES memories(id) ON DELETE CASCADE,
    target_id     UUID NOT NULL REFERENCES memories(id) ON DELETE CASCADE,
    relation_type TEXT NOT NULL DEFAULT 'reference',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(source_id, target_id, relation_type)
);

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

CREATE TABLE chat_history (
    id         SERIAL PRIMARY KEY,
    role       TEXT NOT NULL,
    content    TEXT NOT NULL DEFAULT '',
    tool_calls JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
