# Black Ridge 架构文档

## 系统架构

Black Ridge 是 Maze 系统中的 **Agent 节点**，运行在每个工作容器内。它的核心角色是：

- **会话管理者** — 通过 Tmux 管理多个 AI CLI 工具（Claude Code、Codex、Bash Shell）的独立会话
- **管线执行器** — 按三层管线（system/template/user）编排会话的初始化与恢复流程
- **状态持久化器** — 定期保存会话快照（环境变量、终端内容、管线定义），支持崩溃恢复
- **控制面上报者** — 向 Manager（behavior-panel）注册并定期上报心跳，携带完整的 Agent 状态

Agent 以单二进制部署，前端构建产物通过 `go:embed` 嵌入，实现零依赖的 SPA 部署。

### 组件关系

```
┌──────────────────────────────────────────────────────────────┐
│                    cmd/black-ridge/main.go                    │
│  配置加载 → grpc-gateway ServeMux → gRPC Server → lifecycle   │
└──────────┬──────────┬──────────┬──────────┬──────────────────┘
           │          │          │          │
     ┌─────▼──┐ ┌─────▼──┐ ┌────▼─────┐ ┌▼──────────┐
     │ Tmux   │ │Heartbeat│ │ AutoSave │ │ gRPC      │
     │Service │ │ Service │ │ Service  │ │ Server    │
     └────┬───┘ └────┬───┘ └────┬─────┘ └──┬────────┘
          │          │          │            │
          ▼          ▼          ▼            ▼
     ┌─────────────────────────────┐ ┌──────────────┐
     │    tmux (UNIX socket)       │ │ grpc-gateway  │
     │  PTY / session / environment│ │ (ServeMux fwd)│
     └─────────────────────────────┘ └──────────────┘
```

## 启动流程

`main.go` 按以下顺序初始化：

1. **日志初始化** — 创建 slog 实例（`logutil.New("agent")`），统一 JSON 输出
2. **配置加载** — `config.LoadFromExe()` 搜索可执行文件所在目录及当前工作目录的 YAML 配置文件，执行解析 → 环境变量覆盖 → 校验
3. **HTTP 服务创建** — `http.Server` + `http.ServeMux` 创建 HTTP 服务
4. **TmuxService 初始化** — 注入 TmuxConfig、状态目录、日志、TrustBootstrapper（Claude 信任注入）
5. **LocalConfigStore 初始化** — 加载/创建 `~/.maze/config.json`
6. **grpc-gateway ServeMux 创建** — `gatewayutil.NewServeMux()` 创建带自定义响应格式包装器的 ServeMux
7. **路由注册** — `setup.go` 装配访问日志/CORS、WebSocket 端点、SPA 静态文件和 grpc-gateway 兜底路由
8. **gRPC Server 启动** — 创建 gRPC Server，注册 3 个 Service（Session/Template/Config），应用 `UnaryAuthInterceptor` 认证拦截
9. **Service 注册到 gateway** — 将 3 个 Service 注册到 grpc-gateway ServeMux（进程内直连，不走网络）
10. **心跳服务启动** — `HeartbeatService.Start()` 在独立 goroutine 中运行，向 Manager 注册并定期心跳
11. **自动保存服务启动** — `AutoSaveService.Start()` 在独立 goroutine 中定期保存所有活跃会话的管线状态
12. **统一启停** — `lifecycle.Manager` 监听 SIGINT/SIGTERM，统一关闭 HTTP、gRPC、Heartbeat、AutoSave
13. **安全警告** — 若 `server.auth_token` 或 `controller.auth_token` 为空，打印 DEV 模式警告

## Tmux 会话管理

### 接口定义

`TmuxService` 接口定义了完整的会话管理能力：

- `ListSessions()` — 列出所有活跃会话
- `CreateSession(name, command, workingDir, configs, restoreStrategy, templateID, restoreCommand)` — 创建会话
- `KillSession(name)` — 终止会话
- `GetSession(name)` — 获取单个会话详情
- `CapturePane(name, lines)` — 捕获终端输出
- `SendKeys(name, command)` — 发送按键
- `SendSignal(name, signal)` — 发送信号
- `AttachSession(name, rows, cols)` — 附加到 PTY
- `ResizeSession(name, rows, cols)` — 调整终端尺寸
- `GetSessionEnv(name)` — 获取环境变量
- `ExecutePipeline(sessionName, pipeline)` — 执行管线
- `BuildPipeline(workingDir, command, configs)` — 构建管线
- `SavePipelineState(...)` / `SaveAllPipelineStates()` — 保存状态
- `GetSavedSessions()` / `GetSessionState(name)` — 读取状态
- `RestoreSession(name)` — 恢复会话
- `DeleteSessionWorkspace(name, workspaceRoot)` — 清理工作目录

