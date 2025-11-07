# 部署指南

本文档提供了电商微服务系统的完整部署指南。

## 目录

- [环境要求](#环境要求)
- [本地开发环境](#本地开发环境)
- [Docker 部署](#docker-部署)
- [Kubernetes 部署](#kubernetes-部署)
- [监控和日志](#监控和日志)
- [CI/CD 流程](#cicd-流程)

## 环境要求

### 基础环境

- Go 1.21+
- Docker 20.10+
- Docker Compose 2.0+
- Kubernetes 1.25+
- kubectl
- Helm 3.0+

### 中间件

- MySQL 8.0+
- Redis 7.0+
- MongoDB 7.0+
- Elasticsearch 8.11+
- Neo4j 5.0+
- ClickHouse latest
- Kafka 3.0+
- MinIO latest

## 本地开发环境

### 1. 克隆项目

```bash
git clone https://github.com/your-org/ecommerce-microservices.git
cd ecommerce-microservices
```

### 2. 安装依赖

```bash
make deps
make install-tools
```

### 3. 启动基础设施

```bash
cd deployments/docker
docker-compose up -d mysql redis mongodb elasticsearch kafka
```

### 4. 初始化数据库

```bash
make db-init
```

### 5. 运行服务

```bash
# 运行单个服务
go run cmd/user/main.go

# 或使用 make
make build-service
./bin/user
```

## Docker 部署

### 1. 构建镜像

```bash
# 构建所有服务镜像
make docker-build

# 或构建单个服务
docker build -t ecommerce/user-service:latest \
  --build-arg SERVICE_NAME=user \
  -f deployments/docker/base.Dockerfile .
```

### 2. 启动所有服务

```bash
make docker-up

# 或
cd deployments/docker
docker-compose up -d
```

### 3. 查看日志

```bash
make docker-logs

# 查看特定服务日志
docker-compose logs -f user-service
```

### 4. 停止服务

```bash
make docker-down
```

## Kubernetes 部署

### 1. 准备 Kubernetes 集群

确保你有一个可用的 Kubernetes 集群，并配置好 kubectl。

```bash
kubectl cluster-info
```

### 2. 安装 Istio（可选）

```bash
# 下载 Istio
curl -L https://istio.io/downloadIstio | sh -
cd istio-*
export PATH=$PWD/bin:$PATH

# 安装 Istio
istioctl install --set profile=demo -y

# 启用自动注入
kubectl label namespace ecommerce istio-injection=enabled
```

### 3. 部署基础设施

```bash
# 创建命名空间
kubectl apply -f deployments/k8s/base/namespace.yaml

# 部署配置和密钥
kubectl apply -f deployments/k8s/base/configmap.yaml
kubectl apply -f deployments/k8s/base/secret.yaml

# 部署中间件
kubectl apply -f deployments/k8s/middleware/
```

### 4. 部署微服务

```bash
# 部署所有服务
make k8s-deploy

# 或手动部署
kubectl apply -f deployments/k8s/services/
```

### 5. 部署 Istio 配置

```bash
kubectl apply -f deployments/istio/
```

### 6. 验证部署

```bash
# 查看 Pod 状态
kubectl get pods -n ecommerce

# 查看服务
kubectl get svc -n ecommerce

# 查看 Ingress
kubectl get ingress -n ecommerce
```

### 7. 访问服务

```bash
# 获取 Ingress IP
kubectl get ingress -n ecommerce

# 或使用端口转发
kubectl port-forward -n ecommerce svc/user-service 8080:80
```

## 监控和日志

### Prometheus + Grafana

1. **访问 Prometheus**

```bash
# 端口转发
kubectl port-forward -n ecommerce svc/prometheus 9090:9090

# 访问 http://localhost:9090
```

2. **访问 Grafana**

```bash
# 端口转发
kubectl port-forward -n ecommerce svc/grafana 3000:3000

# 访问 http://localhost:3000
# 默认用户名/密码: admin/admin
```

3. **导入仪表盘**

在 Grafana 中导入预配置的仪表盘：
- `deployments/monitoring/grafana/dashboards/`

### Jaeger 链路追踪

```bash
# 端口转发
kubectl port-forward -n ecommerce svc/jaeger-query 16686:16686

# 访问 http://localhost:16686
```

### ELK 日志系统

1. **访问 Kibana**

```bash
# 端口转发
kubectl port-forward -n ecommerce svc/kibana 5601:5601

# 访问 http://localhost:5601
```

2. **配置索引模式**

在 Kibana 中创建索引模式：
- `ecommerce-logs-*`
- `ecommerce-errors-*`

## CI/CD 流程

### Jenkins 配置

1. **安装 Jenkins**

```bash
# 使用 Helm 安装
helm repo add jenkins https://charts.jenkins.io
helm install jenkins jenkins/jenkins -n jenkins --create-namespace
```

2. **配置 Jenkins**

- 安装必要插件：Kubernetes、Docker、Git
- 配置 Kubernetes 凭据
- 配置 Docker Registry 凭据
- 创建 Pipeline 任务，使用项目根目录的 `Jenkinsfile`

3. **触发构建**

```bash
# 推送代码到 Git 仓库会自动触发构建
git push origin main
```

### 手动部署流程

#### 开发环境

```bash
# 1. 构建镜像
make docker-build

# 2. 推送镜像
make docker-push

# 3. 部署到 dev 环境
kubectl set image deployment/user-service \
  user-service=registry.ecommerce.com/user-service:latest \
  -n ecommerce-dev

# 4. 验证部署
kubectl rollout status deployment/user-service -n ecommerce-dev
```

#### 生产环境（金丝雀发布）

```bash
# 1. 部署新版本
kubectl apply -f deployments/k8s/services/user/deployment-v2.yaml

# 2. 逐步切换流量（使用 Istio）
# 10% 流量
kubectl patch virtualservice user-service \
  --type merge \
  -p '{"spec":{"http":[{"route":[{"destination":{"host":"user-service","subset":"v1"},"weight":90},{"destination":{"host":"user-service","subset":"v2"},"weight":10}]}]}}'

# 观察指标，逐步增加到 100%

# 3. 删除旧版本
kubectl delete deployment user-service-v1
```

## 故障排查

### 查看 Pod 日志

```bash
kubectl logs -f <pod-name> -n ecommerce
```

### 进入 Pod 调试

```bash
kubectl exec -it <pod-name> -n ecommerce -- /bin/sh
```

### 查看事件

```bash
kubectl get events -n ecommerce --sort-by='.lastTimestamp'
```

### 查看资源使用

```bash
kubectl top pods -n ecommerce
kubectl top nodes
```

## 备份和恢复

### 数据库备份

```bash
# MySQL
kubectl exec -n ecommerce mysql-0 -- mysqldump -u root -p ecommerce > backup.sql

# MongoDB
kubectl exec -n ecommerce mongodb-0 -- mongodump --out=/backup
```

### 配置备份

```bash
# 备份所有配置
kubectl get all,configmap,secret -n ecommerce -o yaml > backup.yaml
```

## 扩缩容

### 手动扩缩容

```bash
kubectl scale deployment user-service --replicas=5 -n ecommerce
```

### 自动扩缩容

HPA 已配置在 `deployments/k8s/services/*/hpa.yaml`

```bash
# 查看 HPA 状态
kubectl get hpa -n ecommerce
```

## 安全建议

1. **使用 Secret 管理敏感信息**
2. **启用 RBAC**
3. **使用 Network Policy 限制网络访问**
4. **定期更新镜像和依赖**
5. **启用 Pod Security Policy**
6. **使用 Istio mTLS 加密服务间通信**

## 性能优化

1. **启用 HPA 自动扩缩容**
2. **配置资源限制和请求**
3. **使用 Redis 缓存**
4. **启用 CDN 加速静态资源**
5. **数据库读写分离**
6. **使用连接池**

## 联系方式

如有问题，请联系：
- Email: devops@ecommerce.com
- Slack: #ecommerce-platform
