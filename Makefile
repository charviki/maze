PROJECT_ROOT := $(shell pwd)

MODULES = \
	fabrication/cradle \
	mesa-hub/behavior-panel/server \
	sweetwater/black-ridge/server

# ===== 环境配置 =====
# PLATFORM: docker 或 kubernetes（部署平台选择，所有命令共用）
PLATFORM ?= kubernetes
# ENV: dev / test / prod（运行环境，决定端口、命名空间、overlay 路径）
ENV ?= dev

# ===== ENV 派生变量 =====
ifeq ($(ENV),dev)
  K8S_NAMESPACE := maze-dev
  K8S_OVERLAY := overlays/dev
  COMPOSE_PROJECT := maze-dev
  HOST_DATA_DIR := $(HOME)/.maze-dev
  PORT_MANAGER := 7090
  PORT_WEB := 7080
else ifeq ($(ENV),test)
  K8S_NAMESPACE := maze-test
  K8S_OVERLAY := overlays/test
  COMPOSE_PROJECT := maze-test
  HOST_DATA_DIR := $(HOME)/.maze-test
  PORT_MANAGER := 9090
  PORT_WEB := 9080
else ifeq ($(ENV),prod)
  K8S_NAMESPACE := maze-prod
  K8S_OVERLAY := overlays/production
  COMPOSE_PROJECT := maze-prod
  HOST_DATA_DIR := $(HOME)/.maze-prod
  PORT_MANAGER := 8090
  PORT_WEB := 10800
endif

# Export variables so docker compose can resolve ${VAR:-default} in YAML
export PORT_WEB PORT_MANAGER HOST_DATA_DIR

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

# K8s overlay 目录（由 K8S_OVERLAY 派生）
K8S_OVERLAY_DIR := $(PROJECT_ROOT)/fabrication/kubernetes/$(K8S_OVERLAY)

# ===== 集成测试 =====
TEST_DIR := $(PROJECT_ROOT)/fabrication/tests/integration
TEST_NAME ?=

# ============================================================
#  通用命令（不区分环境）
# ============================================================

.PHONY: help \
        build build-manager build-web build-agent build-deps \
        vet test check check-frontend \
        gen-proto gen-client gen-sdk gen \
        up down status \
        deploy undeploy \
        proxy proxy-web proxy-manager \
        update-manager update-web update-agent update-all \
        test-integration

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "  PLATFORM=docker|kubernetes (default: kubernetes)"
	@echo "  ENV=dev|test|prod (default: dev)"
	@echo ""
	@echo "  示例:"
	@echo "    make up                                # K8s 部署（默认）"
	@echo "    make up PLATFORM=docker                # Docker Compose 启动"
	@echo "    make up PLATFORM=kubernetes ENV=prod"
	@echo "    make test-integration PLATFORM=docker"
	@echo "    make gen                                # 一键生成 proto + HTTP client"
	@echo "    make test-integration                  # K8s 集成测试"

# ============================================================
#  代码生成（proto → Go + HTTP client）
# ============================================================

CRADLE_API_DIR := $(PROJECT_ROOT)/fabrication/cradle/api

gen-proto: ## buf generate 生成 Go 类型 + gRPC stub + grpc-gateway + OpenAPI spec
	cd $(CRADLE_API_DIR) && buf generate

gen-client: gen-proto ## 重新生成 OpenAPI Go HTTP client（需要 openapi-generator + Java）
	cd $(CRADLE_API_DIR) && \
		if [ -z "$$JAVA_HOME" ]; then \
			echo "Error: JAVA_HOME not set. Install Java and set JAVA_HOME."; exit 1; \
		fi && \
		export PATH="$$JAVA_HOME/bin:$$PATH" && \
		openapi-generator generate \
			-i gen/openapiv2/maze.swagger.json \
			-g go \
			-o gen/http \
			--package-name client \
			--additional-properties=isGoSubmodule=true,withGoMod=false,enumClassPrefix=true

