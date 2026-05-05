CREATE SCHEMA IF NOT EXISTS auth;

CREATE TYPE auth.permission_request_status AS ENUM (
    'pending',
    'approved',
    'denied',
    'revoked',
    'expired'
);

CREATE TYPE auth.permission_grant_status AS ENUM (
    'active',
    'revoked',
    'expired'
);

CREATE TABLE auth.subjects (
    subject_key  TEXT PRIMARY KEY,
    subject_type TEXT NOT NULL,
    display_name TEXT NOT NULL,
    is_system    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE auth.casbin_rules (
    id    BIGSERIAL PRIMARY KEY,
    ptype TEXT NOT NULL,
    v0    TEXT,
    v1    TEXT,
    v2    TEXT,
    v3    TEXT,
    v4    TEXT,
    v5    TEXT
);

CREATE TABLE auth.permission_requests (
    id             BIGSERIAL PRIMARY KEY,
    -- 对外暴露的稳定 ID 与内部聚簇主键分离，避免 UUID 主键带来的写入局部性问题。
    public_id      TEXT GENERATED ALWAYS AS ('pa_' || id::text) STORED UNIQUE,
    subject_key    TEXT NOT NULL REFERENCES auth.subjects(subject_key) ON DELETE RESTRICT,
    targets_json   JSONB NOT NULL,
    reason         TEXT NOT NULL,
    status         auth.permission_request_status NOT NULL DEFAULT 'pending',
    reviewed_by    TEXT REFERENCES auth.subjects(subject_key) ON DELETE SET NULL,
    review_comment TEXT,
    reviewed_at    TIMESTAMPTZ,
    expires_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE auth.permission_grants (
    id                BIGSERIAL PRIMARY KEY,
    public_id         TEXT GENERATED ALWAYS AS ('pg_' || id::text) STORED UNIQUE,
    subject_key       TEXT NOT NULL REFERENCES auth.subjects(subject_key) ON DELETE RESTRICT,
    resource          TEXT NOT NULL,
    action            TEXT NOT NULL,
    source_request_id BIGINT NOT NULL REFERENCES auth.permission_requests(id) ON DELETE CASCADE,
    casbin_rule_id    BIGINT UNIQUE REFERENCES auth.casbin_rules(id) ON DELETE SET NULL,
    status            auth.permission_grant_status NOT NULL DEFAULT 'active',
    revoked_by        TEXT REFERENCES auth.subjects(subject_key) ON DELETE SET NULL,
    revoked_reason    TEXT,
    expires_at        TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE auth.audit_logs (
    id                 BIGSERIAL PRIMARY KEY,
    action             TEXT NOT NULL,
    actor_subject_key  TEXT REFERENCES auth.subjects(subject_key) ON DELETE SET NULL,
    target_subject_key TEXT REFERENCES auth.subjects(subject_key) ON DELETE SET NULL,
    request_id         BIGINT REFERENCES auth.permission_requests(id) ON DELETE SET NULL,
    details_json       JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
