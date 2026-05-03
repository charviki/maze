# Behavior Panel 架构文档

## 概述

Behavior Panel 是 Mesa-Hub 的控制中心模块，由 **Agent Manager**（Go 后端，`net/http` + grpc-gateway + gRPC Server 三层架构）和 **Web 前端**（React + @maze/fabrication 组件库）组成。Manager 作为代理网关和 Host 编排引擎，统一管理所有 Agent 节点，代理前端到 Agent 的所有 HTTP 和 WebSocket 请求，并通过声明式 HostSpec 持久化 + Reconciler 实现 Host 的全生命周期自动化管理。

所有 REST API 由 proto 注解驱动，通过 grpc-gateway ServeMux 处理，WebSocket 路由由 `http.ServeMux` 直接管理。gRPC Server 运行在独立端口（`:9090`），gateway 进程内直连，经过 interceptor chain（认证→分层令牌→审计）。

## Manager 角色：代理网关

Manager 采用**代理网关模式**（API Gateway Pattern），核心职责：

1. **节点注册中心** — Agent 启动时向 Manager 注册，上报地址、能力声明和状态
2. **请求代理** — 前端所有 API 请求（Session、Template、LocalConfig）经 Manager 代理到目标 Agent
3. **WebSocket 终端代理** — 前端终端连接经 Manager 双向代理到 Agent
4. **审计日志** — 记录所有代理操作，提供操作可追溯性
5. **Host 编排引擎** — 异步创建 Host（202 Accepted），声明式 HostSpec 持久化，工具组合镜像缓存
6. **灾难恢复** — Reconciler 启动恢复 + 60s 健康巡检 + failed 自动重试（最多 3 次），确保实际状态趋近期望状态
7. **前端不直连 Agent** — 所有通信经过 Manager，保证可观测性

## 数据流

### 节点注册流程（含令牌校验）

Manager 为每个动态创建的 Host 预存专属令牌（`hostTokens` map），Agent 注册时校验令牌身份：

```
CreateHost() → StoreHostToken(name, hostToken) → 注入容器 AGENT_CONTROLLER_AUTH_TOKEN
                                                              ↓
Agent 启动 → gRPC AgentService.Register (Authorization: Bearer <hostToken>)
            → UnaryHostTokenInterceptor 校验令牌（来自 gatewayutil）：
                1. 开发模式（全局 auth_token 为空）→ 放行
                2. 已知 Host（hostTokens 中有预存令牌）→ 精确匹配
                3. 未知 Host → 回退到全局 auth_token 校验
            → NodeRegistry.Register() → 标记 dirty → 返回 Node 信息
```

令牌校验逻辑定义在 [gatewayutil/auth.go](../../../fabrication/cradle/gatewayutil/auth.go) 的 `UnaryHostTokenInterceptor` 中。

### 心跳机制

```
Agent 定时 → gRPC AgentService.Heartbeat (Authorization: Bearer <hostToken>, name, status 快照)
            → UnaryHostTokenInterceptor 校验令牌身份
            → NodeRegistry.Heartbeat() → 更新 LastHeartbeat + AgentStatus → 标记 dirty
            → 超过 30 秒无心跳 → ListNodes/GetNode 时标记为 offline
```

离线阈值定义在 [node.go](../server/internal/model/node.go) 中的 `nodeOfflineThreshold = 30 * time.Second`。

### 代理请求流程（HTTP → gRPC 转发）

```
前端 → Manager REST API（grpc-gateway 路由）
     → grpc-gateway ServeMux 匹配 proto 注解路由
     → 进程内 gRPC 调用 Manager Server（SessionService/TemplateService/ConfigService）
     → Server.proxy 转发到 Agent gRPC Server
     → UnaryAuditInterceptor 记录审计日志（operator=frontend, action, target_node, result, status_code）
```

### WebSocket 终端代理

```
前端 ws:// → Manager /api/v1/nodes/:name/sessions/:id/ws
           → gorilla/websocket.Upgrader.Upgrade() 升级为 WebSocket
           → scheme 替换 (http→ws, https→wss) 构建 Agent WS URL
           → validateAgentURL() SSRF 校验
           → gorilla/websocket.Dial() 连接 Agent
           → 双 goroutine 双向转发 (Binary + Text)
           → 任一端断开 → 另一端自动关闭
```

WebSocket Origin 校验使用配置化的 `allowedOrigins` 列表，为空时允许所有来源（开发模式）。详见 [proxy_ws.go](../server/internal/transport/proxy_ws.go)。

## 节点注册表（NodeRegistry）

### 数据结构

Node 结构定义在 [node.go](../server/internal/model/node.go)，包含：`Name`, `Address`, `ExternalAddr`, `AuthToken`, `Status`(online/offline), `RegisteredAt`, `LastHeartbeat`, `Capabilities`, `AgentStatus`, `Metadata`。

