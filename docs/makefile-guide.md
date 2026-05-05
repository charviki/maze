# Makefile 组织与编排规范

## 目标

这份文档定义仓库内 Makefile 的组织方式、职责边界和命令编排规则，解决 4 件事：

- 根 `Makefile` 负责什么
- 哪些命令应该下沉到子目录
- 共享变量应该放在哪里
- 新增命令时如何判断归属

## 总体结构

仓库采用三层结构：

```text
Makefile
mk/
  env.mk
  go.mk
  frontend.mk
  images.mk
  runtime.mk
<module-or-tool>/
  Makefile
```

三层职责分别是：

- 根 `Makefile`：统一入口、共享变量加载、跨模块编排
- `mk/*.mk`：根级公共规则模块
- 子目录 `Makefile`：强领域命令的本地实现

## 根 Makefile 规范

根 `Makefile` 只保留以下内容：

- 全局变量入口，例如：
  - `PROJECT_ROOT`
  - `PLATFORM`
  - `ENV`
- `include` 公共模块
- 面向使用者的稳定命令名
- 跨模块聚合命令
- 转发到子目录的代理命令

根 `Makefile` 不应继续承载：

- 大段 Docker / Kubernetes 编排脚本细节
- 集成测试的完整环境编排细节
- 只属于单一目录的代码生成实现
- 与某个子模块强耦合的局部流程

判断标准：

- 如果一个命令主要在“编排整个系统”，它应该留在根级
- 如果一个命令天然只服务某个目录，它应该下沉到该目录

## `mk/` 公共模块规范

`mk/` 不是 GNU Make 官方标准目录，而是社区常见约定，用来存放被根 `Makefile` `include` 的 Make 规则片段。

本仓库中，`mk/` 只放“跨多个模块复用、但又不适合放进某个业务目录”的公共规则。

### `mk/env.mk`

职责：

- 统一维护 `ENV=dev|test|prod` 的派生变量
- 统一导出：
  - `K8S_NAMESPACE`
  - `K8S_OVERLAY`
  - `COMPOSE_PROJECT`
  - `HOST_DATA_DIR`
  - `PORT_DIRECTOR_CORE`
  - `PORT_WEB`
  - `PORT_POSTGRES`
- 统一定义：
  - `COMPOSE_FILE`
  - `COMPOSE_TEST_FILE`
  - `K8S_OVERLAY_DIR`
  - 镜像名
  - Dockerfile 路径

规则：

- 所有根级规则和子目录 `Makefile` 都必须从这里获取共享变量
- 不允许在多个 Makefile 里重复拷贝一套 `ENV` 派生逻辑

### `mk/go.mk`

职责：

- 承载 Go 聚合命令：
  - `build-go`
  - `vet`
  - `lint`
  - `lint-fix`
  - `vulncheck`
  - `test`
  - `coverage`

规则：

- 这类命令本质上是在“遍历多个 Go module”
- 不要求每个 Go 模块都维护自己的根级命令入口
- 模块局部的开发命令仍然可以在模块目录自己执行 `go test ./...`

### `mk/frontend.mk`

职责：

- 承载前端聚合命令：
  - `check-frontend`
  - `format-js`
  - `format-js-check`

规则：

- 当一个前端命令横跨多个前端目录时，放在这里
- 只属于某个前端目录的构建或生成命令，应放到对应目录的 `package.json` 或子目录 `Makefile`

### `mk/images.mk`

职责：

- 承载镜像构建与更新命令：
  - `build`
  - `build-director-core`
  - `build-web`
  - `build-agent`
  - `build-deps`
  - `update-director-core`
  - `update-web`
  - `update-agent`
  - `update-all`

规则：

- 这些命令虽然依赖不同目录的 Dockerfile，但使用方式是“从仓库根目录统一构建”
- 不把镜像构建入口拆散到多个业务目录，避免使用者记忆成本上升

### `mk/runtime.mk`

职责：

- 承载系统运行时命令：
  - `up`
  - `deploy`
  - `down`
  - `undeploy`
  - `status`
  - `proxy`
  - `proxy-web`
  - `proxy-director-core`
  - `proxy-db`

