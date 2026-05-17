-- +goose Up
ALTER TABLE host_specs ADD COLUMN skills JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE host_specs ADD COLUMN mcp_servers JSONB NOT NULL DEFAULT '[]'::jsonb;

-- +goose Down
ALTER TABLE host_specs DROP COLUMN IF EXISTS mcp_servers;
ALTER TABLE host_specs DROP COLUMN IF EXISTS skills;
