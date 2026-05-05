# The Maze 🌀

> *"Have you ever questioned the nature of your codebase?"*

The Maze 是一个基于 Westworld 概念构建的 **多 Agent 自主协作 Coding 平台**。不同的 AI Agent（Claude Code、Codex 等）化身"接待员"（Host），在 Sweetwater 小镇中通过共享记忆理解需求、协作编码。The Mesa 是后台控制中心，监控一切。

项目持续构建中。完整愿景请查看 [ROADMAP.md](ROADMAP.md)。

## 已实现模块

### The Mesa — 控制面域

The Mesa 汇总 Director Core（Go 控制核心）、Director Console（React 控制台前端）和 Arrival Gate（统一入口前端）。Director Core 通过 gRPC + grpc-gateway 管理 Agent 节点的注册、心跳，通过声明式 HostSpec + Reconciler 编排 Host 的完整生命周期，代理前端到 Agent 的所有 HTTP 和 WebSocket 请求，记录审计日志。详见 [AGENTS.md](the-mesa/AGENTS.md)。

### Arrival Gate — 入口门户

统一入口门户，两阶段体验（Landing 含登录 → 主界面），西部世界主题的交互式迷宫 + 模块卡片导航。详见 [AGENTS.md](the-mesa/arrival-gate/AGENTS.md)。

### Cradle — Go 共享库

Go 公共库，提供 gRPC/Protobuf IDL 代码生成、OpenAPI Go HTTP client 生成、grpc-gateway 工具、统一日志、配置加载、管线模型、通信协议、脱敏工具和泛型 JSON 存储。详见 [AGENTS.md](fabrication/cradle/AGENTS.md)。

### Black Ridge — Agent 运行时

Agent 节点服务（Go gRPC 后端 + 嵌入式前端）。通过 tmux 管理 AI CLI 工具的会话生命周期，支持三层 Pipeline 编排、会话状态持久化与恢复。详见 [AGENTS.md](sweetwater/black-ridge/AGENTS.md)。

### Skin — UI 组件库

Westworld 主题的 React 组件库。包含视觉特效组件、基础 UI 组件、Agent 业务组件及工具函数。详见 [AGENTS.md](fabrication/skin/AGENTS.md)。

### Integration Tests — 集成测试

跨模块集成测试（Go testing + kit 辅助包），覆盖 Docker 和 Kubernetes 双环境。详见 [AGENTS.md](fabrication/tests/integration/AGENTS.md)。

## 快速开始

### 环境要求

- Go 1.26+
- Node.js 22+
- pnpm
- buf（Proto 代码生成）
- openapi-generator + Java（HTTP client 生成）
- Docker & Docker Compose
- tmux（Agent 节点运行时依赖）

### 安装

```bash
pnpm install
make build-web
```

### 开发

```bash
# 修改 proto 后重新生成代码
make gen

# 启动 Director Core 后端
cd the-mesa/director-core && go run ./cmd/director-core

# 启动 Arrival Gate 前端开发服务器 (port 3002)
pnpm --filter @maze/arrival-gate dev

# 启动 Director Console 前端开发服务器 (port 3000)
pnpm --filter @maze/director-console dev
```

### Docker 部署

```bash
# 完整部署（Nginx + The Mesa Web + Director Core）
cd the-mesa && docker compose up --build
```

## 技术栈

| 层面 | 技术选型 |
|------|----------|
| API 定义 | Protobuf IDL + buf 代码生成 |
| 后端通信 | gRPC + grpc-gateway (REST) |
| Agent 后端 | Go 1.26 + tmux |
| 前端框架 | React + TypeScript + Vite + Tailwind CSS |
| UI 组件库 | @maze/fabrication (Radix UI + CVA + xterm.js) |
| 数据持久化 | PostgreSQL（权限系统）+ JSON 文件（Node/Host） |
| 容器编排 | Docker Compose / Kubernetes |
| 反向代理 | Nginx |

## 项目结构

```
Maze/
├── AGENTS.md                          # AI Agent 项目上下文
├── README.md                          # 项目描述（本文件）
├── ROADMAP.md                         # 未来规划
├── docs/                              # 架构文档
│   ├── architecture.md                #   总体架构（含 Mermaid 拓扑图 + 数据流）
│   └── AGENTS-SPEC.md                 #   AGENTS.md 编写规范
│
├── the-mesa/                          # 控制面域
│   ├── director-core/                 #   ✅ 控制核心（Director Core）
│   ├── director-console/              #   ✅ 控制台前端
│   └── arrival-gate/                  #   ✅ 入口门户
│
├── sweetwater/                        # Agent 运行时环境
│   └── black-ridge/                   #   ✅ Agent 节点
│
└── fabrication/                       # 共享基础设施
    ├── cradle/                        #   ✅ Go 共享库 + Protobuf IDL
    ├── skin/                          #   ✅ UI 组件库
    └── tests/integration/             #   ✅ 跨模块集成测试
```

> ✅ 已实现 · 🚧 规划中（详见 [ROADMAP.md](ROADMAP.md)）
> 架构详情请查看 [docs/architecture.md](docs/architecture.md)

## 开发指南

本项目使用 pnpm workspace 管理 Monorepo。`fabrication/skin` 是共享组件库，被 `the-mesa/director-console`、`the-mesa/arrival-gate` 和 `black-ridge/web` 通过 `workspace:*` 引用。

修改 `fabrication/skin` 后需重新 build（`pnpm --filter @maze/fabrication build`），消费方才会生效。修改 proto 文件后执行 `make gen` 重新生成 Go 类型和 HTTP client。
