FROM public.ecr.aws/docker/library/golang:1.24-bookworm AS backend-builder

WORKDIR /src

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.org

# 先复制 cradle 共享库
COPY fabrication/cradle/ mesa-hub/cradle/

# 复制 behavior-panel 源码
COPY mesa-hub/behavior-panel/server/ ./

# 将 go.mod 中的 replace 路径从相对路径调整为容器内路径，然后下载依赖并构建
RUN sed -i 's|../../../fabrication/cradle|./mesa-hub/cradle|g' go.mod \
    && go mod download \
    && CGO_ENABLED=0 go build -o /bin/agent-manager .

FROM public.ecr.aws/docker/library/debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=backend-builder /bin/agent-manager /usr/local/bin/agent-manager
COPY mesa-hub/behavior-panel/server/config.yaml /usr/local/bin/config.yaml

EXPOSE 8080

ENTRYPOINT ["agent-manager"]
