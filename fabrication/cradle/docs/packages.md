# Cradle 子包文档

## configutil

配置文件搜索、加载、合并与原子写入。

### 导出类型
- `ConfigLayer` — 统一配置层结构，包含环境变量 (`Env map[string]string`) 和配置文件列表 (`Files []ConfigFile`)
- `ConfigFile` — 配置文件项，包含 `Path` 和 `Content` 字段
- `SessionSchema` — Session 创建时的字段定义，包含 `EnvDefs` 和 `FileDefs`
- `EnvDef` — 环境变量定义（Key, Label, Required, Placeholder, Sensitive）
- `FileDef` — 配置文件定义（Path, Label, Required, DefaultContent）

### 导出函数
- `SearchConfigPaths(filename string) (string, error)` — 按优先级搜索配置文件路径：当前目录 → 可执行文件目录 → 可执行文件上级目录
- `LoadYAML(path string, target interface{}) error` — 从指定路径加载 YAML 文件并反序列化
- `LoadFromExe(target interface{}, filename ...string) (string, error)` — 搜索并加载配置文件，默认文件名 `config.yaml`
- `MergeConfigLayers(layers ...ConfigLayer) ConfigLayer` — 合并多个配置层，后者覆盖前者，空值跳过
- `AtomicWriteFile(path string, data []byte, perm os.FileMode) error` — 原子写入文件，先写临时文件再 rename，防止崩溃导致文件损坏

### 使用示例
```go
var cfg MyConfig
path, err := configutil.LoadFromExe(&cfg, "config.yaml")
merged := configutil.MergeConfigLayers(baseLayer, overrideLayer)
err := configutil.AtomicWriteFile("/path/to/file.json", data, 0644)
```

---

## httputil

HTTP 响应封装与 CORS 跨域处理。

### 导出函数
- `Success(c *app.RequestContext, data interface{})` — 返回 HTTP 200，JSON 格式 `{status: "ok", data: ...}`
- `Error(c *app.RequestContext, code int, msg string)` — 返回指定状态码，JSON 格式 `{status: "error", message: ...}`
- `CORS() app.HandlerFunc` — 返回允许所有来源的 CORS 中间件
- `CORSWithOrigins(allowedOrigins []string) app.HandlerFunc` — 返回基于允许来源列表的 CORS 中间件；列表为空时允许所有来源；OPTIONS 预检返回 204；非法来源的 OPTIONS 返回 403
- `CheckOrigin(allowedOrigins []string) func(c *app.RequestContext) bool` — 返回 WebSocket CheckOrigin 函数；列表为空时始终允许

### 使用示例
```go
httputil.Success(c, map[string]string{"name": "agent-1"})
h.Use(httputil.CORSWithOrigins([]string{"https://example.com"}))
```

---

## logutil

结构化日志接口与基于 slog 的实现。

### 导出接口
- `Logger` — 统一日志接口，包含 `Infof/Warnf/Errorf/Fatalf` 及 `WithNode/WithSession/WithAction/WithFields` 链式方法

### 导出类型
- `SlogLogger` — 基于 `log/slog` 的 Logger 实现，同时满足 `logutil.Logger` 和 `hlog.FullLogger` 接口。输出 JSON 格式

### 导出函数
- `New(component string) *SlogLogger` — 创建 JSON 格式结构化日志实例，默认 INFO 级别，输出到 stdout
- `NewNop() *SlogLogger` — 创建丢弃所有输出的 Logger，用于测试（Fatalf 仍会调用 os.Exit(1)）

### SlogLogger 导出方法
- `SetLevel(level hlog.Level)` — 设置日志级别
- `SetOutput(w io.Writer)` — 替换输出目标
- `WithNode(name string) Logger` — 附加 `node_name` 字段，返回新实例
- `WithSession(id string) Logger` — 附加 `session_id` 字段，返回新实例
- `WithAction(action string) Logger` — 附加 `action` 字段，返回新实例
- `WithFields(fields ...string) Logger` — 附加多个键值对字段

### 使用示例
```go
logger := logutil.New("manager")
logger.WithNode("agent-1").WithAction("register").Infof("agent registered")
```

---

## middleware

HTTP 中间件集合（鉴权与 CORS）。

### 导出函数/变量
- `Auth(token string) app.HandlerFunc` — Bearer Token 鉴权中间件。token 为空时跳过鉴权；非空时校验 `Authorization: Bearer <token>`，失败返回 401
- `CORS` (变量) — 委托 `httputil.CORS()`
- `CORSWithOrigins` (变量) — 委托 `httputil.CORSWithOrigins()`

