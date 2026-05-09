-- +goose Up
CREATE TYPE auth.credential_type AS ENUM (
    'user_refresh',
    'host_token'
);

CREATE TYPE auth.credential_status AS ENUM (
    'active',
    'revoked',
    'expired'
);

CREATE TABLE auth.credentials (
    id           BIGSERIAL PRIMARY KEY,
    type         auth.credential_type NOT NULL,
    token_hash   TEXT NOT NULL,
    subject_key  TEXT NOT NULL REFERENCES auth.subjects(subject_key) ON DELETE CASCADE,
    expires_at   TIMESTAMPTZ,
    status       auth.credential_status NOT NULL DEFAULT 'active',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at   TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_auth_credentials_token_hash ON auth.credentials(token_hash);
CREATE INDEX idx_auth_credentials_subject_status ON auth.credentials(subject_key, status);

-- +goose Down
DROP INDEX IF EXISTS idx_auth_credentials_subject_status;
DROP INDEX IF EXISTS idx_auth_credentials_token_hash;
DROP TABLE IF EXISTS auth.credentials;
DROP TYPE IF EXISTS auth.credential_status;
DROP TYPE IF EXISTS auth.credential_type;
