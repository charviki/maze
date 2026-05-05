PROJECT_ROOT := $(shell pwd)

MODULES = \
	fabrication/cradle \
	mesa-hub/behavior-panel/server \
	sweetwater/black-ridge/server

COVERAGE_MODULES = \
	fabrication/cradle \
	fabrication/tests/integration \
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
  PORT_POSTGRES := 5432
else ifeq ($(ENV),test)
  K8S_NAMESPACE := maze-test
  K8S_OVERLAY := overlays/test
  COMPOSE_PROJECT := maze-test
  HOST_DATA_DIR := $(HOME)/.maze-test
  PORT_MANAGER := 9090
  PORT_WEB := 9080
  PORT_POSTGRES := 5433
else ifeq ($(ENV),prod)
  K8S_NAMESPACE := maze-prod
  K8S_OVERLAY := overlays/production
  COMPOSE_PROJECT := maze-prod
  HOST_DATA_DIR := $(HOME)/.maze-prod
  PORT_MANAGER := 8090
  PORT_WEB := 10800
  PORT_POSTGRES := 5434
endif

# Export variables so docker compose can resolve ${VAR:-default} in YAML
export PORT_WEB PORT_MANAGER PORT_POSTGRES HOST_DATA_DIR

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
        vet test coverage check check-frontend \
        gen-proto gen-client gen-sdk gen \
        up down status \
        deploy undeploy \
        proxy proxy-web proxy-manager proxy-db \
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
	find fabrication/skin/src/api/gen -name '*.ts' -exec python3 -c 'from pathlib import Path; import sys; \
path = Path(sys.argv[1]); prefix = "// @ts-nocheck\n"; content = path.read_text(); \
lines = [line for line in content.splitlines() if line != "// @ts-nocheck"]; \
normalized = "\n".join(lines); normalized += "\n" if content.endswith("\n") else ""; \
path.write_text(prefix + normalized.lstrip("\n"))' {} \;
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

