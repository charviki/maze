include ../../../mk/env.mk

.PHONY: help gen gen-proto gen-client

help: ## 显示代码生成帮助
	@echo "  make gen-proto"
	@echo "  make gen-client"
	@echo "  make gen"

gen-proto: ## buf generate 生成 Go 类型 + gRPC stub + grpc-gateway + OpenAPI spec
	@buf generate

gen-client: gen-proto ## 重新生成 OpenAPI Go HTTP client（需要 openapi-generator + Java）
	@if [ -z "$$JAVA_HOME" ]; then \
		echo "Error: JAVA_HOME not set. Install Java and set JAVA_HOME."; \
		exit 1; \
	fi
	@export PATH="$$JAVA_HOME/bin:$$PATH" && \
		openapi-generator generate \
			-i gen/openapiv2/maze.swagger.json \
			-g go \
			-o gen/http \
			--package-name client \
			--additional-properties=isGoSubmodule=true,withGoMod=false,enumClassPrefix=true

gen: gen-client ## 一键生成 proto + HTTP client
