# Black Ridge AGENTS.md

## 职责

如同剧中 Sweetwater 附近的 Black Ridge 是 Host 执行任务、与外界交互的受限区域，Black Ridge 是 AI CLI 工具运行和执行编码任务的 Agent 节点。它通过 Tmux 管理 AI CLI（Claude Code、Codex、Bash Shell）的会话生命周期，提供终端交互、Pipeline 编排、状态持久化与恢复能力，并向 Manager 注册和上报心跳。

## 核心原则
- **管线编排** — 会话创建/恢复由 system/template/user 三层管线驱动，所有步骤按顺序原子执行
- **接口隔离** — TmuxService 通过接口抽象，外部 Handler 不感知 tmux 命令细节
- **安全默认** — 敏感步骤（env/file）执行前关闭 shell 回显，所有 API 端点经 Bearer Token 鉴权
- **可观测性** — 心跳定期上报完整状态快照（Session 详情、内存、本地配置），配置文件使用乐观并发控制

## 依赖关系
- 依赖: [cradle](../../fabrication/cradle/AGENTS.md) (Go 共享库, 提供日志/配置/协议/管线/中间件), [@maze/fabrication](../../fabrication/skin/AGENTS.md) (UI 组件库, 前端构建)
- 被依赖: [behavior-panel](../../mesa-hub/behavior-panel/AGENTS.md) (Manager 通过代理调用 Agent API)

## 关键文件
| 路径 | 职责 | 文档同步 |
|------|------|----------|
| server/main.go | 入口：配置加载、HTTP+gRPC 双协议服务初始化、信号处理 | [architecture.md](docs/architecture.md) |
| server/biz/grpc/server.go | gRPC Server 框架（Session/Config/Template Service） | — |
| server/biz/grpc/template.go | TemplateService gRPC 实现（ListTemplates/CreateTemplate/GetTemplate/UpdateTemplate/DeleteTemplate + Config 管理） | — |
| server/biz/grpc/session.go | SessionService + ConfigService gRPC 实现 | — |
| server/biz/config/config.go | 全局配置结构与校验 | [architecture.md](docs/architecture.md) |
| server/biz/service/tmux.go | Tmux 会话管理与管线执行 | [architecture.md](docs/architecture.md) |
| server/biz/service/heartbeat.go | Manager 心跳注册与状态上报 | [architecture.md](docs/architecture.md) |
| server/biz/service/autosave.go | 定时管线状态自动保存 | [architecture.md](docs/architecture.md) |
| server/biz/service/config_files.go | 配置文件读写与乐观并发控制 | [architecture.md](docs/architecture.md) |
| server/biz/model/template.go | 模板存储与内置模板加载 | [architecture.md](docs/architecture.md) |
| server/biz/router/register.go | API 路由注册 | [api.md](docs/api.md) |
| server/biz/handler/session.go | Session CRUD Handler | [api.md](docs/api.md) |
| server/biz/handler/terminal.go | 终端交互与 WebSocket Handler | [api.md](docs/api.md) |
| Dockerfile | 多阶段构建（前端+后端+运行时） | [architecture.md](docs/architecture.md) |
| entrypoint.sh | 容器启动初始化脚本 | [architecture.md](docs/architecture.md) |

## 详细文档
| 文档 | 内容 |
|------|------|
| [docs/architecture.md](docs/architecture.md) | 系统架构、启动流程、Tmux 管理、管线编排、状态持久化、心跳、自动保存、配置管理、模板系统、部署 |
| [docs/api.md](docs/api.md) | 完整 REST API 参考，含请求/响应格式、WebSocket 协议、数据模型 |
