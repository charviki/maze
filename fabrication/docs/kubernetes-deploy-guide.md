# Kubernetes 部署指南

Maze 支持在 Kubernetes 上运行，通过 `fabrication/kubernetes/` 中的 Makefile 和 Kustomize overlay 管理部署。内置三种环境配置：

- **dev**（默认）— 开发环境，hostPath 持久化，动态构建 Agent 镜像
- **test** — 集成测试环境，hostPath 持久化，`make test-integration` 自动管理
- **production** — 生产环境，PVC 持久化，`imagePullPolicy: IfNotPresent`

## 前置条件

- Docker Desktop 已启用 Kubernetes（Settings → Kubernetes → Enable Kubernetes）
- `kubectl` 可正常连接集群（`kubectl cluster-info` 返回正常）
- 项目代码在本地已就绪

## 快速开始

```bash
# 一键部署（构建镜像 + 部署 K8s 资源 + 等待就绪）
make up

# 启动本地代理
make proxy

# 浏览器打开 http://localhost:7080

# 清理
make down
```

## 环境变量体系

Makefile 使用两个正交变量控制部署行为：

- **`PLATFORM`** = `docker` | `kubernetes`（默认 `kubernetes`）— 运行时类型
- **`ENV`** = `dev` | `test` | `prod`（默认 `dev`）— 环境级别

```bash
make up                              # kubernetes × dev  → maze-dev namespace
make up PLATFORM=docker              # docker × dev      → Docker Compose maze-dev
make up ENV=prod                     # kubernetes × prod → maze-prod namespace
make test-integration                # 固定 ENV=test
make test-integration PLATFORM=docker # docker × test
```

### ENV 派生属性

| 属性          | ENV=dev       | ENV=test       | ENV=prod       |
| ------------- | ------------- | -------------- | -------------- |
| K8s Namespace | `maze-dev`    | `maze-test`    | `maze-prod`    |
| 宿主机目录    | `~/.maze-dev` | `~/.maze-test` | `~/.maze-prod` |
| PORT_MANAGER  | `7090`        | `9090`         | `8090`         |
| PORT_GATEWAY  | `7091`        | `9091`         | `8091`         |
| PORT_WEB      | `7080`        | `9080`         | `10800`        |

## Make 命令一览

```bash
make help               # 显示所有命令
```

| 命令                  | 说明                                                           |
| --------------------- | -------------------------------------------------------------- |
| `make up`             | 构建全部镜像 + 部署 K8s 资源 + 等待就绪                        |
| `make down`           | 删除 namespace 及所有资源                                      |
| `make status`         | 查看 Pod / Service / PVC 状态                                  |
| `make proxy`          | 启动 port-forward 代理（Web + Manager + Gateway），Ctrl+C 停止 |
| `make proxy-web`      | 只代理 Web 前端                                                |
| `make proxy-manager`  | 只代理 Manager API                                             |
| `make build`          | 构建全部镜像（Manager + Web + Agent）                          |
| `make build-manager`  | 只构建 Manager 镜像                                            |
| `make build-web`      | 只构建 Web 镜像                                                |
| `make build-agent`    | 只构建 Agent 基础镜像                                          |
| `make deploy`         | 只部署 K8s 资源（不构建镜像）                                  |
| `make update-manager` | 重建 Manager 镜像 + 滚动重启                                   |
| `make update-web`     | 重建 Web 镜像 + 滚动重启                                       |
| `make update-agent`   | 重建 Agent 基础镜像（新创建的 Host 使用新镜像）                |
| `make update-all`     | 全部更新                                                       |

### 环境选择

通过 `ENV` 参数切换 overlay：

```bash
make up                          # 默认 ENV=dev
make up ENV=prod                 # 生产环境
make deploy ENV=prod             # 只部署，不构建镜像
```

## 更新流程

代码修改后，使用 `update-*` 命令重建镜像并滚动重启：

