-- name: CreateDirective :one
INSERT INTO directives (title, description, status, priority, assignee, author, require_doc_ids, narrative_id, archive_id, visibility)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetDirective :one
SELECT * FROM directives WHERE id = $1;

-- name: ListDirectives :many
SELECT * FROM directives
WHERE (status = $1 OR $1 = '')
  AND (assignee = $2 OR $2 = '')
  AND (priority = $3 OR $3 = '')
  AND (visibility = $4 OR $4 = '')
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountDirectives :one
SELECT count(*) FROM directives
WHERE (status = $1 OR $1 = '')
  AND (assignee = $2 OR $2 = '')
  AND (priority = $3 OR $3 = '')
  AND (visibility = $4 OR $4 = '');

-- name: UpdateDirective :one
UPDATE directives
SET title           = COALESCE($2, title),
    description     = COALESCE($3, description),
    status          = COALESCE($4, status),
    priority        = COALESCE($5, priority),
    assignee        = COALESCE($6, assignee),
    author          = COALESCE($7, author),
    require_doc_ids = COALESCE($8, require_doc_ids),
    narrative_id    = COALESCE($9, narrative_id),
    archive_id      = COALESCE($10, archive_id),
    visibility      = COALESCE($11, visibility),
    updated_at      = now()
WHERE id = $1
RETURNING *;

-- name: DeleteDirective :exec
DELETE FROM directives WHERE id = $1;

-- name: ListDirectivesByDocID :many
SELECT * FROM directives
WHERE require_doc_ids @> to_jsonb($1::text)
ORDER BY created_at DESC;
