-- +goose Up

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
    -- 对外暴露的稳定 ID 与内部聚簇主键分离，降低随机主键对写入局部性的影响。
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

CREATE INDEX idx_permission_requests_status_created_at
    ON auth.permission_requests(status, created_at DESC);

CREATE INDEX idx_permission_requests_subject_created_at
    ON auth.permission_requests(subject_key, created_at DESC);

CREATE INDEX idx_permission_grants_subject_status
    ON auth.permission_grants(subject_key, status, created_at DESC);

CREATE INDEX idx_permission_grants_expires_at
    ON auth.permission_grants(expires_at)
    WHERE expires_at IS NOT NULL AND status = 'active';

-- +goose Down
DROP INDEX IF EXISTS idx_permission_grants_expires_at;
DROP INDEX IF EXISTS idx_permission_grants_subject_status;
DROP INDEX IF EXISTS idx_permission_requests_subject_created_at;
DROP INDEX IF EXISTS idx_permission_requests_status_created_at;

DROP TABLE IF EXISTS auth.audit_logs;
DROP TABLE IF EXISTS auth.permission_grants;
DROP TABLE IF EXISTS auth.permission_requests;
DROP TABLE IF EXISTS auth.casbin_rules;
DROP TABLE IF EXISTS auth.subjects;

DROP TYPE IF EXISTS auth.permission_grant_status;
DROP TYPE IF EXISTS auth.permission_request_status;
DROP SCHEMA IF EXISTS auth;