### 实现细节

- **tmux 命令构建** — `tmuxArgs()` 自动添加 `-u`（UTF-8），可选 `-L`（socket 路径）
- **环境变量** — 所有 tmux 命令统一设置 `TERM=xterm-256color`、`COLORTERM=truecolor`、`LANG=C.UTF-8`
- **prompt 检测** — `waitForPrompt()` 轮询终端内容（50ms 间隔，5s 超时），匹配 `[#$>]`、`[❯➜λ%]` 等提示符模式
- **路径展开** — `expandPath()` 将 `~/` 开头路径展开为用户主目录绝对路径
- **错误映射** — tmux 的各种错误输出统一映射为 `ErrSessionNotFound` sentinel error

### TrustBootstrapper

`ClaudeTrustBootstrapper` 在创建会话时向 `~/.claude.json` 注入工作目录信任配置（`hasTrustDialogAccepted=true`、`hasCompletedProjectOnboarding=true`），避免 Claude CLI 启动时弹出交互式确认。

## Pipeline 三层编排

### 三层定义

| 阶段     | 常量            | 来源                          | 内容                  |
| -------- | --------------- | ----------------------------- | --------------------- |
| System   | `PhaseSystem`   | 系统根据 session 配置自动生成 | cd + env + file 步骤  |
| Template | `PhaseTemplate` | 模板定义的启动命令            | command 步骤          |
| User     | `PhaseUser`     | 用户自定义命令                | 前端直接传入 pipeline |

### 四种步骤类型

| 类型      | 常量          | 执行方式                                                                                   |
| --------- | ------------- | ------------------------------------------------------------------------------------------ |
| `cd`      | `StepCD`      | `mkdir -p {dir} && cd {dir}`，通过 SendKeys 注入                                           |
| `env`     | `StepEnv`     | 双重设置：`tmux set-environment` 写入 tmux 环境变量 + `export KEY='VALUE'` 写入 shell      |
| `file`    | `StepFile`    | 先 `mkdir -p` 创建父目录，再通过 heredoc `cat > {path} << 'SESSIONCONFIGEOF'` 写入文件内容 |
| `command` | `StepCommand` | 直接 `SendKeys` 发送命令文本                                                               |

### 安全机制

- **回显控制** — 在第一个 env/file 步骤前执行 `stty -echo` 关闭 shell 回显，防止 token 等敏感值泄露到终端
- **回显恢复** — 在 command 步骤前执行 `stty echo` 恢复回显
- **defer 保障** — 函数退出时通过 defer 确保回显状态恢复

### BuildPipeline 构建逻辑

`BuildPipeline()` 按 order 顺序构建步骤：

1. **sys-cd** — cd 到工作目录（workingDir → Key 字段）
2. **sys-env-{key}** — 遍历 configs 中 type=env 的项，设置环境变量
3. **sys-file-{key}** — 遍历 configs 中 type=file 的项，写入配置文件
4. **tpl-command** — 模板定义的启动命令（command → Value 字段）

## 会话状态持久化和恢复

### SessionState 结构

| 字段              | 类型     | 说明                                   |
| ----------------- | -------- | -------------------------------------- |
| session_name      | string   | 会话名称                               |
| pipeline          | Pipeline | 管线步骤列表                           |
| restore_strategy  | string   | 恢复策略：auto / manual                |
| restore_command   | string   | 恢复命令（支持 `{session_id}` 占位符） |
| working_dir       | string   | 工作目录                               |
| template_id       | string   | 使用的模板 ID                          |
| cli_session_id    | string   | CLI 工具内部 session ID                |
| env_snapshot      | map      | 环境变量快照                           |
| terminal_snapshot | string   | 终端输出快照                           |
| saved_at          | string   | 保存时间（RFC3339）                    |

### 持久化存储

