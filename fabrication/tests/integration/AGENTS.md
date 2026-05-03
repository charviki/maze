# Integration Tests AGENTS.md

## 职责

跨模块集成测试，覆盖 Host 生命周期、Session 管理、终端操作、灾难恢复、配置审计等核心场景，确保 Manager 与 Agent 节点协同正确。

## 运行命令

```bash
make test-integration PLATFORM=docker        # Docker Compose 环境
make test-integration PLATFORM=kubernetes    # Kubernetes 环境
make test-integration PLATFORM=docker TEST_NAME=TestHostCreateOnline  # 单个测试
```

## 关键文件

| 路径                            | 职责                           |
| ------------------------------- | ------------------------------ |
| kit/config.go                   | 测试配置加载                   |
| kit/env.go                      | 环境变量定义                   |
| host_lifecycle_test.go          | Host 创建/删除/状态生命周期    |
| host_boundary_test.go           | Host 边界条件                  |
| session_lifecycle_test.go       | Session 创建/恢复/删除         |
| session_recovery_test.go        | Session 崩溃恢复               |
| terminal_operation_test.go      | 终端 WebSocket 操作            |
| template_management_test.go     | 模板 CRUD 操作                 |
| config_audit_test.go            | 配置 + 审计日志                |
| node_management_test.go         | 节点注册/管理                  |
| disaster_recovery_test.go       | 灾难恢复                       |
| image_cache_test.go             | 镜像缓存验证                   |
| docker-compose.test.yml         | 集成测试 Docker Compose 编排   |
