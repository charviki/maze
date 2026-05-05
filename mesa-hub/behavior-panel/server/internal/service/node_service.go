package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/charviki/maze-cradle/logutil"
)

// NodeService 节点管理业务逻辑（Manager 本地），供 HTTP handler 和 gRPC handler 共用
type NodeService struct {
	registry NodeRegistry
	logger   logutil.Logger
}

// NewNodeService 创建 NodeService
func NewNodeService(registry NodeRegistry, logger logutil.Logger) *NodeService {
	return &NodeService{
		registry: registry,
		logger:   logger,
	}
}

// ListNodes 返回所有已注册节点
func (s *NodeService) ListNodes(ctx context.Context) ([]*Node, error) {
	return s.registry.List(ctx)
}

// GetNode 返回指定节点信息
func (s *NodeService) GetNode(ctx context.Context, name string) (*Node, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	node, err := s.registry.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get node %q: %w", name, err)
	}
	if node == nil {
		return nil, errors.New("node not found")
	}
	return node, nil
}

// DeleteNode 从注册表删除指定节点
func (s *NodeService) DeleteNode(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("name is required")
	}
	ok, err := s.registry.Delete(ctx, name)
	if err != nil {
		return fmt.Errorf("delete node %q: %w", name, err)
	}
	if !ok {
		return errors.New("node not found")
	}
	return nil
}
