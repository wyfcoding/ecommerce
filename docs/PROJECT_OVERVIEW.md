# 电商微服务系统 - 项目总览

## 📚 项目简介

这是一个基于 Go 语言开发的完整电商微服务系统，采用领域驱动设计（DDD）和微服务架构，支持高并发、高可用的电商业务场景。

**项目版本**: v1.0.0  
**Go 版本**: 1.21+  
**架构模式**: 微服务 + DDD + CQRS  
**通信协议**: gRPC + HTTP/REST

---

## 🏗️ 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      API Gateway                            │
│              (路由、认证、限流、熔断)                          │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
┌───────▼────────┐  ┌──────▼──────┐  ┌────────▼────────┐
│  核心业务服务   │  │  营销服务    │  │  基础设施服务    │
├────────────────┤  ├─────────────┤  ├─────────────────┤
│ • 用户服务      │  │ • 优惠券     │  │ • 搜索服务       │
│ • 商品服务      │  │ • 促销活动   │  │ • 推荐服务       │
│ • 订单服务      │  │ • 积分系统   │  │ • 支付服务       │
│ • 库存服务      │  │ • 秒杀活动   │  │ • 物流服务       │
│ • 购物车服务    │  │ • 拼团活动   │  │ • 通知服务       │
│ • 售后服务      │  │              │  │ • 客服系统       │
│                │  │              │  │ • 分析服务       │
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

---

## 📦 服务清单

### 核心业务服务 (8个)

| 服务名称 | 端口 | 功能描述 | 状态 |
|---------|------|---------|------|
| **用户服务** | 8001 | 用户注册、登录、信息管理、地址管理 | ✅ |
| **商品服务** | 8002 | SPU/SKU管理、分类、品牌、属性 | ✅ |
| **订单服务** | 8003 | 订单创建、查询、状态管理、日志 | ✅ |
| **库存服务** | 8004 | 库存查询、锁定、扣减、预警 | ✅ |
| **购物车服务** | 8005 | 购物车CRUD、合并、库存验证 | ✅ |
| **售后服务** | 8006 | 退款、换货、维修、工单系统 | ✅ |
| **支付服务** | 8007 | 支付创建、查询、退款、回调 | ✅ |
| **物流服务** | 8008 | 运费计算、物流跟踪、电子面单 | ✅ |

### 营销服务 (5个)

| 服务名称 | 端口 | 功能描述 | 状态 |
|---------|------|---------|------|
| **优惠券服务** | 8101 | 优惠券模板、领取、使用、计算 | ✅ |
| **促销服务** | 8102 | 促销活动、规则配置、商品管理 | ✅ |
| **积分服务** | 8103 | 积分增减、查询、交易记录、等级 | ✅ |
| **秒杀服务** | 8104 | 秒杀活动、商品、下单、库存扣减 | ⚠️ |
| **拼团服务** | 8105 | 拼团活动、发起、加入、成团 | ✅ |

### 基础设施服务 (10个)

| 服务名称 | 端口 | 功能描述 | 状态 |
|---------|------|---------|------|
| **搜索服务** | 8201 | 商品搜索、Elasticsearch集成 | ✅ |
| **推荐服务** | 8202 | 个性化推荐、Neo4j图数据库 | ✅ |
| **通知服务** | 8203 | 邮件、短信、推送通知 | ✅ |
| **客服系统** | 8204 | 在线聊天、工单、FAQ、机器人 | ✅ |
| **分析服务** | 8205 | 销售统计、用户分析、ClickHouse | ✅ |
| **定时任务** | 8206 | 订单处理、数据同步、报表生成 | ✅ |
| **风控服务** | 8207 | 反欺诈、风险评分、决策引擎 | ✅ |
| **内容审核** | 8208 | 文本审核、图片审核、AI模型 | ✅ |
| **数据采集** | 8209 | 事件采集、Kafka集成 | ✅ |
| **数据处理** | 8210 | Spark/Flink任务触发 | ✅ |

