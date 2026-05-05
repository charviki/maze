FROM public.ecr.aws/docker/library/golang:1.26-bookworm AS backend-builder

ENV GOTOOLCHAIN=auto

WORKDIR /src

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.org

COPY fabrication/cradle/ fabrication/cradle/

COPY the-mesa/director-core/go.mod the-mesa/director-core/go.sum ./
RUN sed -i 's|../../fabrication/cradle|./fabrication/cradle|g' go.mod
RUN --mount=type=cache,id=go-pkg-mod,target=/go/pkg/mod \
    go mod download

COPY the-mesa/director-core/ ./
RUN sed -i 's|../../fabrication/cradle|./fabrication/cradle|g' go.mod

RUN --mount=type=cache,id=go-pkg-mod,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -o /bin/director-core ./cmd/director-core

FROM public.ecr.aws/docker/library/docker:cli AS docker-cli

FROM public.ecr.aws/docker/library/debian:bookworm-slim

RUN sed -i 's|deb.debian.org|mirrors.tuna.tsinghua.edu.cn|g' /etc/apt/sources.list.d/debian.sources \
    && apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=docker-cli /usr/local/bin/docker /usr/local/bin/docker
COPY --from=docker-cli /usr/local/libexec/docker/cli-plugins/docker-buildx /usr/local/libexec/docker/cli-plugins/docker-buildx
COPY --from=backend-builder /bin/director-core /usr/local/bin/director-core
COPY the-mesa/director-core/config.yaml /usr/local/bin/config.yaml

EXPOSE 8080

ENTRYPOINT ["director-core"]
