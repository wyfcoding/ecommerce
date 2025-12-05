#!/bin/bash

# Batch translate Chinese log messages to English
# This script finds and replaces common Chinese log patterns with English equivalents

set -e

echo "Starting batch log translation..."

# Define translation patterns (Chinese -> English)
declare -A translations
translations["创建成功"]="created successfully"
translations["更新成功"]="updated successfully"
translations["删除成功"]="deleted successfully"
translations["获取成功"]="fetched successfully"
translations["保存成功"]="saved successfully"
translations["执行成功"]="executed successfully"
translations["处理成功"]="processed successfully"
translations["发送成功"]="sent successfully"
translations["接收成功"]="received successfully"
translations["连接成功"]="connected successfully"

translations["创建失败"]="failed to create"
translations["更新失败"]="failed to update"
translations["删除失败"]="failed to delete"
translations["获取失败"]="failed to fetch"
translations["保存失败"]="failed to save"
translations["执行失败"]="failed to execute"
translations["处理失败"]="failed to process"
translations["发送失败"]="failed to send"
translations["接收失败"]="failed to receive"
translations["连接失败"]="failed to connect"
translations["查询失败"]="failed to query"
translations["解析失败"]="failed to parse"
translations["验证失败"]="failed to validate"
translations["初始化失败"]="failed to initialize"

translations["开始处理"]="processing started"
translations["处理完成"]="processing completed"
translations["开始执行"]="execution started"
translations["执行完成"]="execution completed"

translations["数据库连接成功"]="database connected successfully"
translations["数据库连接失败"]="failed to connect to database"
translations["Redis连接成功"]="redis connected successfully"
translations["Redis连接失败"]="failed to connect to redis"

translations["用户创建成功"]="user created successfully"
translations["用户更新成功"]="user updated successfully"
translations["用户删除成功"]="user deleted successfully"
translations["订单创建成功"]="order created successfully"
translations["订单更新成功"]="order updated successfully"
translations["支付成功"]="payment succeeded"
translations["支付失败"]="payment failed"

translations["服务启动成功"]="service started successfully"
translations["服务停止"]="service stopped"
translations["服务启动失败"]="service failed to start"
translations["服务停止失败"]="service failed to stop"

translations["初始化完成"]="initialization completed"
translations["关闭连接"]="connection closed"

# Apply translations to all Go files
for chinese in "${!translations[@]}"; do
    english="${translations[$chinese]}"
    echo "Translating: $chinese -> $english"
    
    # Use find + sed to replace in all .go files
    find pkg internal cmd -name "*.go" -type f -exec sed -i '' "s/${chinese}/${english}/g" {} \;
done

echo "Batch translation completed!"
echo "Please review changes and run: go build ./..."
