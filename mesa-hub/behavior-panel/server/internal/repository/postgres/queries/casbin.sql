-- name: InsertCasbinRule :one
INSERT INTO auth.casbin_rules (ptype, v0, v1, v2)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: DeleteCasbinRule :exec
DELETE FROM auth.casbin_rules
WHERE id = $1;

-- name: ListAllCasbinRules :many
SELECT id, ptype, v0, v1, v2, v3, v4, v5
FROM auth.casbin_rules
ORDER BY id ASC;
