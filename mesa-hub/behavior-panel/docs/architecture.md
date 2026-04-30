# Behavior Panel 架构文档

## 概述

Behavior Panel 是 Mesa-Hub 的控制中心模块，由 **Agent Manager**（Go 后端，基于 Hertz 框架）和 **Web 前端**（React + @maze/fabrication 组件库）组成。Manager 作为代理网关，统一管理所有 Agent 节点，代理前端到 Agent 的所有 HTTP 和 WebSocket 请求。

## Manager 角色：代理网关

Manager 采用**代理网关模式**（API Gateway Pattern），核心职责：

1. **节点注册中心** — Agent 启动时向 Manager 注册，上报地址、能力声明和状态
2. **请求代理** — 前端所有 API 请求（Session、Template、LocalConfig）经 Manager 代理到目标 Agent
3. **WebSocket 终端代理** — 前端终端连接经 Manager 双向代理到 Agent
4. **审计日志** — 记录所有代理操作，提供操作可追溯性
5. **动态 Host 生命周期管理** — 通过 Docker socket 动态创建/销毁 Host 容器，选配工具链和资源限制
6. **前端不直连 Agent** — 所有通信经过 Manager，保证可观测性

## 数据流

### 节点注册流程

```
Agent 启动 → POST /api/v1/nodes/register (name, address, capabilities, status, metadata)
            → NodeRegistry.Register() → 标记 dirty → 返回 Node 信息
```

### 心跳机制

```
Agent 定时 → POST /api/v1/nodes/heartbeat (name, status 快照含 CPU/内存/Session 详情)
            → NodeRegistry.Heartbeat() → 更新 LastHeartbeat + AgentStatus → 标记 dirty
            → 超过 30 秒无心跳 → ListNodes/GetNode 时标记为 offline
```

离线阈值定义在 [node.go](../server/biz/model/node.go) 中的 `nodeOfflineThreshold = 30 * time.Second`。

### 代理请求流程（HTTP）

```
前端 → Manager API (/api/v1/nodes/:name/sessions/*)
     → SessionProxyHandler.proxyToAgent()
     → 1. 从 NodeRegistry 查找节点
     → 2. 构建 Agent URL (node.Address + agentPath)
     → 3. validateAgentURL() SSRF 校验
     → 4. 透传 Auth Token (Authorization: Bearer)
     → 5. 发送请求到 Agent (30s 超时)
     → 6. 回写响应给前端
     → 7. 记录审计日志 (operator=frontend, action, target_node, result, status_code)
```

### WebSocket 终端代理

```
前端 ws:// → Manager /api/v1/nodes/:name/sessions/:id/ws
           → HertzUpgrader.Upgrade() 升级为 WebSocket
           → scheme 替换 (http→ws, https→wss) 构建 Agent WS URL
           → validateAgentURL() SSRF 校验
           → gorilla/websocket.Dial() 连接 Agent
           → 双 goroutine 双向转发 (Binary + Text)
           → 任一端断开 → 另一端自动关闭
```

WebSocket Origin 校验使用配置化的 `allowedOrigins` 列表，为空时允许所有来源（开发模式）。详见 [proxy_ws.go](../server/biz/handler/proxy_ws.go)。

## 节点注册表（NodeRegistry）

### 数据结构

Node 结构定义在 [node.go](../server/biz/model/node.go)，包含：`Name`, `Address`, `ExternalAddr`, `Status`(online/offline), `RegisteredAt`, `LastHeartbeat`, `Capabilities`, `AgentStatus`, `Metadata`。

### 持久化策略：dirty flush

- **存储文件** — `nodes.json`（JSON 格式，位于 workspace.base_dir 目录）
- **dirty 标记** — Register/Heartbeat/Delete 操作设置 `dirty = true`
- **后台 flush loop** — 每 30 秒检查 dirty 标记，有变更时执行原子写入（先写临时文件再 rename，通过 `configutil.AtomicWriteFile` 实现）
- **优雅关闭** — `WaitSave()` 停止 flush loop 并执行最终刷盘，确保最后 30 秒内的数据不丢失
- **启动恢复** — Manager 重启后从 `nodes.json` 加载已有节点，Agent 下次心跳时更新状态
- **并发安全** — 读写锁（`sync.RWMutex`）保护所有内存操作

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
- HTTP 请求：`@maze/fabrication` 的 `createRequest` 封装

### API 客户端分层
- **controller.ts** — Manager 节点管理 API（listNodes, getNode, deleteNode），直连 Manager
- **agent.ts** — 通过 Manager 代理的 Agent API（sessions, templates, local-config, ws），路径前缀 `/api/v1/nodes/:name/`
- **manager.ts** — Manager 模板管理 API（注释说明节点配置和 Session 元数据已移除，由 Agent 本地管理）

