-- name: CreateRule :one
INSERT INTO rules (name, content)
VALUES ($1, $2)
RETURNING *;

-- name: GetRuleByName :one
SELECT * FROM rules WHERE name = $1;

-- name: ListRules :many
SELECT * FROM rules ORDER BY name ASC;

-- name: UpdateRule :one
UPDATE rules
SET content = $2, updated_at = NOW()
WHERE name = $1
RETURNING *;

-- name: DeleteRuleByName :exec
DELETE FROM rules WHERE name = $1;