coverage: ## 生成 Go 覆盖率报告（对每个模块运行 go test -coverprofile）
	@echo "\033[0;32m[coverage]\033[0m Generating coverage reports..."
	@for m in $(COVERAGE_MODULES); do \
		echo "\033[0;32m[coverage]\033[0m $$m"; \
		cd $$m && go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | tail -1 || exit 1; \
		cd $(PROJECT_ROOT); \
	done
	@echo "\033[0;32m[coverage]\033[0m Coverage reports generated."

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
	# K8s 正常部署路径：
	# 1. 直接渲染仓库中的静态 overlay
	# 2. 仅等待当前 overlay 中实际存在的 deployment，再等待 Pod ready
	# 3. 全部就绪后提示精确的代理命令，避免默认 ENV=dev 误导
	#
	# 注意：这里和集成测试一样采用严格失败策略，避免 manager 在数据库还未就绪时过早启动。
	@echo "\033[0;32m[INFO]\033[0m Deploying to Kubernetes ($(ENV))..."
	@mkdir -p $(HOST_DATA_DIR)/docker/agents
	@if [ "$(ENV)" = "dev" ]; then \
		mkdir -p /tmp/maze-dev/kubernetes/agents; \
	elif [ "$(ENV)" = "test" ]; then \
		mkdir -p /tmp/maze-test/kubernetes/agents; \
	fi
	@kubectl create namespace $(K8S_NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	@kubectl kustomize "$(K8S_OVERLAY_DIR)" | kubectl apply -f -
	@echo "\033[0;32m[INFO]\033[0m Waiting for deployments to roll out..."
	@if kubectl get deployment/postgresql -n $(K8S_NAMESPACE) >/dev/null 2>&1; then \
		kubectl rollout status deployment/postgresql -n $(K8S_NAMESPACE) --timeout=180s || \
			(echo "\033[1;33m[WARN]\033[0m PostgreSQL rollout timed out." && kubectl get pods -n $(K8S_NAMESPACE) && exit 1); \
	else \
		echo "\033[0;32m[INFO]\033[0m Skip PostgreSQL rollout wait: deployment/postgresql not managed by overlay $(K8S_OVERLAY)."; \
	fi
	@kubectl rollout status deployment/agent-manager -n $(K8S_NAMESPACE) --timeout=180s || \
		(echo "\033[1;33m[WARN]\033[0m agent-manager rollout timed out." && kubectl get pods -n $(K8S_NAMESPACE) && exit 1)
	@kubectl rollout status deployment/web -n $(K8S_NAMESPACE) --timeout=180s || \
		(echo "\033[1;33m[WARN]\033[0m web rollout timed out." && kubectl get pods -n $(K8S_NAMESPACE) && exit 1)
	@echo "\033[0;32m[INFO]\033[0m Waiting for pods to be ready..."
	@kubectl wait --for=condition=ready pods -l app --timeout=180s -n $(K8S_NAMESPACE) 2>/dev/null || \
		(echo "\033[1;33m[WARN]\033[0m Timeout or partial readiness." && kubectl get pods -n $(K8S_NAMESPACE) && exit 1)
	@echo ""
	@echo "\033[0;32m[INFO]\033[0m Maze is running on Kubernetes! ($(ENV))"
	@echo "  Next: make proxy PLATFORM=$(PLATFORM) ENV=$(ENV)"
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
	@echo "  Postgres: postgresql://localhost:$(PORT_POSTGRES)"
else ifeq ($(PLATFORM),kubernetes)
	@echo "\033[0;32m[INFO]\033[0m Starting port-forward..."
	@echo "  Web:      http://localhost:$(PORT_WEB)"
	@echo "  Manager:  http://localhost:$(PORT_MANAGER)/health"
	@echo "  Postgres: postgresql://localhost:$(PORT_POSTGRES)"
	@bash -c '\
		trap "kill 0" SIGINT SIGTERM; \
		kubectl port-forward svc/web $(PORT_WEB):80 -n $(K8S_NAMESPACE) & \
		kubectl port-forward svc/agent-manager $(PORT_MANAGER):8080 -n $(K8S_NAMESPACE) & \
		kubectl port-forward svc/postgresql $(PORT_POSTGRES):5432 -n $(K8S_NAMESPACE) & \
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

proxy-db: ## 只代理 PostgreSQL
ifeq ($(PLATFORM),kubernetes)
	kubectl port-forward svc/postgresql $(PORT_POSTGRES):5432 -n $(K8S_NAMESPACE)
else
	@echo "Docker mode: postgresql://localhost:$(PORT_POSTGRES)"
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
	# Ensure compose env and spawned host containers are reclaimed even if the test is interrupted.
	@bash -c '\
		set -o pipefail; \
		INFO="\033[0;32m[INFO]\033[0m"; \
		ERROR="\033[0;31m[ERROR]\033[0m"; \
		TIME="\033[0;36m[TIME]\033[0m"; \
		cleanup() { \
			echo -e "$$INFO Stopping test environment..."; \
			docker ps -q --filter label=maze-host | xargs -r docker stop 2>/dev/null || true; \
			docker ps -aq --filter label=maze-host | xargs -r docker rm 2>/dev/null || true; \
			docker compose -f $(COMPOSE_TEST_FILE) -p $(COMPOSE_PROJECT) down -v --remove-orphans >/dev/null 2>&1 || true; \
		}; \
		trap cleanup EXIT INT TERM; \
		START=$$(date +%s); \
		echo -e "$$INFO [1/4] Building Agent base image..."; \
		docker build -f $(AGENT_DOCKERFILE) -t $(AGENT_IMAGE) $(PROJECT_ROOT); \
		ELAPSED=$$(($$(date +%s) - $$START)); \
		echo -e "$$TIME Agent base image built in $${ELAPSED}s"; \
		echo -e "$$INFO [2/4] Starting test environment (docker-compose)..."; \
		docker images --filter label=maze.dockerfile-hash -q | xargs -r docker rmi -f 2>/dev/null; \
		mkdir -p $(HOST_DATA_DIR)/docker/agents; \
		START2=$$(date +%s); \
		docker compose -f $(COMPOSE_TEST_FILE) -p $(COMPOSE_PROJECT) up -d --build; \
		ELAPSED2=$$(($$(date +%s) - $$START2)); \
		echo -e "$$TIME Test environment started in $${ELAPSED2}s"; \
		echo -e "$$INFO [3/4] Waiting for Manager to be ready..."; \
		START3=$$(date +%s); \
		MANAGER_READY=0; \
		for i in $$(seq 1 60); do \
			if curl -sf http://localhost:$(PORT_MANAGER)/health > /dev/null 2>&1; then \
				MANAGER_READY=1; \
				break; \
			fi; \
			echo "  waiting... ($$i/60)"; sleep 2; \
		done; \
		if [ "$$MANAGER_READY" != "1" ]; then \
			echo -e "$$ERROR Manager did not become ready within 120s."; \
			docker compose -f $(COMPOSE_TEST_FILE) -p $(COMPOSE_PROJECT) ps; \
			docker compose -f $(COMPOSE_TEST_FILE) -p $(COMPOSE_PROJECT) logs --tail=200 agent-manager postgres; \
			exit 1; \
		fi; \
		ELAPSED3=$$(($$(date +%s) - $$START3)); \
		echo -e "$$TIME Manager ready in $${ELAPSED3}s"; \
		TOTAL_SETUP=$$(($$(date +%s) - $$START)); \
		echo -e "$$INFO [4/4] Running integration tests (env=docker)..."; \
		echo -e "$$TIME Total setup time: $${TOTAL_SETUP}s"; \
		TEST_HOST_POOL="$${MAZE_TEST_ENABLE_HOST_POOL:-$(if $(TEST_NAME),0,1)}"; \
		TEST_STREAM="$${MAZE_TEST_STREAM_EVENTS:-1}"; \
		TEST_POOL_CLAUDE_SIZE="$${MAZE_TEST_POOL_CLAUDE_SIZE:-4}"; \
		TEST_POOL_GO_SIZE="$${MAZE_TEST_POOL_GO_SIZE:-1}"; \
		TEST_TARGET="$${TEST_NAME:-<all>}"; \
		echo -e "$$INFO Test mode:"; \
		echo "  platform=$(PLATFORM)"; \
		echo "  test_name=$$TEST_TARGET"; \
		echo "  host_pool=$$TEST_HOST_POOL"; \
		echo "  stream_events=$$TEST_STREAM"; \
		echo "  pool_claude=$$TEST_POOL_CLAUDE_SIZE"; \
		echo "  pool_go=$$TEST_POOL_GO_SIZE"; \
		cd $(TEST_DIR); \
		if [ "$$TEST_STREAM" = "1" ]; then \
			MAZE_TEST_ENV=$(PLATFORM) \
			MAZE_TEST_DATA_DIR=$(HOST_DATA_DIR) \
			MAZE_TEST_AGENT_STORAGE_BACKEND=bind \
			MAZE_TEST_MANAGER_URL=http://localhost:$(PORT_MANAGER) \
			MAZE_TEST_AUTH_TOKEN=test-integration-token \
			MAZE_TEST_ENABLE_HOST_POOL="$$TEST_HOST_POOL" \
			MAZE_TEST_POOL_CLAUDE_SIZE="$$TEST_POOL_CLAUDE_SIZE" \
			MAZE_TEST_POOL_GO_SIZE="$$TEST_POOL_GO_SIZE" \
			MAZE_TEST_STREAM_EVENTS="$$TEST_STREAM" \
			go test -json -count=1 -tags=integration -timeout=10m $(if $(TEST_NAME),-run $(TEST_NAME),) . 2>&1 | \
				go run ./cmd/maze-integration-stream; \
		else \
			MAZE_TEST_ENV=$(PLATFORM) \
			MAZE_TEST_DATA_DIR=$(HOST_DATA_DIR) \
			MAZE_TEST_AGENT_STORAGE_BACKEND=bind \
			MAZE_TEST_MANAGER_URL=http://localhost:$(PORT_MANAGER) \
			MAZE_TEST_AUTH_TOKEN=test-integration-token \
			MAZE_TEST_ENABLE_HOST_POOL="$$TEST_HOST_POOL" \
			MAZE_TEST_POOL_CLAUDE_SIZE="$$TEST_POOL_CLAUDE_SIZE" \
			MAZE_TEST_POOL_GO_SIZE="$$TEST_POOL_GO_SIZE" \
			MAZE_TEST_STREAM_EVENTS="$$TEST_STREAM" \
			go test -v -count=1 -tags=integration -timeout=10m $(if $(TEST_NAME),-run $(TEST_NAME),) .; \
		fi \
	'
else ifeq ($(PLATFORM),kubernetes)
	# K8s 集成测试会自己准备一个一次性的 namespace。
	#
	# 整个流程分成 5 段：
	# 1. 本地构建测试镜像，避免依赖外部镜像仓库
	# 2. 重建隔离 namespace，确保每次测试从干净环境启动
	# 3. 直接渲染静态 test overlay
	# 4. 先等待 deployment rollout，再等待 Pod ready，最后再启动 port-forward
	# 5. 通过本地 manager 端口执行集成测试，失败后统一清理 namespace 和 port-forward 进程
	@bash -c '\
		set -o pipefail; \
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
		echo -e "$$INFO Applying rendered manifests..."; \
		mkdir -p /tmp/maze-test/kubernetes/agents; \
		if ! kubectl kustomize "$(K8S_OVERLAY_DIR)" | kubectl apply -f -; then \
			echo -e "$$ERROR Failed to deploy test environment."; \
			exit 1; \
		fi; \
		echo -e "$$INFO Waiting for deployments to roll out..."; \
		if kubectl get deployment/postgresql -n $$NS >/dev/null 2>&1; then \
			kubectl rollout status deployment/postgresql -n $$NS --timeout=180s || \
				(echo -e "$$WARN PostgreSQL rollout timed out." && kubectl get pods -n $$NS && exit 1); \
		else \
			echo -e "$$INFO Skip PostgreSQL rollout wait: deployment/postgresql not managed by overlay."; \
		fi; \
		kubectl rollout status deployment/agent-manager -n $$NS --timeout=180s || \
			(echo -e "$$WARN agent-manager rollout timed out." && kubectl get pods -n $$NS && exit 1); \
		echo -e "$$INFO Waiting for pods to be ready..."; \
		kubectl wait --for=condition=ready pods -l app --timeout=180s -n $$NS 2>/dev/null || \
			(echo -e "$$WARN Timeout or partial readiness." && kubectl get pods -n $$NS && exit 1); \
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
		TEST_HOST_POOL="$${MAZE_TEST_ENABLE_HOST_POOL:-$(if $(TEST_NAME),0,1)}"; \
		TEST_STREAM="$${MAZE_TEST_STREAM_EVENTS:-1}"; \
		TEST_POOL_CLAUDE_SIZE="$${MAZE_TEST_POOL_CLAUDE_SIZE:-4}"; \
		TEST_POOL_GO_SIZE="$${MAZE_TEST_POOL_GO_SIZE:-1}"; \
		TEST_TARGET="$${TEST_NAME:-<all>}"; \
		echo -e "$$INFO Test mode:"; \
		echo "  platform=$(PLATFORM)"; \
		echo "  test_name=$$TEST_TARGET"; \
		echo "  host_pool=$$TEST_HOST_POOL"; \
		echo "  stream_events=$$TEST_STREAM"; \
		echo "  pool_claude=$$TEST_POOL_CLAUDE_SIZE"; \
		echo "  pool_go=$$TEST_POOL_GO_SIZE"; \
		cd $(TEST_DIR) && \
		if [ "$$TEST_STREAM" = "1" ]; then \
			MAZE_TEST_ENV=$(PLATFORM) \
			MAZE_TEST_DATA_DIR=$(HOST_DATA_DIR) \
			MAZE_TEST_AGENT_STORAGE_BACKEND=hostpath \
			MAZE_TEST_MANAGER_URL=http://localhost:$(PORT_MANAGER) \
			MAZE_TEST_AUTH_TOKEN=test-integration-token \
			MAZE_TEST_NAMESPACE=$(K8S_NAMESPACE) \
			MAZE_TEST_ENABLE_HOST_POOL="$$TEST_HOST_POOL" \
			MAZE_TEST_POOL_CLAUDE_SIZE="$$TEST_POOL_CLAUDE_SIZE" \
			MAZE_TEST_POOL_GO_SIZE="$$TEST_POOL_GO_SIZE" \
			MAZE_TEST_STREAM_EVENTS="$$TEST_STREAM" \
			go test -json -count=1 -tags=integration -timeout=10m $(if $(TEST_NAME),-run $(TEST_NAME),) . 2>&1 | \
				go run ./cmd/maze-integration-stream; \
		else \
			MAZE_TEST_ENV=$(PLATFORM) \
			MAZE_TEST_DATA_DIR=$(HOST_DATA_DIR) \
			MAZE_TEST_AGENT_STORAGE_BACKEND=hostpath \
			MAZE_TEST_MANAGER_URL=http://localhost:$(PORT_MANAGER) \
			MAZE_TEST_AUTH_TOKEN=test-integration-token \
			MAZE_TEST_NAMESPACE=$(K8S_NAMESPACE) \
			MAZE_TEST_ENABLE_HOST_POOL="$$TEST_HOST_POOL" \
			MAZE_TEST_POOL_CLAUDE_SIZE="$$TEST_POOL_CLAUDE_SIZE" \
			MAZE_TEST_POOL_GO_SIZE="$$TEST_POOL_GO_SIZE" \
			MAZE_TEST_STREAM_EVENTS="$$TEST_STREAM" \
			go test -v -count=1 -tags=integration -timeout=10m $(if $(TEST_NAME),-run $(TEST_NAME),) .; \
		fi'
endif
