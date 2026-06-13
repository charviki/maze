# Fabrication

## 职责

共享基础设施层，包含 cradle（Go 共享库）、skin（UI 组件库）、molds（Docker 构建模具）、deps（声明式依赖配置）、kubernetes（K8s 部署清单）。

## 项目结构

molds/ 包含 5 个工具链的供应商 Dockerfile，Host 装配镜像由 `sweetwater/black-ridge/Dockerfile` 承担（多 target stage，按工具集组合动态生成）。deps/ 管理 go/python/js 声明式依赖（版本锁定，升级用 `make deps-bump`）。kubernetes/ 按 overlay 分环境（dev/test/production），由仓库根 `Makefile` 统一编排部署与代理命令。

## 核心原则

- **构建规范** — Dockerfile 必须遵循多阶段构建（COPY 拆分 + Cache Mount + 供应商镜像）

## 命令

- `make build-deps` — 构建所有供应商镜像（并行；SKIP_DEPS=1 可跳过、复用现有镜像）
- `make deps-bump` — 查询并更新供应商依赖锁定版本，之后需 `make build-deps`
- `make up PLATFORM=kubernetes ENV=dev|test|prod` — K8s 部署（用已有镜像；首次先 `make build`）
- `make deploy PLATFORM=kubernetes ENV=dev|test|prod` — K8s 只部署，不构建镜像
- `make proxy PLATFORM=kubernetes ENV=dev|test|prod` — 代理当前环境服务到本地

## 子模块

| 子模块 | 目录 | 文档 |
|--------|------|------|
| Skin | skin/ | [AGENTS.md](skin/AGENTS.md) |
| Cradle | cradle/ | [AGENTS.md](cradle/AGENTS.md) |
| Molds | molds/ | [docker-build-guide.md](docs/docker-build-guide.md) |
| K8s | kubernetes/ | [kubernetes-deploy-guide.md](docs/kubernetes-deploy-guide.md) |

## 详细文档

| 文档 | 内容 |
|------|------|
| [docs/docker-build-guide.md](docs/docker-build-guide.md) | Docker 构建规范（三层优化 + 检查清单） |
| [docs/kubernetes-deploy-guide.md](docs/kubernetes-deploy-guide.md) | K8s 部署指南（环境配置、更新流程、存储路径） |
