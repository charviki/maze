-- name: CreateGitKey :one
INSERT INTO git_keys (name, encrypted_token, token_mask)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetGitKeyByName :one
SELECT * FROM git_keys WHERE name = $1;

-- name: ListGitKeys :many
SELECT * FROM git_keys ORDER BY name ASC;

-- name: DeleteGitKeyByName :exec
DELETE FROM git_keys WHERE name = $1;
