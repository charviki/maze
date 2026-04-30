NAMESPACE ?= maze
SCRIPT_DIR := $(shell dirname $(lastword $(MAKEFILE_LIST)))
PROJECT_ROOT := $(shell cd $(SCRIPT_DIR)/../.. && pwd)

# 环境选择：local（默认）或 production
ENV ?= local
OVERLAY_DIR := $(SCRIPT_DIR)/overlays/$(ENV)

# 镜像名称（仅 local 环境使用）
MANAGER_IMAGE := maze-manager:latest
WEB_IMAGE := maze-web:latest
AGENT_IMAGE := maze-agent:claude-latest

# Dockerfile 路径（绝对路径）
MANAGER_DOCKERFILE := $(PROJECT_ROOT)/mesa-hub/behavior-panel/Dockerfile
WEB_DOCKERFILE := $(PROJECT_ROOT)/mesa-hub/behavior-panel/Dockerfile.web
AGENT_DOCKERFILE := $(PROJECT_ROOT)/sweetwater/black-ridge/Dockerfile

.PHONY: help up down status \
        build build-manager build-web build-agent \
        deploy undeploy \
        proxy proxy-web proxy-manager \
        update-manager update-web update-agent update-all

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "  环境选择: ENV=local (默认) 或 ENV=production"
	@echo "  示例: make up ENV=production"

# ===== 构建镜像 =====

build: build-manager build-web build-agent ## 构建全部镜像

build-manager: ## 构建 Manager 镜像
	@echo "\033[0;32m[INFO]\033[0m Building Manager image..."
	docker build -f $(MANAGER_DOCKERFILE) -t $(MANAGER_IMAGE) $(PROJECT_ROOT)

build-web: ## 构建 Web Nginx 镜像
	@echo "\033[0;32m[INFO]\033[0m Building Web Nginx image..."
	docker build -f $(WEB_DOCKERFILE) -t $(WEB_IMAGE) $(PROJECT_ROOT)

build-agent: ## 构建 Agent 基础镜像（用于 Agent 基础层，动态构建时需要）
	@echo "\033[0;32m[INFO]\033[0m Building Agent base image..."
	docker build -f $(AGENT_DOCKERFILE) -t maze-agent:latest $(PROJECT_ROOT)

# ===== 部署管理 =====

up: build deploy ## 一键部署：构建镜像 + 部署 K8s 资源 + 等待就绪
	@echo ""
	@echo "\033[0;32m[INFO]\033[0m ========================================="
	@echo "\033[0;32m[INFO]\033[0m Maze is running on Kubernetes! ($(ENV))"
	@echo "\033[0;32m[INFO]\033[0m ========================================="
	@echo ""
	@echo "  Next step:"
	@echo "    \033[36mmake proxy\033[0m  # http://localhost:10800"
	@echo ""
	@echo "  Other commands:"
	@echo "    make status       # 查看运行状态"
	@echo "    make update-all   # 代码更新后重建并滚动重启"

deploy: ## 部署 K8s 资源并等待就绪
ifeq ($(ENV),local)
	@echo "\033[0;32m[INFO]\033[0m Creating local directories..."
	@mkdir -p $(HOME)/.maze/docker/agents $(HOME)/.maze/kubernetes/agents
	@echo "\033[0;32m[INFO]\033[0m Deploying K8s resources ($(ENV) overlay)..."
	@kubectl apply -k $(SCRIPT_DIR)/base
	@sed 's|__HOST_HOME__|$(HOME)|g' $(OVERLAY_DIR)/manager-configmap.yaml | kubectl apply -f -
	@sed 's|__HOST_HOME__|$(HOME)|g' $(OVERLAY_DIR)/manager-deployment-patch.yaml | kubectl apply -f -
else
	@echo "\033[0;32m[INFO]\033[0m Deploying K8s resources ($(ENV) overlay)..."
	@echo "\033[1;33m[WARN]\033[0m Production overlay contains REGISTRY/VERSION placeholders."
	@echo "\033[1;33m[WARN]\033[0m Please edit overlays/production/*.yaml before deploying."
	@kubectl apply -k $(OVERLAY_DIR)
