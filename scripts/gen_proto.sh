#!/bin/bash

# 此脚本用于从 Protobuf 文件生成 Go 代码。

set -e

# 获取项目根目录。
ROOT_DIR=$(git rev-parse --show-toplevel)

# 定义 Go Protobuf 生成文件的输出目录。
# 用户要求将生成的 pb 文件放在 api-go 目录下。
GO_PROTO_OUT_DIR="${ROOT_DIR}/api-go"

# 确保输出目录存在。
mkdir -p "${GO_PROTO_OUT_DIR}"

# 在 api 目录下查找所有 .proto 文件。
PROTO_FILES=$(find "${ROOT_DIR}/api" -name "*.proto")

# 检查是否存在任何 proto 文件。
if [ -z "${PROTO_FILES}" ]; then
  echo "No .proto files found." # 日志信息保持英文
  exit 0
fi

# 确保 protoc-gen-go 和 protoc-gen-go-grpc 插件已安装。
if ! command -v protoc-gen-go &> /dev/null; then
  echo "protoc-gen-go not found. Please run:" # 日志信息保持英文
  echo "go install google.golang.org/protobuf/cmd/protoc-gen-go@latest" # 命令保持英文
  exit 1
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
  echo "protoc-gen-go-grpc not found. Please run:" # 日志信息保持英文
  echo "go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest" # 命令保持英文
  exit 1
fi

echo "Generating Go code from protobuf files..." # 日志信息保持英文

# 为每个 .proto 文件生成 Go 代码。
for PROTO_FILE in ${PROTO_FILES}; do
  protoc --proto_path="${ROOT_DIR}" \
         --go_out="${GO_PROTO_OUT_DIR}" --go_opt=paths=source_relative \
         --go-grpc_out="${GO_PROTO_OUT_DIR}" --go-grpc_opt=paths=source_relative \
         "${PROTO_FILE}"
done

echo "Done." # 日志信息保持英文
