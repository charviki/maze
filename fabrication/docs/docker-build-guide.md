# Docker 构建规范

## 概述

Maze 项目使用三层 Docker 构建优化策略，所有 Dockerfile 必须遵循本规范。

## 三层优化

### 第 1 层：BuildKit Cache Mount（Dockerfile 内部加速）

在依赖安装步骤中使用 `--mount=type=cache`，让 BuildKit 缓存下载的依赖包。缓存跨构建持久保留，不受 `docker system prune` 影响。

**Go 依赖**：
```dockerfile
RUN --mount=type=cache,id=go-pkg-mod,target=/go/pkg/mod \
    go mod download
```

**pnpm 依赖**：
```dockerfile
RUN --mount=type=cache,id=pnpm-store,target=/root/.local/share/pnpm/store \
    pnpm install
```

**Cache ID 规范**：
- Go: `go-pkg-mod`（所有 Dockerfile 共用）
- pnpm: `pnpm-store`（所有 Dockerfile 共用）
- 统一 ID 确保跨 Dockerfile 缓存共享

### 第 2 层：拆分 COPY（依赖安装与源码解耦）

**核心原则：先复制 package.json/go.mod，安装依赖，再复制源码**。

这样修改源码时，依赖安装步骤直接命中 Docker 层缓存，不需要重新执行。

**前端 Dockerfile 模板**：
```dockerfile
# 1. 只复制依赖声明文件
COPY pnpm-workspace.yaml package.json pnpm-lock.yaml ./
COPY fabrication/skin/package.json fabrication/skin/package.json
COPY <子项目>/web/package.json <子项目>/web/package.json

# 2. 安装依赖（package.json 不变时命中 Docker 层缓存）
RUN --mount=type=cache,id=pnpm-store,target=/root/.local/share/pnpm/store \
    pnpm install

# 3. 复制源码（改源码只触发从这里开始的层重建）
COPY fabrication/skin/ fabrication/skin/
COPY <子项目>/web/ <子项目>/web/
```

**Go Dockerfile 模板**：
```dockerfile
COPY <子项目>/server/ ./

RUN sed -i 's|相对路径|容器内路径|g' go.mod

RUN --mount=type=cache,id=go-pkg-mod,target=/go/pkg/mod \
    go mod download

RUN CGO_ENABLED=0 go build -o /bin/<binary> .
```

### 第 3 层：供应商镜像（跨镜像依赖共享）

供应商镜像（molds）预装工具链，通过 `COPY --from` 按需组合到目标镜像中。

## 供应商镜像体系

### 目录结构

```
fabrication/
  deps/                    ← 原料清单（声明式依赖配置）
    go.txt                 ← Go 预装工具列表
    python.txt             ← Python 预装包列表
    js.txt                 ← npm 预装包列表
  molds/                   ← 模具（供应商 Dockerfile）
    Dockerfile.claude      ← Claude Code CLI
    Dockerfile.codex       ← Codex CLI
    Dockerfile.go          ← Go 1.24 + 预装工具
    Dockerfile.python      ← Python 3.11 + 预装包
    Dockerfile.node        ← Node.js + pnpm + 预装包
  Dockerfile.host          ← Host 装配镜像（多 target stage）
```

### 供应商镜像规范

1. 每个供应商镜像安装工具到 `/opt/<name>/`
2. 使用与 host 基础镜像相同的 Debian 版本（`node:22-bookworm-slim`），确保 glibc 兼容
3. 依赖列表由 `deps/*.txt` 声明式管理
4. 新增工具只需编辑 `deps/*.txt` 并重建供应商镜像

### 构建命令

```bash
# 构建所有供应商镜像
docker build -f fabrication/molds/Dockerfile.claude -t maze-deps-claude .
docker build -f fabrication/molds/Dockerfile.codex -t maze-deps-codex .
docker build -f fabrication/molds/Dockerfile.go -t maze-deps-go .
docker build -f fabrication/molds/Dockerfile.python -t maze-deps-python .
docker build -f fabrication/molds/Dockerfile.node -t maze-deps-node .

# 构建 Host 镜像（预定义组合）
docker build -f fabrication/Dockerfile.host --target full -t maze-host-full .

# 构建 Host 镜像（动态组合，通过 POC 工具）
cd fabrication/cmd/build-host-poc && go run . maze-host-agent1 claude,go,python
```

### 如何在 Dockerfile 中引用供应商镜像

