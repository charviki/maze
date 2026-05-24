-- name: CreateGitKey :one
INSERT INTO git_keys (name, encrypted_token, token_mask, token_type, host)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetGitKeyByName :one
SELECT * FROM git_keys WHERE name = $1;

-- name: ListGitKeys :many
SELECT * FROM git_keys ORDER BY name ASC;

-- name: UpdateGitKeyByName :one
UPDATE git_keys
SET encrypted_token = COALESCE($2, encrypted_token),
    token_mask = COALESCE($3, token_mask),
    token_type = COALESCE($4, token_type),
    host = COALESCE($5, host),
    updated_at = NOW()
WHERE name = $1
RETURNING *;

-- name: DeleteGitKeyByName :exec
DELETE FROM git_keys WHERE name = $1;

-- name: GetGitKeysByNames :many
SELECT * FROM git_keys WHERE name = ANY($1::text[]);