### 管理服务 (3个)

| 服务名称 | 端口 | 功能描述 | 状态 |
|---------|------|---------|------|
| **管理员服务** | 8301 | 管理员登录、权限、角色、审计 | ✅ |
| **结算服务** | 8302 | 订单结算、商家分账 | ✅ |
| **订阅服务** | 8303 | 会员订阅、计划管理 | ✅ |

### 网关服务 (1个)

| 服务名称 | 端口 | 功能描述 | 状态 |
|---------|------|---------|------|
| **API网关** | 8000 | 路由、认证、限流、熔断、聚合 | ✅ |

---

## 🗄️ 数据存储

### 关系型数据库 - MySQL 8.0

**用途**: 核心业务数据存储

**数据库列表**:
- `ecommerce_user` - 用户、地址
- `ecommerce_product` - 商品、分类、品牌
- `ecommerce_order` - 订单、订单项、物流
- `ecommerce_inventory` - 库存、库存日志
- `ecommerce_cart` - 购物车、购物车项
- `ecommerce_aftersales` - 退款、换货、维修
- `ecommerce_payment` - 支付、退款交易
- `ecommerce_marketing` - 优惠券、促销、积分
- `ecommerce_notification` - 通知日志
- `ecommerce_customer_service` - 客服会话、消息

### 缓存 - Redis 7.0

**用途**: 缓存、分布式锁、限流

**数据类型**:
- **String**: 库存缓存、价格缓存、用户会话
- **Hash**: 购物车数据、商品详情
- **Set**: 秒杀用户集合、拼团成员
- **ZSet**: 排行榜、热门商品
- **List**: 消息队列、日志队列

### 文档数据库 - MongoDB 7.0

**用途**: 非结构化数据存储

**集合列表**:
- `products` - 商品详情（富文本、图片）
- `reviews` - 用户评论
- `logs` - 操作日志
- `events` - 业务事件

### 搜索引擎 - Elasticsearch 8.11

**用途**: 全文搜索、日志分析

**索引列表**:
- `products` - 商品搜索
- `orders` - 订单搜索
- `logs` - 日志搜索

### 图数据库 - Neo4j 5.0

**用途**: 关系推荐、社交网络

**节点类型**:
- `User` - 用户节点
- `Product` - 商品节点
- `Category` - 分类节点

**关系类型**:
- `PURCHASED` - 购买关系
- `VIEWED` - 浏览关系
- `SIMILAR_TO` - 相似关系

### 列式数据库 - ClickHouse

**用途**: 实时数据分析、OLAP

**表列表**:
- `sales_facts` - 销售事实表
- `page_views` - 页面浏览事件
- `user_behaviors` - 用户行为

### 消息队列 - Kafka 3.0

**用途**: 异步消息、事件驱动

**Topic列表**:
- `order.created` - 订单创建事件
- `order.paid` - 订单支付事件
- `inventory.changed` - 库存变更事件
- `notification.send` - 通知发送事件

### 对象存储 - MinIO

**用途**: 文件存储

**Bucket列表**:
- `products` - 商品图片
- `users` - 用户头像
- `documents` - 文档文件

---

## 🔧 技术栈

### 后端框架
- **Web框架**: Gin
- **RPC框架**: gRPC
- **ORM**: GORM
- **配置管理**: Viper
- **日志**: Zap
- **验证**: go-playground/validator
- **JWT**: golang-jwt/jwt

### 中间件
- **服务发现**: Consul (可选)
- **配置中心**: Consul/Etcd (可选)
- **链路追踪**: Jaeger
- **监控**: Prometheus + Grafana
- **日志收集**: ELK Stack

### 开发工具
- **API文档**: Swagger/OpenAPI
- **代码生成**: Wire (依赖注入)
- **Mock**: GoMock
- **测试**: Testify

---

## 📂 项目结构

