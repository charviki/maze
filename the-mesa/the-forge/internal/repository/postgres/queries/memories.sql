-- name: CreateMemory :one
INSERT INTO memories (archive_id, parent_id, kind, title, content, summary, type, tags, author, visibility, shared_with, attachments)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetMemory :one
SELECT * FROM memories WHERE id = $1;

-- name: ListMemories :many
SELECT * FROM memories
WHERE (archive_id = $1 OR $1 IS NULL)
  AND (parent_id = $2 OR ($2 IS NULL AND parent_id IS NULL))
  AND (kind = $3 OR $3 = '')
  AND (type = $4 OR $4 = '')
  AND (visibility = $5 OR $5 = '')
  AND (author = $6 OR $6 = '')
ORDER BY created_at DESC
LIMIT $7 OFFSET $8;

-- name: CountMemories :one
SELECT count(*) FROM memories
WHERE (archive_id = $1 OR $1 IS NULL)
  AND (parent_id = $2 OR ($2 IS NULL AND parent_id IS NULL))
  AND (kind = $3 OR $3 = '')
  AND (type = $4 OR $4 = '')
  AND (visibility = $5 OR $5 = '')
  AND (author = $6 OR $6 = '');

-- name: UpdateMemory :one
UPDATE memories
SET title       = COALESCE($2, title),
    content     = COALESCE($3, content),
    summary     = COALESCE($4, summary),
    type        = COALESCE($5, type),
    tags        = COALESCE($6, tags),
    visibility  = COALESCE($7, visibility),
    shared_with = COALESCE($8, shared_with),
    attachments = COALESCE($9, attachments),
    parent_id   = COALESCE($10, parent_id),
    kind        = COALESCE($11, kind),
    updated_at  = now()
WHERE id = $1
RETURNING *;

-- name: DeleteMemory :exec
DELETE FROM memories WHERE id = $1;

-- name: SearchMemories :many
SELECT * FROM memories
WHERE search_vector @@ plainto_tsquery('simple', $1)
  AND (archive_id = $2 OR $2 IS NULL)
  AND (visibility = $3 OR $3 = '')
  AND (author = $4 OR $4 = '')
ORDER BY ts_rank(search_vector, plainto_tsquery('simple', $1)) DESC;

-- name: GetMemoryChildren :many
SELECT * FROM memories
WHERE parent_id = $1
ORDER BY kind ASC, title ASC;

-- name: GetMemoryRootChildren :many
SELECT * FROM memories
WHERE parent_id IS NULL
  AND archive_id = $1
ORDER BY kind ASC, title ASC;

-- Ancestors 查询使用递归 CTE，sqlc 无法解析，在 repository 层手写实现。
