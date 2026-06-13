FROM public.ecr.aws/docker/library/golang:1.26-bookworm AS builder
WORKDIR /go
RUN cp -r /usr/local/go /opt/go

ENV GOROOT=/opt/go
ENV PATH="/opt/go/bin:${PATH}"
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.org

COPY fabrication/deps/go.txt /tmp/go-deps.txt
# 复用 go build/mod 缓存；grep 过滤注释/空行后逐条 go install（go.txt 含版本说明注释）；
# 去掉 || true 让锁版本失败时立即可见
RUN --mount=type=cache,id=go-install-cache,target=/root/.cache/go-build \
    --mount=type=cache,id=go-pkg-mod,target=/go/pkg/mod \
    grep -vE '^[[:space:]]*(#|$)' /tmp/go-deps.txt | xargs -I {} go install {}

FROM scratch
COPY --from=builder /opt/go /opt/go
