.PHONY: check-frontend format-js format-js-check

check-frontend: ## 前端三道检查：tsc → eslint → vitest（每个模块按序执行，任何一步失败即中止）
	@cd $(PROJECT_ROOT)/fabrication/skin && pnpm exec tsc --noEmit || exit 1 && pnpm exec eslint . || exit 1 && pnpm exec vitest run || exit 1
	@cd $(PROJECT_ROOT)/the-mesa/arrival-gate && pnpm exec tsc -b --noEmit || exit 1 && pnpm exec eslint . || exit 1 && pnpm exec vitest run || exit 1
	@cd $(PROJECT_ROOT)/the-mesa/director-console && pnpm exec tsc -b --noEmit || exit 1 && pnpm exec eslint . || exit 1 && pnpm exec vitest run || exit 1
	@cd $(PROJECT_ROOT)/sweetwater/black-ridge/web && pnpm exec tsc -b --noEmit || exit 1 && pnpm exec eslint . || exit 1 && pnpm exec vitest run || exit 1
	@echo "\033[0;32m[check-frontend]\033[0m All frontend checks passed!"

format-js: ## 格式化所有 TS/TSX/JSON/MD 文件
	@echo "\033[0;32m[format-js]\033[0m Formatting with Prettier..."
	@cd $(PROJECT_ROOT) && pnpm run format

format-js-check: ## 检查 TS 格式（CI 用）
	@echo "\033[0;32m[format-js-check]\033[0m Checking formatting..."
	@cd $(PROJECT_ROOT) && pnpm run format:check
