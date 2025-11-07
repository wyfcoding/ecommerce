.PHONY: help proto build run test clean docker-build docker-up docker-down
 
# 默认目标
help:
	@echo "Available targets:"
	@echo "  proto         - Generate protobuf code"
	@echo "  build         - Build all services"
	@echo "  run-user      - Run user service"
	@echo "  run-product   - Run product service"
	@echo "  run-order     - Run order service"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker images"
	@echo "  docker-up     - Start Docker Compose services"
	@echo "  docker-down   - Stop Docker Compose services"

# 生成protobuf代码
proto:
	@echo "Generating protobuf code..."
	@./scripts/gen_proto.sh

# 构建所有服务
build:
	@echo "Building all services..."
	@go build -o bin/user-service ./cmd/user
	@go build -o bin/product-service ./cmd/product
	@go build -o bin/order-service ./cmd/order
	@go build -o bin/cart-service ./cmd/cart
	@echo "Build completed."

# 构建单个服务
build-user:
	@echo "Building user service..."
	@go build -o bin/user-service ./cmd/user

build-product:
	@echo "Building product service..."
	@go build -o bin/product-service ./cmd/product

build-order:
	@echo "Building order service..."
	@go build -o bin/order-service ./cmd/order

# 运行服务
run-user:
	@echo "Running user service..."
	@go run ./cmd/user/main.go

run-product:
	@echo "Running product service..."
	@go run ./cmd/product/main.go

run-order:
	@echo "Running order service..."
	@go run ./cmd/order/main.go

# 运行测试
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# 清理构建产物
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean completed."

# Docker相关
docker-build:
	@echo "Building Docker images..."
	@docker-compose build

docker-up:
	@echo "Starting Docker Compose services..."
	@docker-compose up -d

docker-down:
	@echo "Stopping Docker Compose services..."
	@docker-compose down

# 代码格式化
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

# 代码检查
lint:
	@echo "Running linters..."
	@golangci-lint run

# 安装依赖
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# 数据库迁移
migrate-up:
	@echo "Running database migrations..."
	@mysql -h localhost -u root -p < scripts/init-db.sql

# 生成mock
mock:
	@echo "Generating mocks..."
	@mockgen -source=internal/user/repository/repository.go -destination=internal/user/repository/mock/repository_mock.go -package=mock

# 生成Wire依赖注入代码
wire:
	@echo "Generating wire code..."
	@cd cmd/user && wire
	@cd cmd/product && wire
	@cd cmd/order && wire
	@echo "Wire generation completed."

# 运行所有服务（开发模式）
run-all:
	@echo "Starting all services..."
	@docker-compose up -d mysql redis kafka elasticsearch
	@sleep 5
	@make run-user &
	@make run-product &
	@make run-order &
	@make run-cart &
	@echo "All services started."

# 停止所有服务
stop-all:
	@echo "Stopping all services..."
	@pkill -f "go run ./cmd"
	@docker-compose down
	@echo "All services stopped."

# 查看服务日志
logs-user:
	@docker-compose logs -f user

logs-product:
	@docker-compose logs -f product

logs-order:
	@docker-compose logs -f order

# 健康检查
health:
	@echo "Checking service health..."
	@curl -s http://localhost:8001/health || echo "User service is down"
	@curl -s http://localhost:8002/health || echo "Product service is down"
	@curl -s http://localhost:8003/health || echo "Order service is down"

# 性能测试
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# 代码覆盖率
coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 安全扫描
security:
	@echo "Running security scan..."
	@gosec ./...

# 更新依赖
update-deps:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# 生成API文档
docs:
	@echo "Generating API documentation..."
	@swag init -g cmd/gateway/main.go

# 初始化开发环境
init-dev:
	@echo "Initializing development environment..."
	@make deps
	@make docker-up
	@sleep 10
	@make migrate-up
	@echo "Development environment initialized."

# 完整构建（包括测试和检查）
ci:
	@echo "Running CI pipeline..."
	@make fmt
	@make lint
	@make test
	@make build
	@echo "CI pipeline completed."
