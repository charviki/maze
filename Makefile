PROJECT_ROOT := $(shell pwd)

MODULES = \
	fabrication/cradle \
	mesa-hub/behavior-panel/server \
	sweetwater/black-ridge/server

# ===== 环境配置 =====
# ENV: docker 或 kubernetes（全局环境选择，所有命令共用）
ENV ?= kubernetes
# K8S_ENV: local 或 production（仅 kubernetes 环境使用，控制 K8s overlay）
K8S_ENV ?= local
# K8S_NAMESPACE: K8s namespace（仅 kubernetes 环境使用）
K8S_NAMESPACE ?= maze

# ===== 镜像配置 =====
MANAGER_IMAGE := maze-manager:latest
WEB_IMAGE := maze-web:latest
AGENT_IMAGE := maze-agent:latest

MANAGER_DOCKERFILE := $(PROJECT_ROOT)/mesa-hub/behavior-panel/Dockerfile
WEB_DOCKERFILE := $(PROJECT_ROOT)/mesa-hub/behavior-panel/Dockerfile.web
AGENT_DOCKERFILE := $(PROJECT_ROOT)/sweetwater/black-ridge/Dockerfile
MOLDS_DIR := $(PROJECT_ROOT)/fabrication/molds

# Docker Compose 文件
COMPOSE_FILE := $(PROJECT_ROOT)/mesa-hub/behavior-panel/docker-compose.yml
COMPOSE_TEST_FILE := $(PROJECT_ROOT)/fabrication/tests/integration/docker-compose.test.yml
COMPOSE_TEST_PROJECT := maze-test

# K8s overlay 目录
K8S_OVERLAY_DIR := $(PROJECT_ROOT)/fabrication/kubernetes/overlays/$(K8S_ENV)

# ===== 端口配置 =====
PORT_WEB ?= 10800
PORT_MANAGER ?= 8090

# ===== 集成测试 =====
TEST_DIR := $(PROJECT_ROOT)/fabrication/tests/integration
TEST_NAME ?=

# ============================================================
#  通用命令（不区分环境）
# ============================================================

.PHONY: help \
        build build-manager build-web build-agent build-deps \
        vet test check \
        up down status \
        deploy undeploy \
        proxy proxy-web proxy-manager \
        update-manager update-web update-agent update-all \
        test-integration

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "  环境选择: ENV=kubernetes (默认) 或 ENV=docker"
	@echo "  K8s overlay: K8S_ENV=local (默认) 或 K8S_ENV=production"
	@echo ""
	@echo "  示例:"
	@echo "    make up                                # K8s 部署（默认）"
	@echo "    make up ENV=docker                     # Docker Compose 启动"
	@echo "    make up ENV=kubernetes K8S_ENV=production"
	@echo "    make test-integration ENV=docker"
	@echo "    make test-integration                  # K8s 集成测试"

# ============================================================
#  Go 编译 / 检查 / 测试
# ============================================================

build-go: ## 编译所有 Go 模块
	@for m in $(MODULES); do \
		echo "\033[0;32m[build]\033[0m $$m"; \
		go build ./$$m/... || exit 1; \
	done

vet: ## Go 静态检查
	@for m in $(MODULES); do \
		echo "\033[0;32m[vet]\033[0m $$m"; \
		go vet ./$$m/... || exit 1; \
	done

test: ## 运行所有 Go 单元测试
	@for m in $(MODULES); do \
		echo "\033[0;32m[test]\033[0m $$m"; \
		go test ./$$m/... -count=1 || exit 1; \
	done

check: build-go vet test ## 编译 + 静态检查 + 单元测试
	@echo "\033[0;32m[check]\033[0m All checks passed!"

# ============================================================
#  Docker 镜像构建
# ============================================================

build: build-deps build-manager build-web build-agent ## 构建全部 Docker 镜像

build-manager: ## 构建 Manager 镜像
	@echo "\033[0;32m[INFO]\033[0m Building Manager image..."
	docker build -f $(MANAGER_DOCKERFILE) -t $(MANAGER_IMAGE) $(PROJECT_ROOT)

build-web: ## 构建 Web Nginx 镜像
	@echo "\033[0;32m[INFO]\033[0m Building Web Nginx image..."
	docker build -f $(WEB_DOCKERFILE) -t $(WEB_IMAGE) $(PROJECT_ROOT)

build-agent: ## 构建 Agent 基础镜像
	@echo "\033[0;32m[INFO]\033[0m Building Agent base image..."
	docker build -f $(AGENT_DOCKERFILE) -t $(AGENT_IMAGE) $(PROJECT_ROOT)

build-deps: ## 构建所有供应商镜像（claude/codex/go/python/node）
	@echo "\033[0;32m[INFO]\033[0m Building supplier images..."
	@for dep in claude codex go python node; do \
		echo "\033[0;32m[INFO]\033[0m   Building maze-deps-$${dep}:latest..."; \
		docker build -f $(MOLDS_DIR)/Dockerfile.$${dep} -t maze-deps-$${dep}:latest $(PROJECT_ROOT); \
	done

