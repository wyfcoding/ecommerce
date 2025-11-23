#!/bin/bash

# Install plugins if not installed (optional, can be removed if environment is pre-configured)
# go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# API root directory
API_ROOT="./api"

# Find all proto files
PROTO_FILES=$(find $API_ROOT -name "*.proto")

for PROTO_FILE in $PROTO_FILES; do
    echo "Generating code for $PROTO_FILE..."
    protoc --proto_path=. \
           --proto_path=$API_ROOT \
           --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           $PROTO_FILE
done

echo "Protobuf generation completed."
