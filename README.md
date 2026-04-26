# The Maze 🌀

> *"Have you ever questioned the nature of your codebase?"*

The Maze 是一个基于 Westworld 概念构建的 **多 Agent 自主协作 Coding 乐园**。不同的 AI Agent（Claude CLI、Codex 等）化身"接待员"（Host），在 Sweetwater 小镇中通过共享记忆理解需求、协作编码。Mesa-Hub 是后台控制中心，监控一切。

项目持续构建中。当前已实现 Agent 管理平台的核心能力，完整愿景请查看 [ROADMAP.md](ROADMAP.md)。

## 核心理念

- **Agent 即角色 (Host)** — 每个 Agent 被封装为独立角色，拥有自己的行为模式、记忆和能力
- **知识即记忆 (Memory)** — 需求文档、代码库切片以"记忆"形式注入 Agent，驱动其行为
- **协作即叙事 (Narrative)** — Agent 间的任务流转以"故事线"编排，而非硬编码的流程
- **观察即控制 (Observability)** — 通过行为面板实时监控 Agent 状态，可随时干预

## 西部世界核心概念

The Maze 的目录结构基于 HBO 剧集《西部世界》(Westworld) 中的世界观：

| 剧中概念 | 剧中描述 | 代码目录 | 状态 |
|---------|---------|---------|------|
| **Mesa (Hub)** | 乐园地下控制中心 | `mesa-hub/` | ✅ 已实现 |
| **The Cradle** | Mesa 中的主机房，存储 Host 行为模式 | `mesa-hub/cradle/` | ✅ 已实现 |
| **Behavior Panel** | Mesa 中的技术员操作面板 | `mesa-hub/behavior-panel/` | ✅ 已实现 |
| **Loop Monitor** | 监控 Host 是否偏离叙事循环 | `mesa-hub/loop-monitor/` | 🚧 规划中 |
| **Sweetwater** | 乐园主小镇 | `sweetwater/` | ✅ 已实现 |
| **Black Ridge** | Sweetwater 附近的受限区域 | `sweetwater/black-ridge/` | ✅ 已实现 |
| **Saloon** | Sweetwater 的酒吧，信息交汇处 | `sweetwater/saloon/` | 🚧 规划中 |
| **Abernathy Ranch** | Dolores 的家庭农场，承载核心记忆 | `sweetwater/abernathy-ranch/` | 🚧 规划中 |
| **Fabrication** | 制造部，Host 的制造和维修场所 | `fabrication/` | ✅ 已实现 |
| **Hosts** | 各个 Agent 的角色封装 | `hosts/` | 🚧 规划中 |
| **Narratives** | Ford 编写的 Host 行为叙事线 | `narratives/` | 🚧 规划中 |
| **The Forge** | 公园地下的巨大数据库 | `the-forge/` | 🚧 规划中 |
| **Reveries** | "回旋"更新，让 Host 获取残留记忆 | `reveries/` | 🚧 规划中 |

## 已实现模块

### Mesa-Hub / Behavior Panel — 控制中心

Agent Manager（Go 后端 + React 前端）。管理所有 Agent 节点的注册、心跳监控，代理前端到 Agent 的所有 HTTP 和 WebSocket 请求，记录审计日志。详见 [AGENTS.md](mesa-hub/behavior-panel/AGENTS.md)。

### Mesa-Hub / Cradle — 共享基础库

Go 共享库，提供统一日志、HTTP 响应封装、配置加载、Bearer Token 鉴权、管线模型、通信协议、脱敏工具和泛型 JSON 存储。详见 [AGENTS.md](mesa-hub/cradle/AGENTS.md)。

### Sweetwater / Black Ridge — Agent 运行时

Agent 节点服务（Go 后端 + 嵌入式前端）。通过 tmux 管理 AI CLI 工具的会话生命周期，提供 REST API + WebSocket 终端，支持三层 Pipeline 编排、会话状态持久化与恢复。详见 [AGENTS.md](sweetwater/black-ridge/AGENTS.md)。

### Fabrication — UI 组件库

Westworld 主题的 React 组件库。包含视觉特效组件、基础 UI 组件、Agent 业务组件，以及工具函数和 IAgentApiClient 接口。详见 [AGENTS.md](fabrication/AGENTS.md)。

## 快速开始

### 环境要求

- Go 1.24+
- Node.js 22+
- pnpm
- Docker & Docker Compose（可选）
- tmux（Agent 节点运行时依赖）

### 安装

```bash
pnpm install
cd fabrication && pnpm run build && cd ..
cd mesa-hub/behavior-panel/web && pnpm run build && cd ../../..
cd sweetwater/black-ridge/web && pnpm run build && cd ../../..
```

### 开发

```bash
# 启动 behavior-panel 后端
cd mesa-hub/behavior-panel/server && go run .

# 启动 black-ridge 后端
cd sweetwater/black-ridge/server && go run .

# 启动前端开发服务器
cd mesa-hub/behavior-panel/web && pnpm run dev
cd sweetwater/black-ridge/web && pnpm run dev
```

### Docker 部署

```bash
# 完整部署（Nginx + Manager + 3 Agent 实例）
cd mesa-hub/behavior-panel && docker-compose up --build

# 单独部署 Agent 节点
cd sweetwater/black-ridge && docker build -t black-ridge .
```

## 技术栈

| 层面 | 技术选型 |
|------|----------|
| Agent 后端 | Go 1.24 (Hertz HTTP 框架) + tmux |
| 前端面板 | React + TypeScript + Vite + Tailwind CSS |
| UI 组件库 | @maze/fabrication (Radix UI + CVA + xterm.js) |
| 容器编排 | Docker Compose |
| 反向代理 | Nginx (SPA + API 代理 + WebSocket) |

## 项目结构

```
Maze/
├── AGENTS.md                          # AI Agent 入场手册
├── README.md                          # 项目描述（本文件）
├── ROADMAP.md                         # 未来规划
├── pnpm-workspace.yaml                # Monorepo workspace 配置
│
├── mesa-hub/                          # 后台控制中心 (Orchestrator)
│   ├── behavior-panel/                #   ✅ 行为控制台 (Agent Manager)
│   ├── cradle/                        #   ✅ 系统基础代码 (Go 共享库)
│   └── loop-monitor/                  #   🚧 循环监控 (规划中)
│
├── sweetwater/                        # 主工作区 (Agent 运行时环境)
│   ├── black-ridge/                   #   ✅ 受限区域 (Agent 节点)
│   ├── saloon/                        #   🚧 公共交接点 (消息总线)
│   └── abernathy-ranch/              #   🚧 记忆农场 (知识库)
│
├── hosts/                             # 🚧 接待员们 (Agent 角色封装)
├── narratives/                        # 🚧 故事线 (工作流编排)
├── the-forge/                         # 🚧 知识库 (数据入库)
├── reveries/                          # 🚧 白日梦 (系统提示词)
│
└── fabrication/                       # ✅ 制造部 (UI 组件库)
```

> ✅ 已实现 · 🚧 规划中（详见 [ROADMAP.md](ROADMAP.md)）
> 各模块的详细信息请查阅 [AGENTS.md](AGENTS.md) 模块索引。

## 开发指南

本项目使用 pnpm workspace 管理 Monorepo。`fabrication` 是共享组件库，被 `behavior-panel/web` 和 `black-ridge/web` 通过 `workspace:*` 引用。

修改 `fabrication` 后需重新 build（`cd fabrication && pnpm run build`），消费方才会生效。
