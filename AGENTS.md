# AGENTS.md

The Maze — AI Agent 管理平台。Manager (behavior-panel) 代理网关 + Host 编排引擎，统一管控 Agent 节点 (black-ridge)。声明式 HostSpec 持久化 + Reconciler 自动化管理。

## 核心原则

- **代理网关** — 前端不直连 Agent，所有请求经 Manager 代理转发 + 审计日志
- **声明式编排** — HostSpec 持久化 + Reconciler 确保实际状态趋近期望状态
- **先读后改** — 修改模块代码前，先读该模块的 AGENTS.md 了解上下文
- **分阶段实施** — 大范围改动拆分为小步骤，每步验证通过后再推进下一步
- **拒绝技术债** — 识别设计污染或不必要的兼容妥协时，主动指出并给出替代方案
- **查验全量覆盖** — `make build-go` + `make lint` + `make test` + `make check-frontend` 全部通过

## 交付铁律

1. `make build-go` 零错误
2. `make lint` 零警告
3. `make test` 全部 PASS
4. `make check-frontend` 全部通过（tsc → eslint → vitest）
5. `make build-web` 通过
6. `make test-integration PLATFORM=docker` + `PLATFORM=kubernetes` 全部通过

## 模块索引

| 模块                | 目录                             | 详细文档                                                 |
| ----------------- | ------------------------------ | ---------------------------------------------------- |
| Portal            | mesa-hub/portal/               | [AGENTS.md](mesa-hub/portal/AGENTS.md)               |
| Behavior Panel    | mesa-hub/behavior-panel/       | [AGENTS.md](mesa-hub/behavior-panel/AGENTS.md)       |
| Cradle            | fabrication/cradle/            | [AGENTS.md](fabrication/cradle/AGENTS.md)            |
| Black Ridge       | sweetwater/black-ridge/        | [AGENTS.md](sweetwater/black-ridge/AGENTS.md)        |
| Skin              | fabrication/skin/              | [AGENTS.md](fabrication/skin/AGENTS.md)              |
| Fabrication       | fabrication/                   | [AGENTS.md](fabrication/AGENTS.md)                   |
| Integration Tests | fabrication/tests/integration/ | [AGENTS.md](fabrication/tests/integration/AGENTS.md) |

## 架构文档

详见 [docs/architecture.md](docs/architecture.md)（模块拓扑图 + 请求路由 + Host 生命周期 + Session 代理流）

## 文档维护

- 新增/修改任何 AGENTS.md → 必须遵循 [docs/AGENTS-SPEC.md](docs/AGENTS-SPEC.md)（禁止添加文件表格、禁止搬运函数签名）
- 架构设计变更 → 更新 docs/architecture.md
- 模块对外接口变更 → 更新对应模块 AGENTS.md 的依赖关系
- 文档中禁止复制代码中的类型签名、函数签名、API 端点清单