规则：

- 这类命令编排的是“整套系统”，不是某个单一模块
- PostgreSQL 生命周期和数据保留策略也统一在这里收口

## 子目录 Makefile 规范

子目录 `Makefile` 只在满足以下任一条件时引入：

- 命令天然属于该目录
- 在该目录原地执行比在根目录执行更自然
- 命令实现细节较重，不应堆在根 `Makefile`

### `fabrication/tests/integration/Makefile`

应承载：

- `test-integration`
- Docker 测试环境启动与回收
- Kubernetes 测试 namespace 生命周期
- port-forward
- 测试运行变量装配

根目录保留：

- 同名代理命令，通过 `$(MAKE) -C fabrication/tests/integration ...` 调用

### `fabrication/cradle/api/Makefile`

应承载：

- `gen-proto`
- `gen-client`

根目录保留：

- `gen-proto`
- `gen-client`
- `gen`

其中根目录命令只做聚合和转发，不重复实现生成细节。

### `fabrication/skin/Makefile`

应承载：

- `gen-sdk`

根目录保留：

- `gen-sdk`
- `gen`

## 命令归属判断规则

新增命令时，按下面顺序判断：

1. 这个命令是不是在编排整个系统？
2. 这个命令是不是天然属于某个目录？
3. 这个命令是不是被多个模块共用？
4. 这个命令是否需要在子目录原地执行？

归属规则：

- 系统级编排命令 -> 根 `Makefile` 或 `mk/runtime.mk`
- 跨模块聚合命令 -> `mk/*.mk`
- 强领域局部命令 -> 子目录 `Makefile`
- 包管理器原生命令优先留在 `package.json` / `go` / `buf`，不要为了包装而过度包装

## 变量规范

共享变量统一由 `mk/env.mk` 派生和导出。

根目录与子目录传递变量时，必须保持同名同义：

- `PLATFORM`
- `ENV`
- `TEST_NAME`
- `JAVA_HOME`

规则：

- 子目录 `Makefile` 不得私自重定义这些变量的业务语义
- 若子目录需要新增变量，应尽量使用本领域前缀，避免与全局变量冲突
- 集成测试相关变量优先使用 `MAZE_TEST_*`

## 命名规范

- 目标名优先使用动词或动宾结构：
  - `build-go`
  - `check-frontend`
  - `test-integration`
- 破坏性命令必须显式可见，不允许隐藏在普通命令背后
- 代理命令应与其本地实现同名，降低理解成本

## 帮助信息规范

- 所有面向使用者的目标都应带 `##` 注释，供 `make help` 展示
- `help` 输出要体现命令用途
- 如果命令来自子模块或属于破坏性操作，应在说明中明确标注

## Shell 编写规范

- 默认使用 `@` 抑制噪音，但关键信息必须明确输出
- 多步命令失败必须尽快退出，不允许静默忽略错误
- 需要长期维护的复杂脚本，应优先下沉到子目录 `Makefile` 或独立脚本，而不是持续堆大单条命令
- 不要在命令里复制大量环境派生逻辑

## 不要这样做

- 不要把所有命令继续堆回根 `Makefile`
- 不要为了“看起来模块化”给每个目录都强行加 `Makefile`
- 不要在根和子目录重复维护同一套实现
- 不要让根目录和子目录对同一个变量含义不同
- 不要把破坏性清理伪装成普通 `down`

## 推荐迁移顺序

1. 先抽出 `mk/env.mk`
2. 再抽出 `mk/go.mk`、`mk/frontend.mk`、`mk/images.mk`、`mk/runtime.mk`
3. 优先给 `fabrication/tests/integration` 增加本地 `Makefile`
4. 再给 `fabrication/cradle/api` 和 `fabrication/skin` 增加本地 `Makefile`
5. 最后整理根 `Makefile` 的 `help` 与命令转发

## 成功标准

- 根 `Makefile` 以入口和聚合为主，不再充满大段实现细节
- 强领域目录可以在原地运行自己的核心命令
- 根目录仍然保留统一入口
- 共享变量只有一个权威来源
- 破坏性操作与安全操作在命令语义上清晰分离
