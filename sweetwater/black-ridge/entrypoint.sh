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

# 动态发现 /opt/ 下所有工具的 bin 目录，注入 .bashrc 和 .profile。
# 这样无论供应商镜像提供了哪些工具，tmux session 内的 PATH 都能正确设置。
# 需要同时写 .bashrc 和 .profile：tmux new-session 默认启动 login shell（读 .profile），
# 而 tmux 内新开 pane 可能是 non-login interactive shell（读 .bashrc）。
OPT_PATHS=""
for dir in /opt/*/bin; do
    if [ -d "$dir" ]; then
        OPT_PATHS="${OPT_PATHS}${dir}:"
    fi
done

if [ -n "$OPT_PATHS" ]; then
    BASHRC_LINE="export PATH=\"${OPT_PATHS}\${PATH}\""
    BASHRC_FILE="$AGENT_HOME/.bashrc"
    PROFILE_FILE="$AGENT_HOME/.profile"

    # .bashrc（non-login interactive shell）
    if ! grep -qF "$BASHRC_LINE" "$BASHRC_FILE" 2>/dev/null; then
        echo "$BASHRC_LINE" >> "$BASHRC_FILE"
        echo "[entrypoint] appended $OPT_PATHS to $BASHRC_FILE"
    fi

    # .profile（login shell，tmux new-session 默认启动 login shell）
    if ! grep -qF "$BASHRC_LINE" "$PROFILE_FILE" 2>/dev/null; then
        echo "$BASHRC_LINE" >> "$PROFILE_FILE"
        echo "[entrypoint] appended $OPT_PATHS to $PROFILE_FILE"
    fi

    chown agent:agent "$BASHRC_FILE" "$PROFILE_FILE" 2>/dev/null || true
fi

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

cat > "$AGENT_HOME/.tmux.conf" << 'TMUXEOF'
set -g history-limit 50000
set -g mouse on
set -as terminal-features 'xterm*:mouse'

# 滚轮事件：自动进入 copy-mode 并滚动（如果应用自己处理鼠标则直接转发）
bind -T root WheelUpPane if-shell -Ft= '#{mouse_any_flag}' 'send-keys -M' 'copy-mode -e; send-keys -M'
bind -T root WheelDownPane if-shell -Ft= '#{mouse_any_flag}' 'send-keys -M' 'send-keys -M'

# copy-mode 滚动速度：每次 2 行
bind -T copy-mode WheelUpPane select-pane \; send-keys -X -N 2 scroll-up
bind -T copy-mode WheelDownPane select-pane \; send-keys -X -N 2 scroll-down
bind -T copy-mode-vi WheelUpPane select-pane \; send-keys -X -N 2 scroll-up
bind -T copy-mode-vi WheelDownPane select-pane \; send-keys -X -N 2 scroll-down
TMUXEOF

if [ "${SKIP_TMUX_INIT:-0}" != "1" ]; then
    tmux new-session -d -s _init 2>/dev/null || true
    sleep 0.5
    tmux kill-session -t _init 2>/dev/null || true
fi

exec "$@"
