-- name: InsertHostSpec :execrows
INSERT INTO host_specs (name, display_name, tools, resources, auth_token, status)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (name) DO NOTHING;

-- name: GetHostSpecByName :one
SELECT * FROM host_specs WHERE name = $1;

-- name: ListHostSpecs :many
SELECT * FROM host_specs ORDER BY name ASC;

-- name: UpdateHostSpecStatus :one
UPDATE host_specs
SET status = $2, error_msg = $3, updated_at = NOW()
WHERE name = $1
RETURNING *;

-- name: DeleteHostSpecByName :exec
DELETE FROM host_specs WHERE name = $1;

-- name: IncrementHostSpecRetry :one
UPDATE host_specs
SET retry_count = retry_count + 1, updated_at = NOW()
WHERE name = $1
RETURNING *;
