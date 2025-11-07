#!/bin/bash

# 生成 Protobuf Go 代码的脚本

set -e

# 获取项目根目录
ROOT_DIR=$(pwd)

# 在 api 目录下查找所有 .proto 文件
PROTO_FILES=$(find "${ROOT_DIR}/api" -name "*.proto")

# 检查是否存在任何 proto 文件
if [ -z "${PROTO_FILES}" ]; then
  echo "No .proto files found."
  exit 0
fi

# 确保 protoc-gen-go 和 protoc-gen-go-grpc 插件已安装
if ! command -v protoc-gen-go &> /dev/null; then
  echo "protoc-gen-go not found. Please run:"
  echo "go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
  exit 1
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
  echo "protoc-gen-go-grpc not found. Please run:"
  echo "go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
  exit 1
fi

echo "Generating Go code from protobuf files..."

# 为每个 .proto 文件生成 Go 代码
for PROTO_FILE in ${PROTO_FILES}; do
  echo "Processing: ${PROTO_FILE}"
  protoc --proto_path="${ROOT_DIR}" \
         --go_out="${ROOT_DIR}" --go_opt=paths=source_relative \
         --go-grpc_out="${ROOT_DIR}" --go-grpc_opt=paths=source_relative \
         "${PROTO_FILE}"
done

echo "Protobuf code generation completed."
