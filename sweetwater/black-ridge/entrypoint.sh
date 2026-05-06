#!/bin/bash
set -e

AGENT_HOME="${AGENT_HOME:-/home/agent}"

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
