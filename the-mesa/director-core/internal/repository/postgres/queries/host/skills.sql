-- name: CreateSkill :one
INSERT INTO skills (name, description, config)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSkillByName :one
SELECT * FROM skills WHERE name = $1;

-- name: ListSkills :many
SELECT * FROM skills ORDER BY name ASC;

-- name: UpdateSkill :one
UPDATE skills
SET description = $2, config = $3, updated_at = NOW()
WHERE name = $1
RETURNING *;

-- name: DeleteSkillByName :exec
DELETE FROM skills WHERE name = $1;