```bash
# 只更新 Manager
make update-manager

# 只更新 Web 前端
make update-web

# 更新 Agent 基础镜像（已有 Agent Pod 不会自动更新，需通过 Manager API 删除重建）
make update-agent

# 全部更新
make update-all
```

## 本地访问

K8s Service 都是 ClusterIP（集群内网），需要通过 port-forward 代理到本地：

```bash
# 同时代理 Web 和 Manager（最常用）
make proxy
# Web:      http://localhost:7080
# Manager:  http://localhost:7090/health
# Gateway:  http://localhost:7091

# 只代理某一个
make proxy-web        # 只代理前端
make proxy-manager    # 只代理 Manager API

# 自定义端口
make proxy PORT_WEB=3000 PORT_MANAGER=9000
```

## Agent 动态构建

K8s 模式下 **复用 Docker 模式的动态构建逻辑**，不需要提前预构建所有工具组合的镜像：

1. 用户通过前端或 API 选择工具组合（如 claude + go + python）
2. Manager 调用 `builder.GenerateHostDockerfile()` 生成 Dockerfile
3. Manager 通过 Docker socket 执行 `docker build` 构建镜像
4. Manager 通过 K8s API 创建 Deployment 引用该镜像

这要求 Manager Pod 挂载宿主机的 Docker socket（dev overlay 已配置）。构建出的镜像直接出现在 Docker Desktop 的本地镜像仓库，K8s 通过 `imagePullPolicy: Never` 使用。

```
用户选工具 → Manager 生成 Dockerfile → docker build → K8s Deployment
                                        ↑
                              Docker socket (hostPath 挂载)
```

## 存储路径

Maze 使用 `~/.maze-{ENV}/` 管理持久化数据，三个环境完全隔离：

```
~/.maze-dev/                          # 开发环境 (ENV=dev)
├── docker/                           # Docker Compose 模式
│   ├── nodes.json                    # Manager 持久化
│   ├── host_specs.json               # Manager 持久化
│   ├── audit.log                     # Manager 持久化
│   ├── host_logs/                    # Manager 运行日志
│   └── agents/                       # Agent 工作目录
└── kubernetes/                       # K8s 模式
    ├── nodes.json                    # Manager 持久化
    ├── host_specs.json               # Manager 持久化
    ├── audit.log                     # Manager 持久化
    ├── host_logs/                    # Manager 运行日志
    └── agents/                       # Agent 工作目录

~/.maze-test/                         # 集成测试 (ENV=test)
├── docker/                           # Docker 集成测试
│   ├── nodes.json                    # Manager 持久化
│   ├── host_specs.json               # Manager 持久化
│   ├── audit.log                     # Manager 持久化
│   ├── host_logs/                    # Manager 运行日志
│   └── agents/                       # Agent 工作目录
└── kubernetes/                       # K8s 集成测试
    ├── nodes.json                    # Manager 持久化
    ├── host_specs.json               # Manager 持久化
    ├── audit.log                     # Manager 持久化
    ├── host_logs/                    # Manager 运行日志
    └── agents/                       # Agent 工作目录

~/.maze-prod/                         # 生产环境 (ENV=prod)
├── docker/                           # Docker 生产
│   ├── nodes.json                    # Manager 持久化
│   ├── host_specs.json               # Manager 持久化
│   ├── audit.log                     # Manager 持久化
│   ├── host_logs/                    # Manager 运行日志
│   └── agents/                       # Agent 工作目录
└── kubernetes/                       # K8s 生产
    ├── nodes.json                    # Manager 持久化
    ├── host_specs.json               # Manager 持久化
    ├── audit.log                     # Manager 持久化
    ├── host_logs/                    # Manager 运行日志
    └── agents/                       # Agent 工作目录
```

