# PostgreSQL 规范

## 目标

这份文档定义仓库内 Go 模块接入 PostgreSQL 的统一规范，重点解决 4 件事：

- 表结构由谁拥有
- SQL 放在哪里
- 代码分层如何收敛
- 部署、测试和迁移如何保持一致

## ownership

- 共享层只提供 PostgreSQL 通用能力：
  - 连接池
  - migration runner
  - 事务辅助
  - 通用工具封装
- 共享层不拥有任何业务表结构
- 业务 schema 必须由业务模块自己持有

换句话说：

- 公共库可以提供“怎么连 PG、怎么跑 migration”
- 业务模块必须自己定义“建什么表、查什么 SQL、如何演进”

## 推荐目录

以某个 Go 服务模块为例，推荐：

```text
<module>/server/internal/repository/postgres/
  embed.go
  migrations/
  queries/
  sqlc/
```

说明：

- `migrations/`：版本化 DDL 变更
- `queries/`：业务 SQL
- `sqlc/`：生成代码
- `embed.go`：把 migration 嵌入二进制

要求：

- migration、query、生成代码都放在业务模块内部
- 不要把多个模块的 SQL 混放到共享层
- 不要把业务表结构塞进公共库

## 分层规则

- `service` owning 业务模型、业务规则、repository interface
- `repository/postgres` 只负责 PostgreSQL 持久化实现
- `transport` 不直接依赖 PostgreSQL 细节
- SQL 行、sqlc 类型、数据库错误，不能直接向上透传到业务层之外

边界要求：

- `service` 定义自己需要的 repository interface
- `repository/postgres` 实现这些接口
- `repository/postgres` 负责数据库表示和业务模型之间的转换

## migration 规范

- 使用递增版本号命名，如：
  - `00001_init_xxx.sql`
  - `00002_add_xxx_index.sql`
- 一次 migration 只做一组紧密相关变更
- 文件名必须表达业务意图，不用 `misc`、`temp`、`fix` 这类模糊词
- migration 失败必须直接启动失败，不允许静默降级

建议：

- 建表、索引、约束和必要数据修正都通过 migration 管理
- 不要依赖手工在数据库里补结构
- 不要把“线上先手改，再回填 SQL”当常规流程

## SQL 规范

- `queries/` 只放业务访问 SQL
- 每个文件围绕单一资源或聚合组织
- query name 要表达业务意图，不要只写 `Get`、`List2`、`UpdateNew`
- 复杂约束优先用数据库约束表达，不要只靠应用层记忆

优先使用：

- 主键
- 外键
- 唯一约束
- 必要索引

不要依赖：

- 只在代码里“约定唯一”
- 只在代码里“约定状态合法”

## ID 与主键

- 数据库内部主键默认优先使用自增整型
- 对外暴露的业务 ID 单独设计，不直接暴露内部主键
- 内部主键用于 join、外键、索引局部性和实现细节
- 对外业务 ID 用于 API、审计、外部引用

这意味着：

- “内部主键”和“对外业务 ID”要分开
- 不要把数据库内部主键直接当公共 API 标识

## 事务规范

- 事务边界由 `service` 发起
- 具体 PostgreSQL 实现通过 `context` 或事务执行器承接事务
- 不把 `pgx.Tx` 这类数据库事务对象直接泄漏到业务层
- 一次业务动作内的多步写操作，必须保证原子性

典型场景：

- 创建主记录 + 创建关联记录
- 更新业务状态 + 写审计日志
- 写授权结果 + 写运行时策略映射

## trigger 规范

- 默认不优先使用 trigger
- `updated_at`、状态同步、普通派生字段，优先在显式 SQL 和代码层维护
- 只有在必须由数据库强制执行、且应用层无法稳定保证的约束下，才评估 trigger

原因：

- trigger 会增加理解成本
- trigger 会增加调试和迁移复杂度
- 首次接入 PG 时应优先保持链路清晰

## sqlc 规范

- PostgreSQL 查询代码优先通过 `sqlc` 生成
- 生成目录放在模块自己的 `internal/repository/postgres/sqlc/`
- sqlc 产物不跨模块共享
- query SQL 是源码，生成代码不是手写业务层

