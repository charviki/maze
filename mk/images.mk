.PHONY: build build-director-core build-the-forge build-web build-agent build-deps build-deps-claude build-deps-codex build-deps-go build-deps-python build-deps-node update-director-core update-the-forge update-web update-agent update-all deps-bump
.PHONY: docker-clean docker-clean-containers docker-clean-images docker-clean-cache docker-clean-volumes

build: build-deps build-director-core build-the-forge build-web build-agent ## 构建全部 Docker 镜像

build-director-core: ## 构建 Director Core 镜像
	@echo "\033[0;32m[INFO]\033[0m Building Director Core image..."
	@docker build -f $(DIRECTOR_CORE_DOCKERFILE) -t $(DIRECTOR_CORE_IMAGE) $(PROJECT_ROOT)

build-the-forge: ## 构建 The Forge 镜像
	@echo "\033[0;32m[INFO]\033[0m Building The Forge image..."
	@docker build -f $(THE_FORGE_DOCKERFILE) -t $(THE_FORGE_IMAGE) $(PROJECT_ROOT)

build-web: ## 构建 Web Nginx 镜像
	@echo "\033[0;32m[INFO]\033[0m Building Web Nginx image..."
	@docker build -f $(WEB_DOCKERFILE) -t $(WEB_IMAGE) $(PROJECT_ROOT)

build-agent: ## 构建 Agent 基础镜像
	@echo "\033[0;32m[INFO]\033[0m Building Agent base image..."
	@docker build -f $(AGENT_DOCKERFILE) -t $(AGENT_IMAGE) $(PROJECT_ROOT)

# 拆成独立 target 以便 make -j 并行（5 个 deps 互相独立、无共享写入）。
build-deps-claude:
	@echo "\033[0;32m[INFO]\033[0m Building maze-deps-claude:latest..."
	@docker build -f $(MOLDS_DIR)/Dockerfile.claude -t maze-deps-claude:latest $(PROJECT_ROOT)
build-deps-codex:
	@echo "\033[0;32m[INFO]\033[0m Building maze-deps-codex:latest..."
	@docker build -f $(MOLDS_DIR)/Dockerfile.codex -t maze-deps-codex:latest $(PROJECT_ROOT)
build-deps-go:
	@echo "\033[0;32m[INFO]\033[0m Building maze-deps-go:latest..."
	@docker build -f $(MOLDS_DIR)/Dockerfile.go -t maze-deps-go:latest $(PROJECT_ROOT)
build-deps-python:
	@echo "\033[0;32m[INFO]\033[0m Building maze-deps-python:latest..."
	@docker build -f $(MOLDS_DIR)/Dockerfile.python -t maze-deps-python:latest $(PROJECT_ROOT)
build-deps-node:
	@echo "\033[0;32m[INFO]\033[0m Building maze-deps-node:latest..."
	@docker build -f $(MOLDS_DIR)/Dockerfile.node -t maze-deps-node:latest $(PROJECT_ROOT)

build-deps: ## 构建所有供应商镜像（claude/codex/go/python/node），并行加速
ifeq ($(SKIP_DEPS),1)
	@echo "\033[0;33m[INFO]\033[0m SKIP_DEPS=1，跳过 deps 构建，复用现有 maze-deps-*:latest"
else
	@$(MAKE) -j5 build-deps-claude build-deps-codex build-deps-go build-deps-python build-deps-node
endif

update-director-core: build-director-core ## 重建 Director Core 镜像 + 重启
ifeq ($(PLATFORM),docker)
	@$(DOCKER_COMPOSE) up -d director-core
else ifeq ($(PLATFORM),kubernetes)
	@kubectl rollout restart deployment/director-core -n $(K8S_NAMESPACE)
	@kubectl rollout status deployment/director-core -n $(K8S_NAMESPACE) --timeout=120s
endif

update-the-forge: build-the-forge ## 重建 The Forge 镜像 + 重启
ifeq ($(PLATFORM),docker)
	@$(DOCKER_COMPOSE) up -d the-forge
