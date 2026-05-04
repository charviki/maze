# Auth Overview

## 目标

当前权限系统只服务于 `Behavior Panel / Manager`，首批目标是完成一条可运行、可审计、可回收的最小闭环：

- `admin` 全权启动与接入
- 权限申请单创建
- 审批通过/拒绝
- 生效授权查询
- 撤销与过期回收
- 审计记录

本阶段不引入细粒度角色治理、自动审批、角色继承或跨服务统一身份。

## 核心对象

### Subject

- 外部稳定主体统一使用 `subject_key`
- 格式固定为 `<type>:<name>`
- 当前首批支持：
  - `user:admin`
  - `host:<host-name>`

`subject_key` 是授权和审计使用的真实主键，显示名称只用于展示，不参与权限判定。

### PermissionApplication

`PermissionApplication` 表示一张“权限申请单”，记录主体申请哪些权限目标。

关键字段：

- `id`
- `internal_id`（仅内部持久化使用）
- `subject_key`
- `targets`
- `reason`
- `status`
- `reviewed_by`
- `review_comment`
- `expires_at`

状态流转：

- `pending`
- `approved`
- `denied`
- `revoked`
- `expired`

### PermissionGrant

`PermissionGrant` 表示审批通过后真正生效的授权结果。

一张申请单会拆成一条或多条 grant：

```text
PermissionApplication (1)
  -> PermissionGrant (N)
    -> Casbin Rule (N)
```

关键字段：

- `id`
- `internal_id`（仅内部持久化使用）
- `subject_key`
- `resource`
- `action`
- `source_request_id`
- `status`
- `expires_at`

其中：

- `id` 是对外暴露的稳定字符串 ID，当前由 PostgreSQL `public_id` 生成
- `internal_id` 是数据库内部自增主键，用于 join、FK 和索引局部性优化
- `source_request_id` 是数据库内部的申请单整型外键
- `casbin_rule_id` 只用于内部关联运行时策略，不对外暴露为业务 API 主键

## 代码分层

当前 `Behavior Panel` 的权限实现已经收敛到 `server/internal/`：

- `internal/service`：权限闭环状态机、janitor，以及由业务侧 owning 的 store / tx 接口
- `internal/transport`：`PermissionService` 的 gRPC/gateway 适配
- `internal/repository`：权限领域模型与状态枚举
- `internal/repository/postgres`：PostgreSQL/sqlc 具体实现与 migration

这样做的原因是把“业务规则”和“PostgreSQL 细节”拆开，避免 service 直接依赖 sqlc 类型和查询参数结构，同时确保 repository 接口由使用方而不是实现方塑形。

## API 资源模型

首批公开 2 类资源：

- `permission-applications`
- `subjects/{subject_key}/permissions`

对应路径：

- `POST /api/v1/permission-applications`
- `GET /api/v1/permission-applications`
- `GET /api/v1/permission-applications/{permission_application_id}`
- `POST /api/v1/permission-applications/{permission_application_id}:review`
- `POST /api/v1/permission-applications/{permission_application_id}:revoke`
- `GET /api/v1/subjects/{subject_key}/permissions`

设计原则：

- URL 表达资源
- HTTP 方法表达标准操作
- `:review` 和 `:revoke` 只用于无法自然表达为 CRUD 的状态变更

## 运行时授权

Casbin 首批模型固定为三元组：

```text
sub, obj, act
```

运行时只做两件事：

- 根据 `subject_key` 判断主体是谁
- 根据 `resource + action` 判断是否放行

当前 bootstrap 策略只有一个：

- `user:admin -> * / *`

这意味着现阶段所有已认证的管理端请求都会先映射到 `user:admin`，其他系统后续可先沿用这条接入路径。

## 认证与授权边界

当前全局 Bearer token 和 Host 专属 token只负责认证：

- Bearer token：认证 Manager 管理端请求
- Host token：认证 Agent 注册/心跳

它们不直接充当权限主体。认证通过后，系统再把请求映射成稳定的 `subject_key` 进入授权层。

## 业务 ID 与链路追踪

`PermissionApplication` 的 `id` 是对外业务 ID，不是链路追踪 ID。

- 业务资源：`permission_application_id`
- 观测链路：`trace_id` / `span_id`

数据库内部还会维护一个整型 `internal_id` 作为主键，但它不进入 API payload。

链路追踪信息走：

- HTTP header
- gRPC metadata

不会进入权限业务 payload。
