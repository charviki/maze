-- +goose Up
CREATE TABLE auth.users (
    id           BIGSERIAL PRIMARY KEY,
    username     TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    subject_key  TEXT NOT NULL UNIQUE REFERENCES auth.subjects(subject_key) ON DELETE RESTRICT,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_users_username ON auth.users(username);
CREATE INDEX idx_auth_users_subject_key ON auth.users(subject_key);

-- +goose Down
DROP INDEX IF EXISTS idx_auth_users_subject_key;
DROP INDEX IF EXISTS idx_auth_users_username;
DROP TABLE IF EXISTS auth.users;
