# AGENTS.md

The Maze — 基于 Westworld 概念构建的 AI Agent 管理平台。Manager (behavior-panel) 作为代理网关和 Host 编排引擎，统一管控多个 Agent 节点 (black-ridge)，每个节点通过 tmux 运行 AI CLI 工具（Claude Code、Codex 等）。前端通过 Manager 代理所有操作，确保可观测性。Host 的完整生命周期（创建→部署→监控→恢复→销毁）由 Manager 的声明式 HostSpec 持久化 + Reconciler 自动化管理。

## 核心原则

- **代理网关 (Proxy Gateway)** — 前端不直连 Agent，所有操作经 Manager 代理转发，每次操作记录审计日志
- **可观测性 (Observability)** — Agent 节点的注册、心跳、会话操作必须有据可查，禁止绕过审计的暗箱操作
- **Pipeline 驱动 (Pipeline-Driven)** — 会话创建/恢复由三层 Pipeline（system/template/user）编排，而非散落的 shell 命令
- **共享基础 (Shared Foundation)** — Go 公共库 (cradle) 和 UI 组件库 (fabrication) 跨模块复用，修改时须考虑向后兼容
- **先读后改 (Read Before Modify)** — 修改某个模块的代码前，必须先用 Read 工具读取该模块的 AGENTS.md，了解其上下文和约束
- **构建规范 (Build Standards)** — 新增或修改 Dockerfile 必须遵循 [Docker 构建规范](fabrication/docs/docker-build-guide.md)（拆分 COPY + Cache Mount + 供应商镜像多阶段构建）
- **声明式编排 (Declarative Orchestration)** — Host 规格持久化为 HostSpec，Reconciler 确保实际状态趋近期望状态（启动恢复 + 健康巡检 + 自动重试）
- **分阶段实施 (Incremental Delivery)** — 大范围改动拆分为小步骤，每步实现后立即验证（编译/测试），确认无误后再推进下一步，避免一次性大改导致返工和方案漂移
- **拒绝反模式 (Reject Anti-Patterns)** — 识别设计污染、不必要的兼容妥协或反模式时，主动向开发者指出问题并提供更合理的替代方案，不做盲目的实现者
- **拒绝技术债 (Reject Technical Debt)** — 代码改动应选择合理方案，不应为沿用旧模式或因改动量大而默认采用兼容妥协；遇到设计取舍时主动向开发者反馈确认，避免代码腐烂
- **查验全量覆盖 (Full Linter Coverage)** — 禁止无理由禁用 eslint/tsc 规则；确需 `eslint-disable` 时必须补充注释说明原因，避免裸抑制扩散

## 交付铁律

每次代码变更交付前，必须按以下清单逐项完成：

1. **Go 编译通过** — `make build-go` 零错误
2. **Go 静态检查** — `make lint` 零警告
3. **Go 全量测试** — `make test` 全部 PASS，新增逻辑必须补充单测
4. **前端三道检查** — `make check-frontend` 全部通过（skin 用 `tsc --noEmit`，behavior-panel/black-ridge 用 `tsc -b --noEmit`，严格按 tsc → eslint → vitest 顺序）
5. **前端 Docker 构建** — `make build-web` 通过（`tsc -b` 会触发 `tsc --noEmit` 无法覆盖的编译路径）
6. **文档同步** — 就近检查并更新对应模块的 AGENTS.md、docs/、关键文件表格
7. **双环境验证** — `make test-integration PLATFORM=docker` 和 `make test-integration PLATFORM=kubernetes` 全部通过

## 模块索引

| 模块              | 目录                           | 职责                                                                | 详细文档                                                                                |
| ----------------- | ------------------------------ | ------------------------------------------------------------------- | --------------------------------------------------------------------------------------- |
| Portal            | mesa-hub/portal/               | 统一入口门户（Landing 含登录 → 主界面）                             | [AGENTS.md](mesa-hub/portal/AGENTS.md)                                                  |
| Behavior Panel    | mesa-hub/behavior-panel/       | Agent 管理面板 + Host 编排引擎（Go + React）                        | [AGENTS.md](mesa-hub/behavior-panel/AGENTS.md) + [docs/](mesa-hub/behavior-panel/docs/) |
| Cradle            | fabrication/cradle/            | Go 共享库（HTTP/Pipeline/Config/Auth + gRPC/Protobuf IDL 代码生成） | [AGENTS.md](fabrication/cradle/AGENTS.md) + [docs/](fabrication/cradle/docs/)           |
| Black Ridge       | sweetwater/black-ridge/        | Agent 运行时节点（tmux + Web UI）                                   | [AGENTS.md](sweetwater/black-ridge/AGENTS.md) + [docs/](sweetwater/black-ridge/docs/)   |
| Skin              | fabrication/skin/              | Westworld 主题 UI 组件库                                            | [AGENTS.md](fabrication/skin/AGENTS.md) + [docs/](fabrication/skin/docs/)               |
| Fabrication       | fabrication/                   | 制造部（UI 组件库 + Go 共享库 + Docker 构建 + K8s 部署基础设施）    | [AGENTS.md](fabrication/AGENTS.md) + [docs/](fabrication/docs/)                         |
| Integration Tests | fabrication/tests/integration/ | 跨模块集成测试（Go testing + kit 辅助包）                           | [AGENTS.md](fabrication/tests/integration/AGENTS.md)                                    |

## 文档维护

- 新增/变更模块入口文件 → 更新对应模块 AGENTS.md 的关键文件表格（每模块控制在 5-10 行）
- 架构设计变更 → 更新对应模块的 architecture.md（如有）
- 跨模块变更 → 更新本文件的模块索引
- 函数签名、实现文件、API 端点清单无需更新 AGENTS.md/docs/