```
ecommerce-microservices/
├── api/                    # API定义（Protobuf）
│   └── v1/
│       ├── user.proto
│       ├── product.proto
│       └── ...
├── cmd/                    # 服务入口
│   ├── user/
│   │   └── main.go
│   ├── product/
│   │   └── main.go
│   └── ...
├── internal/               # 内部代码
│   ├── user/              # 用户服务
│   │   ├── model/         # 数据模型
│   │   ├── repository/    # 数据访问层
│   │   ├── service/       # 业务逻辑层
│   │   └── handler/       # HTTP/gRPC处理层
│   ├── product/           # 商品服务
│   ├── order/             # 订单服务
│   ├── inventory/         # 库存服务 ✨
│   ├── cart/              # 购物车服务
│   ├── aftersales/        # 售后服务 ✨
│   ├── payment/           # 支付服务
│   │   └── gateway/       # 支付网关 ✨
│   ├── logistics/         # 物流服务
│   │   └── gateway/       # 物流网关 ✨
│   ├── marketing/         # 营销服务
│   ├── loyalty/           # 积分服务
│   ├── flashsale/         # 秒杀服务
│   ├── groupbuy/          # 拼团服务 ✨
│   ├── search/            # 搜索服务
│   ├── recommendation/    # 推荐服务
│   ├── notification/      # 通知服务
│   ├── customerservice/   # 客服系统 ✨
│   ├── analytics/         # 分析服务
│   ├── scheduler/         # 定时任务 ✨
│   ├── risk_security/     # 风控服务
│   ├── content_moderation/# 内容审核
│   ├── data_ingestion/    # 数据采集
│   ├── data_processing/   # 数据处理
│   ├── settlement/        # 结算服务
│   ├── subscription/      # 订阅服务
│   ├── admin/             # 管理员服务
│   └── gateway/           # API网关
├── pkg/                    # 公共包
│   ├── app/               # 应用框架
│   ├── config/            # 配置管理
│   ├── database/          # 数据库连接
│   ├── errors/            # 错误处理
│   ├── hash/              # 密码哈希
│   ├── idgen/             # ID生成器
│   ├── jwt/               # JWT工具
│   ├── logging/           # 日志工具
│   ├── middleware/        # 中间件
│   ├── pagination/        # 分页工具
│   ├── response/          # 响应工具
│   ├── server/            # 服务器
│   ├── tracing/           # 链路追踪
│   ├── metrics/           # 监控指标
│   ├── validator/         # 验证工具
│   └── utils/             # 工具函数
├── configs/                # 配置文件
│   ├── config.yaml
│   └── config.toml
├── deployments/            # 部署文件
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   └── k8s/
│       ├── base/
│       ├── services/
│       └── middleware/
├── docs/                   # 文档
│   ├── MISSING_FEATURES.md          # 功能清单
│   ├── IMPLEMENTATION_PROGRESS.md   # 实施进度
│   ├── QUICK_START.md               # 快速开始
│   ├── COMPLETION_SUMMARY.md        # 完成总结
│   ├── FINAL_IMPLEMENTATION_SUMMARY.md # 最终总结
│   └── PROJECT_OVERVIEW.md          # 项目总览（本文档）
├── scripts/                # 脚本文件
│   ├── init-db.sql
│   └── gen_proto.sh
├── tests/                  # 测试文件
│   ├── unit/
│   └── integration/
├── Makefile               # Make命令
├── go.mod                 # Go模块
├── go.sum
├── Jenkinsfile            # CI/CD配置
└── README_DEPLOYMENT.md   # 部署文档
```

---

## 🚀 快速开始

### 1. 环境准备

```bash
# 安装 Go 1.21+
go version

# 安装 Docker 和 Docker Compose
docker --version
docker-compose --version
```

### 2. 克隆项目

```bash
git clone https://github.com/your-org/ecommerce-microservices.git
cd ecommerce-microservices
```

### 3. 启动基础设施

```bash
cd deployments/docker
docker-compose up -d
```

### 4. 初始化数据库

