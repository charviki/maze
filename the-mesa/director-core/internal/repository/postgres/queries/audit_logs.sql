-- name: InsertAuditLog :exec
INSERT INTO auth.audit_logs (
    action,
    actor_subject_key,
    target_subject_key,
    request_id,
    details_json
) VALUES ($1, $2, $3, $4, $5);
