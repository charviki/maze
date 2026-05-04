# Behavior Panel AGENTS.md

## 职责

如同剧中技术员在 Mesa 控制中心通过行为面板监控和管理所有 Host，Behavior Panel 是人类开发者管控所有 Agent 节点的 Web 入口。作为系统的代理网关和 Host 编排引擎，它负责注册和管理 Agent 节点，动态编排 Host 的完整生命周期（创建→部署→监控→恢复→销毁），代理前端到 Agent 的所有 HTTP 和 WebSocket 请求，并记录审计日志确保每一步操作可追溯。

## 核心原则

- **代理网关模式** — 前端不直连 Agent，所有请求经 Manager 代理，保持可观测性
- **审计优先** — 每次代理操作都记录审计日志（operator=frontend, action, target_node, result）
- **SSRF 防护** — 代理目标 URL 必须通过协议和内网 IP 校验，Docker 环境下可配置放行
- **异步编排** — Host 创建为异步流程（202 Accepted），后台构建部署，前端轮询状态
- **声明式恢复** — HostSpec 持久化 + Reconciler 启动恢复 + 健康巡检，确保实际状态趋近期望状态
- **优雅关闭** — 监听 SIGINT/SIGTERM，依次停止 Reconciler、刷盘 HostSpec/NodeRegistry、关闭审计日志文件

## 依赖关系

- 依赖: [cradle](../../fabrication/cradle/AGENTS.md) (Go 共享库: logutil, httputil, protocol, configutil, middleware, gatewayutil), [@maze/fabrication](../../fabrication/skin/AGENTS.md) (UI 组件库)
- 被依赖: 无（是控制中心，不被其他模块依赖）

## 关键文件

| 路径                                | 职责                                                                                                                        | 文档同步                                               |
| ----------------------------------- | --------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------ |
| server/cmd/behavior-panel/main.go   | 入口：装配 grpc-gateway / gRPC interceptor chain / PermissionService，并把 HTTP + gRPC 交给 `lifecycle.Manager` 管理       | [architecture.md#启动流程](docs/architecture.md)       |
| server/cmd/behavior-panel/setup.go  | HTTP Server 装配、健康检查、访问日志 middleware，以及 `repository/audit`、`agentclient`、WebSocket 代理等基础依赖初始化     | [architecture.md#代理网关](docs/architecture.md)       |
| server/internal/service/permission_service.go | 权限闭环：申请单创建、审批、撤销、grant 回收等核心业务状态机                                                    | [auth-overview.md](docs/auth-overview.md)              |
| server/internal/service/permission_domain.go  | 权限业务模型、状态枚举与业务错误：申请单、grant、审计对象以及跨层共享的领域语义                                   | [auth-overview.md](docs/auth-overview.md)              |
| server/internal/service/permission_store.go   | 权限存储边界：由 service owning 的 store/tx 接口与写入参数定义                                                   | [auth-overview.md](docs/auth-overview.md)              |
| server/internal/agentclient/        | Agent gRPC 客户端层：连接池 + Session/Template/Config 代理，承接 Manager → Agent 的出站调用                               | [architecture.md#代理请求流程http--grpc-转发](docs/architecture.md) |
| server/internal/repository/audit/   | 审计日志 repository 实现：append-only JSON Lines 持久化、分页、过滤、优雅关闭                                               | [architecture.md#审计日志auditlogrepository](docs/architecture.md) |
| server/internal/transport/permission_service.go | PermissionService gRPC/gateway 适配：proto 转换、错误映射、主体提取                                              | [auth-integration.md](docs/auth-integration.md)        |
| server/internal/repository/postgres/ | Behavior Panel 自有 PostgreSQL 实现：migration、sqlc query、生成代码与 bootstrap 持久化细节                            | [auth-overview.md](docs/auth-overview.md)              |
| server/internal/transport/          | gRPC Server 框架 + HTTP/WebSocket 协议适配；不再承载审计存储或 Agent 出站客户端实现                                         | [architecture.md#代理网关](docs/architecture.md)       |
| server/internal/service/host_node.go + server/internal/repository/file/ | Host/Node 域分层：`service` owning `Node`、`BuildHostInfo()` 与 repository interface，`repository/file` 提供 NodeRegistry/HostSpec 的 JSON 持久化实现 | [architecture.md#节点注册表](docs/architecture.md)     |
| server/internal/config/config.go    | 配置加载：YAML + 反射式 env override + 校验；统一 Manager 元数据根目录与 Docker Agent 数据根目录语义                       | [architecture.md#配置](docs/architecture.md)           |
| server/internal/service/deploy.go   | 公共构建部署方法 BuildAndDeploy                                                                                              | [architecture.md#动态Host](docs/architecture.md)       |
| server/internal/reconciler/reconciler.go | Reconciler：启动恢复 + 60s 健康巡检 + deploying 保护窗口 + failed 自动重试                                             | [architecture.md#Reconciler](docs/architecture.md)     |
| server/internal/hostbuilder/        | Host 构建规划：工具注册表 + Dockerfile 动态生成器（工具排序稳定化 + ToolsetImageTag 组合镜像标签）                         | [architecture.md#动态Host](docs/architecture.md)       |
| server/internal/runtime/            | HostRuntime 抽象 + Docker/K8s 实现 + 构建信号量                                                                             | [architecture.md#动态Host](docs/architecture.md)       |
| web/src/App.tsx                     | 前端入口：HostList + AgentPanel + RadarView + CreateHostDialog + HostLogPanel                                               | [architecture.md#前端](docs/architecture.md)           |
| web/src/api/                        | API 客户端：controller.ts（SDK NodeServiceApi/HostServiceApi）、agent.ts（SDK SessionServiceApi/TemplateServiceApi/ConfigServiceApi，Manager 代理路径） | —                                                      |
| nginx.conf + docker-compose.yml     | Nginx 路由 + 完整部署编排                                                                                                   | [architecture.md#部署拓扑](docs/architecture.md)       |

## 路由架构

Manager 采用 **`net/http` + grpc-gateway + gRPC** 的统一路由架构：

- **`net/http.ServeMux`** 负责 `/health`、WebSocket 路由和访问日志/CORS/Auth 中间件
- **grpc-gateway ServeMux** 作为 `/` 兜底 handler 接管所有 REST API 请求，由 proto 注解驱动路由
- **gRPC Server** 运行在 `:9090`，gateway 进程内直连，经过 interceptor chain（认证→分层令牌→审计）
- 所有 REST API（Host/Node/Audit/Session/Template/Config/Agent）由 proto `google.api.http` 注解定义

## 详细文档

| 文档                                         | 内容                                                                            |
| -------------------------------------------- | ------------------------------------------------------------------------------- |
| [docs/architecture.md](docs/architecture.md) | Manager 架构、数据流、持久化策略、Host 编排引擎、Reconciler、部署拓扑、安全设计 |
| [docs/auth-overview.md](docs/auth-overview.md) | 权限系统对象模型、API 资源、认证与授权边界                                      |
| [docs/auth-integration.md](docs/auth-integration.md) | 其他调用方如何基于 `admin` 和 `subject_key` 接入当前权限系统                    |
| [docs/auth-operations.md](docs/auth-operations.md) | 权限申请、审批、撤销、过期回收、审计与排障操作                                  |
