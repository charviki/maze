# Integration Tests

## 职责

跨模块集成测试，覆盖 Host 生命周期、Session 管理、终端操作、灾难恢复、配置审计、知识库全链路（Archive / Doc / DocLink）等核心场景，确保 Manager 与 Agent 节点及 The Forge 知识库服务协同正确。

## 项目结构

测试用例在根目录按场景划分（host_lifecycle_test.go、session_lifecycle_test.go、forge_*_test.go 等），kit/ 提供测试配置加载、环境等待和 API 测试客户端（含 ForgeTestClient）。编排文件 docker-compose.test.yml 定义 Docker 测试环境。

## 命令

- `make test-integration PLATFORM=docker` — Docker Compose 环境
- `make test-integration PLATFORM=kubernetes` — Kubernetes 环境
- `make test-integration PLATFORM=docker TEST_NAME=TestHostCreateOnline` — 单个测试