- 状态文件存储在 `{state_dir}/{session_name}.json`（默认 `/home/agent/.session-state/`）
- 使用原子写入（`configutil.AtomicWriteFile`）防止写入中途崩溃导致文件损坏
- `saveMu` 互斥锁保证并发安全

### 恢复流程

1. 读取状态文件，反序列化 `SessionState`
2. 若 tmux session 仍活跃，先 `KillSession` 终止
3. 创建新的 tmux session（`new-session -d`）
4. 等待 shell 就绪（`waitForPrompt`）
5. 分离 pipeline 中的 command 步骤和其余步骤
6. 对非 command 步骤执行 `ExecutePipeline`（cd/env/file）
7. 确定恢复命令：优先使用 `RestoreCommand`（含 `--dangerously-skip-permissions` 等恢复专用标志），降级到 pipeline 中的 command
8. 替换 `{session_id}` 占位符为实际的 CLI session ID
9. 发送恢复命令

### 保存触发点

| 触发方式   | 实现                                                        |
| ---------- | ----------------------------------------------------------- |
| 自动保存   | `AutoSaveService` 定时触发 `SaveAllPipelineStates()`        |
| 创建时保存 | `CreateSession()` 成功后调用 `SavePipelineState()`          |
| 手动保存   | `POST /api/v1/sessions/save` 触发 `SaveAllPipelineStates()` |
| 删除前保存 | `DeleteSession()` 先调用 `SaveAllPipelineStates()` 再终止   |

## 心跳机制

### 注册流程

`HeartbeatService.Start()` 实现状态机：

1. **未注册 → 注册** — 发送 `POST {controller.addr}/api/v1/nodes/register`
2. **已注册 → 心跳** — 发送 `POST {controller.addr}/api/v1/nodes/heartbeat`
3. **失败处理** — 标记为未注册，指数退避重试（基础 10s，倍增 2x，上限 5min）

### 注册请求内容

| 字段         | 内容                                                                                |
| ------------ | ----------------------------------------------------------------------------------- |
| Name         | `server.name`，为空时使用 hostname                                                  |
| Address      | `server.advertised_addr`，为空时使用 `http://{hostname}{listen_addr}`               |
| ExternalAddr | `server.external_addr`，为空时使用 `http://localhost{listen_addr}`                  |
| Capabilities | supported_templates=["claude","bash"], max_sessions=10, tools=["tmux","filesystem"] |
| Status       | 完整状态快照（活跃 session 数、内存使用、Session 详情、本地配置）                   |
| Metadata     | version="0.1.0", hostname, started_at                                               |

### 请求认证

当 `controller.auth_token` 非空时，请求头添加 `Authorization: Bearer {token}`。

## 自动保存

- **间隔** — 由 `autosave.interval` 配置，默认 60 秒
- **实现** — `time.NewTicker` 驱动，每次 tick 调用 `tmuxService.SaveAllPipelineStates()`
- **优雅停止** — 通过 `stopCh` channel 接收关闭信号

## 配置管理

### 两层配置模型

| 层级    | 作用域                                | 操作接口                               |
| ------- | ------------------------------------- | -------------------------------------- |
| Global  | Agent 节点全局（如 `~/.claude.json`） | `GET/PUT /api/v1/templates/:id/config` |
| Project | Session 工作目录内（如 `CLAUDE.md`）  | `GET/PUT /api/v1/sessions/:id/config`  |

### 乐观并发控制

`ConfigFileService` 实现基于 MD5 hash 的乐观并发控制：

1. **读取时** — 对文件内容计算 `md5:{hex}` hash，返回 `ConfigFileSnapshot`
2. **保存时** — 客户端提交 `base_hash`，服务端先读取当前 hash
3. **冲突检测** — 若当前 hash ≠ base_hash，返回 gRPC `FailedPrecondition` + 冲突详情 JSON
4. **写入** — 所有文件 hash 校验通过后，统一原子写入

## 模板系统

### 内置模板

三个内置模板从 `go:embed templates/*.yaml` 加载：

| 模板 ID | 名称        | 启动命令                                             | 恢复命令                                                                   |
| ------- | ----------- | ---------------------------------------------------- | -------------------------------------------------------------------------- |
| claude  | Claude Code | `IS_SANDBOX=1 claude --dangerously-skip-permissions` | `IS_SANDBOX=1 claude --dangerously-skip-permissions --resume {session_id}` |
| codex   | Codex       | `codex --full-auto`                                  | （无）                                                                     |
| bash    | Bash Shell  | （无）                                               | （无）                                                                     |

