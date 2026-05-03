# Fabrication AGENTS.md

## 职责

如同剧中的制造部（Fabrication）是 Host 制造、维修和外观定制的场所，Fabrication 为整个系统提供统一的共享基础设施。它包含以下子模块：

- **skin（UI 组件库）** — 西部世界主题的 React 组件库，提供视觉特效组件、基础 UI 组件、Agent 业务组件及工具函数，所有面向使用者的界面都由这里"制造"出来。
- **cradle（Go 共享库）** — Go 公共库，为 Manager 和 Agent 提供统一的日志、配置、HTTP 工具、中间件、通信协议、管线编排及数据持久化能力。
- **molds（供应商镜像）** — Docker 构建模具，使用多阶段构建（`FROM scratch`），只包含 `/opt/*` 工具目录，不含基础系统层。通过 `COPY --from` 按需组合到目标镜像。
- **deps（原料清单）** — 声明式依赖配置，管理供应商镜像的预装工具列表。
- **kubernetes（K8s 部署）** — Kubernetes 部署清单和 Makefile，提供一键部署、滚动更新等运维能力，支持 Docker Desktop 本地 K8s 环境。

## 子模块索引

| 子模块            | 目录        | 职责                                   | 详细文档                                              |
| ----------------- | ----------- | -------------------------------------- | ----------------------------------------------------- |
| Skin (UI)         | skin/       | Westworld 主题 UI 组件库 + OpenAPI 生成的 TS SDK | [AGENTS.md](skin/AGENTS.md) + [docs/](skin/docs/)     |
| Cradle (Go)       | cradle/     | Go 共享库（HTTP/Pipeline/Config/Auth） | [AGENTS.md](cradle/AGENTS.md) + [docs/](cradle/docs/) |
| Molds (模具)      | molds/      | 供应商 Dockerfile                      | [docker-build-guide.md](docs/docker-build-guide.md)   |
| Deps (原料)       | deps/       | 声明式依赖配置                         | [docker-build-guide.md](docs/docker-build-guide.md)   |
| Kubernetes (部署) | kubernetes/ | K8s 部署清单 + Makefile                | 本文件                                                |

## 核心原则

- **构建规范** — 新增或修改 Dockerfile 必须遵循 [Docker 构建规范](docs/docker-build-guide.md)

## 关键文件

| 路径                                    | 职责                                                         | 文档同步                                                      |
| --------------------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------- |
| molds/Dockerfile.\*                     | 供应商 Dockerfile（5 个工具链）                              | [docker-build-guide.md](docs/docker-build-guide.md)           |
| deps/\*.txt                             | 声明式依赖配置（go/python/js）                               | [docker-build-guide.md](docs/docker-build-guide.md)           |
| Dockerfile.host                         | Host 装配镜像（多 target stage）                             | [docker-build-guide.md](docs/docker-build-guide.md)           |
| kubernetes/Makefile                     | K8s 部署 Makefile（up/down/update）                          | 本文件                                                        |
| kubernetes/overlays/dev/                | 开发环境 overlay（hostPath + 动态构建，namespace maze-dev）  | [kubernetes-deploy-guide.md](docs/kubernetes-deploy-guide.md) |
| kubernetes/overlays/test/               | 集成测试 overlay（hostPath + 动态构建，namespace maze-test） | [kubernetes-deploy-guide.md](docs/kubernetes-deploy-guide.md) |
| kubernetes/overlays/production/         | 生产环境 overlay（PVC + 远程镜像，namespace maze-prod）      | [kubernetes-deploy-guide.md](docs/kubernetes-deploy-guide.md) |

## 详细文档

| 文档                                                     | 内容                                                                |
| -------------------------------------------------------- | ------------------------------------------------------------------- |
| [skin/docs/components.md](skin/docs/components.md)       | 组件清单 + 工具函数 + Hook + 接口使用方式                           |
| [docs/docker-build-guide.md](docs/docker-build-guide.md) | Docker 构建规范（三层优化 + 供应商镜像 + 新增 Dockerfile 检查清单） |
| [docs/kubernetes-deploy-guide.md](docs/kubernetes-deploy-guide.md) | K8s 部署完整指南（环境配置、更新流程、存储路径）               |
