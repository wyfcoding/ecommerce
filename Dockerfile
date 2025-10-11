# 阶段 1: 构建 Go 可执行文件
# 使用一个包含完整 Go 工具链的镜像作为构建环境
FROM golang:1.22-alpine AS builder

# 设置构建环境变量，创建静态链接的 Linux 可执行文件
ENV CGO_ENABLED=0 GOOS=linux

WORKDIR /app

# 仅复制依赖管理文件
COPY go.mod go.sum ./
# 下载依赖项。这一步可以利用 Docker 的层缓存，只有在 go.mod/go.sum 变化时才会重新执行。
RUN go mod download

# 复制项目的其余所有源代码
COPY . .

# SERVICE_NAME 将在构建时通过 --build-arg 传入 (例如: --build-arg SERVICE_NAME=user)
ARG SERVICE_NAME
# 检查 SERVICE_NAME 是否被提供
RUN if [ -z "$SERVICE_NAME" ]; then echo "Error: SERVICE_NAME build-arg is required."; exit 1; fi

# 构建指定服务的二进制文件
# -a: 强制重新构建
# -installsuffix cgo: 防止不同 CGO 设置的包冲突
RUN go build -a -installsuffix cgo -o /app/bin/${SERVICE_NAME} ./cmd/${SERVICE_NAME}

# ---

# 阶段 2: 创建轻量级的生产镜像
# 使用 Alpine Linux 作为基础镜像，因为它非常小
FROM alpine:latest

# 为容器内的非 root 用户创建一个目录
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# SERVICE_NAME 构建参数，用于从构建阶段复制正确的文件
ARG SERVICE_NAME

# 从构建器阶段复制编译好的二进制文件
# 同时设置文件的所有者为非 root 用户
# 注意：如果使用非root用户，需要确保复制的文件权限正确
COPY --chown=appuser:appgroup --from=builder /app/bin/${SERVICE_NAME} .

# 设置容器以非 root 用户运行
USER appuser

# 端口声明 (仅为文档目的)
# 服务的实际端口应由外部配置决定
EXPOSE 8000
EXPOSE 9000

# 健康检查指令
# K8s的探针是更好的选择，但这是一个好的备用方案，尤其适用于本地docker环境。
# 如果启用健康检查，请确保安装了 'curl' 或 'wget'。
# 例如，在 Alpine 镜像中，可以通过 'RUN apk --no-cache add curl' 安装。
# RUN apk --no-cache add curl
# HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
#   CMD curl -f http://localhost:8000/healthz || exit 1

# 设置容器以非 root 用户运行
# USER appuser

# 定义容器启动时运行的命令
# 配置文件路径将由运行时环境（如 Kubernetes）提供
# 示例: ["./user", "--conf", "/app/configs/user.toml"]
CMD ["./${SERVICE_NAME}"]