gen-sdk: ## 生成 TypeScript SDK from OpenAPI spec (fabrication/skin)
	@echo "\033[0;32m[gen-sdk]\033[0m Generating TypeScript SDK..."
	cd fabrication/skin && JAVA_HOME=$(JAVA_HOME) npx openapi-generator-cli generate \
		-i ../cradle/api/gen/http/api/openapi.yaml \
		-c openapi-generator-config.yaml \
		-o src/api/gen \
		-g typescript-fetch
	@# SDK gen 文件包含跨文件引用的内部 helper，在消费端 tsc -b 编译时会触发 noUnusedLocals。
	@# @ts-nocheck 只加在自动生成的文件上，不影响业务代码的类型安全。
	find fabrication/skin/src/api/gen -name '*.ts' -exec sed -i '' '/^\/\/ @ts-nocheck$$/d' {} \;
	find fabrication/skin/src/api/gen -name '*.ts' -exec sed -i '' '1i\
// @ts-nocheck' {} \;
	@echo "\033[0;32m[gen-sdk]\033[0m SDK generated at fabrication/skin/src/api/gen/"

gen: gen-client gen-sdk ## 一键生成 proto + HTTP client + TypeScript SDK

# ============================================================
#  Go 编译 / 检查 / 测试
# ============================================================

build-go: ## 编译所有 Go 模块
	@for m in $(MODULES); do \
		echo "\033[0;32m[build]\033[0m $$m"; \
		cd $$m && go build ./... || exit 1; \
		cd $(PROJECT_ROOT); \
	done

vet: ## Go 静态检查（快速模式，仅 go vet）
	@for m in $(MODULES); do \
		echo "\033[0;32m[vet]\033[0m $$m"; \
		cd $$m && go vet ./... || exit 1; \
		cd $(PROJECT_ROOT); \
	done

lint: ## Go 全量代码检查（golangci-lint v2，含 gosec + staticcheck + 30+ linter）
	@echo "\033[0;32m[lint]\033[0m Running golangci-lint on all Go modules..."
	@for m in $(MODULES); do \
		echo "  --> Linting $$m"; \
		cd $$m && golangci-lint run --timeout=5m ./... || exit 1; \
		cd $(PROJECT_ROOT); \
	done
	@echo "  --> Linting integration tests"
	@cd fabrication/tests/integration && golangci-lint run --timeout=5m ./... || exit 1; cd $(PROJECT_ROOT)

lint-fix: ## Go 自动修复可修 lint 问题
	@echo "\033[0;32m[lint-fix]\033[0m Auto-fixing lint issues..."
	@for m in $(MODULES); do \
		cd $$m && golangci-lint run --fix --timeout=5m ./... || exit 1; \
		cd $(PROJECT_ROOT); \
	done

vulncheck: ## Go 漏洞扫描（govulncheck）
	@echo "\033[0;32m[vulncheck]\033[0m Running govulncheck..."
	@for m in $(MODULES); do \
		echo "  --> Scanning $$m"; \
		cd $$m && govulncheck ./... || exit 1; \
		cd $(PROJECT_ROOT); \
	done

test: ## 运行所有 Go 单元测试
	@for m in $(MODULES); do \
		echo "\033[0;32m[test]\033[0m $$m"; \
		cd $$m && go test ./... -count=1 || exit 1; \
		cd $(PROJECT_ROOT); \
	done

check: build-go lint test ## 编译 + golangci-lint + 单元测试（Go 交付铁律）
	@echo "\033[0;32m[check]\033[0m All Go checks passed!"

# ===== 前端检查 =====

FRONTEND_MODULES := fabrication/skin mesa-hub/behavior-panel/web sweetwater/black-ridge/web

check-frontend: ## 前端三道检查：tsc → eslint → vitest（每个模块按序执行，任何一步失败即中止）
	@cd fabrication/skin && npx tsc --noEmit || exit 1 && npx eslint . || exit 1 && npx vitest run || exit 1; \
	cd $(PROJECT_ROOT)
	@cd mesa-hub/behavior-panel/web && npx tsc -b --noEmit || exit 1 && npx eslint . || exit 1 && npx vitest run || exit 1; \
	cd $(PROJECT_ROOT)
	@cd sweetwater/black-ridge/web && npx tsc -b --noEmit || exit 1 && npx eslint . || exit 1 && npx vitest run || exit 1; \
	cd $(PROJECT_ROOT)
	@echo "\033[0;32m[check-frontend]\033[0m All frontend checks passed!"

format-js: ## 格式化所有 TS/TSX/JSON/MD 文件
	@echo "\033[0;32m[format-js]\033[0m Formatting with Prettier..."
	pnpm run format

format-js-check: ## 检查 TS 格式（CI 用）
	@echo "\033[0;32m[format-js-check]\033[0m Checking formatting..."
	pnpm run format:check

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

