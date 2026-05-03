# Cradle 子包职责一览

| 子包 | 职责 |
|------|------|
| configutil | YAML 配置搜索/加载/合并、反射式 env override、路径展开、共享 ServerConfig |
| httputil | 统一 JSON 响应封装、CORS、WebSocket upgrader/relay、SSRF 校验 |
| logutil | 结构化日志接口（基于 slog） |
| middleware | 标准 `net/http` 中间件（Bearer Token 鉴权、CORS） |
| gatewayutil | grpc-gateway 响应格式包装器、ServeMux 工厂、gRPC 认证/审计 interceptor |
| grpcutil | gRPC 生命周期适配器（`ManagedGRPCServer`） |
| lifecycle | 多服务器统一启停管理（errgroup + signal + 优雅关闭） |
| pipeline | Session 管线步骤定义与层级过滤 |
| protocol | 领域模型：Agent 注册/心跳、Host 部署、审计日志（JSON 持久化） |
| maskutil | 敏感值脱敏 |
| storeutil | 泛型 JSON 持久化存储 |
