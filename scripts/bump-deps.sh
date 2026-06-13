#!/usr/bin/env bash
# 查询所有 deps 最新稳定版本，并更新锁定版本：
#   - fabrication/deps/go.txt, js.txt
#   - fabrication/molds/Dockerfile.claude, Dockerfile.codex
# 用法：make deps-bump  （之后运行 make build-deps 重建受影响的 deps 镜像）
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# 沿用项目其它 Dockerfile 的国内代理，保证 go list 能拉到版本索引
export GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"

# 取 go module 最新稳定版（过滤掉 rc/beta/alpha/pre-release，这些不适合锁进生产 deps）
latest_go() {
	go list -m -versions "$1" 2>/dev/null \
		| tr ' ' '\n' \
		| grep -E '^v[0-9]' \
		| grep -vE 'rc|beta|alpha|pre' \
		| tail -1
}

# 取 npm 包最新版本
latest_npm() {
	npm view "$1" version 2>/dev/null
}

# 校验查询结果非空：避免网络抖动时把空版本写进文件（会破坏 docker build）
require() {
	if [ -z "${2:-}" ]; then
		echo "[ERROR] 未能获取 $1 的版本，已中止（未修改任何文件）" >&2
		exit 1
	fi
}

echo "==> 查询 go 工具链最新稳定版本..."
GOPLS=$(latest_go golang.org/x/tools/gopls);                 require gopls "$GOPLS"
GOLANGCI=$(latest_go github.com/golangci/golangci-lint/v2);  require golangci-lint "$GOLANGCI"
DLV=$(latest_go github.com/go-delve/delve);                  require dlv "$DLV"
GOIMPORTS=$(latest_go golang.org/x/tools);                   require goimports "$GOIMPORTS"

echo "==> 查询 js 依赖最新版本..."
TS=$(latest_npm typescript);     require typescript "$TS"
TSNODE=$(latest_npm ts-node);    require ts-node "$TSNODE"
PRETTIER=$(latest_npm prettier); require prettier "$PRETTIER"
ESLINT=$(latest_npm eslint);     require eslint "$ESLINT"

echo "==> 查询 CLI 最新版本..."
CLAUDE=$(latest_npm @anthropic-ai/claude-code); require claude-code "$CLAUDE"
CODEX=$(latest_npm @openai/codex);              require codex "$CODEX"

# go.txt / js.txt 格式简单（pkg@version 每行一条），直接重写整个文件，
# 避免跨平台 sed（macOS 需 -i ''，GNU 需 -i）的坑
cat > fabrication/deps/go.txt <<EOF
# 锁定到具体版本，避免 @latest 每次拉到不同版本导致 deps 层漂移
# （层 ID 变化会连带 agent 镜像 COPY --from=maze-deps-go 失效、整条链重建）。
# 升级版本：make deps-bump
golang.org/x/tools/gopls@${GOPLS}
github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI}
github.com/go-delve/delve/cmd/dlv@${DLV}
golang.org/x/tools/cmd/goimports@${GOIMPORTS}
EOF

cat > fabrication/deps/js.txt <<EOF
# 锁定到具体版本，避免 @latest 每次拉到不同版本导致 deps 层漂移。
# 升级版本：make deps-bump
typescript@${TS}
ts-node@${TSNODE}
prettier@${PRETTIER}
eslint@${ESLINT}
EOF

# Dockerfile 里的版本用 perl 原地替换（perl -i 行为跨平台一致，优于 sed）；
# 版本号通过环境变量 $ENV{} 传入，避免 bash/perl 之间的引号转义地狱
CLAUDE_VERSION="$CLAUDE" perl -i -pe 's{(\@anthropic-ai/claude-code\@)[0-9.]+}{$1$ENV{CLAUDE_VERSION}}' fabrication/molds/Dockerfile.claude
CODEX_VERSION="$CODEX" perl -i -pe 's{(\@openai/codex\@)[0-9.]+}{$1$ENV{CODEX_VERSION}}' fabrication/molds/Dockerfile.codex

echo ""
echo "==> 版本已更新："
echo "    gopls=${GOPLS}  golangci-lint=${GOLANGCI}  dlv=${DLV}  goimports=${GOIMPORTS}"
echo "    typescript=${TS}  ts-node=${TSNODE}  prettier=${PRETTIER}  eslint=${ESLINT}"
echo "    claude-code=${CLAUDE}  codex=${CODEX}"
echo ""
echo "请检查 git diff，确认后运行：make build-deps"
