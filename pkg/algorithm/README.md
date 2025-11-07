## 电商微服务业务算法库

### 📚 已实现的算法

#### 1. 仓库智能分配算法 (`warehouse_allocation.go`)
**业务场景**: 订单下单时选择最优仓库发货

**使用位置**:
- `internal/warehouse/service/warehouse.go` - AllocateWarehouse()
- `internal/order/service/order.go` - CreateOrder()

**算法特点**:
- 综合考虑距离、库存、成本
- 支持订单拆分到多个仓库
- 贪心算法 + 评分机制

**使用示例**:
```go
allocator := algorithm.NewWarehouseAllocator()

// 准备仓库数据
warehouses := map[uint64]map[uint64]*algorithm.WarehouseInfo{
    1: { // 仓库1
        101: {ID: 1, Lat: 39.9, Lon: 116.4, Stock: 100, ShipCost: 500},
    },
}

// 分配
results := allocator.AllocateOptimal(userLat, userLon, items, warehouses)
```

---

#### 2. 优惠券最优组合算法 (`coupon_optimizer.go`)
**业务场景**: 用户下单时自动计算最优优惠组合

**使用位置**:
- `internal/coupon/service/coupon.go` - CalculateOptimalCoupons()
- `internal/order/service/order.go` - CalculateOrderPrice()

**算法特点**:
- 支持折扣券、满减券、立减券
- 动态规划求最优解
- 贪心算法快速近似解

**使用示例**:
```go
optimizer := algorithm.NewCouponOptimizer()

coupons := []algorithm.Coupon{
    {ID: 1, Type: algorithm.CouponTypeDiscount, DiscountRate: 0.8},
    {ID: 2, Type: algorithm.CouponTypeReduction, Threshold: 10000, ReductionAmount: 2000},
}

// 最优组合
couponIDs, finalPrice, discount := optimizer.OptimalCombination(originalPrice, coupons)
```

---

#### 3. 拼团智能匹配算法 (`groupbuy_matcher.go`)
**业务场景**: 用户参与拼团时自动匹配最合适的团

**使用位置**:
- `internal/groupbuy/service/groupbuy.go` - JoinGroup()

**算法特点**:
- 多种匹配策略（最快成团、最近地域、智能匹配）
- 综合评分机制
- 支持批量匹配

**使用示例**:
```go
matcher := algorithm.NewGroupBuyMatcher()

// 智能匹配
bestGroup := matcher.SmartMatch(activityID, userLat, userLon, userRegion, groups)

// 批量匹配
matches := matcher.BatchMatch(activityID, users, groups)
```

---

#### 4. 秒杀防刷算法 (`anti_bot.go`)
**业务场景**: 秒杀活动中识别和拦截机器人

**使用位置**:
- `internal/flashsale/service/flashsale.go` - Kill()
- `internal/gateway/middleware/antibot.go`

**算法特点**:
- 滑动窗口检测请求频率
- 行为模式分析
- 风险评分机制

**使用示例**:
```go
detector := algorithm.NewAntiBotDetector()

behavior := algorithm.UserBehavior{
    UserID: 123,
    IP: "192.168.1.1",
    UserAgent: "Mozilla/5.0...",
    Timestamp: time.Now(),
    Action: "kill",
}

// 判断是否为机器人
if isBot, reason := detector.IsBot(behavior); isBot {
    return errors.New("检测到异常行为: " + reason)
}

// 获取风险评分
score := detector.GetRiskScore(behavior)
if score > 80 {
    // 高风险，需要验证码
}
```

---

### 🚀 待实现的算法

#### 5. 物流路径规划算法
**文件**: `logistics_routing.go`

**业务场景**:
- 配送员路径优化（TSP问题）
- 多点配送最短路径
- 实时路况调整

**算法**:
- Dijkstra最短路径
- A*启发式搜索
- 遗传算法求解TSP

**使用位置**:
- `internal/logistics/service/routing.go`

---

#### 6. 商品推荐算法
**文件**: `recommendation.go`

**业务场景**:
- 猜你喜欢
- 看了又看
- 买了又买

