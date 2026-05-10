-- name: CreateArchive :one
INSERT INTO archives (name, description, icon, author)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetArchive :one
SELECT * FROM archives WHERE id = $1;

-- name: ListArchives :many
SELECT * FROM archives ORDER BY created_at DESC;

-- name: UpdateArchive :one
UPDATE archives
SET name = $2, description = $3, icon = $4, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteArchive :exec
DELETE FROM archives WHERE id = $1;
