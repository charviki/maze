# 错误处理规范

## 设计原则

- 错误必须结构化：gRPC code + ErrorInfo reason + 可选的 detail（BadRequest / PreconditionFailure）
- reason 是机器可读的稳定枚举，message 是人可读的描述
- 用 Google 标准 `errdetails` 类型表达语义，不自定义 detail message
- transport 层只做映射，不含业务逻辑

## Google gRPC 错误模型

`google.rpc.Status` 三层结构：code + message + details[]。

标准 detail 类型选用规则：

| 场景 | detail 类型 | 用途 |
|------|-----------|------|
| 所有业务错误 | `ErrorInfo` | reason + domain + metadata |
| 字段校验错误 | `BadRequest` | FieldViolations[]（field + description） |
| 前置条件冲突 | `PreconditionFailure` | Violations[]（type + subject + description） |

项目已引入 `google.golang.org/genproto/googleapis/rpc/errdetails`，现有 `fabrication/cradle/auth/errors.go` 中使用 `errdetails.ErrorInfo` 构造认证错误。

## 各层职责

| 层 | 职责 | 现有模式 |
|----|------|---------|
| **Repository** | DB error → domain sentinel error | `mapNotFoundError` 模式不变 |
| **Service** | 返回 domain error（sentinel error 或自定义类型如 ValidationError） | 业务规则决定返回什么错误 |
| **Transport** | domain error → gRPC status（code + ErrorInfo reason + detail） | 参考 `fabrication/cradle/auth/errors.go` 的 `statusErrorWithReason` 模式 |
| **Gateway** | gRPC status → HTTP JSON（rpcStatusBody 格式） | grpc-gateway 自动映射 |

## ErrorReason 枚举规范

- 定义在 `fabrication/cradle/api/proto/maze/v1/errors.proto`，通过 `buf generate` 生成 Go 类型
- 枚举值按错误类别分组编号：

| 类别 | 编号范围 | 示例 |
|------|---------|------|
| Authentication | 1–8 | TOKEN_MISSING=1, TOKEN_INVALID=2, TOKEN_EXPIRED=3 |
| NotFound | 10–19 | ARCHIVE_NOT_FOUND=10, HOST_NOT_FOUND=11 |
| AlreadyExists | 20–21 | HOST_ALREADY_EXISTS=20 |
| Validation | 30–31 | VALIDATION_FAILED=30 |
| Precondition | 40–43 | CONFIG_CONFLICT=40 |

- 新增 reason 只需加 proto enum 值，不需改 helper 函数

## 错误构造 API

`fabrication/cradle/errutil` 包提供以下构造函数：

| 函数 | 用途 | 附加的 detail |
|------|------|-------------|
| `NewError(code, reason, msg)` | 通用错误 | ErrorInfo |
| `NewValidationError(code, reason, msg, violations)` | 字段校验 | ErrorInfo + BadRequest |
| `NewPreconditionError(code, reason, msg, violations)` | 前置条件 | ErrorInfo + PreconditionFailure |

提取函数：`ReasonFromError`、`FieldViolationsFromError`、`PreconditionViolationsFromError`。

## HTTP 错误响应契约

所有错误经 grpc-gateway 转换后，HTTP 响应体遵循 `rpcStatusBody` 格式，camelCase 序列化。

```json
// 校验错误
{
  "code": 3,
  "message": "invalid argument",
  "reason": "VALIDATION_FAILED",
  "details": {
    "fieldViolations": [
      {"field": "name", "description": "name is required"}
    ]
  }
}

// Config 冲突（前置条件）
{
  "code": 9,
  "message": "config conflict",
  "reason": "CONFIG_CONFLICT",
  "details": {
    "preconditionViolations": [
      {"type": "CONFIG_CONFLICT", "subject": "/path/to/file", "description": "abc123"}
    ]
  }
}

// 普通 NotFound
{
  "code": 5,
  "message": "archive not found",
  "reason": "ARCHIVE_NOT_FOUND"
}
```

### gRPC code → HTTP status 映射

grpc-gateway 自动完成映射，项目不自定义映射逻辑：

| gRPC code | HTTP status |
|-----------|------------|
| INVALID_ARGUMENT (3) | 400 |
| NOT_FOUND (5) | 404 |
| ALREADY_EXISTS (6) | 409 |
| ABORTED (10) | 409 |
| UNAUTHENTICATED (16) | 401 |
| PERMISSION_DENIED (7) | 403 |
| FAILED_PRECONDITION (9) | 400 |

## 前端错误消费规范

- 按 `reason` 字段分流逻辑，不解析 message 字符串
- `fieldViolations` 用于表单字段高亮
- `preconditionViolations` 用于冲突解决 UI

## 迁移策略

- **渐进式**：新模块用新规范，旧模块按 Phase 迁移
- **兼容期**：gateway 同时输出新字段和旧字段
- **迁移完成**后移除兼容代码

现有 `fabrication/cradle/auth/errors.go` 中的 ErrorReason 字符串常量模式将在后续 Phase 中迁移为 proto enum + `errutil` 统一构造。

## 参考文献

- [gRPC Error Handling](https://grpc.io/docs/guides/error/) — gRPC 官方错误处理指南
- [google/rpc/error_details.proto](https://github.com/googleapis/googleapis/blob/master/google/rpc/error_details.proto) — 标准 detail 类型定义
- [errdetails Go 包](https://pkg.go.dev/google.golang.org/genproto/googleapis/rpc/errdetails) — Go 语言 errdetails 包文档
- [Proto Best Practices](https://protobuf.dev/best-practices/dos-donts/) — Protobuf 最佳实践
