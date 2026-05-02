# Black Ridge API 文档

## 认证说明

所有 `/api/v1/*` 端点均经过 Bearer Token 鉴权中间件保护。

```
Authorization: Bearer {server.auth_token}
```

当 `server.auth_token` 为空时，服务运行在 DEV 模式，所有端点开放无认证。`/health` 端点无需认证。

## 通用响应格式

- 成功：`{"status": "ok", "data": { ... }}`
- 失败：`{"status": "error", "message": "错误描述"}`
- 冲突（HTTP 409）：`{"status": "error", "code": "config_conflict", "message": "...", "conflicts": [...]}`

---

## Session CRUD

### GET /api/v1/sessions

列出所有活跃 tmux 会话。**响应 data**：`Session[]`

### POST /api/v1/sessions

创建新的 tmux 会话。

**请求体**（CreateSessionRequest）：
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 会话名称（唯一） |
| command | string | 否 | 启动命令（通常由模板提供） |
| working_dir | string | 否 | 工作目录（相对路径基于 workspace.root_dir） |
| session_confs | ConfigItem[] | 否 | Session 级配置项（需配合 template_id） |
| restore_strategy | string | 否 | 恢复策略：auto（默认）/ manual |
| template_id | string | 否 | 模板 ID |

**ConfigItem**：`{ type: "env"|"file", key: string, value: string }`

重复名称返回 HTTP 409。

### GET /api/v1/sessions/:id

获取指定会话详情。不存在返回 404。

### DELETE /api/v1/sessions/:id

终止并删除会话。执行：保存所有状态 → 终止 tmux → 清理工作目录 → 清理状态文件。

### GET /api/v1/sessions/:id/config

返回 session 工作目录下的项目级配置文件快照。

**前置条件**：session 必须有 template_id。**响应 data**：`SessionConfigView`

### PUT /api/v1/sessions/:id/config

保存 session 项目级配置文件。**请求体**：`SaveConfigRequest`

```json
{ "files": [{ "path": "CLAUDE.md", "content": "...", "base_hash": "md5:abc..." }] }
```

冲突时返回 HTTP 409。

---

## 终端操作

### GET /api/v1/sessions/:id/output

获取终端输出。**查询参数**：`lines`（默认 50）。**响应 data**：`{ session_id, lines, output }`

### POST /api/v1/sessions/:id/input

发送命令文本（自动追加回车）。**请求体**：`{ "command": "ls -la" }`

### POST /api/v1/sessions/:id/signal

发送控制信号。**请求体**：`{ "signal": "sigint" }`

支持：`sigint`（Ctrl+C）、`up`、`down`、`enter`

### GET /api/v1/sessions/:id/env

获取环境变量。**响应 data**：`map[string]string`

### GET /api/v1/sessions/:id/ws

WebSocket 实时终端连接。

**消息方向**：

- PTY → WebSocket：Binary message（PTY 原始输出）
- WebSocket → PTY：Text message（直接写入 PTY，不追加回车）
- resize 控制：`{ "type": "resize", "cols": 120, "rows": 40 }`

默认终端尺寸：24x80。Origin 校验使用 `allowed_origins` 配置。

---

## 管线保存/恢复

### GET /api/v1/sessions/saved

返回所有已保存的 session。**响应 data**：`SessionState[]`

活跃 session 的 `restore_strategy` 标记为 `"running"`。

### POST /api/v1/sessions/:id/restore

触发管线重放恢复。**响应 data**：`null`

### POST /api/v1/sessions/save

触发全量保存。**响应 data**：`{ "saved_at": "2026-04-26T10:30:00Z" }`

---

## Template

### GET /api/v1/templates

列出所有模板（含内置和自定义）。**响应 data**：`SessionTemplate[]`

### POST /api/v1/templates

创建新模板（自动 `builtin=false`）。ID 已存在返回 409。

### GET /api/v1/templates/:id

获取指定模板。不存在返回 404。

### PUT /api/v1/templates/:id

更新模板元信息。固定路径配置不受影响（需走 `/templates/:id/config`）。

### DELETE /api/v1/templates/:id

删除模板。**内置模板禁止删除**，返回 403。

### GET /api/v1/templates/:id/config

返回模板全局配置文件的真实内容快照。**响应 data**：`TemplateConfigView`

### PUT /api/v1/templates/:id/config

保存模板全局配置文件。写入真实文件后自动更新模板中的 content 定义。冲突时返回 409。

---

## Local Config

### GET /api/v1/local-config

获取 Agent 本地记忆配置。**响应 data**：`LocalAgentConfig { working_dir, env }`

### PUT /api/v1/local-config

更新本地记忆配置。

- `working_dir` 为只读，不允许修改（返回 400）
- `env` 为合并式更新：非空值更新/新增，空字符串删除 key

---

## Health Check

### GET /health

无需认证。**响应**：`{"status": "ok"}`

---

## 数据模型

### Session

| 字段         | 类型   | 说明                              |
| ------------ | ------ | --------------------------------- |
| id           | string | 会话 ID（等于 tmux session name） |
| name         | string | 会话名称                          |
| status       | string | 状态（固定 "running"）            |
| created_at   | string | 创建时间                          |
| window_count | int    | tmux window 数量                  |

### SessionState

| 字段              | 类型     | 说明                                   |
| ----------------- | -------- | -------------------------------------- |
| session_name      | string   | 会话名称                               |
| pipeline          | Pipeline | 管线步骤列表                           |
| restore_strategy  | string   | auto / manual / running                |
| restore_command   | string   | 恢复命令（支持 `{session_id}` 占位符） |
| working_dir       | string   | 工作目录                               |
| template_id       | string   | 模板 ID                                |
| cli_session_id    | string   | CLI 内部 session ID                    |
| env_snapshot      | map      | 环境变量快照                           |
| terminal_snapshot | string   | 终端输出快照                           |
| saved_at          | string   | 保存时间                               |

### SessionTemplate

| 字段                 | 类型          | 说明                     |
| -------------------- | ------------- | ------------------------ |
| id                   | string        | 模板唯一标识             |
| name                 | string        | 显示名称                 |
| command              | string        | 启动命令                 |
| restore_command      | string        | 恢复命令                 |
| session_file_pattern | string        | CLI session 文件匹配模式 |
| description          | string        | 描述                     |
| icon                 | string        | 图标                     |
| builtin              | bool          | 是否内置                 |
| defaults             | ConfigLayer   | 默认配置                 |
| session_schema       | SessionSchema | Session 级配置定义       |

### ConfigFileSnapshot

| 字段    | 类型   | 说明                     |
| ------- | ------ | ------------------------ |
| path    | string | 文件路径                 |
| content | string | 文件内容                 |
| exists  | bool   | 文件是否存在             |
| hash    | string | 内容 hash（`md5:{hex}`） |

### PipelineStep

| 字段  | 类型   | 说明                                           |
| ----- | ------ | ---------------------------------------------- |
| id    | string | 步骤 ID（如 sys-cd, sys-env-KEY, tpl-command） |
| type  | string | cd / env / file / command                      |
| phase | string | system / template / user                       |
| order | int    | 执行顺序                                       |
| key   | string | 键（cd 时为目录，env 时为变量名）              |
| value | string | 值                                             |

### LocalAgentConfig

| 字段        | 类型   | 说明                 |
| ----------- | ------ | -------------------- |
| working_dir | string | 基础工作目录（只读） |
| env         | map    | 默认环境变量         |
