FROM public.ecr.aws/docker/library/golang:1.26-bookworm AS builder

ENV GOTOOLCHAIN=auto

WORKDIR /src

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.org

COPY fabrication/cradle/ fabrication/cradle/

COPY the-mesa/the-forge/go.mod the-mesa/the-forge/go.sum ./
RUN go mod edit -replace github.com/charviki/maze/fabrication/cradle=./fabrication/cradle
RUN --mount=type=cache,id=go-pkg-mod,target=/go/pkg/mod \
    go mod download

COPY the-mesa/the-forge/ ./
RUN go mod edit -replace github.com/charviki/maze/fabrication/cradle=./fabrication/cradle

RUN --mount=type=cache,id=go-pkg-mod,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -o /bin/the-forge ./cmd/the-forge

FROM public.ecr.aws/docker/library/debian:bookworm-slim

RUN sed -i 's|deb.debian.org|mirrors.tuna.tsinghua.edu.cn|g' /etc/apt/sources.list.d/debian.sources \
    && apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/the-forge /usr/local/bin/the-forge
COPY the-mesa/the-forge/config.yaml /usr/local/bin/config.yaml

EXPOSE 8080 9090

ENTRYPOINT ["the-forge"]
