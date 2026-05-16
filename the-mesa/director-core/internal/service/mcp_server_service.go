package service

import (
	"context"
	"fmt"

	"github.com/charviki/maze/fabrication/cradle/logutil"
	"github.com/charviki/maze/fabrication/cradle/protocol"
)

// MCPServerRepo defines the interface for MCP server persistence.
type MCPServerRepo interface {
	Create(ctx context.Context, server *protocol.MCPServer) (*protocol.MCPServer, error)
	Get(ctx context.Context, name string) (*protocol.MCPServer, error)
	List(ctx context.Context) ([]*protocol.MCPServer, error)
	Update(ctx context.Context, server *protocol.MCPServer) (*protocol.MCPServer, error)
	Delete(ctx context.Context, name string) error
}

// MCPServerService implements business logic for MCP server operations.
type MCPServerService struct {
	repo   MCPServerRepo
	logger logutil.Logger
}

// NewMCPServerService creates a new MCPServerService.
func NewMCPServerService(repo MCPServerRepo, logger logutil.Logger) *MCPServerService {
	return &MCPServerService{repo: repo, logger: logger}
}

// Create creates a new MCP server.
func (s *MCPServerService) Create(ctx context.Context, server *protocol.MCPServer) (*protocol.MCPServer, error) {
	if err := validateMCPServerFields(server); err != nil {
		return nil, err
	}
	existing, err := s.repo.Get(ctx, server.Name)
	if err != nil {
		return nil, fmt.Errorf("create mcp server %q: check existing: %w", server.Name, err)
	}
	if existing != nil {
		return nil, ErrAlreadyExists
	}
	result, err := s.repo.Create(ctx, server)
	if err != nil {
		return nil, fmt.Errorf("create mcp server %q: %w", server.Name, err)
	}
	return result, nil
}

// Get returns an MCP server by name.
func (s *MCPServerService) Get(ctx context.Context, name string) (*protocol.MCPServer, error) {
	server, err := s.repo.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get mcp server %q: %w", name, err)
	}
	if server == nil {
		return nil, ErrNotFound
	}
	return server, nil
}

// List returns all MCP servers.
func (s *MCPServerService) List(ctx context.Context) ([]*protocol.MCPServer, error) {
	servers, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list mcp servers: %w", err)
	}
	return servers, nil
}

// Update updates an existing MCP server.
func (s *MCPServerService) Update(ctx context.Context, server *protocol.MCPServer) (*protocol.MCPServer, error) {
	if err := validateMCPServerFields(server); err != nil {
		return nil, err
	}
	result, err := s.repo.Update(ctx, server)
	if err != nil {
		return nil, fmt.Errorf("update mcp server %q: %w", server.Name, err)
	}
	if result == nil {
		return nil, ErrNotFound
	}
	return result, nil
}

// Delete deletes an MCP server by name.
func (s *MCPServerService) Delete(ctx context.Context, name string) error {
	if err := s.repo.Delete(ctx, name); err != nil {
		return fmt.Errorf("delete mcp server %q: %w", name, err)
	}
	return nil
}

func validateMCPServerFields(server *protocol.MCPServer) error {
	switch server.Type {
	case "stdio":
		if server.Command == "" {
			return fmt.Errorf("%w: command is required for stdio type", ErrInvalidInput)
		}
	case "sse", "streamable-http":
		if server.URL == "" {
			return fmt.Errorf("%w: url is required for %s type", ErrInvalidInput, server.Type)
		}
	}
	return nil
}
