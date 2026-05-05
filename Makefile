PROJECT_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

include $(PROJECT_ROOT)/mk/env.mk
include $(PROJECT_ROOT)/mk/go.mk
include $(PROJECT_ROOT)/mk/frontend.mk
include $(PROJECT_ROOT)/mk/images.mk
include $(PROJECT_ROOT)/mk/runtime.mk

.DEFAULT_GOAL := help

.PHONY: help gen gen-proto gen-client gen-sdk test-integration

help: ## 显示帮助信息
	@grep -h -E '^[a-zA-Z0-9_.-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "  PLATFORM=docker|kubernetes (default: $(PLATFORM))"
	@echo "  ENV=dev|test|prod (default: $(ENV))"
	@echo "  TEST_NAME=<regex> (integration only)"
	@echo ""
	@echo "  示例:"
	@echo "    make up"
	@echo "    make up PLATFORM=docker"
	@echo "    make proxy PLATFORM=kubernetes ENV=dev"
	@echo "    make destroy PLATFORM=kubernetes ENV=test"
	@echo "    make test-integration PLATFORM=docker"
	@echo "    make gen"

# 根入口只暴露稳定命令名，真实实现下沉到各自目录，
# 这样可以避免根和子目录重复维护同一套生成与集成测试逻辑。
gen-proto: ## buf generate 生成 Go 类型 + gRPC stub + grpc-gateway + OpenAPI spec
	@$(MAKE) -C $(CRADLE_API_DIR) gen-proto PLATFORM=$(PLATFORM) ENV=$(ENV)

gen-client: ## 重新生成 OpenAPI Go HTTP client（需要 openapi-generator + Java）
	@$(MAKE) -C $(CRADLE_API_DIR) gen-client PLATFORM=$(PLATFORM) ENV=$(ENV)

gen-sdk: ## 生成 TypeScript SDK from OpenAPI spec (fabrication/skin)
	@$(MAKE) -C $(SKIN_DIR) gen-sdk PLATFORM=$(PLATFORM) ENV=$(ENV)

gen: gen-client gen-sdk ## 一键生成 proto + HTTP client + TypeScript SDK

test-integration: ## 运行集成测试（PLATFORM=docker 或 PLATFORM=kubernetes，TEST_NAME=TestX 运行单个测试）
	@$(MAKE) -C $(TEST_DIR) test-integration PLATFORM=$(PLATFORM) ENV=test TEST_NAME="$(TEST_NAME)"
