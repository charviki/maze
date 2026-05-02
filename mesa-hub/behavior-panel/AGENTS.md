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
- 依赖: [cradle](../../fabrication/cradle/AGENTS.md) (Go 共享库: logutil, httputil, protocol, configutil, middleware), [@maze/fabrication](../../fabrication/skin/AGENTS.md) (UI 组件库)
- 被依赖: 无（是控制中心，不被其他模块依赖）

## 关键文件
| 路径 | 职责 | 文档同步 |
|------|------|----------|
| server/main.go | 入口：加载配置、启动 HTTP+gRPC 双协议服务、优雅关闭 | [architecture.md#启动流程](docs/architecture.md) |
| server/biz/grpc/server.go | gRPC Server 框架（Node/Host/Audit/Agent Service + TemplateService） | — |
| server/biz/grpc/template.go | TemplateService gRPC 实现（Template CRUD + Config 代理到 Agent） | — |
| server/router.go | 顶层路由注册、健康检查、NoRoute 重定向 | [api.md](docs/api.md) |
| server/biz/router/register.go | 所有 API 路由定义、Store/Handler/Reconciler 初始化、中间件注册 | [api.md](docs/api.md) |
| server/biz/handler/node.go | 节点注册/心跳/查询/删除 Handler | [api.md#节点管理](docs/api.md) |
| server/biz/handler/host.go | Host 异步生命周期管理（CreateHost 202/ListHosts/GetHost/GetBuildLog/GetRuntimeLog/DeleteHost） | [api.md#Host管理](docs/api.md) |
| server/biz/handler/session_proxy.go | SessionProxyHandler 结构定义和构造函数 | [architecture.md#代理网关](docs/architecture.md) |
| server/biz/handler/proxy_http.go | HTTP 代理实现、审计日志、SSRF 校验 | [architecture.md#代理网关](docs/architecture.md) |
| server/biz/handler/proxy_ws.go | WebSocket 双向代理（前端-Manager-Agent） | [architecture.md#WebSocket终端](docs/architecture.md) |
| server/biz/handler/audit_logger.go | 审计日志：append-only JSON Lines 持久化 | [architecture.md#审计日志](docs/architecture.md) |
| server/biz/model/node.go | NodeRegistry：节点注册表、dirty flush 持久化 | [architecture.md#节点注册表](docs/architecture.md) |
| server/biz/model/host_spec.go | HostSpecManager：Host 规格持久化（CRUD + dirty flush + GetMerged/ListMerged 合并视图） | [architecture.md#HostSpec持久化](docs/architecture.md) |
| server/biz/service/deploy.go | 公共构建部署方法 BuildAndDeploy（消除 handler/reconciler 的重复逻辑） | [architecture.md#动态Host](docs/architecture.md) |
| server/biz/reconciler/reconciler.go | Reconciler：启动恢复 + 60s 健康巡检 + deploying 保护窗口 + failed 自动重试 | [architecture.md#Reconciler](docs/architecture.md) |
| server/biz/config/config.go | 配置加载：YAML + 环境变量覆盖 + 校验；统一 Manager 元数据根目录与 Docker Agent 数据根目录语义 | [architecture.md#配置](docs/architecture.md) |
| server/biz/builder/registry.go | 工具配置注册表（claude/codex/go/python/node） | [architecture.md#动态Host](docs/architecture.md) |
| server/biz/builder/host.go | Dockerfile 动态生成器（工具排序稳定化 + ToolsetImageTag 组合镜像标签） | [architecture.md#动态Host](docs/architecture.md) |
| server/biz/runtime/runtime.go | HostRuntime 抽象接口（DeployHost/RemoveHost/InspectHost/GetRuntimeLogs/IsHealthy） | [architecture.md#动态Host](docs/architecture.md) |
| server/biz/runtime/docker.go | Docker 运行时（docker CLI + socket + BuildKit + 组合镜像缓存 + Dockerfile hash 校验；Agent 数据目录固定到 `agents/<host>`） | [architecture.md#动态Host](docs/architecture.md) |
| server/biz/runtime/kubernetes.go | K8s 运行时（client-go Deployment/Service + BuildKit + 组合镜像缓存 + Dockerfile hash 校验） | [architecture.md#动态Host](docs/architecture.md) |
| server/biz/runtime/semaphore.go | 构建信号量（并发限流，防止镜像重建风暴） | [architecture.md#动态Host](docs/architecture.md) |
| web/src/App.tsx | 前端入口：HostList + AgentPanel + RadarView + CreateHostDialog + HostLogPanel | [architecture.md#前端](docs/architecture.md) |
| web/src/api/controller.ts | Manager API 客户端（listHosts/getHost/createHost/getHostBuildLog/getHostRuntimeLog/deleteHost） | [api.md#Host管理](docs/api.md) |
| web/src/api/agent.ts | 通过 Manager 代理的 Agent API 客户端 | [api.md](docs/api.md) |
| web/src/components/NodeList.tsx | Host 列表组件（数据源切到 listHosts + 全生命周期状态视觉） | [architecture.md#前端](docs/architecture.md) |
| nginx.conf | Nginx 路由配置（Portal 首页 + Behavior Panel SPA + API 代理 + WebSocket） | [architecture.md#部署拓扑](docs/architecture.md) |
| docker-compose.yml | 完整部署编排：web + agent-manager（Manager 元数据挂载到 `docker/`，Agent 数据挂载到 `docker/agents/`） | [architecture.md#部署拓扑](docs/architecture.md) |

## 详细文档
| 文档 | 内容 |
|------|------|
| [docs/architecture.md](docs/architecture.md) | Manager 架构、数据流、持久化策略、Host 编排引擎、Reconciler、部署拓扑、安全设计 |
| [docs/api.md](docs/api.md) | 所有 HTTP API 端点：请求参数、响应格式、认证要求 |
