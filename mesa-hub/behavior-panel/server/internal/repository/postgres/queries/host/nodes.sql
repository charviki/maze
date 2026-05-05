-- name: UpsertNode :one
INSERT INTO nodes (name, address, external_addr, grpc_address, capabilities, agent_status, metadata, status, registered_at, last_heartbeat)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'online', NOW(), NOW())
ON CONFLICT (name) DO UPDATE SET
    address = EXCLUDED.address,
    external_addr = EXCLUDED.external_addr,
    grpc_address = EXCLUDED.grpc_address,
    capabilities = EXCLUDED.capabilities,
    agent_status = EXCLUDED.agent_status,
    metadata = EXCLUDED.metadata,
    status = 'online',
    registered_at = NOW(),
    last_heartbeat = NOW()
RETURNING *;

-- name: UpdateNodeHeartbeat :one
UPDATE nodes
SET agent_status = $2,
    last_heartbeat = NOW(),
    status = 'online'
WHERE name = $1
RETURNING *;

-- name: GetNodeByName :one
SELECT * FROM nodes WHERE name = $1;

-- name: ListNodes :many
SELECT * FROM nodes ORDER BY name ASC;

-- name: DeleteNodeByName :exec
DELETE FROM nodes WHERE name = $1;

-- name: CountNodes :one
SELECT COUNT(*) FROM nodes;

-- name: CountOnlineNodes :one
SELECT COUNT(*) FROM nodes
WHERE last_heartbeat > NOW() - INTERVAL '30 seconds';

-- name: UpsertHostToken :exec
INSERT INTO host_tokens (name, token)
VALUES ($1, $2)
ON CONFLICT (name) DO UPDATE SET token = EXCLUDED.token;

-- name: GetHostTokenByName :one
SELECT token FROM host_tokens WHERE name = $1;

-- name: DeleteHostTokenByName :exec
DELETE FROM host_tokens WHERE name = $1;

-- name: UpdateNodeStatusOffline :exec
UPDATE nodes SET status = 'offline'
WHERE name = $1 AND status = 'online' AND last_heartbeat < NOW() - INTERVAL '30 seconds';
