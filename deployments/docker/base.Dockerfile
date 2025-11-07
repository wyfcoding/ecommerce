# 基础镜像 - 多阶段构建
FROM golang:1.21-alpine AS builder

# 安装必要工具
RUN apk add --no-cache git make gcc musl-dev

WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建参数
ARG SERVICE_NAME
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# 编译
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o /app/bin/${SERVICE_NAME} \
    ./cmd/${SERVICE_NAME}

# 运行时镜像
FROM alpine:latest

# 安装 CA 证书和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

WORKDIR /app

# 从构建阶段复制二进制文件
ARG SERVICE_NAME
COPY --from=builder /app/bin/${SERVICE_NAME} /app/service
COPY --from=builder /app/configs /app/configs

# 修改所有权
RUN chown -R app:app /app

# 切换到非 root 用户
USER app

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 暴露端口
EXPOSE 8080 9090

# 启动服务
ENTRYPOINT ["/app/service"]
