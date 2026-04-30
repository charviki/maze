# 阶段 1：构建前端资产
FROM public.ecr.aws/docker/library/node:22-bookworm-slim AS frontend-builder

WORKDIR /src

RUN npm install -g pnpm

# 只复制依赖声明文件
COPY pnpm-workspace.yaml package.json pnpm-lock.yaml ./
COPY fabrication/skin/package.json fabrication/skin/package.json
COPY sweetwater/black-ridge/web/package.json sweetwater/black-ridge/web/

# 安装依赖（package.json 不变时命中 Docker 层缓存）
RUN --mount=type=cache,id=pnpm-store,target=/root/.local/share/pnpm/store \
    pnpm install

# 复制源码（改源码只触发从这里开始的层重建）
COPY fabrication/skin/ fabrication/skin/
COPY sweetwater/black-ridge/web/ sweetwater/black-ridge/web/

WORKDIR /src/sweetwater/black-ridge/web
RUN pnpm run build

# 阶段 2：构建 Go 后端
FROM public.ecr.aws/docker/library/golang:1.24-bookworm AS backend-builder

WORKDIR /src

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.org

COPY fabrication/cradle/ mesa-hub/cradle/
COPY sweetwater/black-ridge/server/go.mod sweetwater/black-ridge/server/go.sum ./

RUN rm -rf ./web-dist
COPY --from=frontend-builder /src/sweetwater/black-ridge/web-dist ./web-dist/

RUN sed -i 's|../../../fabrication/cradle|./mesa-hub/cradle|g' go.mod
RUN --mount=type=cache,id=go-pkg-mod,target=/go/pkg/mod \
    go mod download

COPY sweetwater/black-ridge/server/ ./
RUN sed -i 's|../../../fabrication/cradle|./mesa-hub/cradle|g' go.mod

RUN --mount=type=cache,id=go-pkg-mod,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -o /bin/agent .

# 阶段 3：运行时镜像
FROM public.ecr.aws/docker/library/node:22-bookworm-slim

ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8

RUN apt-get update && apt-get install -y \
    tmux \
    git \
    curl \
    ca-certificates \
    procps \
    jq \
    && rm -rf /var/lib/apt/lists/*

COPY --from=maze-deps-claude:latest /opt/claude /opt/claude
ENV PATH="/opt/claude/bin:${PATH}"

RUN useradd -m -s /bin/bash agent

RUN mkdir -p /home/agent/.claude && \
    echo '{"hasCompletedOnboarding":true,"firstStartTime":"","opusProMigrationComplete":true,"sonnet1m45MigrationComplete":true,"migrationVersion":11}' > /home/agent/.claude.json

COPY --from=backend-builder /bin/agent /usr/local/bin/agent
COPY sweetwater/black-ridge/config.example.yaml /usr/local/bin/config.yaml

RUN printf '%s\n' \
    'set -g history-limit 50000' \
    'set -g mouse on' \
    "set -as terminal-features 'xterm*:mouse'" \
    '' \
    '# Scroll: auto enter copy-mode, forward to app if it handles mouse' \
    "bind -T root WheelUpPane if-shell -Ft= '#{mouse_any_flag}' 'send-keys -M' 'copy-mode -e; send-keys -M'" \
    "bind -T root WheelDownPane if-shell -Ft= '#{mouse_any_flag}' 'send-keys -M' 'send-keys -M'" \
    '' \
    '# Copy-mode scroll speed: 2 lines per event' \
    'bind -T copy-mode WheelUpPane select-pane \; send-keys -X -N 2 scroll-up' \
    'bind -T copy-mode WheelDownPane select-pane \; send-keys -X -N 2 scroll-down' \
    'bind -T copy-mode-vi WheelUpPane select-pane \; send-keys -X -N 2 scroll-up' \
    'bind -T copy-mode-vi WheelDownPane select-pane \; send-keys -X -N 2 scroll-down' \
    > /home/agent/.tmux.conf

RUN chown -R agent:agent /home/agent

COPY sweetwater/black-ridge/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

USER agent

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["agent"]
