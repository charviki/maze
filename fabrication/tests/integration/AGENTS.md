# Integration Tests

## 职责

跨模块集成测试，覆盖 Host 生命周期、Session 管理、终端操作、灾难恢复、配置审计等核心场景，确保 Manager 与 Agent 节点协同正确。

## 项目结构

测试用例在根目录按场景划分（host_lifecycle_test.go、session_lifecycle_test.go 等），kit/ 提供测试配置加载和环境变量定义。编排文件 docker-compose.test.yml 定义 Docker 测试环境。

## 命令

- `make test-integration PLATFORM=docker` — Docker Compose 环境
- `make test-integration PLATFORM=kubernetes` — Kubernetes 环境
- `make test-integration PLATFORM=docker TEST_NAME=TestHostCreateOnline` — 单个测试
