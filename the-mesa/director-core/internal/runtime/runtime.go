package runtime

import (
	"context"

	"github.com/charviki/maze-cradle/protocol"
)

// HostRuntime Host 运行时抽象接口
// 屏蔽底层容器编排差异（Docker / K8s），Handler 层只操作此接口
type HostRuntime interface {
	// DeployHost 部署一个 Host：构建镜像 + 创建容器/工作负载
	// dockerfileContent 由 builder 层生成，runtime 负责构建和部署
	DeployHost(ctx context.Context, spec *protocol.HostDeploySpec, dockerfileContent string) (*protocol.CreateHostResponse, error)

	// StopHost 停止运行时资源（容器/Deployment/Service），保留持久化数据
	// 用于 reconciler 重新部署和 deployHostAsync 清理旧容器等场景
	StopHost(ctx context.Context, name string) error

	// RemoveHost 销毁 Host：停止运行时 + 清理持久化数据 + 删除镜像
	// 仅用于用户主动删除 Host 的场景
	RemoveHost(ctx context.Context, name string) error

	// InspectHost 查询 Host 状态
	InspectHost(ctx context.Context, name string) (*protocol.ContainerInfo, error)

	// GetRuntimeLogs 获取容器/Pod 运行日志（最近 tailLines 行）
	GetRuntimeLogs(ctx context.Context, name string, tailLines int) (string, error)

	// IsHealthy 检查容器/Pod 是否健康运行
	IsHealthy(ctx context.Context, name string) (bool, error)
}
