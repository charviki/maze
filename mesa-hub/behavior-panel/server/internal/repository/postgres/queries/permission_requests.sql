-- name: InsertPermissionRequest :one
INSERT INTO auth.permission_requests (
    subject_key,
    targets_json,
    reason,
    expires_at
) VALUES ($1, $2, $3, $4)
RETURNING id, public_id, subject_key, targets_json, reason, status, reviewed_by, review_comment, reviewed_at, expires_at, created_at, updated_at;

-- name: GetPermissionRequestByPublicID :one
SELECT id, public_id, subject_key, targets_json, reason, status, reviewed_by, review_comment, reviewed_at, expires_at, created_at, updated_at
FROM auth.permission_requests
WHERE public_id = $1;

-- name: ListPermissionRequestsAll :many
SELECT id, public_id, subject_key, targets_json, reason, status, reviewed_by, review_comment, reviewed_at, expires_at, created_at, updated_at
FROM auth.permission_requests
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountPermissionRequestsAll :one
SELECT COUNT(*)
FROM auth.permission_requests;

-- name: ListPermissionRequestsByStatus :many
SELECT id, public_id, subject_key, targets_json, reason, status, reviewed_by, review_comment, reviewed_at, expires_at, created_at, updated_at
FROM auth.permission_requests
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountPermissionRequestsByStatus :one
SELECT COUNT(*)
FROM auth.permission_requests
WHERE status = $1;

-- name: ReviewPermissionRequest :one
UPDATE auth.permission_requests
SET status = sqlc.arg(status),
    reviewed_by = sqlc.narg(reviewed_by),
    review_comment = sqlc.narg(review_comment),
    reviewed_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND status = sqlc.arg(expected_status)
RETURNING id, public_id, subject_key, targets_json, reason, status, reviewed_by, review_comment, reviewed_at, expires_at, created_at, updated_at;

-- name: UpdatePermissionRequestStatus :one
UPDATE auth.permission_requests
SET status = $2,
    reviewed_by = $3,
    review_comment = $4,
    reviewed_at = CASE
        WHEN $3 IS NULL THEN reviewed_at
        ELSE NOW()
    END,
    updated_at = NOW()
WHERE id = $1
RETURNING id, public_id, subject_key, targets_json, reason, status, reviewed_by, review_comment, reviewed_at, expires_at, created_at, updated_at;
