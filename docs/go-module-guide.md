# Go 分层与模块规范

## 先看现状

当前仓库的 Go 代码按“**模块独立 + 根目录统一编排**”组织：

- `fabrication/cradle/`：Go 共享库
- `the-mesa/director-core/`：The Mesa 控制核心的 Go 模块
- `sweetwater/black-ridge/server/`：Black Ridge 的 Go 模块
- `fabrication/tests/integration/`：跨模块集成测试模块

根目录通过 `go.work` 统一收敛这些主模块；根 `Makefile` 的 `MODULES` 只负责需要参与常规 Go 构建和静态检查的服务模块。

## 目录约定

- 新的 Go 服务模块，默认保持 `xxx/server/` 为 Go 模块根目录；像 `the-mesa/director-core/` 这类已独立收拢的模块可直接以模块顶层作为 Go 模块根
- 入口放在 `<module>/cmd/<service>/` 或 `server/cmd/<service>/`
- 业务代码放在 `<module>/internal/` 或 `server/internal/`
- 运行时清单放模块外层目录：
  - Dockerfile
  - `docker-compose.yml`
  - `fabrication/kubernetes/overlays/...`
- 不要把部署清单、脚本、前端构建产物塞进 `server/internal/`
- 新增服务模块时，优先把 `internal/` 按职责拆开：
  - 业务层
  - 协议适配层
  - 持久化实现层
  - 其他基础设施实现层

## 分层规范

### `service`

- `service` 是业务层，拥有业务模型、状态枚举、业务错误和 repository interface
- 业务规则、状态机、事务边界从 `service` 发起
- `service` 不依赖具体存储实现，也不依赖 transport DTO
- 业务层决定自己需要哪些持久化能力，并在这里定义对应边界接口

### `transport`

- `transport` 只负责协议适配：
  - 解析请求
  - 调用 `service`
  - 做 proto / HTTP / gRPC 模型转换
  - 做错误码映射
- `transport` 不拥有业务模型，不直接依赖具体 repository 实现
- 新增接口时，必须明确“transport 输入模型”到“service 输入模型”的转换点

### `repository`

- `repository` 只负责持久化、查询和恢复，不负责定义业务语义
- `repository` 可以依赖 `service` 中定义的业务模型和接口
- `repository` 不反向塑造业务层，不新增一套与业务重复的“repository model”
- `repository` 不在存储层拼装业务视图，业务投影应留在业务层或适配层完成

### 其他实现层

- 模块可以按需要拆出其他实现层，例如运行时适配、外部服务客户端、构建器、审计实现等
- 这些目录名称由模块自己决定，不做强制
- 但它们都属于业务层之外的实现细节，不能反向拥有业务模型或业务规则

## 依赖方向

统一遵循：

- `transport` -> `service`
- `repository` -> `service`
- 其他实现层 -> `service` 或共享库
- 禁止 `service` -> 具体 `repository`
- 禁止 `transport` -> 具体 `repository`

换句话说，依赖要收敛到业务层，其他层只提供业务层需要的能力。

## 模型与转换

- 业务模型由 `service` owning
- `transport` 使用 proto / request / response 模型
- `repository` 负责“存储表示 <-> 业务模型”的转换
- 不能把 repository 自己的模型一路透传到 transport 或 service
- 不能因为“已经分目录”就省略模型转换；边界上的转换是分层成立的前提

## 指针和值类型

- 默认优先值类型，保持语义直接、减少 nil 分支
- 只有在以下场景才使用指针：
  - 需要表达“存在 / 不存在”
  - 需要区分“零值”和“未提供”
  - 需要原地修改对象
  - 结构体较大，复制成本明显
- 像时间、可选配置、外键引用这类有 presence 语义的字段，优先用指针
- 跨层接口不要随意暴露可变内部对象；如果 repository 保存的是内部状态，对外返回时要做 defensive copy

## 接口设计

- repository interface 定义在 `service`，由业务层决定自己需要什么能力
- 接口名与分层语义保持一致：
  - 已经使用 `repository` 语义，就不要再混用 `Store`
- 接口只暴露业务真正需要的最小能力，不把底层存储细节带上来
- 事务优先由 `service` 发起，并通过 `context` 透传到具体实现，避免把事务对象泄漏到业务层

## 创建步骤

1. 在目标模块目录创建 `server/go.mod`；如果模块直接以顶层作为根，则创建 `<module>/go.mod`
2. 按现有模式建立目录：
   - `server/cmd/<service>/` 或 `cmd/<service>/`
   - `server/internal/service/` 或 `internal/service/`
   - `server/internal/transport/` 或 `internal/transport/`
   - `server/internal/repository/` 或 `internal/repository/`
3. 其余实现目录按模块需要补充，不强制统一命名
4. 先确定业务模型 owner，再落 transport / repository 转换
5. 配置优先复用 `cradle/configutil`、日志优先复用 `cradle/logutil`
6. 若提供 HTTP / gRPC 服务，优先沿用现有 `lifecycle.Manager` 管理启动与关闭

## 接入根工作区

- 将新模块路径加入根 `go.work`
- 若该模块需要参与常规 Go 构建 / lint / test，再加入根 `Makefile` 的 `MODULES`
- 在模块目录执行 `go mod tidy`

## 测试约定

- 单元测试与源码同目录，命名为 `*_test.go`
- 跨模块集成测试统一放到 `fabrication/tests/integration/`
- 不要把集成测试混进服务模块内部
- 分层调整时，优先补边界测试：
  - transport 转换测试
  - service 业务规则测试
  - repository defensive copy / 持久化测试

## 交付前检查

- 模块内先执行 `go build ./...`
- 根目录执行 `make build-go`
- 根目录执行 `make lint`
- 根目录执行 `make test`

## 文档同步

- 更新对应模块的 `AGENTS.md`
- 涉及架构调整时，同步更新模块 `docs/architecture.md`
- 若新增一级模块，更新根 `AGENTS.md` 的模块索引
- 若调整分层边界或目录语义，优先同步这份指南，避免后续继续沿用旧约定