**算法**:
- 协同过滤（UserCF/ItemCF）
- 矩阵分解（SVD）
- 热度衰减算法

**使用位置**:
- `internal/recommendation/service/recommend.go`

---

#### 7. 动态定价算法
**文件**: `dynamic_pricing.go`

**业务场景**:
- 根据库存、时间、竞品动态调价
- 差异化定价
- 促销价格优化

**算法**:
- 需求弹性模型
- 竞价算法
- 强化学习

**使用位置**:
- `internal/pricing/service/pricing.go`

---

#### 8. 风控算法
**文件**: `risk_control.go`

**业务场景**:
- 欺诈订单识别
- 异常交易检测
- 信用评分

**算法**:
- 孤立森林（异常检测）
- 逻辑回归（分类）
- 规则引擎

**使用位置**:
- `internal/risk_security/service/risk.go`

---

#### 9. 搜索排序算法
**文件**: `search_ranking.go`

**业务场景**:
- 商品搜索结果排序
- 个性化排序
- 多维度权重

**算法**:
- BM25算法
- Learning to Rank
- 多臂老虎机（A/B测试）

**使用位置**:
- `internal/search/service/search.go`

---

#### 10. 库存预测算法
**文件**: `inventory_forecast.go`

**业务场景**:
- 销量预测
- 补货建议
- 安全库存计算

**算法**:
- 移动平均（MA）
- 指数平滑（ETS）
- ARIMA时间序列

**使用位置**:
- `internal/analytics/service/forecast.go`

---

#### 11. 用户画像算法
**文件**: `user_profiling.go`

**业务场景**:
- 用户分群
- RFM模型
- 生命周期价值（LTV）

**算法**:
- K-Means聚类
- RFM分析
- 回归预测

**使用位置**:
- `internal/analytics/service/user_analysis.go`

---

#### 12. 库存分配优化（背包问题）
**文件**: `inventory_optimization.go`

**业务场景**:
- 有限库存下的订单优先级
- 高价值订单优先
- VIP用户优先

**算法**:
- 0-1背包问题
- 多重背包
- 贪心算法

**使用位置**:
- `internal/inventory/service/allocation.go`

---

### 📊 算法复杂度对比

| 算法 | 时间复杂度 | 空间复杂度 | 适用场景 |
|------|-----------|-----------|---------|
| 仓库分配（贪心） | O(n log n) | O(n) | 实时订单分配 |
| 优惠券组合（DP） | O(2^n) | O(n) | 优惠券数量<20 |
| 优惠券组合（贪心） | O(n log n) | O(n) | 快速计算 |
| 拼团匹配 | O(n log n) | O(n) | 实时匹配 |
| 防刷检测 | O(1) | O(n) | 高并发场景 |
| Dijkstra | O(V²) | O(V) | 路径规划 |
| 协同过滤 | O(n²) | O(n²) | 离线计算 |
| K-Means | O(nkt) | O(n) | 用户分群 |

---

### 🎯 使用建议

#### 1. 仓库分配
- 订单量大时使用贪心算法
- 需要精确最优解时使用动态规划
- 实时性要求高时使用缓存预计算

#### 2. 优惠券组合
- 优惠券<10个用完全枚举
- 优惠券10-20个用动态规划
- 优惠券>20个用贪心算法

#### 3. 拼团匹配
- 普通场景用智能匹配
- 地域性强的用最近匹配
- 时间紧迫用最快成团

#### 4. 防刷检测
- 秒杀场景必须使用
- 结合验证码双重防护
- 定期更新检测规则

---

### 🔧 扩展方向

1. **机器学习集成**
   - 使用TensorFlow/PyTorch训练模型
   - Go调用Python模型服务
   - 实时特征工程

2. **分布式计算**
   - Spark处理大规模数据
   - 实时流计算（Flink）
   - 离线批处理

3. **A/B测试**
   - 多臂老虎机算法
   - Thompson采样
   - UCB算法

4. **图算法**
   - 社交网络分析
   - 商品关联图
   - 知识图谱

---

**更新时间**: 2024-01-XX
