# AGENTS.md

The Maze — 基于 Westworld 概念构建的 AI Agent 管理平台。Manager (behavior-panel) 统一管控多个 Agent 节点 (black-ridge)，每个节点通过 tmux 运行 AI CLI 工具（Claude Code、Codex 等）。前端通过 Manager 代理所有操作，确保可观测性。

## 核心原则

- **代理网关 (Proxy Gateway)** — 前端不直连 Agent，所有操作经 Manager 代理转发，每次操作记录审计日志
- **可观测性 (Observability)** — Agent 节点的注册、心跳、会话操作必须有据可查，禁止绕过审计的暗箱操作
- **Pipeline 驱动 (Pipeline-Driven)** — 会话创建/恢复由三层 Pipeline（system/template/user）编排，而非散落的 shell 命令
- **共享基础 (Shared Foundation)** — Go 公共库 (cradle) 和 UI 组件库 (fabrication) 跨模块复用，修改时须考虑向后兼容
- **先读后改 (Read Before Modify)** — 修改模块代码前，先阅读该模块的 AGENTS.md 了解上下文
- **验证后交付 (Verify Before Deliver)** — 修改代码后必须验证编译通过、测试通过，不要仅靠推理，自验证通过后再交付

## 模块索引

| 模块 | 目录 | 职责 | 详细文档 |
|------|------|------|----------|
| Behavior Panel | mesa-hub/behavior-panel/ | Agent 管理面板（Go + React） | [AGENTS.md](mesa-hub/behavior-panel/AGENTS.md) + [docs/](mesa-hub/behavior-panel/docs/) |
| Cradle | mesa-hub/cradle/ | Go 共享库（HTTP/Pipeline/Config/Auth） | [AGENTS.md](mesa-hub/cradle/AGENTS.md) + [docs/](mesa-hub/cradle/docs/) |
| Black Ridge | sweetwater/black-ridge/ | Agent 运行时节点（tmux + Web UI） | [AGENTS.md](sweetwater/black-ridge/AGENTS.md) + [docs/](sweetwater/black-ridge/docs/) |
| Fabrication | fabrication/ | Westworld 主题 UI 组件库 | [AGENTS.md](fabrication/AGENTS.md) + [docs/](fabrication/docs/) |

## 文档维护

- 修改代码后，就近检查并更新对应模块的 AGENTS.md 或 docs/（参考关键文件表格的"文档同步"列）
- 跨模块变更时更新本文件
