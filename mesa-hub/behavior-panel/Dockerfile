FROM public.ecr.aws/docker/library/golang:1.24-bookworm AS backend-builder

WORKDIR /src

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.org

COPY fabrication/cradle/ mesa-hub/cradle/

COPY mesa-hub/behavior-panel/server/ ./

RUN sed -i 's|../../../fabrication/cradle|./mesa-hub/cradle|g' go.mod

RUN --mount=type=cache,id=go-pkg-mod,target=/go/pkg/mod \
    go mod download

RUN CGO_ENABLED=0 go build -o /bin/agent-manager .

FROM public.ecr.aws/docker/library/debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=backend-builder /bin/agent-manager /usr/local/bin/agent-manager
COPY mesa-hub/behavior-panel/server/config.yaml /usr/local/bin/config.yaml

EXPOSE 8080

ENTRYPOINT ["agent-manager"]
