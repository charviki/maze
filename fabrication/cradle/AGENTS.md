# Cradle

## 职责

Go 共享库，为 Manager 和 Agent 提供统一的日志、配置、HTTP 工具、中间件、通信协议（gRPC/Protobuf）、管线编排及数据持久化能力。

## 项目结构

Protobuf IDL 定义在 api/proto/，由 buf 生成 Go 类型/gRPC stub/gateway handler/OpenAPI spec（api/gen/），openapi-generator 再生成 Go HTTP client。工具包按 domain 划分：configutil（配置）、httputil（HTTP 封装）、logutil（日志）、middleware/gatewayutil（认证/审计）、grpcutil/lifecycle（gRPC 生命周期）、pipeline（管线）、protocol（领域模型）、maskutil（脱敏）、storeutil（持久化）。PostgreSQL 共享能力在 db/。

## 核心原则

- **IDL 驱动** — API 类型契约通过 Protobuf 定义，buf + openapi-generator 全链路生成
- **安全默认** — 敏感值脱敏、原子写入、Bearer Token 鉴权内置在工具层
- **零副作用** — 纯工具库，不依赖外部服务，不启动网络监听

## 命令

- `make gen` — 一键生成 proto + HTTP client（需要 buf + openapi-generator + Java）
- `make gen-proto` — 仅生成 Go 类型 + gRPC + gateway + OpenAPI
- `make gen-client` — 仅生成 Go HTTP client

## 依赖

- 依赖: Go 1.26 + yaml.v3 + gRPC + protobuf + grpc-gateway + buf
- 被依赖: [behavior-panel](../../mesa-hub/behavior-panel/AGENTS.md), [black-ridge](../../sweetwater/black-ridge/AGENTS.md)

## 详细文档

| 文档 | 内容 |
|------|------|
| [docs/packages.md](docs/packages.md) | 子包职责一览 |
