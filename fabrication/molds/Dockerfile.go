FROM public.ecr.aws/docker/library/golang:1.24-bookworm AS builder
WORKDIR /go
RUN cp -r /usr/local/go /opt/go

ENV GOROOT=/opt/go
ENV PATH="/opt/go/bin:${PATH}"
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.org

COPY fabrication/deps/go.txt /tmp/go-deps.txt
RUN xargs -a /tmp/go-deps.txt -I {} go install {} || true

FROM scratch
COPY --from=builder /opt/go /opt/go
