# The Mesa

## 职责

The Mesa 控制面域，汇总 Arrival Gate 入口壳、Director Console 控制台前端、Director Core 控制核心、The Forge 知识库服务，以及域级 Docker Compose / Nginx / 测试资产。

## 项目结构

入口前端直接位于 `arrival-gate/` 顶层，控制台前端直接位于 `director-console/` 顶层，控制核心 Go 模块直接位于 `director-core/` 顶层，知识库服务直接位于 `the-forge/` 顶层（含 Go 后端 `cmd/` + `internal/` + React 前端 `web/`），域级组装资产位于 `the-mesa/` 根目录。
Web 入口使用 `/arrival-gate/` 与 `/director-console/`，Director Core 的二进制名和环境变量前缀统一为 `director-core` / `DIRECTOR_CORE_*`。

## 核心原则

- **SSRF 防护** — 代理目标 URL 必须通过协议和内网 IP 校验
- **异步编排** — Host 创建为 202 Accepted，后台构建部署，前端轮询状态
- **优雅关闭** — 依次停止 Reconciler → 刷盘 HostSpec/NodeRegistry → 关闭审计日志

## 命令

- `make build-go` / `make lint` / `make test` — Director Core 与其他 Go 模块编译/检查/测试
- `make check-frontend` — Arrival Gate / Director Console / The Forge / 其他前端 tsc → eslint → vitest
- `make build-web` — The Mesa 组合前端 Docker 构建
- `make up PLATFORM=kubernetes ENV=dev|test|prod` — K8s 构建并部署（dev/test/prod 三套环境）
- `cd the-mesa && docker compose up --build` — 本地 Docker Compose 组合运行

## 依赖

- 依赖: [cradle](../fabrication/cradle/AGENTS.md), [@maze/fabrication](../fabrication/skin/AGENTS.md)
- 被依赖: 无

## 详细文档

| 文档 | 内容 |
|------|------|
| [docs/architecture.md](docs/architecture.md) | The Mesa 组合架构、数据流、部署拓扑 |
| [director-core/docs/auth-overview.md](director-core/docs/auth-overview.md) | 权限系统对象模型、API 资源、认证与授权边界 |
| [director-core/docs/auth-integration.md](director-core/docs/auth-integration.md) | 调用方如何基于 admin 和 subject_key 接入权限系统 |
| [director-core/docs/auth-operations.md](director-core/docs/auth-operations.md) | 权限申请、审批、撤销、过期回收、审计与排障操作 |