### 使用示例
```go
h.Use(middleware.Auth("my-secret-token"))
h.Use(middleware.CORS)
```

---

## pipeline

管线步骤定义与层级过滤。

### 导出类型
- `PipelinePhase` — 步骤层级：`PhaseSystem`（系统级）、`PhaseTemplate`（模板级）、`PhaseUser`（用户级）
- `PipelineStepType` — 步骤类型：`StepCD`（切换目录）、`StepEnv`（环境变量）、`StepFile`（写入文件）、`StepCommand`（执行命令）
- `PipelineStep` — 步骤结构体：ID、Type、Phase、Order、Key、Value
- `Pipeline` — 管线类型，即 `[]PipelineStep`

### Pipeline 导出方法
- `SystemSteps() Pipeline` — 过滤 system 层步骤
- `TemplateSteps() Pipeline` — 过滤 template 层步骤
- `UserSteps() Pipeline` — 过滤 user 层步骤
- `Sorted() Pipeline` — 按 Order 字段稳定排序（插入排序）

### 使用示例
```go
var p pipeline.Pipeline
for _, step := range p.Sorted().SystemSteps() {
    fmt.Println(step.ID, step.Type, step.Value)
}
```

---

## protocol

Agent 与 Manager 之间的通信协议定义。

### 导出类型（register.go）
- `AgentCapabilities` — Agent 能力：SupportedTemplates、MaxSessions、Tools
- `AgentStatus` — Agent 状态：ActiveSessions、CPUUsage、MemoryUsageMB、WorkspaceRoot、SessionDetails、LocalConfig
- `SessionDetail` — Session 状态：ID、Template、WorkingDir、UptimeSeconds
- `AgentMetadata` — 静态元数据：Version、Hostname、StartedAt
- `LocalAgentConfig` — 本地记忆配置：WorkingDir、Env
- `RegisterRequest` — 注册请求：Name、Address、ExternalAddr、Capabilities、Status、Metadata
- `HeartbeatRequest` — 心跳请求：Name、Status

### 导出类型（audit.go）
- `AuditLogEntry` — 审计日志条目：ID、Timestamp、Operator、TargetNode、Action、PayloadSummary、Result、StatusCode

### 导出类型（host.go）
- `ToolConfig` — 工具配置：ID、Image、SourcePath、DestPath、BinPaths、EnvVars、Description、Category
- `ResourceLimits` — 资源限制：CPULimit、MemoryLimit（均为可选字符串）
- `HostDeploySpec` — 运行时无关的 Host 部署规格：Name、Tools、Resources、AuthToken
- `CreateHostRequest` — 创建 Host 请求：Name、Tools、DisplayName、Resources
- `CreateHostResponse` — 创建 Host 响应：Name、Tools、ImageTag、ContainerID、Status、BuildLog
- `ContainerInfo` — 容器信息：ID、Name、Status、Image、CreatedAt

### 使用示例
```go
req := protocol.RegisterRequest{
    Name:    "agent-1",
    Address: "http://agent-1:8080",
    Capabilities: protocol.AgentCapabilities{
        SupportedTemplates: []string{"claude"},
        MaxSessions:        5,
    },
}
```

---

## maskutil

敏感值脱敏工具。

### 导出函数
- `MaskedValue(val string) string` — 长度 <= 8 返回 `****`；长度 > 8 保留前 4 后 4 字符，中间替换为 `****`

### 使用示例
```go
maskutil.MaskedValue("sk-1234567890abcdef") // "sk-1****cdef"
```

---

## storeutil

泛型 JSON 文件持久化存储，线程安全。

### 导出类型
- `JSONStore[T any]` — 泛型 JSON 存储，内置读写锁

### 导出函数
- `NewJSONStore[T any](path string, data T, logger logutil.Logger) *JSONStore[T]` — 创建并加载，文件不存在时使用零值

### JSONStore 导出方法
- `Save() error` — 原子持久化到 JSON 文件
- `Get() T` — 返回数据副本
- `GetData() *T` — 返回数据只读引用（性能敏感场景）
- `Update(fn func(data *T), persist bool) error` — 写锁下更新，可选持久化
- `View(fn func(data *T))` — 读锁下只读回调

### 使用示例
```go
store := storeutil.NewJSONStore[map[string]string]("/data/nodes.json", map[string]string{}, logger)
store.Update(func(data *map[string]string) {
    (*data)["agent-1"] = "http://agent-1:8080"
}, true)
```
