-- name: UpsertSubject :one
INSERT INTO auth.subjects (
    subject_key,
    subject_type,
    display_name,
    is_system,
    updated_at
) VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (subject_key) DO UPDATE
SET display_name = EXCLUDED.display_name,
    is_system = EXCLUDED.is_system,
    updated_at = NOW()
RETURNING subject_key, subject_type, display_name, is_system, created_at, updated_at;

-- name: GetSubject :one
SELECT subject_key, subject_type, display_name, is_system, created_at, updated_at
FROM auth.subjects
WHERE subject_key = $1;
