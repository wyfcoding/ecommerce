# 电商微服务系统

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com)

一个基于 Go 语言开发的完整电商微服务系统，采用领域驱动设计（DDD）和微服务架构。

## 📚 目录

- [特性](#特性)
- [系统架构](#系统架构)
- [技术栈](#技术栈)
- [快速开始](#快速开始)
- [服务列表](#服务列表)
- [项目结构](#项目结构)
- [开发指南](#开发指南)
- [部署](#部署)
- [文档](#文档)
- [贡献](#贡献)
- [许可证](#许可证)

## ✨ 特性

### 核心业务
- 🛍️ **用户系统** - 注册、登录、第三方登录（微信、QQ）、地址管理
- 📦 **商品系统** - SPU/SKU管理、分类、品牌、属性、搜索、推荐
- 🛒 **购物车** - 添加、删除、更新、合并、库存验证
- 📋 **订单系统** - 创建、支付、发货、收货、取消、退款
- 📊 **库存系统** - 多仓库管理、智能分配、调拨、盘点
- 💰 **支付系统** - 支付宝、微信支付、退款
- 🚚 **物流系统** - 运费计算、物流跟踪、电子面单
- 🔄 **售后系统** - 退款、换货、维修、工单

### 营销功能
- 🎫 **优惠券** - 满减、折扣、新人券、会员券
- 🎉 **促销活动** - 限时折扣、满减、买赠
- ⚡ **秒杀** - 高并发秒杀、库存预扣、限流熔断
- 👥 **拼团** - 发起、参与、自动成团
- 🎁 **积分系统** - 积分增减、等级、任务、积分商城、抽奖

### 基础设施
- 🔍 **搜索服务** - Elasticsearch全文搜索
- 🤖 **推荐服务** - 基于Neo4j的个性化推荐
- 📧 **通知服务** - 邮件、短信、推送
- 💬 **消息中心** - 站内信、消息模板、推送管理
- 👨‍💼 **客服系统** - 在线聊天、智能客服、工单、FAQ
- 📈 **数据分析** - 销售、订单、用户、商品等7种报表
- ⏰ **定时任务** - 订单自动处理、数据同步、报表生成
- 🛡️ **风控系统** - 反欺诈、风险评分
- 🔐 **内容审核** - 文本、图片审核

### 管理功能
- 👤 **管理员** - 权限管理、角色管理、操作审计
- 💼 **结算系统** - 订单结算、商家分账
- 📝 **CMS系统** - 内容管理、页面管理

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                      API Gateway                            │
│              (路由、认证、限流、熔断、聚合)                    │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
┌───────▼────────┐  ┌──────▼──────┐  ┌────────▼────────┐
│  核心业务服务   │  │  营销服务    │  │  基础设施服务    │
├────────────────┤  ├─────────────┤  ├─────────────────┤
│ • 用户服务      │  │ • 优惠券     │  │ • 搜索服务       │
│ • 商品服务      │  │ • 促销活动   │  │ • 推荐服务       │
│ • 订单服务      │  │ • 积分系统   │  │ • 通知服务       │
│ • 库存服务      │  │ • 秒杀活动   │  │ • 消息中心       │
│ • 购物车服务    │  │ • 拼团活动   │  │ • 客服系统       │
│ • 售后服务      │  │ • 积分商城   │  │ • 数据分析       │
│ • 支付服务      │  │              │  │ • 定时任务       │
│ • 物流服务      │  │              │  │ • 风控系统       │
│ • 仓库服务      │  │              │  │ • 内容审核       │
│ • OAuth服务     │  │              │  │                 │
└────────────────┘  └─────────────┘  └─────────────────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
┌─────────────────────────────────────────────────────────────┐
│                      数据存储层                              │
├─────────────────────────────────────────────────────────────┤
│ MySQL │ Redis │ MongoDB │ ES │ Neo4j │ ClickHouse │ Kafka │
└─────────────────────────────────────────────────────────────┘
```

## 🛠️ 技术栈

### 后端
- **语言**: Go 1.21+
- **Web框架**: Gin
- **RPC框架**: gRPC
- **ORM**: GORM
- **配置管理**: Viper
- **日志**: Zap
- **验证**: go-playground/validator
- **JWT**: golang-jwt/jwt

### 数据存储
- **MySQL 8.0** - 核心业务数据
- **Redis 7.0** - 缓存、分布式锁、限流
- **MongoDB 7.0** - 非结构化数据
- **Elasticsearch 8.11** - 全文搜索
- **Neo4j 5.0** - 图数据库（推荐系统）
- **ClickHouse** - OLAP分析
- **Kafka 3.0** - 消息队列
- **MinIO** - 对象存储

### 中间件
- **服务发现**: Consul (可选)
- **配置中心**: Consul/Etcd (可选)
- **链路追踪**: Jaeger
- **监控**: Prometheus + Grafana
- **日志收集**: ELK Stack

## 🚀 快速开始

### 环境要求

- Go 1.21+
- Docker & Docker Compose
- Make

### 安装

```bash
# 克隆项目
git clone https://github.com/your-org/ecommerce-microservices.git
cd ecommerce-microservices

# 安装依赖
go mod download

# 启动基础设施（MySQL、Redis、MongoDB等）
docker-compose up -d

# 初始化数据库
make db-migrate
make db-seed
```

### 运行

```bash
# 启动所有服务
make run-all

# 或单独启动服务
make run-user      # 用户服务
make run-product   # 商品服务
make run-order     # 订单服务
```

### 访问

- **API网关**: http://localhost:8000
- **用户服务**: http://localhost:8001
- **商品服务**: http://localhost:8002
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000

## 📦 服务列表

### 核心业务服务 (10个)
| 服务 | 端口 | 说明 |
|------|------|------|
| User Service | 8001 | 用户管理、认证 |
| Product Service | 8002 | 商品管理 |
| Order Service | 8003 | 订单管理 |
| Inventory Service | 8004 | 库存管理 |
| Cart Service | 8005 | 购物车 |
| AfterSales Service | 8006 | 售后服务 |
| Payment Service | 8007 | 支付服务 |
| Logistics Service | 8008 | 物流服务 |
| Warehouse Service | 8009 | 仓库管理 |
| OAuth Service | 8010 | 第三方登录 |

### 营销服务 (5个)
| 服务 | 端口 | 说明 |
|------|------|------|
| Marketing Service | 8101 | 优惠券、促销 |
| Loyalty Service | 8102 | 积分系统 |
| FlashSale Service | 8103 | 秒杀活动 |
| GroupBuy Service | 8104 | 拼团活动 |
| PointsMall Service | 8105 | 积分商城 |

### 基础设施服务 (10个)
| 服务 | 端口 | 说明 |
|------|------|------|
| Search Service | 8201 | 搜索服务 |
| Recommendation Service | 8202 | 推荐服务 |
| Notification Service | 8203 | 通知服务 |
| MessageCenter Service | 8204 | 消息中心 |
| CustomerService Service | 8205 | 客服系统 |
| Analytics Service | 8206 | 数据分析 |
| Report Service | 8207 | 报表系统 |
| Scheduler Service | 8208 | 定时任务 |
| RiskSecurity Service | 8209 | 风控系统 |
| ContentModeration Service | 8210 | 内容审核 |

### 管理服务 (3个)
| 服务 | 端口 | 说明 |
|------|------|------|
| Admin Service | 8301 | 管理员服务 |
| Settlement Service | 8302 | 结算服务 |
| Subscription Service | 8303 | 订阅服务 |

### 网关服务 (1个)
| 服务 | 端口 | 说明 |
|------|------|------|
| API Gateway | 8000 | API网关 |

## 📁 项目结构

```
ecommerce-microservices/
├── api/                    # API定义（Protobuf）
│   └── v1/
├── cmd/                    # 服务入口
│   ├── user/
│   ├── product/
│   └── ...
├── internal/               # 内部代码
│   ├── user/              # 用户服务
│   │   ├── model/         # 数据模型
│   │   ├── repository/    # 数据访问层
│   │   ├── service/       # 业务逻辑层
│   │   └── handler/       # HTTP/gRPC处理层
│   ├── product/           # 商品服务
│   ├── order/             # 订单服务
│   └── ...
├── pkg/                    # 公共包
│   ├── config/            # 配置管理
│   ├── database/          # 数据库连接
│   ├── errors/            # 错误处理
│   ├── jwt/               # JWT工具
│   ├── logging/           # 日志工具
│   └── ...
├── configs/                # 配置文件
├── deployments/            # 部署文件
│   ├── docker/
│   └── k8s/
├── docs/                   # 文档
├── scripts/                # 脚本文件
├── docker-compose.yml      # Docker Compose配置
├── Makefile               # Make命令
└── go.mod                 # Go模块
```

## 👨‍💻 开发指南

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `golangci-lint` 进行代码检查
- 注释覆盖率 90%+

### Git 规范

- 分支策略：Git Flow
- 提交信息：语义化提交
  - `feat:` 新功能
  - `fix:` 修复bug
  - `docs:` 文档更新
  - `refactor:` 代码重构
  - `test:` 测试相关

### 测试

```bash
# 运行单元测试
make test

# 运行集成测试
make test-integration

# 查看测试覆盖率
make test-coverage
```

## 🚢 部署

### Docker部署

```bash
# 构建镜像
make docker-build

# 启动服务
docker-compose up -d
```

### Kubernetes部署

```bash
# 部署到Kubernetes
kubectl apply -f deployments/k8s/

# 查看服务状态
kubectl get pods -n ecommerce
```

详细部署文档请参考：[部署指南](README_DEPLOYMENT.md)

## 📖 文档

- [架构设计](docs/ARCHITECTURE.md)
- [快速开始](docs/QUICK_START.md)
- [项目总览](docs/PROJECT_OVERVIEW.md)
- [API文档](docs/API.md)
- [部署指南](README_DEPLOYMENT.md)

## 🤝 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'feat: Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 📞 联系方式

- **Email**: dev@ecommerce.com
- **GitHub**: https://github.com/your-org/ecommerce-microservices
- **文档**: https://docs.ecommerce.com

---

**Star ⭐ 本项目如果对你有帮助！**
