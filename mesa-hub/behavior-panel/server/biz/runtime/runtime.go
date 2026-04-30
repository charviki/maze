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
	// RemoveHost 销毁 Host：停止容器/工作负载并清理资源
	RemoveHost(ctx context.Context, name string) error
	// InspectHost 查询 Host 状态
	InspectHost(ctx context.Context, name string) (*protocol.ContainerInfo, error)
}