# ============================================================
#  部署（Docker Compose / Kubernetes 自动切换）
# ============================================================

up: build deploy ## 一键部署：构建镜像 + 启动服务

deploy: ## 部署服务（根据 ENV 自动选择 Docker Compose 或 K8s）
ifeq ($(ENV),docker)
	@echo "\033[0;32m[INFO]\033[0m Starting with Docker Compose..."
	@mkdir -p $(HOME)/.maze/docker/agents
	docker compose -f $(COMPOSE_FILE) up -d
	@echo ""
	@echo "\033[0;32m[INFO]\033[0m Maze is running on Docker Compose!"
	@echo "  Web:      http://localhost:$(PORT_WEB)"
	@echo "  Manager:  http://localhost:$(PORT_MANAGER)/health"
else ifeq ($(ENV),kubernetes)
	@echo "\033[0;32m[INFO]\033[0m Deploying to Kubernetes ($(K8S_ENV))..."
	@mkdir -p $(HOME)/.maze/docker/agents $(HOME)/.maze/kubernetes/agents
ifeq ($(K8S_ENV),local)
	@kubectl apply -k $(PROJECT_ROOT)/fabrication/kubernetes/base
	@sed 's|__HOST_HOME__|$(HOME)|g' $(K8S_OVERLAY_DIR)/manager-configmap.yaml | kubectl apply -f -
	@sed 's|__HOST_HOME__|$(HOME)|g' $(K8S_OVERLAY_DIR)/manager-deployment-patch.yaml | kubectl apply -f -
else
	@echo "\033[1;33m[WARN]\033[0m Production overlay contains REGISTRY/VERSION placeholders."
	@kubectl apply -k $(K8S_OVERLAY_DIR)
endif
	@echo "\033[0;32m[INFO]\033[0m Waiting for pods to be ready..."
	@kubectl wait --for=condition=ready pods -l app --timeout=120s -n $(K8S_NAMESPACE) 2>/dev/null || \
		(echo "\033[1;33m[WARN]\033[0m Timeout or partial readiness." && kubectl get pods -n $(K8S_NAMESPACE))
	@echo ""
	@echo "\033[0;32m[INFO]\033[0m Maze is running on Kubernetes! ($(K8S_ENV))"
	@echo "  Next: make proxy"
endif

down: ## 停止并删除所有服务
ifeq ($(ENV),docker)
	@echo "\033[0;32m[INFO]\033[0m Stopping Docker Compose services..."
	docker compose -f $(COMPOSE_FILE) down
else ifeq ($(ENV),kubernetes)
	@echo "\033[0;32m[INFO]\033[0m Removing Kubernetes deployment..."
	kubectl delete namespace $(K8S_NAMESPACE) --ignore-not-found=true
endif

undeploy: down ## down 的别名

status: ## 查看运行状态
ifeq ($(ENV),docker)
	@docker compose -f $(COMPOSE_FILE) ps
else ifeq ($(ENV),kubernetes)
	@echo "=== Pods ==="
	@kubectl get pods -n $(K8S_NAMESPACE) 2>/dev/null || echo "  No pods found"
	@echo ""
	@echo "=== Services ==="
	@kubectl get svc -n $(K8S_NAMESPACE) 2>/dev/null || echo "  No services found"
	@echo ""
	@echo "=== PVCs ==="
	@kubectl get pvc -n $(K8S_NAMESPACE) 2>/dev/null || echo "  No PVCs found"
endif

# ============================================================
#  本地访问代理
# ============================================================

proxy: ## 启动 port-forward（K8s）或直接访问（Docker 已暴露端口）
ifeq ($(ENV),docker)
	@echo "\033[0;32m[INFO]\033[0m Docker mode: ports already exposed."
	@echo "  Web:      http://localhost:$(PORT_WEB)"
	@echo "  Manager:  http://localhost:$(PORT_MANAGER)/health"
else ifeq ($(ENV),kubernetes)
	@echo "\033[0;32m[INFO]\033[0m Starting port-forward..."
	@echo "  Web:      http://localhost:$(PORT_WEB)"
	@echo "  Manager:  http://localhost:$(PORT_MANAGER)/health"
	@bash -c '\
		trap "kill 0" SIGINT SIGTERM; \
		kubectl port-forward svc/web $(PORT_WEB):80 -n $(K8S_NAMESPACE) & \
		kubectl port-forward svc/agent-manager $(PORT_MANAGER):8080 -n $(K8S_NAMESPACE) & \
		wait'
endif

proxy-web: ## 只代理 Web 前端
ifeq ($(ENV),kubernetes)
	kubectl port-forward svc/web $(PORT_WEB):80 -n $(K8S_NAMESPACE)
else
	@echo "Docker mode: http://localhost:$(PORT_WEB)"
endif

proxy-manager: ## 只代理 Manager API
ifeq ($(ENV),kubernetes)
	kubectl port-forward svc/agent-manager $(PORT_MANAGER):8080 -n $(K8S_NAMESPACE)
else
	@echo "Docker mode: http://localhost:$(PORT_MANAGER)/health"
endif

# ============================================================
#  滚动更新（K8s）/ 重启（Docker）
# ============================================================

