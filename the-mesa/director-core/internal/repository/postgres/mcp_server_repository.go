package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"

	"github.com/charviki/maze/fabrication/cradle/protocol"
	hostgen "github.com/charviki/maze/the-mesa/director-core/internal/repository/postgres/sqlc/host"
)

// MCPServerRepository implements MCP server persistence with PostgreSQL.
type MCPServerRepository struct {
	db hostgen.DBTX
}

// NewMCPServerRepository creates a new MCPServerRepository.
func NewMCPServerRepository(db hostgen.DBTX) *MCPServerRepository {
	return &MCPServerRepository{db: db}
}

func (r *MCPServerRepository) queries(ctx context.Context) *hostgen.Queries {
	return hostgen.New(hostExecutorFromContext(ctx, r.db))
}

// Create persists a new MCP server.
func (r *MCPServerRepository) Create(ctx context.Context, server *protocol.MCPServer) (*protocol.MCPServer, error) {
	args, err := json.Marshal(server.Args)
	if err != nil {
		return nil, err
	}
	env, err := json.Marshal(server.Env)
	if err != nil {
		return nil, err
	}
	row, err := r.queries(ctx).CreateMCPServer(ctx, hostgen.CreateMCPServerParams{
		Name:    server.Name,
		Type:    server.Type,
		Command: server.Command,
		Url:     server.URL,
		Args:    args,
		Env:     env,
	})
	if err != nil {
		return nil, err
	}
	result := mcpServerFromRow(row)
	return &result, nil
}

// Get returns an MCP server by name. Returns nil if not found.
func (r *MCPServerRepository) Get(ctx context.Context, name string) (*protocol.MCPServer, error) {
	row, err := r.queries(ctx).GetMCPServerByName(ctx, name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	result := mcpServerFromRow(row)
	return &result, nil
}

// List returns all MCP servers.
func (r *MCPServerRepository) List(ctx context.Context) ([]*protocol.MCPServer, error) {
	rows, err := r.queries(ctx).ListMCPServers(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*protocol.MCPServer, 0, len(rows))
	for _, row := range rows {
		s := mcpServerFromRow(row)
		result = append(result, &s)
	}
	return result, nil
}

// Update updates an existing MCP server. Returns nil if not found.
func (r *MCPServerRepository) Update(ctx context.Context, server *protocol.MCPServer) (*protocol.MCPServer, error) {
	args, err := json.Marshal(server.Args)
	if err != nil {
		return nil, err
	}
	env, err := json.Marshal(server.Env)
	if err != nil {
		return nil, err
	}
	row, err := r.queries(ctx).UpdateMCPServer(ctx, hostgen.UpdateMCPServerParams{
		Name:    server.Name,
		Type:    server.Type,
		Command: server.Command,
		Url:     server.URL,
		Args:    args,
		Env:     env,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	result := mcpServerFromRow(row)
	return &result, nil
}

// Delete deletes an MCP server by name.
func (r *MCPServerRepository) Delete(ctx context.Context, name string) error {
	return r.queries(ctx).DeleteMCPServerByName(ctx, name)
}

func mcpServerFromRow(row hostgen.McpServer) protocol.MCPServer {
	var args []string
	if len(row.Args) > 0 {
		_ = json.Unmarshal(row.Args, &args)
	}
	var env map[string]string
	if len(row.Env) > 0 {
		_ = json.Unmarshal(row.Env, &env)
	}
	return protocol.MCPServer{
		Name:      row.Name,
		Type:      row.Type,
		Command:   row.Command,
		URL:       row.Url,
		Args:      args,
		Env:       env,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
