# Fabrication

## 职责

共享基础设施层，包含 cradle（Go 共享库）、skin（UI 组件库）、molds（Docker 构建模具）、deps（声明式依赖配置）、kubernetes（K8s 部署清单）。

## 项目结构

molds/ 包含 5 个工具链的供应商 Dockerfile，Dockerfile.host 为 Host 装配镜像（多 target stage）。deps/ 管理 go/python/js 声明式依赖。kubernetes/ 按 overlay 分环境（dev/test/production），每个环境通过 Makefile 提供 up/down/update 命令。

## 核心原则

- **构建规范** — Dockerfile 必须遵循多阶段构建（COPY 拆分 + Cache Mount + 供应商镜像）

## 命令

- `make build-all` — 构建所有供应商镜像
- `make deploy-{dev|test|production}` — K8s 部署

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