### 关键交互
- `NodeList` 组件使用 `usePollingWithBackoff` 轮询节点列表
- 选中节点后构建 `createAgentApi('', nodeName)` 实例，传给 `AgentPanel`
- WebSocket URL 由 `buildWsUrl()` 根据当前页面协议（ws/wss）动态构建

## 部署拓扑

```
浏览器 → Nginx (:10800)
            ├── /          → 前端静态资源 (React SPA)
            ├── /api/*     → proxy_pass → agent-manager:8080
            └── /health    → proxy_pass → agent-manager:8080

agent-manager (:8080)
  ├── /api/v1/nodes/register    ← Agent 注册
  ├── /api/v1/nodes/heartbeat   ← Agent 心跳
  ├── /api/v1/nodes/:name/*     → 代理到 Agent HTTP API
  └── /api/v1/nodes/:name/sessions/:id/ws → WebSocket 双向代理

Agent 实例 (xN)
  ├── agent-claude-1 (:8081)
  ├── agent-claude-2 (:8082)
  └── agent-codex-1  (:8083)
```

### Docker Compose 编排（来自 docker-compose.yml）
- **web** — Nginx 容器，构建前端资产并托管静态文件 + 反向代理 API
- **agent-manager** — Go 后端容器，监听 8080 端口
  - 挂载 `/var/run/docker.sock` 以动态创建/销毁 Host 容器
  - 挂载 `~/agents:/agents` 以管理 Host 持久化目录
  - 安装 `docker-ce-cli` 用于执行 docker build/run/stop/rm
- **agent-1/2/3** — Agent 实例容器，各自独立的持久卷（`~/agents/agent-*`）

### Nginx 配置要点（来自 nginx.conf）
- WebSocket 支持：`proxy_http_version 1.1` + `Upgrade/Connection` 头处理
- 长连接超时：`proxy_read_timeout 3600s`（避免终端连接被断开）
- SPA 路由：`try_files $uri $uri/ /index.html`

## 配置

### 配置来源（优先级从高到低）
1. 环境变量（`AGENT_MANAGER_SERVER_LISTEN_ADDR` 等）
2. YAML 配置文件（`config.yaml`，位于可执行文件同目录）
3. 默认值（`validate()` 函数填充）

### 配置项
| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|----------|--------|------|
| server.listen_addr | AGENT_MANAGER_SERVER_LISTEN_ADDR | :8080 | HTTP 监听地址 |
| server.auth_token | AGENT_MANAGER_SERVER_AUTH_TOKEN | "" | API 鉴权 Token，空为开发模式 |
| server.allowed_origins | AGENT_MANAGER_SERVER_ALLOWED_ORIGINS | [] | CORS/WebSocket 允许的来源列表 |
| server.allow_private_networks | AGENT_MANAGER_ALLOW_PRIVATE_NETWORKS | false | 是否允许代理到内网 IP |
| workspace.base_dir | AGENT_MANAGER_WORKSPACE_BASE_DIR | ~/agents | 宿主机持久化根目录（用于 docker -v 挂载路径） |
| workspace.mount_dir | AGENT_MANAGER_WORKSPACE_MOUNT_DIR | 同 base_dir | Manager 容器内挂载路径（用于文件操作） |
| docker.socket_path | AGENT_MANAGER_DOCKER_SOCKET_PATH | /var/run/docker.sock | Docker socket 路径 |
| docker.network | AGENT_MANAGER_DOCKER_NETWORK | "" | Docker 网络名（Host 容器加入此网络） |
| docker.agent_base_image | AGENT_MANAGER_DOCKER_AGENT_BASE_IMAGE | "" | Agent 基础镜像名（含 agent 二进制和 entrypoint） |
| docker.manager_addr | AGENT_MANAGER_DOCKER_MANAGER_ADDR | http://agent-manager:8080 | Manager 在容器网络中的地址 |

### 开发模式
当 `auth_token` 为空时进入开发模式：
- 所有 API 端点无鉴权保护（Auth 中间件空 token 时 pass-through）
- CORS 和 WebSocket 允许所有来源
- Manager 启动时打印 DEV mode 警告日志

## 安全设计

### Auth Token 认证
- 使用 `cradlemw.Auth()` 中间件，验证 `Authorization: Bearer <token>` 头
- 注册和心跳端点单独添加 Auth 保护，防止恶意注册
- 其余 API 通过 `protected` 路由组统一保护
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
