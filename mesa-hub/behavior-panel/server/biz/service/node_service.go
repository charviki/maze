package service

import (
	"context"
	"fmt"

	"github.com/charviki/maze-cradle/logutil"
	"github.com/charviki/mesa-hub-behavior-panel/biz/model"
)

// NodeService 节点管理业务逻辑（Manager 本地），供 HTTP handler 和 gRPC handler 共用
type NodeService struct {
	registry *model.NodeRegistry
	logger   logutil.Logger
}

// NewNodeService 创建 NodeService
func NewNodeService(registry *model.NodeRegistry, logger logutil.Logger) *NodeService {
	return &NodeService{
		registry: registry,
		logger:   logger,
	}
}

// ListNodes 返回所有已注册节点
func (s *NodeService) ListNodes(ctx context.Context) ([]*model.Node, error) {
	return s.registry.List(), nil
}

// GetNode 返回指定节点信息
func (s *NodeService) GetNode(ctx context.Context, name string) (*model.Node, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	node := s.registry.Get(name)
	if node == nil {
		return nil, fmt.Errorf("node not found")
	}
	return node, nil
}

// DeleteNode 从注册表删除指定节点
func (s *NodeService) DeleteNode(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if !s.registry.Delete(name) {
		return fmt.Errorf("node not found")
	}
	return nil
}