### 持久化策略：dirty flush

- **存储文件** — `nodes.json`（节点数据）+ `host_tokens.json`（Host 令牌映射），均为 JSON 格式，位于 workspace.base_dir 目录
- **host_tokens.json** — 存储每个 Host 的专属认证令牌（name → token 映射）。令牌在 `CreateHost` 时预存，Agent 注册时用于身份校验。独立的文件存储是因为令牌需要在 Agent 注册前就存在，而 Node 在注册时才创建
- **dirty 标记** — Register/Heartbeat/Delete/StoreHostToken/RemoveHostToken 操作设置 `dirty = true`
- **后台 flush loop** — 每 30 秒检查 dirty 标记，有变更时执行原子写入（先写临时文件再 rename，通过 `configutil.AtomicWriteFile` 实现）
- **优雅关闭** — `WaitSave()` 停止 flush loop 并执行最终刷盘，确保最后 30 秒内的数据不丢失
- **启动恢复** — Manager 重启后从 `nodes.json` 加载已有节点，Agent 下次心跳时更新状态
- **并发安全** — 读写锁（`sync.RWMutex`）保护所有内存操作

## HostSpec 持久化（HostSpecManager）

### 数据结构

HostSpec 结构定义在 [host_spec.go](../server/biz/model/host_spec.go)，包含：`Name`, `Tools`, `Resources`, `AuthToken`, `Status`(pending/deploying/online/offline/failed), `ErrorMsg`, `RetryCount`, `CreatedAt`, `UpdatedAt`。

### 状态流转

```
pending → deploying → online（Agent 注册成功）
                    → offline（Agent 心跳超时）
                    → failed（构建/部署失败）
failed → deploying（Reconciler 自动重试，最多 3 次）
```

### 持久化策略：dirty flush

- **存储文件** — `host_specs.json`（JSON 格式），位于 workspace.base_dir 目录
- **dirty 标记** — Create/UpdateStatus/Delete/IncrementRetry 操作设置 `dirty = true`
- **后台 flush loop** — 每 30 秒检查 dirty 标记，有变更时执行原子写入
- **优雅关闭** — `WaitSave()` 停止 flush loop 并执行最终刷盘
- **启动恢复** — Manager 重启后从 `host_specs.json` 加载已有 HostSpec
- **ListMerged** — 将 HostSpec 与 NodeRegistry 合并，返回带地址/心跳的完整视图

### 与 NodeRegistry 的关系

- NodeRegistry 记录已注册的 Agent 节点（运行时状态）
- HostSpecManager 记录期望的 Host 规格（声明式意图）
- ListMerged 合并两者：HostSpec.Status 为 online/offline 时从 NodeRegistry 获取地址和心跳

## Reconciler（声明式恢复）

Reconciler 在 [reconciler.go](../server/biz/reconciler/reconciler.go) 中实现，负责确保实际运行状态趋近期望状态。

### 启动恢复（RecoverOnStartup）

Manager 启动时执行一次：

1. 遍历所有 HostSpec，预存 AuthToken 到 NodeRegistry
2. 根据 HostSpec.Status 决定恢复策略：
   - `pending`/`deploying` → 检查运行时是否存在，不存在则重新部署
   - `online`/`offline` → 检查运行时是否存在，不存在则重新部署
   - `failed` → 如果 RetryCount < 3，递增重试计数并重新部署

### 健康巡检（StartHealthCheck）

后台 goroutine 每 60 秒执行一次：

- `deploying` — **保护窗口**（5 分钟内不干预，让 deployHostAsync 完成）；超过保护期仍未上线则重新部署
- `online`/`offline` — 检查运行时健康，不健康则重新部署
- `failed` — RetryCount < 3 时自动重试
- `pending` — 超过 5 分钟视为后台任务丢失，标记为 failed

### 重新部署（redeploy）

1. 预存 Host AuthToken 到 NodeRegistry
2. 清理旧容器/Pod（RemoveHost，忽略不存在的错误）
3. 更新状态为 deploying
4. 生成 Dockerfile（工具排序稳定化）
5. 检查工具组合镜像缓存（相同工具组合共享镜像，跳过 docker build）
6. 调用 DeployHost 执行构建和部署

### K8s 环境保护

- `IsHealthy` 只检查 Deployment 是否存在（不检查 Pod 状态）
- K8s 的 `rollout restart` 不会触发 Manager 干预（Deployment 对象存在即视为健康）

## 构建优化

### 工具排序稳定化

`GenerateHostDockerfile` 对工具列表排序后再生成，确保相同组合产生相同 Dockerfile，最大化 Docker 层缓存命中。

### 工具组合镜像缓存

