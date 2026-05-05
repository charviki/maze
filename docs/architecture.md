# The Maze 系统架构

## 模块依赖拓扑

```mermaid
graph TD
    subgraph "The Mesa (控制面)"
        ArrivalGate["Arrival Gate<br/>统一入口门户<br/>React + Vite"]
        MesaControl["Director Core / Director Console<br/>代理网关 + Host 编排引擎<br/>Go + React"]
    end

    subgraph "Sweetwater (运行时)"
        BR["Black Ridge<br/>Agent 运行时节点<br/>Go + tmux + React"]
    end

    subgraph "Fabrication (共享基础设施)"
        Cradle["Cradle<br/>Go 共享库<br/>HTTP/Pipeline/Config/gRPC"]
        Skin["Skin<br/>UI 组件库<br/>React + Tailwind"]
        IT["Integration Tests<br/>跨模块集成测试<br/>Go testing"]
    end

    ArrivalGate -->|消费组件| Skin
    MesaControl -->|import Go 库| Cradle
    MesaControl -->|消费组件| Skin
    BR -->|import Go 库| Cradle
    BR -->|消费组件| Skin
    IT -->|测试| MesaControl
    IT -->|测试| BR

    MesaControl -.->|gRPC 代理| BR
```

---

## 请求路由全景

```mermaid
graph LR
    Browser["浏览器<br/>Arrival Gate / Director Console 前端"] -->|HTTP/WS| Nginx["Nginx<br/>路由分发"]
    Nginx -->|"/arrival-gate/"| ArrivalGate
    Nginx -->|"/director-console/"| DirectorConsole["Director Console SPA"]
    Nginx -->|"/api/*"| DirectorCore["Director Core<br/>net/http + grpc-gateway + gRPC"]

    DirectorCore -->|"gRPC 进程内"| DirectorCoreLogic["Director Core 业务逻辑<br/>Host/Node/权限/审计"]
    DirectorCore -->|"gRPC 出站"| BR_BE["Black Ridge 后端<br/>net/http + grpc-gateway + gRPC"]

    BR_BE -->|tmux| CLI["AI CLI<br/>Claude Code / Codex / Bash"]

    DirectorCoreLogic -->|"声明式编排"| Runtime["Docker / K8s<br/>Host 容器运行时"]
```

---

## Host 生命周期

```mermaid
sequenceDiagram
    actor User
    participant BP as Director Core
    participant Runtime as Docker/K8s
    participant BR as Black Ridge
    participant RC as Reconciler

    User->>BP: POST /api/hosts (创建 Host)
    BP->>BP: 持久化 HostSpec (JSON)
    BP->>Runtime: 构建镜像 + 启动容器
    BP-->>User: 202 Accepted

    Runtime-->>BP: 容器启动完成
    BP->>BR: 注册 Agent 节点 (gRPC)

    loop 每 60s
        RC->>BR: 健康检查
        BR-->>RC: 心跳 + 状态快照
        alt 不健康
            RC->>Runtime: 自动重启
        end
    end

    User->>BP: DELETE /api/hosts/:id
    BP->>BR: 注销节点
    BP->>Runtime: 销毁容器
    BP->>BP: 清理 HostSpec
```

---

## Session 创建代理流

```mermaid
sequenceDiagram
    actor User
    participant BP as Director Core<br/>(代理网关)
    participant BR as Black Ridge<br/>(Agent 节点)
    participant TM as tmux
    participant CLI as AI CLI

    User->>BP: POST /api/sessions (创建 Session)
    BP->>BP: 记录审计日志
    BP->>BR: gRPC CreateSession
    BR->>BR: 执行 System Pipeline
    BR->>BR: 执行 Template Pipeline
    BR->>BR: 执行 User Pipeline
    BR->>TM: 创建 tmux 窗口
    TM->>CLI: 启动 AI CLI
    BR-->>BP: Session 创建成功
    BP-->>User: Session 信息

    User->>BP: WebSocket /api/sessions/:id/terminal
    BP->>BR: WebSocket 代理 (pty)
    BR->>TM: 转发终端 I/O
    TM->>CLI: stdin/stdout
```

---

## 模块职责一览

| 模块 | 职责 |
|------|------|
| Arrival Gate | 统一入口门户，西部世界主题 Landing → 主界面 |
| Director Core / Director Console | 代理网关 + Host 编排引擎，管控 Agent 节点全生命周期 |
| Black Ridge | Agent 运行时，tmux + Pipeline 管理 AI CLI 会话 |
| Cradle | Go 共享库，Proto IDL 驱动，提供 HTTP/Config/Auth/Pipeline |
| Skin | Westworld 主题 React 组件库，视觉特效 + Agent 业务组件 |
| Fabrication | 共享基础设施（Docker 构建模具 + K8s 部署清单） |
| Integration Tests | 跨模块集成测试，覆盖 Docker/K8s 双环境 |

## 外部依赖

| 依赖 | 用途 |
|------|------|
| PostgreSQL | Director Core 权限系统持久化 |
| Docker | Host 容器运行时 + 镜像构建 |
| Kubernetes | 生产环境容器编排（可选） |
| tmux | Black Ridge Agent 节点的终端会话管理 |
| AI CLI 工具 | Claude Code、Codex 等（Agent 节点内运行） |
| buf | Protobuf IDL 代码生成 |
| Nginx | 前端路由分发 + API 反向代理 |
