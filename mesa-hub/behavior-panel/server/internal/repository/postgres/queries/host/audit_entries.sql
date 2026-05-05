-- name: InsertAuditEntry :exec
INSERT INTO audit_entries (audit_id, timestamp, operator, target_node, action, payload_summary, result, status_code)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: ListAuditEntries :many
SELECT * FROM audit_entries ORDER BY timestamp DESC;

-- name: ListAuditEntriesPage :many
SELECT * FROM audit_entries
ORDER BY timestamp DESC
LIMIT $1 OFFSET $2;

-- name: CountAuditEntries :one
SELECT COUNT(*) FROM audit_entries;

-- name: QueryAuditEntries :many
SELECT * FROM audit_entries
WHERE ($1::text = '' OR target_node LIKE '%' || $1 || '%')
  AND ($2::text = '' OR action LIKE '%' || $2 || '%')
ORDER BY timestamp DESC;
