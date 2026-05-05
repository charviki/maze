-- name: InsertPermissionGrant :one
INSERT INTO auth.permission_grants (
    subject_key,
    resource,
    action,
    source_request_id,
    expires_at
) VALUES ($1, $2, $3, $4, $5)
RETURNING id, public_id, subject_key, resource, action, source_request_id, casbin_rule_id, status, revoked_by, revoked_reason, expires_at, created_at, updated_at;

-- name: AttachGrantCasbinRule :one
UPDATE auth.permission_grants
SET casbin_rule_id = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, public_id, subject_key, resource, action, source_request_id, casbin_rule_id, status, revoked_by, revoked_reason, expires_at, created_at, updated_at;

-- name: ListActiveSubjectPermissionGrants :many
SELECT
    sqlc.embed(g),
    pr.public_id AS source_request_public_id
FROM auth.permission_grants AS g
JOIN auth.permission_requests AS pr ON pr.id = g.source_request_id
WHERE g.subject_key = $1 AND g.status = 'active'
ORDER BY g.created_at DESC;

-- name: ListPermissionGrantsByRequest :many
SELECT id, public_id, subject_key, resource, action, source_request_id, casbin_rule_id, status, revoked_by, revoked_reason, expires_at, created_at, updated_at
FROM auth.permission_grants
WHERE source_request_id = $1
ORDER BY created_at ASC;

-- name: RevokeActivePermissionGrantsByRequest :many
UPDATE auth.permission_grants
SET status = $2,
    revoked_by = $3,
    revoked_reason = $4,
    updated_at = NOW()
WHERE source_request_id = $1 AND status = 'active'
RETURNING id, public_id, subject_key, resource, action, source_request_id, casbin_rule_id, status, revoked_by, revoked_reason, expires_at, created_at, updated_at;

-- name: ListExpiredPermissionGrants :many
SELECT
    sqlc.embed(g),
    pr.public_id AS source_request_public_id
FROM auth.permission_grants AS g
JOIN auth.permission_requests AS pr ON pr.id = g.source_request_id
WHERE g.status = 'active' AND g.expires_at IS NOT NULL AND g.expires_at <= NOW()
ORDER BY g.expires_at ASC;

-- name: ExpirePermissionGrant :one
UPDATE auth.permission_grants
SET status = 'expired',
    updated_at = NOW()
WHERE id = $1
RETURNING id, public_id, subject_key, resource, action, source_request_id, casbin_rule_id, status, revoked_by, revoked_reason, expires_at, created_at, updated_at;
