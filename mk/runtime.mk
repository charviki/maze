.PHONY: up deploy down destroy undeploy status proxy proxy-web proxy-director-core proxy-db

up: build deploy ## 一键部署：构建镜像 + 启动服务

deploy: ## 部署服务（根据 PLATFORM 自动选择 Docker Compose 或 K8s）
ifeq ($(PLATFORM),docker)
	@echo "\033[0;32m[INFO]\033[0m Starting with Docker Compose..."
	@mkdir -p $(HOST_DATA_DIR)/docker/agents
	@$(DOCKER_COMPOSE) up -d
	@echo ""
	@echo "\033[0;32m[INFO]\033[0m Maze is running on Docker Compose!"
	@echo "  Web:            http://localhost:$(PORT_WEB)"
	@echo "  Director Core:  http://localhost:$(PORT_DIRECTOR_CORE)/health"
	@echo "  Postgres:       postgresql://localhost:$(PORT_POSTGRES)"
	@echo "  Data policy:    PostgreSQL data is stored in Docker volume director-core-postgres-data"
else ifeq ($(PLATFORM),kubernetes)
	# K8s 部署必须基于当前 overlay 的真实资源集合等待 rollout；
	# 否则 test 环境缺少 web 时会被错误地当成失败。
	@echo "\033[0;32m[INFO]\033[0m Deploying to Kubernetes ($(ENV))..."
	@mkdir -p $(HOST_DATA_DIR)/docker/agents
	@if [ -n "$(AGENT_HOSTPATH_BASE)" ]; then mkdir -p $(AGENT_HOSTPATH_BASE); fi
	@kubectl create namespace $(K8S_NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	@kubectl kustomize "$(K8S_OVERLAY_DIR)" | kubectl apply -f -
	@echo "\033[0;32m[INFO]\033[0m Waiting for deployments to roll out..."
	@if kubectl get deployment/postgresql -n $(K8S_NAMESPACE) >/dev/null 2>&1; then \
		kubectl rollout status deployment/postgresql -n $(K8S_NAMESPACE) --timeout=180s || \
			(echo "\033[1;33m[WARN]\033[0m PostgreSQL rollout timed out." && kubectl get pods -n $(K8S_NAMESPACE) && exit 1); \
		kubectl wait --for=condition=ready pod -l app=postgresql --timeout=180s -n $(K8S_NAMESPACE) || \
			(echo "\033[1;33m[WARN]\033[0m PostgreSQL pod not ready." && kubectl get pods -n $(K8S_NAMESPACE) && exit 1); \
	else \
		echo "\033[0;32m[INFO]\033[0m Skip PostgreSQL rollout wait: deployment/postgresql not managed by overlay $(K8S_OVERLAY)."; \
	fi
	@if kubectl get deployment/director-core -n $(K8S_NAMESPACE) >/dev/null 2>&1; then \
		kubectl rollout status deployment/director-core -n $(K8S_NAMESPACE) --timeout=180s || \
			(echo "\033[1;33m[WARN]\033[0m director-core rollout timed out." && kubectl get pods -n $(K8S_NAMESPACE) && exit 1); \
		kubectl wait --for=condition=ready pod -l app=director-core --timeout=180s -n $(K8S_NAMESPACE) || \
			(echo "\033[1;33m[WARN]\033[0m director-core pod not ready." && kubectl get pods -n $(K8S_NAMESPACE) && exit 1); \
	fi
	@if kubectl get deployment/web -n $(K8S_NAMESPACE) >/dev/null 2>&1; then \
		kubectl rollout status deployment/web -n $(K8S_NAMESPACE) --timeout=180s || \
			(echo "\033[1;33m[WARN]\033[0m web rollout timed out." && kubectl get pods -n $(K8S_NAMESPACE) && exit 1); \
		kubectl wait --for=condition=ready pod -l app=web --timeout=180s -n $(K8S_NAMESPACE) || \
			(echo "\033[1;33m[WARN]\033[0m web pod not ready." && kubectl get pods -n $(K8S_NAMESPACE) && exit 1); \
	else \
		echo "\033[0;32m[INFO]\033[0m Skip web rollout wait: deployment/web not managed by overlay $(K8S_OVERLAY)."; \
	fi
	@echo ""
	@echo "\033[0;32m[INFO]\033[0m Maze is running on Kubernetes! ($(ENV))"
	@echo "  Next: make proxy PLATFORM=$(PLATFORM) ENV=$(ENV)"
	@echo "  Data policy: PostgreSQL deployment and data will be preserved by make down"
endif

down: ## 停止服务并默认保留 PostgreSQL 数据（K8s 不删除 namespace）
ifeq ($(PLATFORM),docker)
	@echo "\033[0;32m[INFO]\033[0m Stopping Docker Compose services..."
	@$(DOCKER_COMPOSE) down
	@echo "\033[0;32m[INFO]\033[0m PostgreSQL data preserved in Docker volume director-core-postgres-data."
else ifeq ($(PLATFORM),kubernetes)
	# 默认 down 只移除应用工作负载，并显式保留 PostgreSQL 资源；
	# 这样再次 make up 时可以直接复用数据库数据，不会把 destroy 语义藏在普通命令里。
	@echo "\033[0;32m[INFO]\033[0m Removing Kubernetes application resources while preserving PostgreSQL data..."
	@kubectl delete deployment,service -l app=maze-agent -n $(K8S_NAMESPACE) --ignore-not-found=true
	@kubectl delete deployment/director-core service/director-core configmap/director-core-config -n $(K8S_NAMESPACE) --ignore-not-found=true
	@kubectl delete deployment/web service/web configmap/web-nginx-config ingress/web -n $(K8S_NAMESPACE) --ignore-not-found=true
	@echo "\033[0;32m[INFO]\033[0m PostgreSQL deployment, service, config and data were preserved in namespace $(K8S_NAMESPACE)."
endif

destroy: ## 销毁服务并删除 PostgreSQL 数据（破坏性操作）
ifeq ($(PLATFORM),docker)
	@echo "\033[0;31m[WARN]\033[0m Destroying Docker environment and PostgreSQL data..."
	@$(DOCKER_COMPOSE) down -v --remove-orphans
	@echo "\033[0;31m[WARN]\033[0m PostgreSQL Docker volume director-core-postgres-data removed."
else ifeq ($(PLATFORM),kubernetes)
	# dev/test 的 PostgreSQL 使用 hostPath；仅删除 namespace 不会清理节点上的数据目录。
	# 因此 destroy 需要先停掉 PostgreSQL，再用一次性清理 Pod 擦除 hostPath，最后删除 namespace。
	# 这里不用 heredoc，是为了避免 Make 对 recipe 续行和 shell here-doc 的组合解析不稳定。
	@echo "\033[0;31m[WARN]\033[0m Destroying Kubernetes environment and PostgreSQL data in namespace $(K8S_NAMESPACE)..."
	@kubectl create namespace $(K8S_NAMESPACE) --dry-run=client -o yaml | kubectl apply -f - >/dev/null
	@kubectl delete deployment/postgresql -n $(K8S_NAMESPACE) --ignore-not-found=true
	@kubectl wait --for=delete pod -l app=postgresql --timeout=120s -n $(K8S_NAMESPACE) >/dev/null 2>&1 || true
	@if [ -n "$(POSTGRES_HOSTPATH)" ]; then \
		kubectl delete pod/postgres-hostpath-cleaner -n $(K8S_NAMESPACE) --ignore-not-found=true >/dev/null 2>&1 || true; \
		printf '%s\n' \
			'apiVersion: v1' \
			'kind: Pod' \
			'metadata:' \
			'  name: postgres-hostpath-cleaner' \
			'  namespace: $(K8S_NAMESPACE)' \
			'spec:' \
			'  restartPolicy: Never' \
			'  containers:' \
			'    - name: cleaner' \
			'      image: alpine:3.20' \
			'      command:' \
			'        - /bin/sh' \
			'        - -c' \
			'        - |' \
			'          mkdir -p /cleanup' \
			'          rm -rf /cleanup/* /cleanup/.[!.]* /cleanup/..?* 2>/dev/null || true' \
			'      volumeMounts:' \
			'        - name: postgres-data' \
			'          mountPath: /cleanup' \
			'  volumes:' \
			'    - name: postgres-data' \
			'      hostPath:' \
			'        path: $(POSTGRES_HOSTPATH)' \
			'        type: DirectoryOrCreate' | kubectl apply -f - >/dev/null; \
		kubectl wait --for=condition=Ready pod/postgres-hostpath-cleaner --timeout=120s -n $(K8S_NAMESPACE) >/dev/null; \
		kubectl wait --for=jsonpath='{.status.phase}'=Succeeded pod/postgres-hostpath-cleaner --timeout=120s -n $(K8S_NAMESPACE) >/dev/null; \
		kubectl delete pod/postgres-hostpath-cleaner -n $(K8S_NAMESPACE) --ignore-not-found=true >/dev/null; \
		echo "\033[0;31m[WARN]\033[0m Cleared PostgreSQL hostPath data at $(POSTGRES_HOSTPATH)."; \
	fi
	@kubectl delete namespace $(K8S_NAMESPACE) --ignore-not-found=true
	@echo "\033[0;31m[WARN]\033[0m Namespace $(K8S_NAMESPACE) deleted; PostgreSQL data destroyed."
endif

undeploy: down ## down 的别名（默认保留 PostgreSQL 数据）

status: ## 查看运行状态
ifeq ($(PLATFORM),docker)
	@$(DOCKER_COMPOSE) ps
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

proxy: ## 启动访问代理（K8s 只代理当前环境真实存在的服务）
ifeq ($(PLATFORM),docker)
	@echo "\033[0;32m[INFO]\033[0m Docker mode: ports already exposed."
	@echo "  Web:            http://localhost:$(PORT_WEB)"
	@echo "  Director Core:  http://localhost:$(PORT_DIRECTOR_CORE)/health"
	@echo "  Postgres:       postgresql://localhost:$(PORT_POSTGRES)"
else ifeq ($(PLATFORM),kubernetes)
	@echo "\033[0;32m[INFO]\033[0m Starting port-forward for services managed by overlay $(K8S_OVERLAY)..."
	@bash -c '\
		set -e; \
		PF_STARTED=0; \
		trap "kill 0" SIGINT SIGTERM; \
		start_pf() { \
			name="$$1"; ports="$$2"; desc="$$3"; \
			if kubectl get svc/$$name -n $(K8S_NAMESPACE) >/dev/null 2>&1; then \
				echo "  $$desc"; \
				kubectl port-forward svc/$$name $$ports -n $(K8S_NAMESPACE) & \
				PF_STARTED=1; \
			else \
				echo "  skip $$name: service not managed by overlay $(K8S_OVERLAY)"; \
			fi; \
		}; \
		start_pf web $(PORT_WEB):80 "Web:            http://localhost:$(PORT_WEB)"; \
		start_pf director-core $(PORT_DIRECTOR_CORE):8080 "Director Core:  http://localhost:$(PORT_DIRECTOR_CORE)/health"; \
		start_pf postgresql $(PORT_POSTGRES):5432 "Postgres:       postgresql://localhost:$(PORT_POSTGRES)"; \
		if [ "$$PF_STARTED" -ne 1 ]; then \
			echo "No services available for proxy in namespace $(K8S_NAMESPACE)."; \
			exit 1; \
		fi; \
		wait'
endif

proxy-web: ## 只代理 Web 前端（若当前环境无 web，则提示并退出）
ifeq ($(PLATFORM),kubernetes)
	@if kubectl get svc/web -n $(K8S_NAMESPACE) >/dev/null 2>&1; then \
		kubectl port-forward svc/web $(PORT_WEB):80 -n $(K8S_NAMESPACE); \
	else \
		echo "\033[1;33m[WARN]\033[0m Service web not managed by overlay $(K8S_OVERLAY)."; \
	fi
else
	@echo "Docker mode: http://localhost:$(PORT_WEB)"
endif

proxy-director-core: ## 只代理 Director Core API
ifeq ($(PLATFORM),kubernetes)
	@kubectl port-forward svc/director-core $(PORT_DIRECTOR_CORE):8080 -n $(K8S_NAMESPACE)
else
	@echo "Docker mode: http://localhost:$(PORT_DIRECTOR_CORE)/health"
endif

proxy-db: ## 只代理 PostgreSQL
ifeq ($(PLATFORM),kubernetes)
	@kubectl port-forward svc/postgresql $(PORT_POSTGRES):5432 -n $(K8S_NAMESPACE)
else
	@echo "Docker mode: postgresql://localhost:$(PORT_POSTGRES)"
endif
