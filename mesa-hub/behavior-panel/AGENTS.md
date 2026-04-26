# Behavior Panel AGENTS.md

## 职责

如同剧中技术员在 Mesa 控制中心通过行为面板监控和管理所有 Host，Behavior Panel 是人类开发者管控所有 Agent 节点的 Web 入口。作为系统的代理网关，它负责注册和管理 Agent 节点，代理前端到 Agent 的所有 HTTP 和 WebSocket 请求，并记录审计日志确保每一步操作可追溯。

## 核心原则
- **代理网关模式** — 前端不直连 Agent，所有请求经 Manager 代理，保持可观测性
- **审计优先** — 每次代理操作都记录审计日志（operator=frontend, action, target_node, result）
- **SSRF 防护** — 代理目标 URL 必须通过协议和内网 IP 校验，Docker 环境下可配置放行
- **优雅关闭** — 监听 SIGINT/SIGTERM，依次停止 HTTP 服务、刷盘节点数据、关闭审计日志文件

## 依赖关系
- 依赖: [cradle](../cradle/AGENTS.md) (Go 共享库: logutil, httputil, protocol, configutil, middleware), [@maze/fabrication](../../fabrication/AGENTS.md) (UI 组件库)
- 被依赖: 无（是控制中心，不被其他模块依赖）

## 关键文件
| 路径 | 职责 | 文档同步 |
|------|------|----------|
| server/main.go | 入口：加载配置、启动 Hertz HTTP 服务、优雅关闭 | [architecture.md#启动流程](docs/architecture.md) |
| server/router.go | 顶层路由注册、健康检查、NoRoute 重定向 | [api.md](docs/api.md) |
| server/biz/router/register.go | 所有 API 路由定义、Store/Handler 初始化、中间件注册 | [api.md](docs/api.md) |
| server/biz/handler/node.go | 节点注册/心跳/查询/删除 Handler | [api.md#节点管理](docs/api.md) |
| server/biz/handler/session_proxy.go | SessionProxyHandler 结构定义和构造函数 | [architecture.md#代理网关](docs/architecture.md) |
| server/biz/handler/proxy_http.go | HTTP 代理实现、审计日志、SSRF 校验 | [architecture.md#代理网关](docs/architecture.md) |
| server/biz/handler/proxy_ws.go | WebSocket 双向代理（前端-Manager-Agent） | [architecture.md#WebSocket终端](docs/architecture.md) |
| server/biz/handler/audit_logger.go | 审计日志：append-only JSON Lines 持久化 | [architecture.md#审计日志](docs/architecture.md) |
| server/biz/model/node.go | NodeRegistry：节点注册表、dirty flush 持久化 | [architecture.md#节点注册表](docs/architecture.md) |
| server/biz/config/config.go | 配置加载：YAML + 环境变量覆盖 + 校验 | [architecture.md#配置](docs/architecture.md) |
| web/src/App.tsx | 前端入口：节点列表 + AgentPanel + RadarView | [architecture.md#前端](docs/architecture.md) |
| web/src/api/agent.ts | 通过 Manager 代理的 Agent API 客户端 | [api.md](docs/api.md) |
| web/src/api/controller.ts | Manager 节点管理 API 客户端 | [api.md#节点管理](docs/api.md) |
| nginx.conf | Nginx 反向代理配置（API 代理 + WebSocket 支持） | [architecture.md#部署拓扑](docs/architecture.md) |
| docker-compose.yml | 完整部署编排：web + agent-manager + 多 Agent 实例 | [architecture.md#部署拓扑](docs/architecture.md) |

## 详细文档
| 文档 | 内容 |
|------|------|
| [docs/architecture.md](docs/architecture.md) | Manager 架构、数据流、持久化策略、部署拓扑、安全设计 |
| [docs/api.md](docs/api.md) | 所有 HTTP API 端点：请求参数、响应格式、认证要求 |
