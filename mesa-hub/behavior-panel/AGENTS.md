# Behavior Panel

## 职责

代理网关 + Host 编排引擎，管控 Agent 节点生命周期（创建→部署→监控→恢复），代理前端到 Agent 的所有请求并记录审计日志。

## 项目结构

Go 后端入口在 server/cmd/behavior-panel/，采用 service → repository → transport 三层架构。业务域分四块：Host/Node 编排、权限管理、Agent 出站调用（agentclient/）、审计日志（repository/audit/）。路由采用 net/http + grpc-gateway + gRPC 统一架构，REST API 由 proto 注解驱动。
前端入口 web/src/App.tsx，API 客户端在 web/src/api/，通过 Nginx 代理到后端。

## 核心原则

- **SSRF 防护** — 代理目标 URL 必须通过协议和内网 IP 校验
- **异步编排** — Host 创建为 202 Accepted，后台构建部署，前端轮询状态
- **优雅关闭** — 依次停止 Reconciler → 刷盘 HostSpec/NodeRegistry → 关闭审计日志

## 命令

- `make build-go` / `make lint` / `make test` — Go 编译/检查/测试
- `make check-frontend` — 前端 tsc → eslint → vitest
- `make build-web` — 前端 Docker 构建

## 依赖

- 依赖: [cradle](../../fabrication/cradle/AGENTS.md), [@maze/fabrication](../../fabrication/skin/AGENTS.md)
- 被依赖: 无

## 详细文档

| 文档 | 内容 |
|------|------|
| [docs/architecture.md](docs/architecture.md) | 架构、数据流、Host 编排引擎、Reconciler、部署拓扑 |
| [docs/auth-overview.md](docs/auth-overview.md) | 权限系统对象模型、API 资源、认证与授权边界 |
| [docs/auth-integration.md](docs/auth-integration.md) | 调用方如何基于 admin 和 subject_key 接入权限系统 |
| [docs/auth-operations.md](docs/auth-operations.md) | 权限申请、审批、撤销、过期回收、审计与排障操作 |
