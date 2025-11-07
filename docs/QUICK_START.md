# 快速开始指南

## 📋 环境要求

- Go 1.21+
- Docker & Docker Compose
- Make

## 🚀 快速启动

### 1. 克隆项目

```bash
git clone https://github.com/your-org/ecommerce-microservices.git
cd ecommerce-microservices
```

### 2. 启动基础设施

```bash
# 启动 MySQL, Redis, MongoDB, Elasticsearch 等
docker-compose up -d

# 验证服务
docker-compose ps
```

### 3. 初始化数据库

```bash
# 创建数据库
make db-create

# 运行迁移
make db-migrate

# 初始化测试数据（可选）
make db-seed
```

### 4. 启动服务

```bash
# 启动所有服务
make run-all

# 或单独启动
make run-user      # 用户服务
make run-product   # 商品服务
make run-order     # 订单服务
```

### 5. 验证

```bash
# 检查服务健康状态
curl http://localhost:8001/health  # 用户服务
curl http://localhost:8002/health  # 商品服务
curl http://localhost:8000/health  # API网关
```

## 📝 API测试

### 用户注册

```bash
curl -X POST http://localhost:8000/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "Test123456",
    "email": "test@example.com"
  }'
```

### 用户登录

```bash
curl -X POST http://localhost:8000/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "Test123456"
  }'
```

### 获取商品列表

```bash
curl http://localhost:8000/api/v1/products?page=1&pageSize=10
```

## 🔍 访问地址

- **API网关**: http://localhost:8000
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Jaeger**: http://localhost:16686

## ❓ 常见问题

### 端口冲突

```bash
# 查看端口占用
lsof -i :8001

# 修改配置文件
vim configs/config.yaml
```

### 数据库连接失败

```bash
# 检查MySQL
docker ps | grep mysql

# 测试连接
mysql -h127.0.0.1 -P3306 -uroot -proot123456
```

### 依赖下载失败

```bash
# 设置Go代理
go env -w GOPROXY=https://goproxy.cn,direct
go mod download
```

## 📚 更多文档

- [架构设计](ARCHITECTURE.md)
- [项目总览](PROJECT_OVERVIEW.md)
- [部署指南](../README_DEPLOYMENT.md)
