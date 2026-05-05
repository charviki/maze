.PHONY: build-go vet lint lint-fix vulncheck test coverage check

build-go: ## 编译所有 Go 模块
	@for m in $(MODULES); do \
		echo "\033[0;32m[build]\033[0m $$m"; \
		cd $(PROJECT_ROOT)/$$m && go build ./... || exit 1; \
	done

vet: ## Go 静态检查（快速模式，仅 go vet）
	@for m in $(MODULES); do \
		echo "\033[0;32m[vet]\033[0m $$m"; \
		cd $(PROJECT_ROOT)/$$m && go vet ./... || exit 1; \
	done

lint: ## Go 全量代码检查（golangci-lint v2，含 gosec + staticcheck + 30+ linter）
	@echo "\033[0;32m[lint]\033[0m Running golangci-lint on all Go modules..."
	@for m in $(MODULES); do \
		echo "  --> Linting $$m"; \
		cd $(PROJECT_ROOT)/$$m && golangci-lint run --timeout=5m ./... || exit 1; \
	done
	@echo "  --> Linting integration tests"
	@cd $(TEST_DIR) && golangci-lint run --timeout=5m ./... || exit 1

lint-fix: ## Go 自动修复可修 lint 问题
	@echo "\033[0;32m[lint-fix]\033[0m Auto-fixing lint issues..."
	@for m in $(MODULES); do \
		cd $(PROJECT_ROOT)/$$m && golangci-lint run --fix --timeout=5m ./... || exit 1; \
	done

vulncheck: ## Go 漏洞扫描（govulncheck）
	@echo "\033[0;32m[vulncheck]\033[0m Running govulncheck..."
	@for m in $(MODULES); do \
		echo "  --> Scanning $$m"; \
		cd $(PROJECT_ROOT)/$$m && govulncheck ./... || exit 1; \
	done

test: ## 运行所有 Go 单元测试
	@for m in $(MODULES); do \
		echo "\033[0;32m[test]\033[0m $$m"; \
		cd $(PROJECT_ROOT)/$$m && go test ./... -count=1 || exit 1; \
	done

coverage: ## 生成 Go 覆盖率报告（对每个模块运行 go test -coverprofile）
	@echo "\033[0;32m[coverage]\033[0m Generating coverage reports..."
	@for m in $(COVERAGE_MODULES); do \
		echo "\033[0;32m[coverage]\033[0m $$m"; \
		cd $(PROJECT_ROOT)/$$m && go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | tail -1 || exit 1; \
	done
	@echo "\033[0;32m[coverage]\033[0m Coverage reports generated."

check: build-go lint test ## 编译 + golangci-lint + 单元测试（Go 交付铁律）
	@echo "\033[0;32m[check]\033[0m All Go checks passed!"