```bash
make db-migrate
make db-seed
```

### 5. 启动服务

```bash
# 启动所有服务
make run-all

# 或单独启动
make run-user
make run-product
make run-order
```

### 6. 访问服务

- **API网关**: http://localhost:8000
- **用户服务**: http://localhost:8001
- **商品服务**: http://localhost:8002
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000

---

## 📊 核心功能

### 用户管理
- ✅ 用户注册/登录
- ✅ 用户信息管理
- ✅ 收货地址管理
- ✅ JWT认证

### 商品管理
- ✅ SPU/SKU管理
- ✅ 商品分类
- ✅ 品牌管理
- ✅ 商品属性
- ✅ 商品搜索

### 订单管理
- ✅ 订单创建
- ✅ 订单查询
- ✅ 订单状态管理
- ✅ 订单日志

### 库存管理
- ✅ 库存查询
- ✅ 库存锁定/解锁
- ✅ 库存扣减/恢复
- ✅ 低库存预警

### 购物车
- ✅ 添加/删除商品
- ✅ 更新数量
- ✅ 购物车合并
- ✅ 库存验证

### 售后服务
- ✅ 退款管理
- ✅ 换货管理
- ✅ 维修管理
- ✅ 工单系统

### 支付服务
- ✅ 支付宝支付
- ✅ 微信支付
- ✅ 支付查询
- ✅ 退款处理

### 物流服务
- ✅ 运费计算
- ✅ 物流跟踪
- ✅ 电子面单（顺丰）

### 营销服务
- ✅ 优惠券系统
- ✅ 促销活动
- ✅ 积分系统
- ✅ 秒杀活动
- ✅ 拼团活动

### 客服系统
- ✅ 在线聊天
- ✅ 智能客服
- ✅ FAQ系统
- ✅ 工单管理

### 数据分析
- ✅ 销售统计
- ✅ 用户分析
- ✅ 商品排行
- ✅ 实时报表

---

## 🔐 安全特性

### 认证授权
- JWT Token认证
- 权限控制（RBAC）
- API密钥管理

### 数据安全
- 密码加密（bcrypt）
- 敏感信息加密
- SQL注入防护
- XSS防护

### 业务安全
- 库存防超卖
- 订单防重复提交
- 优惠券防刷
- 风控检测
- 支付签名验证

---

## 📈 性能优化

### 缓存策略
- Redis缓存（库存、价格、用户会话）
- 本地缓存（配置、字典）
- CDN加速（静态资源）

### 数据库优化
- 索引优化
- 分页查询
- 读写分离
- 连接池

### 并发控制
- 分布式锁（Redis）
- 乐观锁（版本号）
- 限流（令牌桶）
- 熔断（Circuit Breaker）

---

## 📝 开发规范

### 代码规范
- 遵循 Go 官方规范
- 使用 golangci-lint 检查
- 注释覆盖率 90%+

### Git 规范
- 分支策略：Git Flow
- 提交信息：语义化提交
- 代码审查：Pull Request

### 测试规范
- 单元测试覆盖率 80%+
- 集成测试
- 压力测试

---

## 📚 相关文档

- [快速开始指南](./QUICK_START.md)
- [功能清单](./MISSING_FEATURES.md)
- [实施进度](./IMPLEMENTATION_PROGRESS.md)
- [完成总结](./COMPLETION_SUMMARY.md)
- [最终总结](./FINAL_IMPLEMENTATION_SUMMARY.md)
- [部署指南](../README_DEPLOYMENT.md)

---

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 📄 许可证

本项目采用 MIT 许可证

---

## 📞 联系方式

- **Email**: dev@ecommerce.com
- **GitHub**: https://github.com/your-org/ecommerce-microservices
- **文档**: https://docs.ecommerce.com

---

**项目状态**: 🟢 进展顺利  
**完成度**: 75%  
**最后更新**: 2024-01-XX

---

**感谢使用本系统！** 🎉
