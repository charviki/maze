#!/bin/bash
set -e

AGENT_HOME="${AGENT_HOME:-/home/agent}"

ensure_claude_json() {
    local json_file="$AGENT_HOME/.claude.json"
    if [ ! -f "$json_file" ]; then
        # 仅在首次启动时写入默认值，若宿主持久卷中已有文件则保留用户状态，
        # 避免 rebuild 容器时暗中覆盖本地配置。
        echo '{"hasCompletedOnboarding":true,"firstStartTime":"","opusProMigrationComplete":true,"sonnet1m45MigrationComplete":true,"migrationVersion":11,"projects":{}}' > "$json_file"
        echo "[entrypoint] initialized default $json_file"
    else
        echo "[entrypoint] keeping existing $json_file from persisted /home/agent volume"
    fi

    # 历史版本可能把 hasCompletedOnboarding 覆盖丢失，导致 Claude 再次进入首次引导。
    node -e '
const fs = require("fs")
const path = process.argv[1]
const cfg = JSON.parse(fs.readFileSync(path, "utf8"))
cfg.hasCompletedOnboarding = true
fs.writeFileSync(path, JSON.stringify(cfg, null, 2))
' "$json_file"
    echo "[entrypoint] ensured hasCompletedOnboarding=true in $json_file"
}

ensure_claude_json

mkdir -p "$AGENT_HOME/.claude"
if [ ! -f "$AGENT_HOME/.claude/settings.json" ]; then
    echo '{}' > "$AGENT_HOME/.claude/settings.json"
    echo "[entrypoint] initialized empty $AGENT_HOME/.claude/settings.json"
else
    echo "[entrypoint] keeping existing $AGENT_HOME/.claude/settings.json from persisted /home/agent volume"
fi

# Claude 对 bypass 提示实际会读取顶层和 permissions 下两个位置的 skip 标记。
# 这里做幂等 merge，避免历史持久卷缺字段时再次弹交互确认或主题选择。
node -e '
const fs = require("fs")
const path = process.argv[1]
const cfg = JSON.parse(fs.readFileSync(path, "utf8"))
cfg.permissions ||= {}
cfg.permissions.allow ||= ["Bash(*)", "Read(*)", "Write(*)", "Edit(*)", "MultiEdit(*)", "WebFetch(*)", "WebSearch(*)"]
cfg.permissions.deny ||= []
cfg.permissions.skipDangerousModePermissionPrompt = true
cfg.skipDangerousModePermissionPrompt = true
cfg.theme ||= "dark"
fs.writeFileSync(path, JSON.stringify(cfg, null, 2))
' "$AGENT_HOME/.claude/settings.json"
echo "[entrypoint] ensured default Claude settings in $AGENT_HOME/.claude/settings.json"

cat > "$AGENT_HOME/.tmux.conf" << TMUXEOF
set -g history-limit 50000
TMUXEOF

if [ "${SKIP_TMUX_INIT:-0}" != "1" ]; then
    tmux new-session -d -s _init 2>/dev/null || true
    sleep 0.5
    tmux kill-session -t _init 2>/dev/null || true
fi

exec "$@"
