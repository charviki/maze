# Director Core

## 职责

The Mesa 控制面核心，负责 Agent 节点注册/心跳管理、声明式 Host 编排（HostSpec 持久化 + Reconciler 自动化）、Session/Template/Config 代理转发、审计日志，以及 JWT 认证 + Casbin RBAC 权限系统。

## 项目结构

`cmd/director-core/` 是服务入口，组装依赖并启动 gRPC + HTTP 双协议服务。`internal/` 遵循标准 Go 分层：`service/`（业务逻辑：Host、Node、Auth、Permission、Audit）、`transport/`（gRPC handler + HTTP gateway + WebSocket 终端代理）、`repository/`（抽象接口，postgres/ 和 file/ 两个实现）、`agentclient/`（向 Black Ridge 节点的 gRPC 出站代理 + 连接池管理）、`runtime/`（Docker / K8s 容器运行时抽象）、`reconciler/`（HostSpec 状态调和循环）、`hostbuilder/`（Host 镜像构建与工具校验）、`config/`（配置加载）。

## 核心原则

- **代理网关** — 前端不直连 Agent，所有请求经 Director Core 代理转发到 Black Ridge 节点，并记录审计日志
- **声明式编排** — HostSpec 持久化到文件系统，Reconciler 定期调和实际状态趋近期望状态，支持 Docker 和 Kubernetes 双运行时
- **Proto 层校验** — 所有 gRPC 请求通过 gatewayutil.NewValidationInterceptor（buf.validate）统一校验，transport/service 层不做手动参数检查
- **双数据库** — 权限系统（auth DB）与 Host 管理（host DB）使用独立 PostgreSQL 实例，通过独立连接池和 migration 管理

## 命令

```bash
# 构建
make build-go  # 或 cd the-mesa/director-core && go build ./cmd/director-core

# 单元测试
make test      # 或 go test ./...
```

## 依赖

- 依赖: [Cradle](../../fabrication/cradle/AGENTS.md)（auth、db、gatewayutil、grpcutil、logutil、configutil、lifecycle、middleware、pipeline、protocol、storeutil）
- 依赖: PostgreSQL（auth + host 两个数据库）
- 依赖: Docker 或 Kubernetes（Host 容器运行时）
- 依赖: Casbin（权限策略引擎）

## 详细文档

| 文档 | 内容 |
|------|------|
| [auth-overview.md](docs/auth-overview.md) | 权限系统对象模型、API 资源、认证与授权边界 |
| [auth-integration.md](docs/auth-integration.md) | 调用方如何基于 admin 和 subject_key 接入权限系统 |
| [auth-operations.md](docs/auth-operations.md) | 权限申请、审批、撤销、过期回收、审计与排障操作 |
