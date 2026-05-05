# 集成测试场景

## 运行方式

```bash
make test-integration PLATFORM=docker        # Docker Compose 环境
make test-integration PLATFORM=kubernetes    # Kubernetes 环境
make test-integration PLATFORM=docker TEST_NAME=TestHostCreateOnline  # 单个测试
```

## 测试场景

| 测试文件 | 覆盖场景 |
|----------|---------|
| host_lifecycle_test.go | Host 创建/删除/状态生命周期 |
| host_boundary_test.go | Host 边界条件（非法参数、重复创建等） |
| session_lifecycle_test.go | Session 创建/恢复/删除 |
| session_recovery_test.go | Session 崩溃恢复 |
| terminal_operation_test.go | WebSocket 终端交互操作 |
| template_management_test.go | 模板 CRUD 操作 |
| config_audit_test.go | 节点配置管理 + 审计日志验证 |
| node_management_test.go | Agent 节点注册/心跳/管理 |
| disaster_recovery_test.go | 灾难恢复（Manager 重启后 Host 恢复） |
| image_cache_test.go | Host 镜像缓存验证 |

## 测试基础设施

- `kit/config.go` — 测试配置加载
- `kit/env.go` — 环境变量定义
- `docker-compose.test.yml` — Docker 测试环境编排
