# PKG 包说明

本目录包含项目的所有基础设施和工具包。

## 📦 核心包

### 应用框架
- **app** - 应用启动框架，提供统一的服务启动方式
- **config** - 配置管理，支持TOML格式配置文件
- **server** - 服务器封装（HTTP/gRPC）

### 数据存储
- **database/mysql** - MySQL数据库连接和配置
- **database/redis** - Redis连接和配置
- **database/mongodb** - MongoDB连接和配置
- **cache** - 缓存抽象层，基于Redis实现
- **elasticsearch** - Elasticsearch客户端

### 消息队列
- **messagequeue/kafka** - Kafka消息队列客户端

### 对象存储
- **minio** - MinIO对象存储客户端

### 认证授权
- **jwt** - JWT令牌生成和验证
- **hash** - 密码哈希（bcrypt）
- **middleware** - HTTP中间件（认证、CORS、日志、恢复）

### 可观测性
- **logging** - 日志系统（基于zap）
- **metrics** - 监控指标（基于Prometheus）
- **tracing** - 链路追踪（基于Jaeger）

### 工具类
- **idgen** - 分布式ID生成器（基于Snowflake）
- **errors** - 统一错误处理
- **response** - HTTP响应封装
- **validator** - 数据验证
- **pagination** - 分页工具
- **utils** - 通用工具函数（字符串、时间、金额）

### 高级功能
- **limiter** - 限流器（本地/分布式）
- **lock** - 分布式锁（基于Redis）
- **circuitbreaker** - 熔断器

## 🎯 使用原则

1. **单一职责** - 每个包只负责一个功能领域
2. **接口抽象** - 提供接口定义，便于替换实现
3. **配置驱动** - 通过配置文件控制行为
4. **可测试性** - 所有包都应该易于测试
5. **文档完善** - 每个包都有清晰的注释

## 📝 包依赖关系

```
app
├── config
├── logging
├── metrics
├── server
└── tracing

database
├── mysql
├── redis
└── mongodb

middleware
├── jwt
├── hash
└── response

utils (无依赖)
validator (无依赖)
errors (无依赖)
```

## 🔧 开发指南

### 添加新包
1. 在pkg目录下创建新包目录
2. 实现核心功能
3. 添加单元测试
4. 更新此README

### 修改现有包
1. 保持向后兼容
2. 更新测试用例
3. 更新文档

## 📊 代码统计

- 总包数：24个
- 总代码行数：~2300行
- 平均每包：~95行

## 🚀 未来计划

- [ ] 添加更多数据库支持（PostgreSQL、TiDB）
- [ ] 完善监控指标
- [ ] 添加更多中间件
- [ ] 性能优化
