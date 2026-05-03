# Cradle AGENTS.md

## 职责

如同剧中 The Cradle 是 Mesa 地下主机房中存储和运转所有 Host 行为模式的核心设施，Cradle 是整个系统运转的基础设施层。它为 Manager 和 Agent 提供统一的日志、配置、HTTP 工具、中间件、通信协议、管线编排及数据持久化能力——所有模块共享的"行为模式"都在这里定义。

## 核心原则

- **接口统一** — 日志、HTTP 响应等通过 cradle 定义统一接口，Manager 与 Agent 共用同一套规范
- **IDL 驱动** — API 类型契约通过 Protobuf IDL 定义，buf 自动生成 Go 类型、gRPC stub、HTTP gateway，确保 Manager 与 Agent 数据结构强一致
- **安全默认** — 敏感值脱敏、原子写入、Bearer Token 鉴权等安全机制内置在工具层
- **零副作用** — 纯工具库，不依赖外部服务，不启动网络监听

## 依赖关系

- 依赖: Go 1.26 标准库 + hertz (HTTP 框架) + yaml.v3 (配置解析) + gRPC + protobuf + grpc-gateway + buf (IDL 代码生成)
- 被依赖: [behavior-panel](../../mesa-hub/behavior-panel/AGENTS.md), [black-ridge](../../sweetwater/black-ridge/AGENTS.md)

## 关键文件

| 路径                             | 职责                                                 | 文档同步                                   |
| -------------------------------- | ---------------------------------------------------- | ------------------------------------------ |
| api/buf.yaml + buf.gen.yaml      | buf 项目配置 + 代码生成配置（Go/gRPC/gateway/OpenAPI） | —                                          |
| api/proto/                       | Protobuf IDL 定义（agent/session/template/config/host/node/audit） | —                                          |
| api/gen/                         | buf 生成的 Go 类型 + gRPC stub + grpc-gateway handler（自动生成，勿手动编辑） | —                                          |
| api/gen/openapiv2/               | buf 生成的 OpenAPI/Swagger 文档（maze.swagger.json 为合并后的完整 spec） | 自动生成，勿手动编辑                       |
| api/gen/http/                    | openapi-generator 生成的 Go HTTP client（自动生成，勿手动编辑）     | —                                          |
| configutil/                      | 配置搜索/加载/合并/层定义/原子写入                    | [packages.md](docs/packages.md)            |
| httputil/                        | 统一 JSON 响应封装 + CORS 中间件                     | [packages.md](docs/packages.md)            |
| logutil/logger.go                | 结构化日志接口与 slog 实现                           | [packages.md](docs/packages.md)            |
| middleware/                      | Bearer Token 鉴权 + CORS 中间件（委托 httputil）     | [packages.md](docs/packages.md)            |
| gatewayutil/                     | grpc-gateway 响应包装器 + ServeMux 工厂 + 认证/审计 interceptor | [packages.md](docs/packages.md)            |
| pipeline/pipeline.go             | 管线步骤定义与层级过滤                               | [packages.md](docs/packages.md)            |
| protocol/                        | 领域模型：Agent 注册/心跳 + Host 部署 + 审计日志（JSON 持久化） | [packages.md](docs/packages.md)            |
| maskutil/mask.go                 | 敏感值脱敏                                           | [packages.md](docs/packages.md)            |
| storeutil/json_store.go          | 泛型 JSON 持久化存储                                 | [packages.md](docs/packages.md)            |

## 详细文档

| 文档                                 | 内容                  |
| ------------------------------------ | --------------------- |
| [docs/packages.md](docs/packages.md) | 各子包说明 + 导出 API |

## 代码生成流程

修改 proto 文件后，需重新生成 Go 类型和 HTTP client：

```bash
make gen    # 一键生成 proto + HTTP client
```

等价于依次执行：
1. `make gen-proto` — `buf generate` 生成 Go 类型 + gRPC stub + grpc-gateway + OpenAPI spec（`maze.swagger.json`）
2. `make gen-client` — 以 `maze.swagger.json` 为输入，`openapi-generator` 生成 Go HTTP client

前置依赖：`buf`、`openapi-generator`、`Java`
