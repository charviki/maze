# Cradle AGENTS.md

## 职责

如同剧中 The Cradle 是 Mesa 地下主机房中存储和运转所有 Host 行为模式的核心设施，Cradle 是整个系统运转的基础设施层。它为 Manager 和 Agent 提供统一的日志、配置、HTTP 工具、中间件、通信协议、管线编排及数据持久化能力——所有模块共享的"行为模式"都在这里定义。

## 核心原则
- **接口统一** — 日志、HTTP 响应等通过 cradle 定义统一接口，Manager 与 Agent 共用同一套规范
- **协议驱动** — Agent 注册、心跳、审计等通信协议集中在 protocol 包中定义，确保 Manager 与 Agent 数据结构一致
- **安全默认** — 敏感值脱敏、原子写入、Bearer Token 鉴权等安全机制内置在工具层
- **零副作用** — 纯工具库，不依赖外部服务，不启动网络监听

## 依赖关系
- 依赖: Go 1.24 标准库 + hertz (HTTP 框架) + yaml.v3 (配置解析)
- 被依赖: [behavior-panel](../../mesa-hub/behavior-panel/AGENTS.md), [black-ridge](../../sweetwater/black-ridge/AGENTS.md)

## 关键文件
| 路径 | 职责 | 文档同步 |
|------|------|----------|
| configutil/loader.go | YAML 配置文件搜索与加载 | [packages.md#configutil](docs/packages.md) |
| configutil/merge.go | 多配置层合并 | [packages.md#configutil](docs/packages.md) |
| configutil/config_layer.go | 配置层数据结构定义 | [packages.md#configutil](docs/packages.md) |
| configutil/atomic_write.go | 原子文件写入 | [packages.md#configutil](docs/packages.md) |
| httputil/response.go | 统一 JSON 响应封装 | [packages.md#httputil](docs/packages.md) |
| httputil/cors.go | CORS 中间件与 WebSocket Origin 校验 | [packages.md#httputil](docs/packages.md) |
| logutil/logger.go | 结构化日志接口与 slog 实现 | [packages.md#logutil](docs/packages.md) |
| middleware/auth.go | Bearer Token 鉴权中间件 | [packages.md#middleware](docs/packages.md) |
| middleware/cors.go | CORS 中间件（委托 httputil） | [packages.md#middleware](docs/packages.md) |
| pipeline/pipeline.go | 管线步骤定义与层级过滤 | [packages.md#pipeline](docs/packages.md) |
| protocol/register.go | Agent 注册与心跳协议 | [packages.md#protocol](docs/packages.md) |
| protocol/host.go | Host 部署协议（HostDeploySpec、ToolConfig、ResourceLimits）+ Host 生命周期协议（HostStatus 常量、HostSpec 结构体、HostInfo 合并视图、CreateHostRequest） | [packages.md#protocol](docs/packages.md) |
| protocol/audit.go | 审计日志条目定义 | [packages.md#protocol](docs/packages.md) |
| maskutil/mask.go | 敏感值脱敏 | [packages.md#maskutil](docs/packages.md) |
| storeutil/json_store.go | 泛型 JSON 持久化存储 | [packages.md#storeutil](docs/packages.md) |

## 详细文档
| 文档 | 内容 |
|------|------|
| [docs/packages.md](docs/packages.md) | 各子包说明 + 导出 API |