deploy: ## 部署服务（根据 PLATFORM 自动选择 Docker Compose 或 K8s）
ifeq ($(PLATFORM),docker)
	@echo "\033[0;32m[INFO]\033[0m Starting with Docker Compose..."
	@mkdir -p $(HOST_DATA_DIR)/docker/agents
	docker compose -f $(COMPOSE_FILE) -p $(COMPOSE_PROJECT) up -d
	@echo ""
	@echo "\033[0;32m[INFO]\033[0m Maze is running on Docker Compose!"
	@echo "  Web:      http://localhost:$(PORT_WEB)"
	@echo "  Manager:  http://localhost:$(PORT_MANAGER)/health"
else ifeq ($(PLATFORM),kubernetes)
	@echo "\033[0;32m[INFO]\033[0m Deploying to Kubernetes ($(ENV))..."
	@mkdir -p $(HOST_DATA_DIR)/docker/agents $(HOST_DATA_DIR)/kubernetes/agents
	@kubectl create namespace $(K8S_NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	@kubectl kustomize $(K8S_OVERLAY_DIR) | sed 's|__HOST_HOME__|$(HOME)|g' | kubectl apply -f -
	@echo "\033[0;32m[INFO]\033[0m Waiting for pods to be ready..."
	@kubectl wait --for=condition=ready pods -l app --timeout=120s -n $(K8S_NAMESPACE) 2>/dev/null || \
		(echo "\033[1;33m[WARN]\033[0m Timeout or partial readiness." && kubectl get pods -n $(K8S_NAMESPACE))
	@echo ""
	@echo "\033[0;32m[INFO]\033[0m Maze is running on Kubernetes! ($(ENV))"
	@echo "  Next: make proxy"
endif

down: ## 停止并删除所有服务
ifeq ($(PLATFORM),docker)
	@echo "\033[0;32m[INFO]\033[0m Stopping Docker Compose services..."
	docker compose -f $(COMPOSE_FILE) -p $(COMPOSE_PROJECT) down
else ifeq ($(PLATFORM),kubernetes)
	@echo "\033[0;32m[INFO]\033[0m Removing Kubernetes deployment..."
	kubectl delete namespace $(K8S_NAMESPACE) --ignore-not-found=true
endif

undeploy: down ## down 的别名

status: ## 查看运行状态
ifeq ($(PLATFORM),docker)
	@docker compose -f $(COMPOSE_FILE) -p $(COMPOSE_PROJECT) ps
else ifeq ($(PLATFORM),kubernetes)
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
ifeq ($(PLATFORM),docker)
	@echo "\033[0;32m[INFO]\033[0m Docker mode: ports already exposed."
	@echo "  Web:      http://localhost:$(PORT_WEB)"
	@echo "  Manager:  http://localhost:$(PORT_MANAGER)/health"
else ifeq ($(PLATFORM),kubernetes)
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
ifeq ($(PLATFORM),kubernetes)
	kubectl port-forward svc/web $(PORT_WEB):80 -n $(K8S_NAMESPACE)
else
	@echo "Docker mode: http://localhost:$(PORT_WEB)"
endif

proxy-manager: ## 只代理 Manager API
ifeq ($(PLATFORM),kubernetes)
	@bash -c '\
		trap "kill 0" SIGINT SIGTERM; \
		kubectl port-forward svc/agent-manager $(PORT_MANAGER):8080 -n $(K8S_NAMESPACE) & \
		wait'
else
	@echo "Docker mode: http://localhost:$(PORT_MANAGER)/health"
endif

# ============================================================
#  滚动更新（K8s）/ 重启（Docker）
# ============================================================

update-manager: build-manager ## 重建 Manager 镜像 + 重启
ifeq ($(PLATFORM),docker)
	docker compose -f $(COMPOSE_FILE) -p $(COMPOSE_PROJECT) up -d agent-manager
else ifeq ($(PLATFORM),kubernetes)
	kubectl rollout restart deployment/agent-manager -n $(K8S_NAMESPACE)
	kubectl rollout status deployment/agent-manager -n $(K8S_NAMESPACE) --timeout=120s
endif

update-web: build-web ## 重建 Web 镜像 + 重启
ifeq ($(PLATFORM),docker)
	docker compose -f $(COMPOSE_FILE) -p $(COMPOSE_PROJECT) up -d web
else ifeq ($(PLATFORM),kubernetes)
	kubectl rollout restart deployment/web -n $(K8S_NAMESPACE)
	kubectl rollout status deployment/web -n $(K8S_NAMESPACE) --timeout=120s
endif

update-agent: build-agent ## 重建 Agent 基础镜像
	@echo "\033[0;32m[INFO]\033[0m Agent base image updated. New Hosts will use the updated image."

update-all: build-manager build-web build-agent ## 全部更新：重建所有镜像 + 重启
ifeq ($(PLATFORM),docker)
	docker compose -f $(COMPOSE_FILE) -p $(COMPOSE_PROJECT) up -d
else ifeq ($(PLATFORM),kubernetes)
	kubectl rollout restart deployment/agent-manager -n $(K8S_NAMESPACE)
	kubectl rollout restart deployment/web -n $(K8S_NAMESPACE)
	kubectl rollout status deployment/agent-manager -n $(K8S_NAMESPACE) --timeout=120s
	kubectl rollout status deployment/web -n $(K8S_NAMESPACE) --timeout=120s
endif
	@echo "\033[0;32m[INFO]\033[0m All services updated."

# ============================================================
#  集成测试（强制 ENV=test）
# ============================================================

# 集成测试无论用户传入什么 ENV，一律使用 test 环境的派生变量
test-integration: override K8S_NAMESPACE := maze-test
test-integration: override K8S_OVERLAY := overlays/test
test-integration: override COMPOSE_PROJECT := maze-test
test-integration: override HOST_DATA_DIR := $(HOME)/.maze-test
test-integration: override PORT_MANAGER := 9090
test-integration: override PORT_WEB := 9080
test-integration: override K8S_OVERLAY_DIR := $(PROJECT_ROOT)/fabrication/kubernetes/overlays/test

test-integration: ## 运行集成测试（PLATFORM=docker 或 PLATFORM=kubernetes，TEST_NAME=TestX 运行单个测试）
ifeq ($(PLATFORM),docker)
	@START=$$(date +%s); \
	echo "\033[0;32m[INFO]\033[0m [1/4] Building Agent base image..."; \
	docker build -f $(AGENT_DOCKERFILE) -t $(AGENT_IMAGE) $(PROJECT_ROOT); \
	ELAPSED=$$(($$(date +%s) - $$START)); \
	echo "\033[0;36m[TIME]\033[0m Agent base image built in $${ELAPSED}s"; \
	echo "\033[0;32m[INFO]\033[0m [2/4] Starting test environment (docker-compose)..."; \
	docker images --filter label=maze.dockerfile-hash -q | xargs -r docker rmi -f 2>/dev/null; \
	mkdir -p $(HOST_DATA_DIR)/docker/agents; \
	START2=$$(date +%s); \
	docker compose -f $(COMPOSE_TEST_FILE) -p $(COMPOSE_PROJECT) up -d --build; \
	ELAPSED2=$$(($$(date +%s) - $$START2)); \
	echo "\033[0;36m[TIME]\033[0m Test environment started in $${ELAPSED2}s"; \
	echo "\033[0;32m[INFO]\033[0m [3/4] Waiting for Manager to be ready..."; \
	START3=$$(date +%s); \
	bash -c 'for i in $$(seq 1 60); do \
		curl -sf http://localhost:$(PORT_MANAGER)/health > /dev/null 2>&1 && break; \
		echo "  waiting... ($$i/60)"; sleep 2; \
	done'; \
	ELAPSED3=$$(($$(date +%s) - $$START3)); \
	echo "\033[0;36m[TIME]\033[0m Manager ready in $${ELAPSED3}s"; \
	TOTAL_SETUP=$$(($$(date +%s) - $$START)); \
	echo "\033[0;32m[INFO]\033[0m [4/4] Running integration tests (env=docker)..."; \
	echo "\033[0;36m[TIME]\033[0m Total setup time: $${TOTAL_SETUP}s"; \
	cd $(TEST_DIR) && \
		MAZE_TEST_ENV=$(PLATFORM) \
		MAZE_TEST_DATA_DIR=$(HOST_DATA_DIR) \
		MAZE_TEST_AGENT_STORAGE_BACKEND=bind \
		MAZE_TEST_MANAGER_URL=http://localhost:$(PORT_MANAGER) \
		MAZE_TEST_AUTH_TOKEN=test-integration-token \
		go test -v -count=1 -tags=integration -timeout=10m $(if $(TEST_NAME),-run $(TEST_NAME),) ./...; \
		TEST_EXIT=$$?; \
		echo "\033[0;32m[INFO]\033[0m Stopping test environment..."; \
		docker ps -q --filter label=maze-host | xargs -r docker stop 2>/dev/null; \
		docker ps -aq --filter label=maze-host | xargs -r docker rm 2>/dev/null; \
		docker compose -f $(COMPOSE_TEST_FILE) -p $(COMPOSE_PROJECT) down -v --remove-orphans; \
		exit $$TEST_EXIT
		exit $$TEST_EXIT
else ifeq ($(PLATFORM),kubernetes)
	@bash -c '\
		INFO="\033[0;32m[INFO]\033[0m"; \
		ERROR="\033[0;31m[ERROR]\033[0m"; \
		WARN="\033[1;33m[WARN]\033[0m"; \
		NS="$(K8S_NAMESPACE)"; \
		PF_PIDS=""; \
		DEPLOYED=false; \
		lsof -ti:$(PORT_MANAGER) | xargs kill -9 2>/dev/null; \
		cleanup() { \
			echo -e "$$INFO Cleaning up test environment..."; \
			if [ -n "$$PF_PIDS" ]; then \
				echo -e "$$INFO Stopping port-forward..."; \
				kill $$PF_PIDS 2>/dev/null; \
			fi; \
			if [ "$$DEPLOYED" = "true" ]; then \
				echo -e "$$INFO Deleting namespace $$NS..."; \
				kubectl delete namespace $$NS --ignore-not-found=true --wait=false 2>/dev/null; \
			fi; \
		}; \
		trap cleanup EXIT INT TERM; \
		echo -e "$$INFO [1/5] Building local images for Kubernetes test env..."; \
		docker build -f $(AGENT_DOCKERFILE) -t $(AGENT_IMAGE) $(PROJECT_ROOT); \
		docker build -f $(MANAGER_DOCKERFILE) -t $(MANAGER_IMAGE) $(PROJECT_ROOT); \
		echo -e "$$INFO [2/5] Recreating test namespace $$NS..."; \
		kubectl delete namespace $$NS --ignore-not-found=true --wait=true 2>/dev/null || true; \
		kubectl create namespace $$NS --dry-run=client -o yaml | kubectl apply -f -; \
		DEPLOYED=true; \
		echo -e "$$INFO [3/5] Deploying test environment to namespace $$NS..."; \
		kubectl kustomize $(K8S_OVERLAY_DIR) | sed "s|__HOST_HOME__|$(HOME)|g" | kubectl apply -f -; \
		if [ $$? -ne 0 ]; then \
			echo -e "$$ERROR Failed to deploy test environment."; \
			exit 1; \
		fi; \
		echo -e "$$INFO Waiting for pods to be ready..."; \
		kubectl wait --for=condition=ready pods -l app --timeout=120s -n $$NS 2>/dev/null || \
			(echo -e "$$WARN Timeout or partial readiness." && kubectl get pods -n $$NS); \
		echo -e "$$INFO [4/5] Starting port-forward..."; \
		kubectl port-forward svc/agent-manager $(PORT_MANAGER):8080 -n $$NS 2>&1 | grep -v "^Handling" & \
		PF_PID1=$$!; \
		PF_PIDS="$$PF_PID1"; \
		for i in $$(seq 1 30); do \
			if curl -sf http://localhost:$(PORT_MANAGER)/health > /dev/null 2>&1; then \
				break; \
			fi; \
			if ! kill -0 $$PF_PID1 2>/dev/null; then \
				echo -e "$$ERROR port-forward process died. Check kubectl and cluster status."; \
				exit 1; \
			fi; \
			sleep 2; \
		done; \
		if ! curl -sf http://localhost:$(PORT_MANAGER)/health > /dev/null 2>&1; then \
			echo -e "$$ERROR port-forward not ready after 60s."; \
			exit 1; \
		fi; \
		echo -e "$$INFO Port-forward active: manager=$(PORT_MANAGER)"; \
		echo -e "$$INFO [5/5] Running tests..."; \
		cd $(TEST_DIR) && \
			MAZE_TEST_ENV=$(PLATFORM) \
			MAZE_TEST_DATA_DIR=$(HOST_DATA_DIR) \
			MAZE_TEST_AGENT_STORAGE_BACKEND=hostpath \
			MAZE_TEST_MANAGER_URL=http://localhost:$(PORT_MANAGER) \
			MAZE_TEST_AUTH_TOKEN=test-integration-token \
			MAZE_TEST_NAMESPACE=$(K8S_NAMESPACE) \
			go test -v -count=1 -tags=integration -timeout=10m $(if $(TEST_NAME),-run $(TEST_NAME),) ./...'
endif
