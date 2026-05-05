-- +goose Up
CREATE TABLE host_specs (
    id           BIGSERIAL PRIMARY KEY,
    name         TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL DEFAULT '',
    tools        JSONB NOT NULL DEFAULT '[]'::jsonb,
    resources    JSONB NOT NULL DEFAULT '{}'::jsonb,
    auth_token   TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending',
    error_msg    TEXT NOT NULL DEFAULT '',
    retry_count  INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_host_specs_status ON host_specs(status);

CREATE TABLE nodes (
    id             BIGSERIAL PRIMARY KEY,
    name           TEXT NOT NULL UNIQUE,
    address        TEXT NOT NULL DEFAULT '',
    external_addr  TEXT NOT NULL DEFAULT '',
    grpc_address   TEXT NOT NULL DEFAULT '',
    status         TEXT NOT NULL DEFAULT 'online',
    capabilities   JSONB NOT NULL DEFAULT '{}'::jsonb,
    agent_status   JSONB NOT NULL DEFAULT '{}'::jsonb,
    metadata       JSONB NOT NULL DEFAULT '{}'::jsonb,
    registered_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_heartbeat TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_nodes_status ON nodes(status);
CREATE INDEX idx_nodes_last_heartbeat ON nodes(last_heartbeat);

CREATE TABLE host_tokens (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    token      TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE audit_entries (
    id              BIGSERIAL PRIMARY KEY,
    audit_id        TEXT NOT NULL UNIQUE,
    timestamp       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    operator        TEXT NOT NULL,
    target_node     TEXT NOT NULL DEFAULT '',
    action          TEXT NOT NULL,
    payload_summary TEXT NOT NULL DEFAULT '',
    result          TEXT NOT NULL,
    status_code     INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_audit_entries_timestamp ON audit_entries(timestamp DESC);
CREATE INDEX idx_audit_entries_action ON audit_entries(action);
CREATE INDEX idx_audit_entries_target_node ON audit_entries(target_node);

-- +goose Down
DROP TABLE IF EXISTS audit_entries;
DROP TABLE IF EXISTS host_tokens;
DROP TABLE IF EXISTS nodes;
DROP TABLE IF EXISTS host_specs;
