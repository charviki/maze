# Go 模块创建指南

## 创建

- 使用 `.skills/go-server-scaffold/` 作为起点，替换 `ServiceName`、`ModulePath`、`EnvPrefix`、`CradleReplace`
- 目录结构保持 `server/` 为模块根，入口放在 `cmd/<service>/`
- 配置优先复用 `cradle/configutil.ServerConfig`，避免重新定义监听地址、鉴权和 CORS 字段

## 编译

- 将新模块路径加入根 `Makefile` 的 `MODULES` 列表
- 在模块目录执行 `go mod tidy`
- 执行 `go build ./...` 确认所有包都可编译

## 测试

- 单元测试与源码同目录放置，命名为 `*_test.go`
- 跨模块集成测试放到 `fabrication/tests/integration/`
- 若暴露 gRPC 能力，优先使用官方 client 或健康检查服务做最小验证

## 检查

- 本地运行 `make build-go`
- 本地运行 `make lint`
- 本地运行 `make test`

## 部署

- Dockerfile 遵循 `fabrication/docs/docker-build-guide.md`
- 若需要嵌入前端产物，优先将 `go:embed` 放在独立包，而不是 `cmd/` 入口目录
- Kubernetes overlay、docker compose 等运行时清单放在模块外层目录维护

## 可观测性

- HTTP 和 gRPC 统一交给 `lifecycle.Manager` 管理
- 访问日志使用标准 `net/http` middleware 记录
- 优先通过 `logutil.New("<component>")` 输出结构化 JSON 日志

## 文档

- 更新对应模块的 `AGENTS.md` 关键文件表
- 涉及架构变化时同步更新模块 `docs/architecture.md`
- 若新增模块，更新根 `AGENTS.md` 模块索引
