include ../../../mk/env.mk

.PHONY: help test-integration

help: ## 显示集成测试帮助
	@echo "  make test-integration PLATFORM=docker"
	@echo "  make test-integration PLATFORM=kubernetes"
	@echo "  make test-integration PLATFORM=docker TEST_NAME=TestHostCreateOnline"

# 集成测试固定在 test 环境运行，避免测试命令与开发环境共享状态。
test-integration: override ENV := test
test-integration: override K8S_NAMESPACE := maze-test
test-integration: override K8S_OVERLAY := overlays/test
test-integration: override K8S_OVERLAY_DIR := $(PROJECT_ROOT)/fabrication/kubernetes/overlays/test
test-integration: override COMPOSE_PROJECT := maze-test
test-integration: override HOST_DATA_DIR := $(HOME)/.maze-test
test-integration: override PORT_DIRECTOR_CORE := 9090
test-integration: override PORT_WEB := 9080
test-integration: override PORT_POSTGRES := 5433
test-integration: override POSTGRES_HOSTPATH := /tmp/maze-test/postgresql/data
test-integration: override AGENT_HOSTPATH_BASE := /tmp/maze-test/kubernetes/agents

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
			$(DOCKER_TEST_COMPOSE) down -v --remove-orphans >/dev/null 2>&1 || true; \
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
		$(DOCKER_TEST_COMPOSE) up -d --build; \
		ELAPSED2=$$(($$(date +%s) - $$START2)); \
		echo -e "$$TIME Test environment started in $${ELAPSED2}s"; \
		echo -e "$$INFO [3/4] Waiting for Director Core to be ready..."; \
		START3=$$(date +%s); \
		DIRECTOR_CORE_READY=0; \
		for i in $$(seq 1 60); do \
			if curl -sf http://localhost:$(PORT_DIRECTOR_CORE)/health > /dev/null 2>&1; then \
				DIRECTOR_CORE_READY=1; \
				break; \
			fi; \
			echo "  waiting... ($$i/60)"; sleep 2; \
		done; \
		if [ "$$DIRECTOR_CORE_READY" != "1" ]; then \
			echo -e "$$ERROR Director Core did not become ready within 120s."; \
			$(DOCKER_TEST_COMPOSE) ps; \
			$(DOCKER_TEST_COMPOSE) logs --tail=200 director-core postgres; \
			exit 1; \
		fi; \
		ELAPSED3=$$(($$(date +%s) - $$START3)); \
		echo -e "$$TIME Director Core ready in $${ELAPSED3}s"; \
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
			MAZE_TEST_DIRECTOR_CORE_URL=http://localhost:$(PORT_DIRECTOR_CORE) \
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
			MAZE_TEST_DIRECTOR_CORE_URL=http://localhost:$(PORT_DIRECTOR_CORE) \
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
	# 5. 通过本地 Director Core 端口执行集成测试，失败后统一清理 namespace 和 port-forward 进程
	@bash -c '\
		set -o pipefail; \
		INFO="\033[0;32m[INFO]\033[0m"; \
		ERROR="\033[0;31m[ERROR]\033[0m"; \
		WARN="\033[1;33m[WARN]\033[0m"; \
		NS="$(K8S_NAMESPACE)"; \
		PF_PIDS=""; \
		DEPLOYED=false; \
		lsof -ti:$(PORT_DIRECTOR_CORE) | xargs kill -9 2>/dev/null; \
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
		docker build -f $(DIRECTOR_CORE_DOCKERFILE) -t $(DIRECTOR_CORE_IMAGE) $(PROJECT_ROOT); \
		echo -e "$$INFO [2/5] Recreating test namespace $$NS..."; \
		kubectl delete namespace $$NS --ignore-not-found=true --wait=true 2>/dev/null || true; \
		kubectl create namespace $$NS --dry-run=client -o yaml | kubectl apply -f -; \
		DEPLOYED=true; \
		echo -e "$$INFO [3/5] Deploying test environment to namespace $$NS..."; \
		echo -e "$$INFO Applying rendered manifests..."; \
		mkdir -p $(AGENT_HOSTPATH_BASE); \
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
		kubectl rollout status deployment/director-core -n $$NS --timeout=180s || \
			(echo -e "$$WARN director-core rollout timed out." && kubectl get pods -n $$NS && exit 1); \
		echo -e "$$INFO Waiting for pods to be ready..."; \
		kubectl wait --for=condition=ready pod -l app=postgresql --timeout=180s -n $$NS 2>/dev/null || \
			(echo -e "$$WARN postgresql not ready." && kubectl get pods -n $$NS && exit 1); \
		kubectl wait --for=condition=ready pod -l app=director-core --timeout=180s -n $$NS 2>/dev/null || \
			(echo -e "$$WARN director-core not ready." && kubectl get pods -n $$NS && exit 1); \
		echo -e "$$INFO [4/5] Starting port-forward..."; \
		kubectl port-forward svc/director-core $(PORT_DIRECTOR_CORE):8080 -n $$NS 2>&1 | grep -v "^Handling" & \
		PF_PID1=$$!; \
		PF_PIDS="$$PF_PID1"; \
		for i in $$(seq 1 30); do \
			if curl -sf http://localhost:$(PORT_DIRECTOR_CORE)/health > /dev/null 2>&1; then \
				break; \
			fi; \
			if ! kill -0 $$PF_PID1 2>/dev/null; then \
				echo -e "$$ERROR port-forward process died. Check kubectl and cluster status."; \
				exit 1; \
			fi; \
			sleep 2; \
		done; \
		if ! curl -sf http://localhost:$(PORT_DIRECTOR_CORE)/health > /dev/null 2>&1; then \
			echo -e "$$ERROR port-forward not ready after 60s."; \
			exit 1; \
		fi; \
		echo -e "$$INFO Port-forward active: director-core=$(PORT_DIRECTOR_CORE)"; \
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
			MAZE_TEST_DIRECTOR_CORE_URL=http://localhost:$(PORT_DIRECTOR_CORE) \
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
			MAZE_TEST_DIRECTOR_CORE_URL=http://localhost:$(PORT_DIRECTOR_CORE) \
			MAZE_TEST_AUTH_TOKEN=test-integration-token \
			MAZE_TEST_NAMESPACE=$(K8S_NAMESPACE) \
			MAZE_TEST_ENABLE_HOST_POOL="$$TEST_HOST_POOL" \
			MAZE_TEST_POOL_CLAUDE_SIZE="$$TEST_POOL_CLAUDE_SIZE" \
			MAZE_TEST_POOL_GO_SIZE="$$TEST_POOL_GO_SIZE" \
			MAZE_TEST_STREAM_EVENTS="$$TEST_STREAM" \
			go test -v -count=1 -tags=integration -timeout=10m $(if $(TEST_NAME),-run $(TEST_NAME),) .; \
		fi'
endif
