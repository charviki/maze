-- name: CreateMCPServer :one
INSERT INTO mcp_servers (name, type, command, url, args, env)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetMCPServerByName :one
SELECT * FROM mcp_servers WHERE name = $1;

-- name: ListMCPServers :many
SELECT * FROM mcp_servers ORDER BY name ASC;

-- name: UpdateMCPServer :one
UPDATE mcp_servers
SET type = $2, command = $3, url = $4, args = $5, env = $6, updated_at = NOW()
WHERE name = $1
RETURNING *;

-- name: DeleteMCPServerByName :exec
DELETE FROM mcp_servers WHERE name = $1;
