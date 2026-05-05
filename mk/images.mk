.PHONY: build build-director-core build-web build-agent build-deps update-director-core update-web update-agent update-all

build: build-deps build-director-core build-web build-agent ## 构建全部 Docker 镜像

build-director-core: ## 构建 Director Core 镜像
	@echo "\033[0;32m[INFO]\033[0m Building Director Core image..."
	@docker build -f $(DIRECTOR_CORE_DOCKERFILE) -t $(DIRECTOR_CORE_IMAGE) $(PROJECT_ROOT)

build-web: ## 构建 Web Nginx 镜像
	@echo "\033[0;32m[INFO]\033[0m Building Web Nginx image..."
	@docker build -f $(WEB_DOCKERFILE) -t $(WEB_IMAGE) $(PROJECT_ROOT)

build-agent: ## 构建 Agent 基础镜像
	@echo "\033[0;32m[INFO]\033[0m Building Agent base image..."
	@docker build -f $(AGENT_DOCKERFILE) -t $(AGENT_IMAGE) $(PROJECT_ROOT)

build-deps: ## 构建所有供应商镜像（claude/codex/go/python/node）
	@echo "\033[0;32m[INFO]\033[0m Building supplier images..."
	@for dep in claude codex go python node; do \
		echo "\033[0;32m[INFO]\033[0m   Building maze-deps-$${dep}:latest..."; \
		docker build -f $(MOLDS_DIR)/Dockerfile.$${dep} -t maze-deps-$${dep}:latest $(PROJECT_ROOT); \
	done

update-director-core: build-director-core ## 重建 Director Core 镜像 + 重启
ifeq ($(PLATFORM),docker)
	@$(DOCKER_COMPOSE) up -d director-core
else ifeq ($(PLATFORM),kubernetes)
	@kubectl rollout restart deployment/director-core -n $(K8S_NAMESPACE)
	@kubectl rollout status deployment/director-core -n $(K8S_NAMESPACE) --timeout=120s
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

update-all: build-director-core build-web build-agent ## 全部更新：重建所有镜像 + 重启
ifeq ($(PLATFORM),docker)
	@$(DOCKER_COMPOSE) up -d
else ifeq ($(PLATFORM),kubernetes)
	@kubectl rollout restart deployment/director-core -n $(K8S_NAMESPACE)
	@kubectl rollout status deployment/director-core -n $(K8S_NAMESPACE) --timeout=120s
	@if kubectl get deployment/web -n $(K8S_NAMESPACE) >/dev/null 2>&1; then \
		kubectl rollout restart deployment/web -n $(K8S_NAMESPACE); \
		kubectl rollout status deployment/web -n $(K8S_NAMESPACE) --timeout=120s; \
	else \
		echo "\033[0;32m[INFO]\033[0m Skip web restart: deployment/web not managed by overlay $(K8S_OVERLAY)."; \
	fi
endif
	@echo "\033[0;32m[INFO]\033[0m All services updated."
