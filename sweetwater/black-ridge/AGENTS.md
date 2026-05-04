# Black Ridge AGENTS.md

## 职责

如同剧中 Sweetwater 附近的 Black Ridge 是 Host 执行任务、与外界交互的受限区域，Black Ridge 是 AI CLI 工具运行和执行编码任务的 Agent 节点。它通过 Tmux 管理 AI CLI（Claude Code、Codex、Bash Shell）的会话生命周期，提供终端交互、Pipeline 编排、状态持久化与恢复能力，并向 Manager 注册和上报心跳。

## 核心原则

- **管线编排** — 会话创建/恢复由 system/template/user 三层管线驱动，所有步骤按顺序原子执行
- **接口隔离** — TmuxService、ConfigFileProvider、SessionStateRepository 通过接口抽象，transport 不直接依赖具体实现，外部 Handler 不感知底层细节
- **安全默认** — 敏感步骤（env/file）执行前关闭 shell 回显，所有 API 端点经 Bearer Token 鉴权（gRPC interceptor）
- **可观测性** — 心跳定期上报完整状态快照（Session 详情、内存、本地配置），配置文件使用乐观并发控制

## 路由架构

Agent 采用 **`net/http` + grpc-gateway + gRPC** 的统一路由架构：

- **`net/http.ServeMux`** 负责 WebSocket 路由、健康检查、访问日志和 SPA fallback
- **grpc-gateway ServeMux** 作为 `/api/` 兜底 handler 接管所有 REST 请求，由 proto 注解驱动路由
- **gRPC Server** 运行在 `:9090`（进程内直连），经过 `UnaryAuthInterceptor` 认证拦截
- 所有 REST API（Session/Template/Config）由 proto `google.api.http` 注解定义，含 `additional_bindings` 支持 Agent 内部路径
- SPA 静态文件通过 `go:embed` 嵌入，非 `/api/` 路径 fallback 到 `index.html`

## 依赖关系

- 依赖: [cradle](../../fabrication/cradle/AGENTS.md) (Go 共享库, 提供日志/配置/协议/管线/中间件), [@maze/fabrication](../../fabrication/skin/AGENTS.md) (UI 组件库, 前端构建)
- 被依赖: [behavior-panel](../../mesa-hub/behavior-panel/AGENTS.md) (Manager 通过代理调用 Agent API)

## 关键文件

| 路径                               | 职责                                                                                                              | 文档同步                                |
| ---------------------------------- | ----------------------------------------------------------------------------------------------------------------- | --------------------------------------- |
| server/cmd/black-ridge/main.go    | 入口：配置加载、grpc-gateway 创建、`lifecycle.Manager` 管理 HTTP/gRPC/后台任务                                   | [architecture.md](docs/architecture.md) |
| server/cmd/black-ridge/setup.go   | HTTP Server 装配：健康检查、访问日志、SPA fallback、WebSocket 终端路由                                            | [architecture.md](docs/architecture.md) |
| server/internal/transport/        | gRPC Server 框架 + Session/Template/Config 协议适配 + WebSocket 终端                                              | —                                       |
| server/internal/config/config.go  | 全局配置结构与校验（YAML + 反射式 env override）                                                                  | [architecture.md](docs/architecture.md) |
| server/internal/service/          | Tmux 会话管理 + 心跳注册 + 自动保存 + 配置文件读写（乐观并发控制）+ 业务模型（Session/SessionState/SessionTemplate）+ 模板持久化 + 状态持久化（SessionStateRepository） | [architecture.md](docs/architecture.md) |
| server/internal/service/session_state_fs.go | SessionStateRepository 接口及文件系统实现                                                                         | [architecture.md](docs/architecture.md) |
| server/internal/webstatic/        | `go:embed` 静态资源包，供 `http.FileServer` 提供 SPA 资源                                                         | [architecture.md](docs/architecture.md) |
| web/src/api/client.ts              | Agent API 客户端（使用生成 SDK 的直连路径方法，*2 后缀）                                                         | —                                       |
| Dockerfile + entrypoint.sh         | 多阶段构建 + 容器启动初始化                                                                                       | [architecture.md](docs/architecture.md) |

## 详细文档

| 文档                                         | 内容                                                                                          |
| -------------------------------------------- | --------------------------------------------------------------------------------------------- |
| [docs/architecture.md](docs/architecture.md) | 系统架构、启动流程、Tmux 管理、管线编排、状态持久化、心跳、自动保存、配置管理、模板系统、部署 |