update-manager: build-manager ## 重建 Manager 镜像 + 重启
ifeq ($(ENV),docker)
	docker compose -f $(COMPOSE_FILE) up -d agent-manager
else ifeq ($(ENV),kubernetes)
	kubectl rollout restart deployment/agent-manager -n $(K8S_NAMESPACE)
	kubectl rollout status deployment/agent-manager -n $(K8S_NAMESPACE) --timeout=120s
endif

update-web: build-web ## 重建 Web 镜像 + 重启
ifeq ($(ENV),docker)
	docker compose -f $(COMPOSE_FILE) up -d web
else ifeq ($(ENV),kubernetes)
	kubectl rollout restart deployment/web -n $(K8S_NAMESPACE)
	kubectl rollout status deployment/web -n $(K8S_NAMESPACE) --timeout=120s
endif

update-agent: build-agent ## 重建 Agent 基础镜像
	@echo "\033[0;32m[INFO]\033[0m Agent base image updated. New Hosts will use the updated image."

update-all: build-manager build-web build-agent ## 全部更新：重建所有镜像 + 重启
ifeq ($(ENV),docker)
	docker compose -f $(COMPOSE_FILE) up -d
else ifeq ($(ENV),kubernetes)
	kubectl rollout restart deployment/agent-manager -n $(K8S_NAMESPACE)
	kubectl rollout restart deployment/web -n $(K8S_NAMESPACE)
	kubectl rollout status deployment/agent-manager -n $(K8S_NAMESPACE) --timeout=120s
	kubectl rollout status deployment/web -n $(K8S_NAMESPACE) --timeout=120s
endif
	@echo "\033[0;32m[INFO]\033[0m All services updated."

# ============================================================
#  集成测试
# ============================================================

test-integration: ## 运行集成测试（ENV=docker 或 ENV=kubernetes，TEST_NAME=TestX 运行单个测试）
ifeq ($(ENV),docker)
	@START=$$(date +%s); \
	echo "\033[0;32m[INFO]\033[0m [1/4] Building Agent base image..."; \
	docker build -f $(AGENT_DOCKERFILE) -t $(AGENT_IMAGE) $(PROJECT_ROOT); \
	ELAPSED=$$(($$(date +%s) - $$START)); \
	echo "\033[0;36m[TIME]\033[0m Agent base image built in $${ELAPSED}s"; \
	echo "\033[0;32m[INFO]\033[0m [2/4] Starting test environment (docker-compose)..."; \
	mkdir -p $(HOME)/.maze-test/docker/agents; \
	START2=$$(date +%s); \
	docker compose -f $(COMPOSE_TEST_FILE) -p $(COMPOSE_TEST_PROJECT) up -d --build; \
	ELAPSED2=$$(($$(date +%s) - $$START2)); \
	echo "\033[0;36m[TIME]\033[0m Test environment started in $${ELAPSED2}s"; \
	echo "\033[0;32m[INFO]\033[0m [3/4] Waiting for Manager to be ready..."; \
	START3=$$(date +%s); \
	bash -c 'for i in $$(seq 1 60); do \
		curl -sf http://localhost:9090/health > /dev/null 2>&1 && break; \
		echo "  waiting... ($$i/60)"; sleep 2; \
	done'; \
	ELAPSED3=$$(($$(date +%s) - $$START3)); \
	echo "\033[0;36m[TIME]\033[0m Manager ready in $${ELAPSED3}s"; \
	TOTAL_SETUP=$$(($$(date +%s) - $$START)); \
	echo "\033[0;32m[INFO]\033[0m [4/4] Running integration tests (env=docker)..."; \
	echo "\033[0;36m[TIME]\033[0m Total setup time: $${TOTAL_SETUP}s"; \
	cd $(TEST_DIR) && \
		MAZE_TEST_ENV=$(ENV) \
		MAZE_TEST_MANAGER_URL=http://localhost:9090 \
		MAZE_TEST_AUTH_TOKEN=test-integration-token \
		go test -v -tags=integration -timeout=10m $(if $(TEST_NAME),-run $(TEST_NAME),) ./...; \
		TEST_EXIT=$$?; \
		echo "\033[0;32m[INFO]\033[0m Stopping test environment..."; \
		docker compose -f $(COMPOSE_TEST_FILE) -p $(COMPOSE_TEST_PROJECT) down -v --remove-orphans; \
		exit $$TEST_EXIT
else ifeq ($(ENV),kubernetes)
	@echo "\033[0;32m[INFO]\033[0m Running integration tests (env=kubernetes)..."
	@echo "\033[0;32m[INFO]\033[0m Ensure Manager is running in namespace maze-test and port-forward is active (9090)."
	@cd $(TEST_DIR) && \
		MAZE_TEST_ENV=$(ENV) \
		MAZE_TEST_MANAGER_URL=http://localhost:9090 \
		MAZE_TEST_AUTH_TOKEN=test-integration-token \
		MAZE_TEST_NAMESPACE=maze-test \
		go test -v -tags=integration -timeout=10m $(if $(TEST_NAME),-run $(TEST_NAME),) ./...
endif
