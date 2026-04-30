FROM public.ecr.aws/docker/library/golang:1.24-bookworm AS backend-builder

WORKDIR /src

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.org

COPY fabrication/cradle/ mesa-hub/cradle/

COPY mesa-hub/behavior-panel/server/go.mod mesa-hub/behavior-panel/server/go.sum ./
RUN sed -i 's|../../../fabrication/cradle|./mesa-hub/cradle|g' go.mod
RUN --mount=type=cache,id=go-pkg-mod,target=/go/pkg/mod \
    go mod download

COPY mesa-hub/behavior-panel/server/ ./
RUN sed -i 's|../../../fabrication/cradle|./mesa-hub/cradle|g' go.mod

RUN --mount=type=cache,id=go-pkg-mod,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -o /bin/agent-manager .

FROM public.ecr.aws/docker/library/debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    gnupg \
    && install -m 0755 -d /etc/apt/keyrings \
    && curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg \
    && chmod a+r /etc/apt/keyrings/docker.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null \
    && apt-get update \
    && apt-get install -y docker-ce-cli \
    && rm -rf /var/lib/apt/lists/*

COPY --from=backend-builder /bin/agent-manager /usr/local/bin/agent-manager
COPY mesa-hub/behavior-panel/server/config.yaml /usr/local/bin/config.yaml

EXPOSE 8080

ENTRYPOINT ["agent-manager"]
