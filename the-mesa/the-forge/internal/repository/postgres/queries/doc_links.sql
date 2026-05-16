-- name: CreateLink :one
INSERT INTO doc_links (source_id, target_id, relation_type)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetLinksBySource :many
SELECT dl.*, d.title AS target_title
FROM doc_links dl
JOIN docs d ON d.id = dl.target_id
WHERE dl.source_id = $1
  AND (dl.relation_type = $2 OR $2 = '')
ORDER BY dl.created_at DESC;

-- name: GetLinksByTarget :many
SELECT dl.*, d.title AS source_title
FROM doc_links dl
JOIN docs d ON d.id = dl.source_id
WHERE dl.target_id = $1
  AND (dl.relation_type = $2 OR $2 = '')
ORDER BY dl.created_at DESC;

-- name: DeleteLink :exec
DELETE FROM doc_links WHERE id = $1;

-- name: GetLink :one
SELECT * FROM doc_links WHERE id = $1;
