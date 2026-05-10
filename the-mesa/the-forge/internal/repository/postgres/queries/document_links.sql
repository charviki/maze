-- name: CreateLink :one
INSERT INTO document_links (source_id, target_id, relation_type)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetLinksBySource :many
SELECT dl.*, m.title AS target_title
FROM document_links dl
JOIN memories m ON m.id = dl.target_id
WHERE dl.source_id = $1
  AND (dl.relation_type = $2 OR $2 = '')
ORDER BY dl.created_at DESC;

-- name: GetLinksByTarget :many
SELECT dl.*, m.title AS source_title
FROM document_links dl
JOIN memories m ON m.id = dl.source_id
WHERE dl.target_id = $1
  AND (dl.relation_type = $2 OR $2 = '')
ORDER BY dl.created_at DESC;

-- name: DeleteLink :exec
DELETE FROM document_links WHERE id = $1;

-- name: GetLink :one
SELECT * FROM document_links WHERE id = $1;
