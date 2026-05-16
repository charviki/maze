# The Forge

## 职责

AI 与人共用的知识库服务，实现 Archive / Doc / DocLink 三层知识模型。一切皆文档——Doc 通过 parent_id 支持树形嵌套，通过 status 字段区分纯知识文档（null）与待办/任务（pending/active/done/failed）。支持全文搜索、Markdown 编辑和三栏布局。

## 项目结构

`cmd/the-forge/` 是服务入口，负责组装依赖并启动 gRPC + HTTP 双协议服务。`internal/` 遵循标准 Go 分层，参照 director-core 模式：`service/`（业务逻辑）、`transport/`（gRPC + HTTP handler）、`repository/postgres/`（PostgreSQL 持久化 + sqlc 生成代码）。`web/` 是 React 前端 SPA。

## 核心原则

- **gRPC + grpc-gateway 双协议** — KnowledgeService 通过 gRPC 实现，grpc-gateway 自动暴露 REST API；鉴权在 gRPC UnaryAuthInterceptor 层处理，HTTP 层不重复校验 JWT（与 director-core 一致）
- **grpcutil.NewServeMux** — 使用 `gatewayutil.NewServeMux()` 而非默认 `gwruntime.NewServeMux()`，确保 proto JSON 序列化输出零值字段
- **搜索用 PostgreSQL 全文索引** — Doc 全文搜索基于 `search_vector` GIN 索引 + `pg_trgm`，不依赖外部搜索引擎
- **三栏布局** — 前端左侧 Sidebar 选 Archive → 中间 DocTree 完整文档树（可拖拽宽）→ 右侧 Content 文档阅读/编辑

## 命令

```bash
# 构建
make build-go  # 或 cd the-mesa/the-forge && go build ./cmd/the-forge

# 单元测试
make test      # 或 go test ./...

# 前端开发服务器 (port 3001，代理 /api → localhost:8080)
pnpm --filter @maze/the-forge dev

# 前端构建
pnpm --filter @maze/the-forge build
```

## 依赖

- 依赖: [Cradle](../../fabrication/cradle/AGENTS.md)（grpcutil、gatewayutil、logutil、configutil、db、auth、middleware）
- 依赖: [Skin](../../fabrication/skin/AGENTS.md)（前端使用 fetchWithAuth、getCurrentUser、AppShell，API 类型使用生成的 SDK）
- 依赖: PostgreSQL（maze_forge 数据库）

## 详细文档

- [API 定义](../../fabrication/cradle/api/proto/maze/v1/knowledge.proto) — KnowledgeService 的 Protobuf IDL
- [数据库 Schema](internal/repository/postgres/migrations/00001_init.sql) — 单文件 DDL（archives + docs + doc_links）
- [MVP 实现方案](../../.claude/mvp-implementation-plan.md) — 最终数据模型与 API