### 模板保护

- 内置模板（`builtin=true`）禁止删除
- `UpdateTemplate` 只修改元信息，固定路径配置走 `/templates/:id/config`
- 每次启动时内置模板从 embed FS 无条件覆盖

## 部署

### Docker 多阶段构建

**阶段 1：frontend-builder（Node 22）**

- 安装 pnpm，复制 Monorepo 工作区配置
- 复制 fabrication 和 black-ridge/web 前端代码
- `pnpm install && pnpm run build`

**阶段 2：backend-builder（Go 1.24）**

- 复制 cradle 和 agent 源码
- 从 frontend-builder 复制 `web-dist/`（通过 `go:embed` 嵌入）
- `CGO_ENABLED=0 go build -o /bin/agent`

**阶段 3：runtime（Node 22）**

- 安装运行时依赖：tmux, git, curl, ca-certificates, procps, jq
- 全局安装 `@anthropic-ai/claude-code`
- 创建 agent 用户，设置 `USER agent`、`EXPOSE 8080`

### entrypoint 初始化流程

1. **初始化 `.claude.json`** — 首次启动写入默认值
2. **初始化 `.claude/settings.json`** — 幂等 merge 默认权限配置
3. **生成 `.tmux.conf`** — 写入 `set -g history-limit 50000`
4. **tmux 预热** — 创建并销毁临时 session，确保 tmux server 启动
5. **执行 CMD** — `exec "$@"` 启动 agent 二进制

## 完整配置项对照表

### server

| YAML 字段              | 环境变量                     | 默认值 | 说明                       |
| ---------------------- | ---------------------------- | ------ | -------------------------- |
| server.listen_addr     | AGENT_SERVER_LISTEN_ADDR     | :8080  | HTTP 监听地址              |
| server.auth_token      | AGENT_SERVER_AUTH_TOKEN      | ""     | API Bearer Token（空=DEV） |
| server.name            | AGENT_NAME                   | ""     | Agent 名称（空=hostname）  |
| server.external_addr   | AGENT_EXTERNAL_ADDR          | ""     | 外部可访问地址             |
| server.advertised_addr | AGENT_ADVERTISED_ADDR        | ""     | 注册时使用的可达地址       |
| server.allowed_origins | AGENT_SERVER_ALLOWED_ORIGINS | []     | 允许的 Origin              |

### tmux

| YAML 字段          | 环境变量                 | 默认值    | 说明             |
| ------------------ | ------------------------ | --------- | ---------------- |
| tmux.socket_path   | AGENT_TMUX_SOCKET_PATH   | ""        | tmux socket 路径 |
| tmux.default_shell | AGENT_TMUX_DEFAULT_SHELL | /bin/bash | 默认 shell       |

### terminal

| YAML 字段              | 环境变量                     | 默认值 | 说明             |
| ---------------------- | ---------------------------- | ------ | ---------------- |
| terminal.default_lines | AGENT_TERMINAL_DEFAULT_LINES | 50     | 终端输出默认行数 |

### controller

| YAML 字段                     | 环境变量                            | 默认值 | 说明           |
| ----------------------------- | ----------------------------------- | ------ | -------------- |
| controller.addr               | AGENT_CONTROLLER_ADDR               | ""     | Manager 地址   |
| controller.enabled            | AGENT_CONTROLLER_ENABLED            | false  | 是否启用心跳   |
| controller.heartbeat_interval | AGENT_CONTROLLER_HEARTBEAT_INTERVAL | 10     | 心跳间隔（秒） |
| controller.auth_token         | AGENT_CONTROLLER_AUTH_TOKEN         | ""     | 心跳认证 Token |

### workspace

| YAML 字段           | 环境变量                 | 默认值                    | 说明         |
| ------------------- | ------------------------ | ------------------------- | ------------ |
| workspace.root_dir  | AGENT_WORKSPACE_ROOT_DIR | /home/agent               | 工作区根目录 |
| workspace.state_dir | （无）                   | {root_dir}/.session-state | 状态文件目录 |

### autosave

| YAML 字段         | 环境变量                | 默认值 | 说明               |
| ----------------- | ----------------------- | ------ | ------------------ |
| autosave.interval | AGENT_AUTOSAVE_INTERVAL | 60     | 自动保存间隔（秒） |