endif
	@echo "\033[0;32m[INFO]\033[0m Waiting for pods to be ready..."
	@kubectl wait --for=condition=ready pods -l app --timeout=120s -n $(NAMESPACE) 2>/dev/null || \
		(echo "\033[1;33m[WARN]\033[0m Timeout or partial readiness. Current status:" && kubectl get pods -n $(NAMESPACE))

down: ## 删除所有 K8s 资源
	@echo "\033[0;32m[INFO]\033[0m Removing Maze deployment..."
	kubectl delete namespace $(NAMESPACE) --ignore-not-found=true

undeploy: down ## down 的别名

status: ## 查看 Pod/Service/PVC 状态
	@echo "=== Pods ==="
	@kubectl get pods -n $(NAMESPACE) 2>/dev/null || echo "  No pods found (namespace may not exist)"
	@echo ""
	@echo "=== Services ==="
	@kubectl get svc -n $(NAMESPACE) 2>/dev/null || echo "  No services found"
	@echo ""
	@echo "=== PVCs ==="
	@kubectl get pvc -n $(NAMESPACE) 2>/dev/null || echo "  No PVCs found"

# ===== 本地访问 =====

PORT_WEB ?= 10800
PORT_MANAGER ?= 8090

proxy: ## 启动 port-forward 代理（Web:10800 + Manager:8090），Ctrl+C 停止
	@echo "\033[0;32m[INFO]\033[0m Starting port-forward..."
	@echo "  Web:      http://localhost:$(PORT_WEB)"
	@echo "  Manager:  http://localhost:$(PORT_MANAGER)/health"
	@kubectl port-forward svc/web $(PORT_WEB):80 -n $(NAMESPACE) &
	@kubectl port-forward svc/agent-manager $(PORT_MANAGER):8080 -n $(NAMESPACE) &
	@wait

proxy-web: ## 只代理 Web 前端
	kubectl port-forward svc/web $(PORT_WEB):80 -n $(NAMESPACE)

proxy-manager: ## 只代理 Manager API
	kubectl port-forward svc/agent-manager $(PORT_MANAGER):8080 -n $(NAMESPACE)

# ===== 滚动更新 =====

update-manager: build-manager ## 更新 Manager：重建镜像 + 滚动重启
	@echo "\033[0;32m[INFO]\033[0m Rolling restart agent-manager..."
	kubectl rollout restart deployment/agent-manager -n $(NAMESPACE)
	kubectl rollout status deployment/agent-manager -n $(NAMESPACE) --timeout=120s

update-web: build-web ## 更新 Web：重建镜像 + 滚动重启
	@echo "\033[0;32m[INFO]\033[0m Rolling restart web..."
	kubectl rollout restart deployment/web -n $(NAMESPACE)
	kubectl rollout status deployment/web -n $(NAMESPACE) --timeout=120s

update-agent: build-agent ## 更新 Agent 基础镜像（新创建的 Host 使用新镜像）
	@echo "\033[0;32m[INFO]\033[0m Agent base image updated."
	@echo "\033[0;32m[INFO]\033[0m Newly created Hosts will use the updated image."
	@echo "\033[0;32m[INFO]\033[0m To update existing agents, delete and recreate via Manager API."

update-all: build-manager build-web build-agent ## 全部更新：重建所有镜像 + 滚动重启 Manager 和 Web
	@echo "\033[0;32m[INFO]\033[0m Rolling restart all deployments..."
	kubectl rollout restart deployment/agent-manager -n $(NAMESPACE)
	kubectl rollout restart deployment/web -n $(NAMESPACE)
	kubectl rollout status deployment/agent-manager -n $(NAMESPACE) --timeout=120s
	kubectl rollout status deployment/web -n $(NAMESPACE) --timeout=120s
	@echo "\033[0;32m[INFO]\033[0m Agent base image updated for new Hosts."
