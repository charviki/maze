include ../../mk/env.mk

.PHONY: help gen-sdk

help: ## 显示 SDK 生成帮助
	@echo "  make gen-sdk"

gen-sdk: ## 生成 TypeScript SDK from OpenAPI spec (fabrication/skin)
	@echo "\033[0;32m[gen-sdk]\033[0m Generating TypeScript SDK..."
	@JAVA_HOME=$(JAVA_HOME) npx openapi-generator-cli generate \
		-i ../cradle/api/gen/http/api/openapi.yaml \
		-c openapi-generator-config.yaml \
		-o src/api/gen \
		-g typescript-fetch
	@# SDK 生成文件包含跨文件引用的内部 helper，在消费端 tsc -b 编译时会触发 noUnusedLocals。
	@# 仅对自动生成文件增加 @ts-nocheck，避免为生成器副作用污染业务代码告警。
	@find src/api/gen -name '*.ts' -exec python3 -c 'from pathlib import Path; import sys; \
path = Path(sys.argv[1]); prefix = "// @ts-nocheck\n"; content = path.read_text(); \
lines = [line for line in content.splitlines() if line != "// @ts-nocheck"]; \
normalized = "\n".join(lines); normalized += "\n" if content.endswith("\n") else ""; \
path.write_text(prefix + normalized.lstrip("\n"))' {} \;
	@echo "\033[0;32m[gen-sdk]\033[0m SDK generated at fabrication/skin/src/api/gen/"
