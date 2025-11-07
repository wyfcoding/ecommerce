# Bug修复计划

## 当前状态
- 编译错误数量：约200+
- 主要问题类型：
  1. 未使用的变量/导入
  2. 类型不匹配
  3. 缺失的model字段
  4. 重复声明
  5. 方法签名不匹配

## 修复策略

### 阶段1：清理简单错误（已完成）
- ✅ 修复module名称不匹配
- ✅ 修复包名冲突
- ✅ 删除缺失的依赖引用
- ✅ 添加缺失的基础包

### 阶段2：修复核心错误（进行中）
需要修复的主要问题：

1. **未使用的变量/导入**
   - pkg/algorithm/warehouse_allocation.go:85 - qty未使用
   - pkg/database/redis.go:22 - fmt未定义
   - 多个model文件中gorm未使用

2. **缺失的model字段**
   - GroupBuyGroup, GroupBuyMember
   - InventoryLog
   - AftersalesApplication
   - 等等

3. **重复声明**
   - UserService
   - SearchService
   - RiskSecurityService
   - DataIngestionService
   - PriceRule, Discount, SkuPriceInfo
   - Review, WishlistItem

4. **方法签名不匹配**
   - response.Error参数数量
   - service方法参数类型

### 阶段3：建议的解决方案

由于错误太多且涉及大量业务逻辑，建议采用以下策略：

#### 方案A：保留核心服务，删除不完整的服务
保留以下核心服务：
- user
- product  
- order
- cart
- inventory
- warehouse
- aftersales
- groupbuy
- messagecenter
- oauth
- pointsmall
- report

删除或注释掉有严重问题的服务。

#### 方案B：创建最小可运行版本
1. 只保留最核心的3-5个服务
2. 确保这些服务完全可编译运行
3. 其他服务作为TODO逐步完善

#### 方案C：修复所有错误（工作量巨大）
需要：
1. 补全所有缺失的model定义
2. 统一所有接口签名
3. 删除所有重复声明
4. 修复所有类型不匹配

## 推荐方案

**采用方案B：创建最小可运行版本**

### 核心服务列表（5个）
1. **user** - 用户服务（基础）
2. **product** - 商品服务（基础）
3. **order** - 订单服务（核心业务）
4. **warehouse** - 仓库服务（已集成算法）
5. **aftersales** - 售后服务（新增完整）

### 实施步骤
1. 创建一个新的分支保存当前代码
2. 在主分支上只保留核心5个服务
3. 修复这5个服务的所有编译错误
4. 确保可以成功编译和运行
5. 逐步添加其他服务

## 下一步行动

请选择：
A. 继续修复所有错误（需要大量时间）
B. 采用最小可运行版本（推荐）
C. 其他建议