else ifeq ($(PLATFORM),kubernetes)
	@if kubectl get deployment/the-forge -n $(K8S_NAMESPACE) >/dev/null 2>&1; then \
		kubectl rollout restart deployment/the-forge -n $(K8S_NAMESPACE); \
		kubectl rollout status deployment/the-forge -n $(K8S_NAMESPACE) --timeout=120s; \
	else \
		echo "\033[0;32m[INFO]\033[0m Skip the-forge restart: deployment/the-forge not managed by overlay $(K8S_OVERLAY)."; \
	fi
endif

update-web: build-web ## 重建 Web 镜像 + 重启（若当前环境不存在 web，则自动跳过）
ifeq ($(PLATFORM),docker)
	@$(DOCKER_COMPOSE) up -d web
else ifeq ($(PLATFORM),kubernetes)
	@if kubectl get deployment/web -n $(K8S_NAMESPACE) >/dev/null 2>&1; then \
		kubectl rollout restart deployment/web -n $(K8S_NAMESPACE); \
		kubectl rollout status deployment/web -n $(K8S_NAMESPACE) --timeout=120s; \
	else \
		echo "\033[0;32m[INFO]\033[0m Skip web restart: deployment/web not managed by overlay $(K8S_OVERLAY)."; \
	fi
endif

update-agent: build-agent ## 重建 Agent 基础镜像
	@echo "\033[0;32m[INFO]\033[0m Agent base image updated. New Hosts will use the updated image."

update-all: build-director-core build-the-forge build-web build-agent ## 全部更新：重建所有镜像 + 重启
ifeq ($(PLATFORM),docker)
	@$(DOCKER_COMPOSE) up -d
else ifeq ($(PLATFORM),kubernetes)
	@kubectl rollout restart deployment/director-core -n $(K8S_NAMESPACE)
	@kubectl rollout status deployment/director-core -n $(K8S_NAMESPACE) --timeout=120s
	@if kubectl get deployment/the-forge -n $(K8S_NAMESPACE) >/dev/null 2>&1; then \
		kubectl rollout restart deployment/the-forge -n $(K8S_NAMESPACE); \
		kubectl rollout status deployment/the-forge -n $(K8S_NAMESPACE) --timeout=120s; \
	fi
	@if kubectl get deployment/web -n $(K8S_NAMESPACE) >/dev/null 2>&1; then \
		kubectl rollout restart deployment/web -n $(K8S_NAMESPACE); \
		kubectl rollout status deployment/web -n $(K8S_NAMESPACE) --timeout=120s; \
	else \
		echo "\033[0;32m[INFO]\033[0m Skip web restart: deployment/web not managed by overlay $(K8S_OVERLAY)."; \
	fi
endif
	@echo "\033[0;32m[INFO]\033[0m All services updated."

docker-clean-containers: ## 清理已退出容器（解锁被其引用的悬空镜像/层）
	@docker container prune -f

# 去 until=168h 过滤：本地高频迭代产生的失效层几乎都在数小时内，7 天过滤等于不清。
# 依赖前置 container-prune：否则被退出容器引用的镜像永远删不掉（悬空镜像被锁住的根因）。
docker-clean-images: docker-clean-containers ## 清理未被运行容器引用的镜像（含悬空）
	@docker image prune -a -f

# 默认只删 Private 失效层（~5GB），保留 Shared 活跃层，不牺牲缓存命中率。
# 要连活跃层一起清用 docker buildx prune --all -f（代价：下次构建稍慢）。
docker-clean-cache: ## 回收 BuildKit 失效缓存（保留活跃层）
	@docker builder prune -f

docker-clean-volumes: ## 清理无容器挂载的匿名卷
	@docker volume prune -f

# 顺序关键：容器→镜像→缓存→卷（docker-clean-images 已含 container-prune 依赖）。
docker-clean: docker-clean-images docker-clean-cache docker-clean-volumes ## 一键清理所有可回收 Docker 资源（容器→镜像→缓存→卷）

deps-bump: ## 查询 deps 最新版本并更新 deps/*.txt 与 Dockerfile.claude/codex，之后需 make build-deps
	@bash $(PROJECT_ROOT)/scripts/bump-deps.sh
