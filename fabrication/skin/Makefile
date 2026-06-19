include ../../mk/env.mk

.PHONY: help gen-sdk

help: ## 显示 SDK 生成帮助
	@echo "  make gen-sdk"

gen-sdk: ## 生成 TypeScript SDK from OpenAPI spec via hey-api（纯 Node，无 Java 依赖）
	@echo "\033[0;32m[gen-sdk]\033[0m Generating TypeScript SDK..."
	@npx @hey-api/openapi-ts
	@echo "\033[0;32m[gen-sdk]\033[0m SDK generated at fabrication/skin/src/api/gen/"