`ToolsetImageTag` 为工具组合生成稳定标签（如 `maze-agent:claude-go`）。构建前先检查组合镜像是否存在：

- 存在 → `docker tag` 复用，跳过构建（相同工具组合的后续 Host 创建从 77 秒降到 ~5 秒）
- 不存在 → 执行 `docker build`，构建后打上组合标签

### BuildKit

`docker build` 启用 `DOCKER_BUILDKIT=1`，利用 BuildKit 并行 COPY 和层缓存加速构建。

## 审计日志（AuditLogger）

### 持久化策略：append-only JSON Lines

- **存储文件** — `audit.log`（JSON Lines 格式，每行一条 JSON 记录）
- **写入方式** — `O_CREATE | O_WRONLY | O_APPEND`，追加写入不覆盖
- **内存上限** — 最多保留 10000 条（`maxAuditEntries`），超出时截断最旧的记录
- **启动恢复** — Manager 重启后从 `audit.log` 加载历史日志
- **降级** — 文件打开失败时降级为纯内存模式，不阻塞启动
- **ID 生成** — 使用 `crypto/rand` 生成唯一 ID（`audit-` 前缀 + 32 位 hex）
- **查询** — 支持全量列表（时间倒序）、分页查询（page/page_size）、按 node/action 过滤

## 前端架构

### 技术栈

- React + TypeScript
- UI 组件库：`@maze/fabrication`（AgentPanel, NodeList, RadarView, BootSequence 等）
- HTTP 请求：`@maze/fabrication` 的 `createSdkConfiguration` + SDK 生成 API 类

### API 客户端分层

- **SDK 生成层** — 由 `openapi-generator`（typescript-fetch）从 OpenAPI 3.0 文档自动生成 TypeScript SDK（位于 `@maze/fabrication` 的 `src/api/gen/`），提供类型安全的 API 类和模型类型
- **共享辅助层** — `@maze/fabrication` 提供 `createSdkConfiguration`（注入自定义 fetch）、`normalizeTemplate`（模板规范化）、`unwrapSdkResponse`（响应解包）
- **controller.ts** — Manager 节点/Host 管理 API，使用生成 SDK 的 NodeServiceApi/HostServiceApi，直连 Manager
- **agent.ts** — 通过 Manager 代理的 Agent API，使用生成 SDK 的 SessionServiceApi/TemplateServiceApi/ConfigServiceApi，路径前缀 `/api/v1/nodes/:name/`

### 关键交互

- `NodeList` 组件使用 `usePollingWithBackoff` 轮询节点列表
- 选中 Host 后构建 `createAgentApi('', hostName)` 实例，传给 `AgentPanel`
- WebSocket URL 由 `buildWsUrl()` 根据当前页面协议（ws/wss）动态构建

## 部署拓扑

```
浏览器 → Nginx (:10800)
            ├── /                → Portal 首页 (Landing 含登录 → 主界面)
            ├── /behavior-panel/ → Behavior Panel SPA (Host 管理)
            ├── /api/*           → proxy_pass → agent-manager:8080
            └── /health          → proxy_pass → agent-manager:8080

agent-manager
  ├── Hertz (:8080)
  │   ├── /health                        ← 健康检查
  │   ├── /api/v1/nodes/:name/sessions/:id/ws → WebSocket 双向代理
  │   └── NoRoute → grpc-gateway ServeMux → REST API（proto 注解驱动）
  └── gRPC Server (:9090)
      ├── 7 个 Service（Host/Node/Audit/Agent/Session/Template/Config）
      └── Interceptor chain：认证 → 分层令牌 → 审计

Agent 实例（动态创建）
  └── 通过 Manager UI 创建，由 Docker/K8s 运行时部署
```

### Docker Compose 编排（来自 docker-compose.yml）

- **web** — Nginx 容器，构建前端资产并托管静态文件 + 反向代理 API
- **agent-manager** — Go 后端容器，监听 8080 端口
  - 挂载 `/var/run/docker.sock` 以动态创建/销毁 Host 容器
  - 挂载 `~/.maze-prod/docker:/data` 以持久化 Manager 元数据，并通过 `/data/agents` 管理 Host 工作目录
  - 安装 `docker-ce-cli` 用于执行 docker build/run/stop/rm
- **Agent 实例** — 全部通过 Manager UI 动态创建，不再在 docker-compose.yml 中静态定义

### Nginx 配置要点（来自 nginx.conf）

- Portal 首页：`/` → 302 重定向到 `/portal/`，`/portal/` → alias + `try_files`
- Behavior Panel：`/behavior-panel/` → `alias` + `try_files $uri $uri/ /behavior-panel/index.html`
- WebSocket 支持：`proxy_http_version 1.1` + `Upgrade/Connection` 头处理
- 长连接超时：`proxy_read_timeout 3600s`（避免终端连接被断开）

