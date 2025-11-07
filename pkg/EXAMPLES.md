# PKG 使用示例

## Cache 缓存使用

```go
import (
    "ecommerce/pkg/cache"
    "ecommerce/pkg/database/redis"
)

// 创建Redis客户端
redisClient, cleanup, _ := redis.NewRedisClient(&redis.Config{
    Addr: "localhost:6379",
})
defer cleanup()

// 创建缓存实例
cache := cache.NewRedisCache(redisClient, "myapp")

// 设置缓存
cache.Set(ctx, "user:1", user, 10*time.Minute)

// 获取缓存
var user User
cache.Get(ctx, "user:1", &user)
```

## Limiter 限流使用

```go
import "ecommerce/pkg/limiter"

// 本地限流器（每秒10个请求）
limiter := limiter.NewLocalLimiter(10, 10)

// 分布式限流器（1分钟内最多100个请求）
limiter := limiter.NewRedisLimiter(redisClient, 100, time.Minute)

// 检查是否允许
if allowed, _ := limiter.Allow(ctx, "user:123"); allowed {
    // 处理请求
}
```

## Lock 分布式锁使用

```go
import "ecommerce/pkg/lock"

lock := lock.NewRedisLock(redisClient)

// 获取锁
token, err := lock.Lock(ctx, "order:123", 30*time.Second)
if err != nil {
    // 获取锁失败
}
defer lock.Unlock(ctx, "order:123", token)

// 执行业务逻辑
```

## CircuitBreaker 熔断器使用

```go
import "ecommerce/pkg/circuitbreaker"

// 创建熔断器（5次失败后熔断，30秒后尝试恢复）
cb := circuitbreaker.NewCircuitBreaker(5, 30*time.Second, 2)

// 执行调用
err := cb.Call(func() error {
    return callExternalService()
})
```

## Pagination 分页使用

```go
import "ecommerce/pkg/pagination"

// 解析分页参数
page := &pagination.Page{
    PageNum:  1,
    PageSize: 20,
}
page.Validate()

// 查询数据
users, total := userRepo.List(page.Offset(), page.Limit())

// 返回分页结果
result := pagination.NewPageResult(total, page, users)
```

## Validator 验证使用

```go
import "ecommerce/pkg/validator"

// 验证手机号
if !validator.IsValidPhone("13800138000") {
    return errors.New("invalid phone")
}

// 验证邮箱
if !validator.IsValidEmail("user@example.com") {
    return errors.New("invalid email")
}

// 验证密码强度
if !validator.IsValidPassword("Pass123!") {
    return errors.New("weak password")
}
```

## Utils 工具使用

```go
import "ecommerce/pkg/utils"

// 生成随机字符串
code := utils.RandomString(6)

// MD5哈希
hash := utils.MD5("hello")

// 金额转换
yuan := utils.FenToYuan(10000) // 100.00元
fen := utils.YuanToFen(100.00) // 10000分

// 时间格式化
str := utils.FormatTime(time.Now())
```

## IDGen ID生成使用

```go
import "ecommerce/pkg/idgen"

// 初始化（在main函数中）
idgen.Init(1) // 机器ID

// 生成ID
id := idgen.GenID()

// 生成订单号
orderNo := idgen.GenOrderNo() // O1234567890

// 生成支付单号
paymentNo := idgen.GenPaymentNo() // P1234567890
```

## Errors 错误处理使用

```go
import "ecommerce/pkg/errors"

// 使用预定义错误
return errors.ErrUserNotFound

// 创建自定义错误
err := errors.New(errors.ErrCodeInvalidArgument, "invalid user id")

// 添加详情
err = err.WithDetails("user id must be positive")

// 转换为gRPC错误
return err.ToGRPCError()
```

## Response HTTP响应使用

```go
import "ecommerce/pkg/response"

// 成功响应
response.Success(c, user)

// 带消息的成功响应
response.SuccessWithMessage(c, "created successfully", user)

// 错误响应
response.BadRequest(c, "invalid parameter")
response.Unauthorized(c, "token expired")
response.NotFound(c, "user not found")
response.InternalError(c, "database error")
```
