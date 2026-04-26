# Behavior Panel API 文档

## 概述

所有 API 端点前缀为 `/api/v1`。认证方式为 `Authorization: Bearer <auth_token>` 请求头。
当 `auth_token` 配置为空时，认证中间件 pass-through（开发模式）。

响应统一格式（由 cradle 的 `httputil` 封装）：
- 成功：`{"status": "ok", "data": <payload>}`
- 失败：`{"status": "error", "message": "<错误描述>"}`

---

## 节点管理

### POST /api/v1/nodes/register
Agent 注册节点。**认证**：需要 Bearer Token

**请求体**（protocol.RegisterRequest）：
```json
{
  "name": "claude-1",
  "address": "http://agent-claude-1:8080",
  "external_addr": "http://localhost:8081",
  "capabilities": {},
  "status": {},
  "metadata": {}
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 节点唯一标识，同名节点会被覆盖 |
| address | string | 是 | Agent 内部地址，Manager 代理用，已含 scheme 前缀 |
| external_addr | string | 否 | Agent 外部可达地址 |
| capabilities | object | 否 | Agent 能力声明（protocol.AgentCapabilities） |
| status | object | 否 | Agent 初始状态快照（protocol.AgentStatus） |
| metadata | object | 否 | Agent 元数据（protocol.AgentMetadata） |

**响应**：返回完整 Node 对象，`status` 为 "online"，`registered_at` 和 `last_heartbeat` 设为当前时间。

### POST /api/v1/nodes/heartbeat
Agent 上报心跳和状态快照。**认证**：需要 Bearer Token

**请求体**（protocol.HeartbeatRequest）：
```json
{
  "name": "claude-1",
  "status": {
    "cpu_usage": 45.2,
    "memory_usage_mb": 512,
    "session_details": []
  }
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 节点名称，必须已注册 |
| status | object | 否 | Agent 状态快照（CPU、内存、Session 详情） |

**响应**：成功返回更新后的 Node 对象。`active_sessions` 从 `status.session_details` 长度同步。节点不存在返回 404。

### GET /api/v1/nodes
列出所有注册节点。**认证**：需要 Bearer Token（开发模式除外）

**响应**：返回 Node 数组，按 name 字母序排列。超过 30 秒无心跳的节点标记为 offline（惰性检测）。

### GET /api/v1/nodes/:name
获取指定节点详情。**认证**：需要 Bearer Token（开发模式除外）

**路径参数**：`name` - 节点名称。**响应**：返回单个 Node 对象。节点不存在返回 404。

### DELETE /api/v1/nodes/:name
删除指定节点。**认证**：需要 Bearer Token（开发模式除外）

**路径参数**：`name` - 节点名称。**响应**：成功返回 `{"status": "ok", "data": null}`。节点不存在返回 404。

---

## Session 代理

以下端点将请求代理到目标 Agent。Manager 在 NodeRegistry 中查找节点地址，构建 Agent URL 并转发请求。代理超时 30 秒。

### GET /api/v1/nodes/:name/sessions
代理列出 Agent 的所有 Session。**审计**：action=list_sessions

### POST /api/v1/nodes/:name/sessions
代理创建 Session。请求体透传到 Agent 的 CreateSessionRequest。**审计**：action=create_session

### GET /api/v1/nodes/:name/sessions/saved
代理获取已保存的 Session 列表。**审计**：action=get_saved_sessions

### GET /api/v1/nodes/:name/sessions/:id
代理获取单个 Session。**审计**：action=get_session

### DELETE /api/v1/nodes/:name/sessions/:id
代理删除 Session。**审计**：action=delete_session

### GET /api/v1/nodes/:name/sessions/:id/config
代理获取 Session 配置。**审计**：action=get_session_config

### PUT /api/v1/nodes/:name/sessions/:id/config
代理更新 Session 配置。请求体透传到 Agent 的 SaveConfigRequest。**审计**：action=update_session_config

### POST /api/v1/nodes/:name/sessions/:id/restore
代理恢复 Session。请求体透传到 Agent。**审计**：action=restore_session

### POST /api/v1/nodes/:name/sessions/save
代理保存所有 Session。**审计**：action=save_all_sessions

---

## Template 代理

### GET /api/v1/nodes/:name/templates
代理列出 Agent 的所有模板。**审计**：action=list_templates

### POST /api/v1/nodes/:name/templates
代理创建模板。请求体透传到 Agent 的 SessionTemplate。**审计**：action=create_template

### GET /api/v1/nodes/:name/templates/:id
代理获取单个模板。**审计**：action=get_template

### PUT /api/v1/nodes/:name/templates/:id
代理更新模板。请求体透传到 Agent。**审计**：action=update_template

### DELETE /api/v1/nodes/:name/templates/:id
代理删除模板。**审计**：action=delete_template

### GET /api/v1/nodes/:name/templates/:id/config
代理获取模板配置。**审计**：action=get_template_config

### PUT /api/v1/nodes/:name/templates/:id/config
代理更新模板配置。请求体透传到 Agent 的 SaveConfigRequest。**审计**：action=update_template_config

---

## Local Config 代理

### GET /api/v1/nodes/:name/local-config
代理获取 Agent 本地配置。**审计**：action=get_local_config

### PUT /api/v1/nodes/:name/local-config
代理更新 Agent 本地配置。请求体透传到 Agent 的 Partial LocalAgentConfig。**审计**：action=update_local_config

---

## WebSocket

### GET /api/v1/nodes/:name/sessions/:id/ws
WebSocket 终端代理。Manager 将前端 WebSocket 连接升级后，建立到 Agent 的 WebSocket 客户端连接，双向转发 Binary 和 Text 消息。

**路径参数**：
| 参数 | 说明 |
|------|------|
| name | 节点名称 |
| id | Session ID |

**代理目标**：`ws://{node.Address}/api/v1/sessions/{id}/ws`（http 自动替换为 ws，https 替换为 wss）

**Origin 校验**：使用 `allowedOrigins` 配置列表，为空时允许所有来源。

**行为**：
- Manager 使用 `HertzUpgrader`（hertz-contrib/websocket）升级前端连接
- 使用 `gorilla/websocket.DefaultDialer` 作为客户端连接 Agent，携带 Auth Token
- 双 goroutine 双向转发 Binary 和 Text 消息
- 任一端断开时另一端自动关闭
- 连接时记录审计日志（action=websocket_connect, status_code=101）

---

## 审计日志

### GET /api/v1/audit/logs
获取审计日志列表。**认证**：需要 Bearer Token（开发模式除外）

**查询参数**：
| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码，从 1 开始 |
| page_size | int | 否 | 50 | 每页条数 |

**无分页参数**：返回全部日志（按时间倒序），格式：`{"status": "ok", "data": [...]}`

**有分页参数**：
```json
{
  "status": "ok",
  "data": {
    "logs": [],
    "total": 150,
    "page": 1,
    "page_size": 50
  }
}
```

**审计日志条目字段**（protocol.AuditLogEntry）：
| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 唯一 ID（audit- 前缀 + 32 位 hex，crypto/rand 生成） |
| timestamp | time | 操作时间，自动填充 |
| operator | string | 操作者（当前固定为 "frontend"） |
| target_node | string | 目标节点名称 |
| action | string | 操作类型（如 list_sessions, create_session, websocket_connect 等） |
| payload_summary | string | 请求体摘要（截断至 200 字符） |
| result | string | 操作结果（success / error: ...） |
| status_code | int | HTTP 状态码 |

---

## 健康检查

### GET /health
返回服务健康状态，无需认证。**响应**：`{"status": "ok"}`