## 配置

### 配置来源（优先级从高到低）

1. 环境变量（`AGENT_MANAGER_SERVER_LISTEN_ADDR` 等）
2. YAML 配置文件（`config.yaml`，位于可执行文件同目录）
3. 默认值（`validate()` 函数填充）

### 配置项

| 配置项                        | 环境变量                              | 默认值                        | 说明                                                                                                       |
| ----------------------------- | ------------------------------------- | ----------------------------- | ---------------------------------------------------------------------------------------------------------- |
| server.listen_addr            | AGENT_MANAGER_SERVER_LISTEN_ADDR      | :8080                         | HTTP 监听地址                                                                                              |
| server.auth_token             | AGENT_MANAGER_SERVER_AUTH_TOKEN       | ""                            | API 鉴权 Token，空为开发模式                                                                               |
| server.allowed_origins        | AGENT_MANAGER_SERVER_ALLOWED_ORIGINS  | []                            | CORS/WebSocket 允许的来源列表                                                                              |
| server.allow_private_networks | AGENT_MANAGER_ALLOW_PRIVATE_NETWORKS  | false                         | 是否允许代理到内网 IP                                                                                      |
| workspace.base_dir            | AGENT_MANAGER_WORKSPACE_BASE_DIR      | ~/.maze/docker                | Manager 元数据根目录（nodes.json / host_specs.json / audit.log / host_logs；容器化部署通常挂载到 `/data`） |
| workspace.mount_dir           | AGENT_MANAGER_WORKSPACE_MOUNT_DIR     | 同 base_dir                   | Manager 容器内的元数据挂载路径（用于文件操作）                                                             |
| docker.socket_path            | AGENT_MANAGER_DOCKER_SOCKET_PATH      | /var/run/docker.sock          | Docker socket 路径                                                                                         |
| docker.network                | AGENT_MANAGER_DOCKER_NETWORK          | ""                            | Docker 网络名（Host 容器加入此网络）                                                                       |
| docker.agent_base_image       | AGENT_MANAGER_DOCKER_AGENT_BASE_IMAGE | ""                            | Agent 基础镜像名（含 agent 二进制和 entrypoint）                                                           |
| docker.agent_data_dir         | AGENT_MANAGER_DOCKER_AGENT_DATA_DIR   | `<workspace.base_dir>/agents` | Docker Agent 宿主机根目录；容器化部署中常显式指向宿主机的 `.../docker/agents`                              |
| docker.manager_addr           | AGENT_MANAGER_DOCKER_MANAGER_ADDR     | http://agent-manager:8080     | Manager 在容器网络中的地址                                                                                 |

### 开发模式

当 `auth_token` 为空时进入开发模式：

- 所有 API 端点无鉴权保护（Auth 中间件空 token 时 pass-through）
- CORS 和 WebSocket 允许所有来源
- Manager 启动时打印 DEV mode 警告日志

## 安全设计

### Auth Token 认证

- **分层令牌校验** — 注册和心跳通过 gRPC interceptor `UnaryHostTokenInterceptor`（来自 `gatewayutil`）进行分层校验：
  1. 开发模式（全局 `auth_token` 为空）→ 放行所有请求
  2. 已知 Host（`hostTokens` 中有预存令牌）→ 精确匹配 Host 专属令牌
  3. 未知 Host（非 Manager 创建）→ 回退到全局 `auth_token` 校验
- Host 专属令牌在 `CreateHost` 时生成（当前阶段使用 hostname），通过 `AGENT_CONTROLLER_AUTH_TOKEN` 注入容器
- Agent 自身 API 鉴权使用全局 `auth_token`（`AGENT_SERVER_AUTH_TOKEN`），与注册令牌独立
- 其余 REST API 认证由 gRPC interceptor chain 统一处理（`UnaryAuthInterceptor`），WebSocket 路由由 Hertz `cradlemw.Auth()` 中间件保护
- Auth Token 为空时中间件 pass-through（开发模式）
- Manager 到 Agent 代理时透传 Auth Token

### SSRF 防护

- `validateAgentURL()` 校验代理目标 URL：
  - 仅允许 http/https/ws/wss 协议
  - 解析主机名 IP，检查是否为内网地址
  - 内网范围：127.0.0.0/8, 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 169.254.0.0/16, ::1/128, fc00::/7
- Docker/Kubernetes 环境下设置 `allow_private_networks: true` 跳过内网检查

### CORS

- 配置 `allowed_origins` 时使用白名单 CORS 中间件（`cradlemw.CORSWithOrigins`）
- 未配置时使用 `cradlemw.CORS()` 允许所有来源（开发模式）
- WebSocket 通过 `httputil.CheckOrigin()` 复用相同的 origins 配置
