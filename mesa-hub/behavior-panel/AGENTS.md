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
| server/main.go                      | 入口：加载配置、创建 grpc-gateway ServeMux、注册 7 个 Service 的 gateway handler、gRPC interceptor chain（认证→分层令牌→审计）、优雅关闭 | [architecture.md#启动流程](docs/architecture.md)       |
| server/router.go                    | 顶层路由注册、健康检查、NoRoute 转发到 grpc-gateway ServeMux                                                                | —                                                      |
| server/biz/grpc/                    | gRPC Server 框架 + TemplateService 实现（Template CRUD + Config 代理到 Agent）                                              | —                                                      |
| server/biz/handler/                 | 依赖初始化 + 代理 Handler（SessionProxy + WebSocket）+ 审计日志                                                              | [architecture.md#代理网关](docs/architecture.md)       |
| server/biz/model/                   | NodeRegistry 节点注册表 + HostSpecManager 规格持久化                                                                         | [architecture.md#节点注册表](docs/architecture.md)     |
| server/biz/config/config.go         | 配置加载：YAML + 环境变量覆盖 + 校验；统一 Manager 元数据根目录与 Docker Agent 数据根目录语义                               | [architecture.md#配置](docs/architecture.md)           |
| server/biz/service/deploy.go        | 公共构建部署方法 BuildAndDeploy                                                                                              | [architecture.md#动态Host](docs/architecture.md)       |
| server/biz/reconciler/reconciler.go | Reconciler：启动恢复 + 60s 健康巡检 + deploying 保护窗口 + failed 自动重试                                                  | [architecture.md#Reconciler](docs/architecture.md)     |
| server/biz/builder/                 | 工具注册表 + Dockerfile 动态生成器（工具排序稳定化 + ToolsetImageTag 组合镜像标签）                                         | [architecture.md#动态Host](docs/architecture.md)       |
| server/biz/runtime/                 | HostRuntime 抽象 + Docker/K8s 实现 + 构建信号量                                                                             | [architecture.md#动态Host](docs/architecture.md)       |
| web/src/App.tsx                     | 前端入口：HostList + AgentPanel + RadarView + CreateHostDialog + HostLogPanel                                               | [architecture.md#前端](docs/architecture.md)           |
| web/src/api/                        | API 客户端：controller.ts（SDK NodeServiceApi/HostServiceApi）、agent.ts（SDK SessionServiceApi/TemplateServiceApi/ConfigServiceApi，Manager 代理路径） | —                                                      |
| nginx.conf + docker-compose.yml     | Nginx 路由 + 完整部署编排                                                                                                   | [architecture.md#部署拓扑](docs/architecture.md)       |

## 路由架构

Manager 采用 **Hertz HTTP 外壳 + grpc-gateway 嵌入** 的统一路由架构：

- **Hertz** 负责 WebSocket 路由（终端代理）和静态中间件（CORS、Auth）
- **grpc-gateway ServeMux** 通过 Hertz `NoRoute` 接管所有 REST API 请求，由 proto 注解驱动路由
- **gRPC Server** 运行在 `:9090`，gateway 进程内直连，经过 interceptor chain（认证→分层令牌→审计）
- 所有 REST API（Host/Node/Audit/Session/Template/Config/Agent）由 proto `google.api.http` 注解定义

## 详细文档

| 文档                                         | 内容                                                                            |
| -------------------------------------------- | ------------------------------------------------------------------------------- |
| [docs/architecture.md](docs/architecture.md) | Manager 架构、数据流、持久化策略、Host 编排引擎、Reconciler、部署拓扑、安全设计 |