在目标 Dockerfile 中使用 `COPY --from` 从供应商镜像复制工具：

```dockerfile
# 引用 Claude Code 供应商镜像
COPY --from=maze-deps-claude:latest /opt/claude /opt/claude
ENV PATH="/opt/claude/bin:${PATH}"

# 引用 Go 工具链供应商镜像
COPY --from=maze-deps-go:latest /opt/go /opt/go
ENV PATH="/opt/go/bin:${PATH}"
ENV GOROOT=/opt/go
```

**注意事项**：
- 供应商镜像必须预先构建（上面的构建命令）
- `COPY --from` 的镜像名格式：`maze-deps-<工具名>:latest`
- 源路径统一为 `/opt/<工具名>/`
- 复制后需要手动设置 `ENV PATH`（因为 ENV 不会从源镜像继承）

### 如何新增一个供应商模具

假设要新增一个 Rust 工具链供应商：

**步骤 1：创建依赖配置文件**

```bash
# fabrication/deps/rust.txt
rust-analyzer@latest
cargo-watch@latest
```

**步骤 2：创建供应商 Dockerfile**

```dockerfile
# fabrication/molds/Dockerfile.rust
FROM public.ecr.aws/docker/library/node:22-bookworm-slim

RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --prefix /opt/rust

ENV PATH="/opt/rust/bin:${PATH}"

COPY fabrication/deps/rust.txt /tmp/rust-deps.txt
RUN xargs -a /tmp/rust-deps.txt -I {} cargo install {} || true
RUN rm -f /tmp/rust-deps.txt
```

**步骤 3：构建并验证**

```bash
docker build -f fabrication/molds/Dockerfile.rust -t maze-deps-rust .
docker run --rm maze-deps-rust rustc --version
```

**步骤 4：在 Dockerfile.host 中添加组合**

在 `fabrication/Dockerfile.host` 中添加对应的扩展层和组合层：

```dockerfile
# 扩展层
FROM base AS with-rust
COPY --from=maze-deps-rust:latest /opt/rust /opt/rust
ENV PATH="/opt/rust/bin:${PATH}"

# 在 full 层中添加
COPY --from=maze-deps-rust:latest /opt/rust /opt/rust
```

**步骤 5：更新 POC 工具注册表**

在 `fabrication/cmd/build-host-poc/main.go` 的 `ToolRegistry` 中添加：

```go
"rust": {
    Image:      "maze-deps-rust:latest",
    SourcePath: "/opt/rust",
    DestPath:   "/opt/rust",
    BinPaths:   []string{"/opt/rust/bin"},
    EnvVars:    map[string]string{},
},
```

### 如何更新预装依赖

修改 `fabrication/deps/*.txt` 后，重新构建对应的供应商镜像即可：

```bash
# 例如：给 Go 工具链新增一个工具
echo "golang.org/x/tools/cmd/stringer@latest" >> fabrication/deps/go.txt
docker build -f fabrication/molds/Dockerfile.go -t maze-deps-go .
```

## 新增 Dockerfile 检查清单

新增或修改 Dockerfile 时，必须遵循以下规范：

### 必做

- [ ] **Cache Mount**：`pnpm install` 和 `go mod download` 使用 `--mount=type=cache`，ID 与其他 Dockerfile 一致
- [ ] **拆分 COPY**：先复制 `package.json`/`go.mod`，执行依赖安装，再复制源码
- [ ] **供应商镜像**：如需 Claude Code 等工具，使用 `COPY --from=maze-deps-<name>:latest`，不直接 `npm install -g`
- [ ] **构建验证**：修改后必须 `docker build` 验证通过

### 禁止

- ❌ 不要在 Dockerfile 中 `npm install -g @anthropic-ai/claude-code`（使用供应商镜像）
- ❌ 不要在 `&&` 链中混用 `--mount=type=cache`（它是 RUN 指令的修饰符，不是 shell 命令）
- ❌ 不要用 `# syntax=docker/dockerfile:1`（需要拉取 Docker Hub 镜像，网络不通时构建失败）

### Python 供应商镜像注意事项

- 必须使用与 host 相同的基础镜像（`node:22-bookworm-slim`），确保 glibc 兼容
- 不能使用 venv（venv 的 symlink 指向外部路径，COPY 到新镜像后失效）
- 必须复制真正的 Python 二进制和标准库（`/usr/bin/python3.11` + `/usr/lib/python3.11`）


