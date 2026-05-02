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
| Skin (UI)         | skin/       | Westworld 主题 UI 组件库               | [AGENTS.md](skin/AGENTS.md) + [docs/](skin/docs/)     |
| Cradle (Go)       | cradle/     | Go 共享库（HTTP/Pipeline/Config/Auth） | [AGENTS.md](cradle/AGENTS.md) + [docs/](cradle/docs/) |
| Molds (模具)      | molds/      | 供应商 Dockerfile                      | [docker-build-guide.md](docs/docker-build-guide.md)   |
| Deps (原料)       | deps/       | 声明式依赖配置                         | [docker-build-guide.md](docs/docker-build-guide.md)   |
| Kubernetes (部署) | kubernetes/ | K8s 部署清单 + Makefile                | 本文件                                                |

## 核心原则

- **主题一致性** — 所有组件必须遵循西部世界科幻视觉风格（切角面板、解密动画、CRT 扫描线等）
- **特效可控** — 动画特效通过 AnimationSettings Context 全局管控，尊重用户 `prefers-reduced-motion` 系统设置
- **API 抽象** — Agent 业务组件仅依赖 `IAgentApiClient` 接口，不绑定具体传输层实现
- **敏感数据保护** — 涉及环境变量和文件内容展示时，必须使用 `maskEnvValue` / `maskFileContent` 脱敏
- **构建规范** — 新增或修改 Dockerfile 必须遵循 [Docker 构建规范](docs/docker-build-guide.md)

## 依赖关系

- 依赖: React, @xterm/xterm (+addon-fit/addon-attach/addon-webgl), @radix-ui/react-dialog, @radix-ui/react-select, @radix-ui/react-slot, class-variance-authority, clsx, tailwind-merge, lucide-react, tailwindcss, tailwindcss-animate
- 被依赖: [behavior-panel/web](../../mesa-hub/behavior-panel/AGENTS.md), [black-ridge/web](../../sweetwater/black-ridge/AGENTS.md)（消费组件和 IAgentApiClient）

## 关键文件

| 路径                                    | 职责                                                         | 文档同步                                                      |
| --------------------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------- |
| skin/src/index.ts                       | 模块入口，统一导出所有组件、工具函数和类型                   | 本文件                                                        |
| skin/src/types.ts                       | 全局类型定义（Session, Pipeline, Template, Config 等）       | 本文件                                                        |
| skin/src/api.ts                         | IAgentApiClient 接口定义，规范所有 Agent 交互方法            | 本文件                                                        |
| skin/src/utils.ts                       | cn 类名合并工具                                              | 本文件                                                        |
| skin/src/utils/mask.ts                  | 环境变量和文件内容脱敏工具                                   | 本文件                                                        |
| skin/src/utils/request.ts               | HTTP 请求工厂函数 createRequest                              | 本文件                                                        |
| skin/src/hooks/usePollingWithBackoff.ts | 带指数退避的轮询 Hook                                        | 本文件                                                        |
| skin/src/components/ui/                 | 视觉特效和基础 UI 组件（20 个文件）                          | [components.md](skin/docs/components.md)                      |
| skin/src/components/agent/              | Agent 业务组件（6 个文件）                                   | [components.md](skin/docs/components.md)                      |
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
