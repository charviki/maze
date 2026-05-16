-- name: CreateDoc :one
INSERT INTO docs (archive_id, parent_id, title, content, summary, status, priority, assignee, tags, author, visibility, shared_with, attachments)
VALUES (@archive_id, @parent_id, @title, @content, @summary, @status, @priority, @assignee, @tags, @author, @visibility, @shared_with, @attachments)
RETURNING *;

-- name: GetDoc :one
SELECT * FROM docs WHERE id = @id;

-- name: ListDocs :many
SELECT * FROM docs
WHERE ($1::uuid IS NULL OR archive_id = $1)
  AND ($2::uuid IS NULL OR parent_id = $2)
  AND ($3 = '' OR status = $3)
  AND ($4 = '' OR visibility = $4)
  AND ($5 = '' OR author = $5)
ORDER BY title ASC, created_at DESC
LIMIT $6 OFFSET $7;

-- name: ListDocsHasStatus :many
SELECT * FROM docs
WHERE ($1::uuid IS NULL OR archive_id = $1)
  AND ($2::uuid IS NULL OR parent_id = $2)
  AND status IS NOT NULL
  AND ($3 = '' OR status = $3)
  AND ($4 = '' OR visibility = $4)
  AND ($5 = '' OR author = $5)
ORDER BY title ASC, created_at DESC
LIMIT $6 OFFSET $7;

-- name: CountDocs :one
SELECT count(*) FROM docs
WHERE ($1::uuid IS NULL OR archive_id = $1)
  AND ($2::uuid IS NULL OR parent_id = $2)
  AND ($3 = '' OR status = $3)
  AND ($4 = '' OR visibility = $4)
  AND ($5 = '' OR author = $5);

-- name: CountDocsHasStatus :one
SELECT count(*) FROM docs
WHERE ($1::uuid IS NULL OR archive_id = $1)
  AND ($2::uuid IS NULL OR parent_id = $2)
  AND status IS NOT NULL
  AND ($3 = '' OR status = $3)
  AND ($4 = '' OR visibility = $4)
  AND ($5 = '' OR author = $5);

-- name: UpdateDoc :one
UPDATE docs
SET title       = COALESCE(sqlc.narg('title'), title),
    content     = COALESCE(sqlc.narg('content'), content),
    summary     = COALESCE(sqlc.narg('summary'), summary),
    status      = CASE WHEN sqlc.narg('clear_status')::boolean THEN NULL ELSE COALESCE(sqlc.narg('status'), status) END,
    priority    = CASE WHEN sqlc.narg('clear_priority')::boolean THEN NULL ELSE COALESCE(sqlc.narg('priority'), priority) END,
    assignee    = COALESCE(sqlc.narg('assignee'), assignee),
    tags        = COALESCE(sqlc.narg('tags'), tags),
    visibility  = COALESCE(sqlc.narg('visibility'), visibility),
    shared_with = COALESCE(sqlc.narg('shared_with'), shared_with),
    attachments = COALESCE(sqlc.narg('attachments'), attachments),
    parent_id   = COALESCE(sqlc.narg('parent_id'), parent_id),
    updated_at  = now()
WHERE id = @id
RETURNING *;

-- name: DeleteDoc :exec
DELETE FROM docs WHERE id = @id;

-- name: SearchDocs :many
SELECT * FROM docs
WHERE search_vector @@ plainto_tsquery('simple', @plainto_tsquery)
  AND (archive_id = @archive_id OR @archive_id IS NULL)
  AND (visibility = @visibility OR @visibility = '')
  AND (author = @author OR @author = '')
ORDER BY ts_rank(search_vector, plainto_tsquery('simple', @plainto_tsquery)) DESC;

-- name: GetDocChildren :many
SELECT * FROM docs
WHERE parent_id = @parent_id
ORDER BY title ASC;

-- name: GetDocRootChildren :many
SELECT * FROM docs
WHERE parent_id IS NULL
  AND archive_id = @archive_id
ORDER BY title ASC;
