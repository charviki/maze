---
name: "go-server-scaffold"
description: "Scaffolds a new Go server module from templates via a render script. Invoke when user wants a new minimal Go service created in this workspace."
---

# Go Server Scaffold

## 目的

为 Maze 仓库中的新 Go 服务生成最小可运行骨架，并将职责明确拆分为：

- `SKILL.md` 负责收集输入、说明约束、指导后续接入
- `bin/render.py` 负责确定性渲染模板、创建目录、替换变量、落盘文件

默认遵循以下约束：

- 使用 `cmd/` + `internal/` 标准目录结构
- 使用 `net/http` + `http.ServeMux`
- 使用 gRPC 官方健康检查服务作为最小示例
- 使用 `cradle/lifecycle` 统一管理 HTTP + gRPC 生命周期
- 使用 `cradle/configutil.ServerConfig` + `ApplyEnvOverrides` 处理配置

## 输入

- `ServiceName`：服务目录名，也是 `cmd/<service>/` 名称
- `ModulePath`：`go.mod` 的 module 路径
- `EnvPrefix`：环境变量前缀，例如 `MAZE_FOO`
- `OutputDir`：目标模块目录，例如 `/path/to/repo/my-service`
- `CradleReplace`：可选，默认指向仓库内 `fabrication/cradle`

## 生成内容

模板文件平铺于 `template/` 目录，渲染脚本通过 `TEMPLATE_MAP` 显式声明每个模板的输出路径：

| 模板文件 | 输出路径 |
|---|---|
| `go.mod.tmpl` | `go.mod` |
| `cmd_main.go.tmpl` | `cmd/<ServiceName>/main.go` |
| `cmd_setup.go.tmpl` | `cmd/<ServiceName>/setup.go` |
| `config.go.tmpl` | `internal/config/config.go` |
| `health_service.go.tmpl` | `internal/service/health_service.go` |
| `grpc_server.go.tmpl` | `internal/transport/grpc_server.go` |
| `http_health.go.tmpl` | `internal/transport/http_health.go` |

## 执行方式

### Skill 交互职责

1. 向用户收集：
   - `ServiceName`
   - `ModulePath`
   - `EnvPrefix`
   - `OutputDir`
   - 是否需要覆盖已有目录
2. 检查这些输入是否完整、命名是否合理
3. 调用渲染脚本，不要手工逐文件创建模板内容
4. 渲染完成后，再提示后续接入和验证动作

### 渲染脚本

固定使用以下脚本：

```bash
python3 .skills/go-server-scaffold/bin/render.py \
  --output-dir <OutputDir> \
  --service-name <ServiceName> \
  --module-path <ModulePath> \
  --env-prefix <EnvPrefix>
```

如果需要显式指定 `cradle` replace：

```bash
python3 .skills/go-server-scaffold/bin/render.py \
  --output-dir <OutputDir> \
  --service-name <ServiceName> \
  --module-path <ModulePath> \
  --env-prefix <EnvPrefix> \
  --cradle-replace /absolute/path/to/fabrication/cradle
```

如需覆盖已有目录：

```bash
python3 .skills/go-server-scaffold/bin/render.py ... --force
```

## 生成后建议

1. 按项目约定把新模块路径加入根 `Makefile` 的 `MODULES` 列表
2. 在新模块目录执行：

```bash
go mod tidy
go build ./...
go run ./cmd/{{.ServiceName}}
```

3. 验证：

```bash
curl http://localhost:8080/health
grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check
```

## 失败策略

- 若 `OutputDir` 已存在且未传 `--force`，脚本必须直接失败，避免覆盖已有模块
- 若模板变量缺失，脚本必须直接失败，避免落盘残留 `{{...}}`
- 若渲染成功但后续 `go mod tidy` / `go build` 失败，由 Skill 向用户报告失败点，而不是继续猜测修复

## 约束

- 不生成 `imagebuilder/`、`reconciler/`、`runtime/` 等业务专属目录
- 不生成 Makefile 或 `.golangci.yml`，统一继承仓库根配置
- 默认通过 `replace` 指向仓库内 `fabrication/cradle`
- 不要让模型手工逐文件写模板内容，优先走脚本渲染
