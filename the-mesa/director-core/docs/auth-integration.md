# Auth Integration

## 适用范围

这份文档说明其他模块或调用方如何基于当前最小权限系统接入 `Director Core`。

当前阶段的接入前提：

- 已启用 PostgreSQL
- `authz.enabled=true`
- Director Core 已成功 bootstrap `user:admin`

## 当前最小接入路径

首批只有一条标准接入路径：

- 管理端请求通过全局 Bearer token 认证
- 认证成功后统一映射到 `user:admin`
- 由 `user:admin` 执行所有授权动作

这条路径的目标不是长期替代细粒度权限，而是给当前系统和后续模块提供稳定的最小接入方式。

## 启用配置

`Director Core` 需要同时满足以下条件：

```yaml
server:
  auth_token: "<director-core-token>"

database:
  host: "<postgres-host>"
  port: 5432
  name: "maze_auth"
  user: "<postgres-user>"
  password: "<postgres-password>"

authz:
  enabled: true
  admin_subject_key: "user:admin"
  admin_display_name: "admin"
```

关键点：

- `authz.enabled` 是显式开关
- 配置不完整时服务直接启动失败
- 不存在“数据库没配所以静默关掉授权”的退化路径

## 接入步骤

### 1. 使用全局 Bearer token 访问 Director Core

所有管理端 REST / gRPC 请求继续携带：

```http
Authorization: Bearer <director-core-token>
```

Director Core 在当前阶段会把这个认证结果映射到：

```text
user:admin
```

### 2. 创建主体

当前系统支持两类主体：

- `user:admin`
- `host:<host-name>`

`user:admin` 会在启动时自动 bootstrap。

如果你要为某个 Host 建权限申请，直接使用：

```text
subject_key = host:demo
```

### 3. 创建权限申请单

示例：

```json
POST /api/v1/permission-applications
{
  "subjectKey": "host:demo",
  "targets": [
    { "resource": "session/*", "action": "read" },
    { "resource": "session/*/terminal", "action": "write" }
  ],
  "reason": "允许 demo host 读取 session 并写入终端",
  "expiresAt": "2026-05-05T10:00:00Z"
}
```

### 4. 审批申请单

示例：

```json
POST /api/v1/permission-applications/{permission_application_id}:review
{
  "approved": true,
  "reviewComment": "批准 demo host 使用 session 能力"
}
```

### 5. 查询主体当前权限

```http
GET /api/v1/subjects/host:demo/permissions
```

返回的是当前生效的 grant，而不是底层 Casbin rule。

## 本地与部署接线

### Docker Compose

- 默认编排会拉起一个 `postgres` service
- `director-core` 通过 `DIRECTOR_CORE_AUTH_DATABASE_*`、`DIRECTOR_CORE_HOST_DATABASE_*` 和 `DIRECTOR_CORE_AUTHZ_*` 环境变量接入
- `authz.enabled=true` 时，Director Core 启动会先跑 migration，再 bootstrap `user:admin`

### Kubernetes

- 当前仓库优先提供“连接外部 PostgreSQL”的接线方式
- `director-core-config` 提供 `DIRECTOR_CORE_AUTH_DATABASE_HOST/PORT/NAME`、`DIRECTOR_CORE_HOST_DATABASE_HOST/PORT/NAME` 和 `DIRECTOR_CORE_AUTHZ_*`
- `director-core-secret` 提供 `DATABASE_USER`、`DATABASE_PASSWORD` 与 `AUTH_TOKEN`
- 如果这些配置缺失，权限系统不会静默降级，而是直接启动失败

## 建议的接入顺序

如果后续模块需要先接进来，建议顺序如下：

1. 先复用 Director Core 的 Bearer token 认证
2. 先按 `user:admin` 跑通全链路
3. 再为模块里的具体主体分配 `subject_key`
4. 再通过权限申请单授予最小能力

## 当前不支持

以下能力不属于首批接入范围：

- 自动审批
- 角色 CRUD
- 细粒度角色矩阵
- domain 维度 Casbin 模型
- 通过裸 `rule_id` 操作底层策略

## 后续扩展建议

等当前闭环稳定后，再单独推进：

- 细粒度角色
- 非 `admin` 的管理端主体
- 统一身份源
- 自动审批策略