要求：

- 更新 `queries/` 后同步重新生成代码
- 不手改生成目录内容
- nullability 语义通过 SQL / 类型配置明确表达

## 配置规范

启用 PostgreSQL 的模块，至少显式配置：

- `host`
- `port`
- `database`
- `user`
- `password`

如果某能力声明“启用后依赖 PostgreSQL”，则：

- 配置缺失时直接失败
- 不允许因为没配 PG 而悄悄降级成“部分可用”

## 部署规范

- 本地开发、测试、生产使用的 PostgreSQL 来源必须明确
- 部署清单中的数据库地址，必须和实际部署资源一致
- 不允许出现“配置写了 PG 地址，但当前环境根本没部署这套资源”的漂移

建议：

- 在 Docker / Kubernetes 环境中，把数据库工作负载和应用配置一起维护
- 部署脚本只等待当前环境真实存在的数据库资源
- 环境未就绪时直接失败，不要跳过

## 生命周期与数据保留规范

PostgreSQL 数据保留必须以“命令语义清晰、默认安全、显式销毁”为原则设计。

先明确两个概念：

- 停止环境：移除运行中的应用工作负载，但保留数据库数据
- 销毁环境：移除运行中的应用工作负载，并删除数据库数据

这两个动作不能再绑定到同一条默认命令上。

### 默认命令语义

- `make down` 必须是默认安全命令
- 默认安全命令必须保留 PostgreSQL 数据
- 只有显式破坏性清理命令才能删除 PostgreSQL 数据

这意味着：

- Docker 下，`make down` 只做 `docker compose down`，不删除 named volume
- Kubernetes 下，`make down` 不再默认删除整个 namespace
- 显式破坏性清理命令需要单独命名，并在输出中明确警告

### Docker 实现方案

Docker 环境的 PostgreSQL 数据保留依赖 named volume。

要求：

- PostgreSQL 数据目录继续挂载到 `/var/lib/postgresql/data`
- `make up PLATFORM=docker` 启动时复用现有 named volume
- `make down PLATFORM=docker` 停止容器和网络，但不删除 volume

显式销毁路径：

- 提供单独命令：
  - `make destroy PLATFORM=docker`
- 该命令必须执行带 volume 删除语义的清理
- 下次启动后 PostgreSQL 必须以全新实例初始化

### Kubernetes 实现方案

Kubernetes 环境不能再把“删除 namespace”作为默认 `down` 语义，因为 namespace 删除会连带删除 PVC。

总体实现分成两条路径：

- 保留数据路径：删除或停用应用工作负载，但保留 PostgreSQL PVC
- 销毁数据路径：删除应用工作负载，并显式删除 PVC 或整个 namespace

要求：

- PostgreSQL 必须作为独立可保留资源看待
- 默认 `down` 不能再通过 `kubectl delete namespace` 实现
- 默认 `down` 应只删除当前 overlay 管理的应用工作负载，或至少跳过 PostgreSQL PVC
- 破坏性清理命令才允许执行 namespace 级删除或 PVC 删除

推荐实现方式：

1. 保持 PostgreSQL 资源与应用资源的逻辑边界清晰
2. `deploy` 仍统一部署整套环境
3. `down` 改为“保留数据卸载”
4. 新增显式“全量销毁”命令承接旧行为

### Kubernetes 资源拆分方案

为了让默认 `down` 能保留 PostgreSQL 数据，部署资源应拆成两类：

- 应用资源：
  - manager
  - web
  - 其他无状态服务
- 数据资源：
  - postgresql deployment/stateful workload
  - postgresql service
  - PVC
  - 初始化 ConfigMap

推荐做法：

- 继续由同一 overlay 管理整套环境
- 但 `down` 时只删除应用资源，不直接删除 PostgreSQL PVC
- 如当前 K8s 清单组织不方便按资源组删除，可以引入数据资源独立清单或 label 分组

关键要求不是“文件怎么分”，而是：

- 默认卸载路径必须能只移除应用资源
- PostgreSQL 数据卷必须留存

### 各环境的具体方案

#### `ENV=dev`

目标：

