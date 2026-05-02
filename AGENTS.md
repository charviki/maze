# AGENTS.md

The Maze — 基于 Westworld 概念构建的 AI Agent 管理平台。Manager (behavior-panel) 作为代理网关和 Host 编排引擎，统一管控多个 Agent 节点 (black-ridge)，每个节点通过 tmux 运行 AI CLI 工具（Claude Code、Codex 等）。前端通过 Manager 代理所有操作，确保可观测性。Host 的完整生命周期（创建→部署→监控→恢复→销毁）由 Manager 的声明式 HostSpec 持久化 + Reconciler 自动化管理。

## 核心原则

- **代理网关 (Proxy Gateway)** — 前端不直连 Agent，所有操作经 Manager 代理转发，每次操作记录审计日志
- **可观测性 (Observability)** — Agent 节点的注册、心跳、会话操作必须有据可查，禁止绕过审计的暗箱操作
- **Pipeline 驱动 (Pipeline-Driven)** — 会话创建/恢复由三层 Pipeline（system/template/user）编排，而非散落的 shell 命令
- **共享基础 (Shared Foundation)** — Go 公共库 (cradle) 和 UI 组件库 (fabrication) 跨模块复用，修改时须考虑向后兼容
- **先读后改 (Read Before Modify)** — 修改模块代码前，先阅读该模块的 AGENTS.md 了解上下文
- **构建规范 (Build Standards)** — 新增或修改 Dockerfile 必须遵循 [Docker 构建规范](fabrication/docs/docker-build-guide.md)（拆分 COPY + Cache Mount + 供应商镜像多阶段构建）
- **声明式编排 (Declarative Orchestration)** — Host 规格持久化为 HostSpec，Reconciler 确保实际状态趋近期望状态（启动恢复 + 健康巡检 + 自动重试）

## 交付铁律（强制执行）

每次代码变更交付前，必须按以下清单逐项完成，缺一不可：

1. **编译通过** — `make build-go` 零错误
2. **静态检查** — `make vet` 零警告
3. **全量测试** — `make test` 全部 PASS，新增逻辑必须补充单测
4. **文档同步** — 就近检查并更新对应模块的 AGENTS.md、docs/、关键文件表格
5. **双环境验证** — 涉及运行时、部署、网络、配置的变更，必须在 Docker Compose 和 Kubernetes 两种环境下都验证通过

快捷命令：`make check` = 编译 + 静态检查 + 测试（交付铁律 1-3 项）

## 模块索引

| 模块              | 目录                           | 职责                                                                | 详细文档                                                                                |
| ----------------- | ------------------------------ | ------------------------------------------------------------------- | --------------------------------------------------------------------------------------- |
| Portal            | mesa-hub/portal/               | 统一入口门户（Landing 含登录 → 主界面）                             | [AGENTS.md](mesa-hub/portal/AGENTS.md)                                                  |
| Behavior Panel    | mesa-hub/behavior-panel/       | Agent 管理面板 + Host 编排引擎（Go + React）                        | [AGENTS.md](mesa-hub/behavior-panel/AGENTS.md) + [docs/](mesa-hub/behavior-panel/docs/) |
| Cradle            | fabrication/cradle/            | Go 共享库（HTTP/Pipeline/Config/Auth + gRPC/Protobuf IDL 代码生成） | [AGENTS.md](fabrication/cradle/AGENTS.md) + [docs/](fabrication/cradle/docs/)           |
| Black Ridge       | sweetwater/black-ridge/        | Agent 运行时节点（tmux + Web UI）                                   | [AGENTS.md](sweetwater/black-ridge/AGENTS.md) + [docs/](sweetwater/black-ridge/docs/)   |
| Fabrication       | fabrication/                   | 制造部（UI 组件库 + Go 共享库 + Docker 构建 + K8s 部署基础设施）    | [AGENTS.md](fabrication/AGENTS.md) + [docs/](fabrication/docs/)                         |
| Integration Tests | fabrication/tests/integration/ | 跨模块集成测试（Go testing + kit 辅助包）                           | —                                                                                       |

## 文档维护

- 新增源码文件（`.go`）→ 必须在对应模块 AGENTS.md 的关键文件表格中添加条目
- 修改函数签名 / 接口 / 核心流程 → 必须检查并更新对应 AGENTS.md 和 docs/ 中的描述
- 跨模块变更 → 更新本文件的模块索引
- 新增机制（如缓存、限流、hash 校验）→ 必须在关键文件表格中记录，并在 docs/ 中补充说明
