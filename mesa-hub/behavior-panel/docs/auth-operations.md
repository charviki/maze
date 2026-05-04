# Auth Operations

## 常见操作

### 查看权限申请单

```http
GET /api/v1/permission-applications?status=pending&page=1&page_size=20
```

适用场景：

- 查看待审批申请
- 查看某阶段的申请状态
- 排查某个主体最近提交过哪些申请

### 查看单个权限申请单

```http
GET /api/v1/permission-applications/{permission_application_id}
```

关注字段：

- `status`
- `reviewed_by`
- `review_comment`
- `targets`
- `expires_at`

### 审批通过

```json
POST /api/v1/permission-applications/{permission_application_id}:review
{
  "permission_application_id": "<id>",
  "approved": true,
  "review_comment": "批准"
}
```

审批通过后会发生：

1. 申请单状态变为 `approved`
2. 为每个 target 生成一条 `permission_grant`
3. 把 grant 同步为 Casbin 规则
4. 写入审计日志

### 审批拒绝

```json
POST /api/v1/permission-applications/{permission_application_id}:review
{
  "permission_application_id": "<id>",
  "approved": false,
  "review_comment": "缺少业务理由"
}
```

拒绝时要求提供 `review_comment`，避免审计记录失去解释性。

### 撤销已批准权限

```json
POST /api/v1/permission-applications/{permission_application_id}:revoke
{
  "permission_application_id": "<id>",
  "revoke_reason": "风险收敛"
}
```

撤销时会：

1. 找到该申请单关联的所有活动 grant
2. 将这些 grant 标记为 `revoked`
3. 删除对应的 Casbin 规则
4. 将申请单状态改成 `revoked`
5. 写入审计日志

## 过期回收

后台 janitor 每 30 秒扫描一次已过期 grant：

1. 找出 `status=active` 且 `expires_at <= now()` 的 grant
2. 将 grant 标记为 `expired`
3. 删除对应 Casbin rule
4. 将关联申请单状态同步为 `expired`
5. 写入审计日志

这意味着：

- 过期控制基于 grant
- 不是直接去改原始申请单内容

## 审计关注点

首批权限系统会记录以下关键事件：

- `permission_request.created`
- `permission_request.approved`
- `permission_request.denied`
- `permission_request.revoked`
- `permission_grant.expired`

排障时建议同时看两类信息：

- 业务维度：`permission_application_id`
- 链路维度：`trace_id`

两者不是同一个字段，也不互相替代。

## 排障建议

### 场景 1：申请单已批准，但主体仍然没有权限

排查顺序：

1. 查看申请单状态是否确实为 `approved`
2. 查看是否生成了对应 grant
3. 查看 grant 是否已过期或已被撤销
4. 查看 Casbin rule 是否成功写入并重新加载
5. 查看请求链路的 `trace_id` 是否有中间失败

### 场景 2：服务启动时报权限系统初始化失败

优先检查：

1. `authz.enabled` 是否为 `true`
2. PostgreSQL 配置是否完整
3. migration 是否成功执行
4. `user:admin` bootstrap 是否成功
5. `user:admin -> * / *` 策略是否已写入

### 场景 3：REST 接口通过，但 gRPC 客户端失败

优先检查：

1. gRPC metadata 是否携带 `authorization`
2. gateway / client 是否透传了 header
3. 是否命中了权限 interceptor
4. 方法是否注册了资源动作映射

## 当前运维边界

首批运维层只保证：

- `admin` bootstrap
- 申请闭环
- grant 回收
- 审计可查

以下能力不在当前运维范围：

- 角色后台管理
- 自动审批规则运营
- 细粒度权限可视化编辑器
