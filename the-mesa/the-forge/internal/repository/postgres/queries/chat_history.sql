-- name: CreateChatMessage :one
INSERT INTO chat_history (role, content, tool_calls)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListChatHistory :many
SELECT * FROM chat_history ORDER BY created_at ASC;

-- name: ClearChatHistory :exec
DELETE FROM chat_history;