| ENV  | Manager 数据                        | Agent 数据                                       | 访问方式                       |
| ---- | ----------------------------------- | ------------------------------------------------ | ------------------------------ |
| dev  | hostPath `~/.maze-dev/kubernetes/`  | hostPath `~/.maze-dev/kubernetes/agents/{name}`  | 宿主机直接可见                 |
| test | hostPath `~/.maze-test/kubernetes/` | hostPath `~/.maze-test/kubernetes/agents/{name}` | 集成测试自动管理               |
| prod | PVC `manager-data`                  | PVC（动态创建）                                  | `kubectl exec` 或 `kubectl cp` |

## Kustomize Overlay 结构

```
fabrication/kubernetes/
├── Makefile                              # 部署管理入口（支持 PLATFORM × ENV 参数）
├── base/                                 # 公共资源（三种环境共享）
│   ├── kustomization.yaml
│   ├── namespace.yaml                    # maze-prod namespace
│   ├── manager-rbac.yaml                 # ServiceAccount + Role + RoleBinding（含 pods/log 权限）
│   ├── manager-secret.yaml               # Auth Token Secret（需填入实际值）
│   ├── manager-pvc.yaml                  # Manager PVC（仅 production 使用）
│   ├── manager-service.yaml              # Manager ClusterIP Service :8080(HTTP) + :9090(gRPC)
│   ├── web-configmap.yaml                # Nginx 配置模板（envsubst 变量注入）
│   ├── web-deployment.yaml               # Web Deployment
│   ├── web-service.yaml                  # Web ClusterIP Service :80
│   └── web-ingress.yaml                  # Ingress（需 Ingress Controller）
└── overlays/
    ├── dev/                              # 开发环境 overlay
    │   ├── kustomization.yaml            # namespace: maze-dev
    │   ├── manager-configmap.yaml        # hostPath + 动态构建 + Docker socket
    │   └── manager-deployment.yaml       # hostPath 挂载 + Docker socket 挂载
    ├── test/                             # 集成测试 overlay
    │   ├── kustomization.yaml            # namespace: maze-test，删除 web/namespace 资源
    │   ├── manager-configmap.yaml        # hostPath + 动态构建 + Docker socket
    │   └── manager-deployment.yaml       # hostPath 挂载 + Docker socket + 声明 8080/9090 端口
    └── production/                       # 生产环境 overlay
        ├── kustomization.yaml
        ├── manager-configmap.yaml        # PVC + IfNotPresent
        └── manager-deployment.yaml       # PVC 挂载 + 远程镜像仓库
```

### 三套 Overlay 差异

| 配置项          | dev                                               | test                                               | production           |
| --------------- | ------------------------------------------------- | -------------------------------------------------- | -------------------- |
| Namespace       | `maze-dev`                                        | `maze-test`                                        | `maze-prod`          |
| Manager 数据卷  | hostPath (`~/.maze-dev/kubernetes/`)              | hostPath (`~/.maze-test/kubernetes/`)              | PVC (`manager-data`) |
| Agent 数据卷    | hostPath (`~/.maze-dev/kubernetes/agents/{name}`) | hostPath (`~/.maze-test/kubernetes/agents/{name}`) | PVC（动态创建）      |
| Agent 镜像      | 动态构建（docker build via Docker socket）        | 动态构建                                           | 预构建镜像选配       |
| VolumeType      | `hostpath`                                        | `hostpath`                                         | `pvc`                |
| ImagePullPolicy | `Never`                                           | `Never`                                            | `IfNotPresent`       |
| Docker socket   | ✅ 挂载                                           | ✅ 挂载                                            | ❌ 不挂载            |
| Web 前端        | ✅                                                | ❌（测试不需要）                                   | ✅                   |
| 目录自动创建    | ✅                                                | ✅                                                 | ❌                   |

## 部署前必做

### 1. 填入 Auth Token

编辑 `base/manager-secret.yaml`：

```yaml
stringData:
  AUTH_TOKEN: 'your-secret-token-here'
```

### 2. 生产环境：替换镜像仓库占位符

编辑 `overlays/production/` 下的文件，将 `REGISTRY` 和 `VERSION` 替换为实际值：

