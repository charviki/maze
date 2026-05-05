# Black Ridge

## 职责

Agent 运行时节点，通过 tmux 管理 AI CLI（Claude Code/Codex）会话生命周期，提供终端交互、Pipeline 编排、状态持久化，向 Manager 注册和上报心跳。

## 项目结构

Go 后端入口在 server/cmd/black-ridge/，业务逻辑分 service/（Tmux 会话管理 + 心跳注册 + 配置读写）和 transport/（gRPC 适配 + WebSocket 终端）。路由采用 net/http + grpc-gateway + gRPC 统一架构，SPA 静态文件通过 go:embed 嵌入。
前端在 web/，嵌入后端镜像。

## 核心原则

- **管线编排** — 会话由 system/template/user 三层 Pipeline 按序驱动
- **接口隔离** — TmuxService 通过接口抽象，Handler 不感知 tmux 命令细节
- **乐观并发** — 配置文件使用乐观并发控制

## 命令

- `make build-go` / `make lint` / `make test` — Go 编译/检查/测试
- `make check-frontend` — 前端检查

## 依赖

- 依赖: [cradle](../../fabrication/cradle/AGENTS.md), [@maze/fabrication](../../fabrication/skin/AGENTS.md)
- 被依赖: [behavior-panel](../../mesa-hub/behavior-panel/AGENTS.md)

## 详细文档

| 文档 | 内容 |
|------|------|
| [docs/architecture.md](docs/architecture.md) | 系统架构、Tmux 管理、管线编排、状态持久化、心跳、自动保存 |