- 支持频繁 `up/down`
- 保留本地调试数据
- 实现尽量简单

方案：

- Kubernetes `dev` 为 PostgreSQL 增加持久化存储
- 存储可以使用：
  - PVC
  - 或单机开发环境下可接受的 `hostPath`
- 挂载路径固定为 `/var/lib/postgresql/data`

建议：

- 如果开发集群是单机或本地集群，优先选择实现最直接的方案
- 但仍需让持久化语义稳定、可解释

#### `ENV=test`

目标：

- 支持普通 `make up/down` 保留测试数据库
- 仍然保留“需要时彻底清空”的能力

方案：

- Kubernetes `test` 也为 PostgreSQL 增加持久化存储
- 默认 `down` 保留数据
- 集成测试仍可继续使用一次性隔离 namespace 的清理策略

注意：

- 日常 test 环境运维和“集成测试临时环境”是两种不同语义
- 集成测试为了保证干净环境，可以继续走专门的临时 namespace 生命周期
- 这不应反向定义日常 `make down PLATFORM=kubernetes ENV=test`

#### `ENV=prod`

目标：

- 保留现有 PVC 语义
- 避免默认运维命令误删数据

方案：

- 继续使用现有 `postgresql-data` PVC
- 默认 `down` 不删除 PVC
- 只有显式破坏性清理命令才允许删除 PVC 或 namespace

### 命令建议

推荐形成两组明确命令：

- 安全命令：
  - `make up`
  - `make down`
  - `make status`
  - `make proxy`
- 破坏性命令：
  - `make destroy`

要求：

- 命令名必须让操作者一眼能看出是否会删数据
- 不允许把删数据能力继续藏在 `down` 内部

### 代理与连接语义

因为 PostgreSQL 变成需要长期保留的数据资源，代理命令也必须体现这一点。

要求：

- `make proxy` 可以继续同时暴露 Web、Manager、PostgreSQL
- `make proxy-db` 必须能够单独暴露 PostgreSQL
- 输出中应明确当前连接地址、端口和环境

### 帮助与输出要求

- `make help` 必须明确区分“保留数据”命令和“销毁数据”命令
- `make down` 输出必须明确提示“PostgreSQL 数据已保留”
- 破坏性命令输出必须明确提示“PostgreSQL 数据将被删除”

### 验证要求

至少验证以下链路：

- Docker:
  - `up -> down -> up` 后数据库数据仍在
- Kubernetes:
  - `up -> down -> up` 后数据库数据仍在
- 破坏性路径:
  - 执行显式销毁命令后数据库数据被清空

### 一句话总结

总体实现方案不是单纯“给 PostgreSQL 加 PVC”，而是同时完成三件事：

- 默认命令语义改成保留数据
- Kubernetes 资源卸载路径改成应用与数据分离
- `dev` / `test` / `prod` 三个环境都提供真实可复用的数据库持久化实现

## 测试规范

- repository 层至少覆盖：
  - 基本 CRUD
  - 事务回滚
  - 约束冲突
  - 并发状态竞争
- service 层要覆盖：
  - 业务状态机
  - 事务边界
  - 数据库错误到业务错误的映射
- 集成测试要覆盖：
  - migration 能否成功执行
  - 应用启动时能否连通 PG
  - 核心读写链路是否真实经过数据库

## seed 规范

- 默认只初始化启动必需的最小数据
- 不预置未来才需要的角色、字典、运营规则
- seed 要能解释“为什么启动必须有它”

## 文档要求

模块首次接入 PostgreSQL 时，至少补齐：

- 为什么要引入 PG
- 表结构 ownership 在哪里
- migration 在哪里
- query SQL 在哪里
- 如何生成 sqlc
- 启动时如何执行 migration
- 本地 / 测试 / 生产如何提供数据库
- 初始化失败如何排查

## 交付检查

改动 PostgreSQL 相关代码后，至少自查：

- migration 是否可重复执行
- 新增约束和索引是否必要且命名清晰
- SQL 是否仍符合分层边界
- repository 是否把数据库表示转换回业务模型
- 测试是否覆盖成功路径和失败路径
- 部署清单和配置是否保持一致
