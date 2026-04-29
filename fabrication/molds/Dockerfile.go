# Go 工具链供应商镜像
# 安装 Go 1.24 和预装工具到 /opt/go/
FROM public.ecr.aws/docker/library/golang:1.24-bookworm

# 将 Go 安装复制到 /opt/go/
RUN cp -r /usr/local/go /opt/go

ENV GOROOT=/opt/go
ENV PATH="/opt/go/bin:${PATH}"
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.org

# 预装常用 Go 工具
COPY fabrication/deps/go.txt /tmp/go-deps.txt
RUN xargs -a /tmp/go-deps.txt -I {} go install {} || true

RUN rm -f /tmp/go-deps.txt
