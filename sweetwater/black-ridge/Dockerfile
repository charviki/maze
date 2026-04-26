# 阶段 1：构建前端资产 (支持 Monorepo)
FROM public.ecr.aws/docker/library/node:22-bookworm-slim AS frontend-builder

WORKDIR /src

# 安装 pnpm
RUN npm install -g pnpm

# 复制 Monorepo 工作区配置和锁定文件
COPY pnpm-workspace.yaml package.json ./

# 复制共享组件库
COPY fabrication ./fabrication

# 复制当前子项目代码
COPY sweetwater/black-ridge/web ./sweetwater/black-ridge/web

# 安装所有依赖并构建
RUN pnpm install

WORKDIR /src/sweetwater/black-ridge/web
RUN pnpm run build

FROM public.ecr.aws/docker/library/golang:1.24-bookworm AS backend-builder

WORKDIR /src

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.org

# 先复制 cradle 共享库
COPY mesa-hub/cradle/ mesa-hub/cradle/

# 复制 agent 源码
COPY sweetwater/black-ridge/server/ ./

RUN rm -rf ./web-dist
COPY --from=frontend-builder /src/sweetwater/black-ridge/web-dist ./web-dist/

# 将 go.mod 中的 replace 路径调整为容器内路径，然后下载依赖并构建
RUN sed -i 's|../../../mesa-hub/cradle|./mesa-hub/cradle|g' go.mod \
    && go mod download \
    && CGO_ENABLED=0 go build -o /bin/agent .

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

RUN npm install -g @anthropic-ai/claude-code

RUN useradd -m -s /bin/bash agent

RUN mkdir -p /home/agent/.claude && \
    echo '{"hasCompletedOnboarding":true,"firstStartTime":"","opusProMigrationComplete":true,"sonnet1m45MigrationComplete":true,"migrationVersion":11}' > /home/agent/.claude.json

COPY --from=backend-builder /bin/agent /usr/local/bin/agent
COPY sweetwater/black-ridge/config.example.yaml /usr/local/bin/config.yaml

RUN printf '%s\n' \
    'set -g history-limit 50000' \
    > /home/agent/.tmux.conf

RUN chown -R agent:agent /home/agent

COPY sweetwater/black-ridge/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

USER agent

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["agent"]