```bash
# manager-configmap.yaml
AGENT_MANAGER_KUBERNETES_AGENT_IMAGE_PREFIX: "registry.example.com/maze-agent"
AGENT_MANAGER_KUBERNETES_AGENT_IMAGE_TAG: "v1.2.3"

# manager-deployment.yaml
image: registry.example.com/maze-manager:v1.2.3
```

## 容器命名规范

| 组件              | Docker Compose (service name) | K8s (Deployment/Service) |
| ----------------- | ----------------------------- | ------------------------ |
| Manager           | `agent-manager`               | `agent-manager`          |
| Web 前端          | `web`                         | `web`                    |
| Agent（动态创建） | `agent-claude-1` 等           | `maze-agent-{name}`      |

## 镜像策略

### 开发/测试环境

Manager 挂载 Docker socket，Agent 镜像按需动态构建，无需提前预构建：

| 镜像                  | 说明                                                         |
| --------------------- | ------------------------------------------------------------ |
| `maze-manager:latest` | Manager 后端（`make build-manager`）                         |
| `maze-web:latest`     | Web Nginx 前端（`make build-web`）                           |
| `maze-agent:latest`   | Agent 基础镜像（`make build-agent`，动态构建的 FROM 基础层） |
| `maze-agent:{name}`   | 按需动态构建（如 `maze-agent:dolores`）                      |

### 生产环境

推送到镜像仓库后，在 `overlays/production/` 中配置镜像前缀和 tag。

## RBAC 权限

Manager 需要以下 K8s API 权限来动态管理 Agent：

| 资源                     | 操作           | 用途                                   |
| ------------------------ | -------------- | -------------------------------------- |
| `apps/deployments`       | CRUD           | 创建/删除 Agent Deployment             |
| `pods`                   | get/list/watch | 查询 Agent Pod 状态                    |
| `pods/log`               | get            | 获取 Agent 运行时日志                  |
| `services`               | CRUD           | 创建/删除 Agent Service                |
| `persistentvolumeclaims` | CRUD           | 创建/删除 Agent PVC（production 模式） |

## Docker Compose vs Kubernetes 差异

| 维度       | Docker Compose                            | Kubernetes (dev)                            | Kubernetes (production)                 |
| ---------- | ----------------------------------------- | ------------------------------------------- | --------------------------------------- |
| 部署方式   | `docker compose up`                       | `make up`                                   | `make up ENV=prod`                      |
| Agent 构建 | Docker socket 动态构建                    | Docker socket 动态构建（相同）              | 预构建镜像选配                          |
| Agent 启动 | `docker run`                              | K8s API 创建 Deployment + Service           | K8s API 创建 Deployment + PVC + Service |
| 持久化     | bind mount (`~/.maze-dev/docker/agents/`) | hostPath (`~/.maze-dev/kubernetes/agents/`) | PVC                                     |
| 网络发现   | Docker DNS                                | K8s Service DNS                             | K8s Service DNS                         |
| 配置注入   | docker-compose.yml                        | ConfigMap + Secret                          | ConfigMap + Secret                      |
| 更新组件   | `docker compose up -d --build`            | `make update-*`                             | 推送新镜像 + `rollout restart`          |

## 集成测试

集成测试固定使用 `ENV=test`，支持 Docker 和 Kubernetes 两种平台：

```bash
# Docker 集成测试
make test-integration PLATFORM=docker

# Kubernetes 集成测试（自动部署 test 环境 + port-forward + 测试 + 清理）
make test-integration PLATFORM=kubernetes

# 运行单个测试
make test-integration PLATFORM=docker TEST_NAME=TestHostCreateOnline
```

Kubernetes 集成测试流程：

1. 自动创建 `maze-test` namespace
2. 用 test overlay 部署 Manager
3. 等待 Pod 就绪
4. 启动 port-forward（9090:8080 + 9091:8081）
5. 运行测试
6. 无论成功/失败/中断，自动清理 port-forward 和 namespace
