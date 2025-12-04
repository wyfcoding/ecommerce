// Package middleware 提供了用于Gin和gRPC的通用中间件。
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimiter 实现了基于IP的请求限流器。
// 它维护一个map，为每个独立的IP地址分配一个 `rate.Limiter` 实例。
type IPRateLimiter struct {
	ips map[string]*rate.Limiter // 存储每个IP地址对应的限流器
	mu  *sync.RWMutex            // 读写锁，用于保护ips map的并发访问
	r   rate.Limit               // 允许每秒发生的事件数
	b   int                      // 令牌桶的容量
}

// NewIPRateLimiter 创建并返回一个新的IPRateLimiter实例。
// r: 限流的速率（例如，每秒允许多少个请求）。
// b: 令牌桶的大小（允许瞬时突发多少个请求）。
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}

	// 启动一个后台goroutine来定期清理旧的限流器。
	go i.cleanup()

	return i
}

// GetLimiter 获取给定IP地址的限流器。
// 如果该IP的限流器不存在，则会创建一个新的。
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()         // 加锁，确保map的安全访问
	defer i.mu.Unlock() // 函数返回时解锁

	limiter, exists := i.ips[ip]
	if !exists {
		// 如果IP的限流器不存在，则创建一个新的并存储
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

// cleanup 定期清理不再活跃的IP限流器。
// 这是一个简化的实现，当前版本仅在map过大时清空。
// 在生产环境中，需要更精细的策略，例如基于IP的上次访问时间来移除不活跃的条目，以防止内存泄漏。
func (i *IPRateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute) // 每分钟清理一次
		i.mu.Lock()
		// 简化的清理逻辑：如果IP数量超过阈值，直接清空map。
		// 更完善的实现应该追踪每个limiter的最后使用时间，并移除过期或不活跃的。
		if len(i.ips) > 10000 { // 假设10000是一个阈值
			i.ips = make(map[string]*rate.Limiter)
		}
		i.mu.Unlock()
	}
}

// RateLimitMiddleware 创建一个Gin中间件，用于对客户端IP进行请求限流。
// 如果请求超过限流，则返回 HTTP 429 Too Many Requests 状态码。
func RateLimitMiddleware(limiter *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()          // 获取客户端IP地址
		l := limiter.GetLimiter(ip) // 获取该IP对应的限流器
		if !l.Allow() {
			// 如果请求不允许通过（超过限流），则返回错误响应并中止请求链
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort() // 停止处理当前请求
			return
		}
		c.Next() // 请求通过限流，继续处理请求
	}
}

// RateLimit 是一个便捷函数，用于创建基于IP的限流中间件。
// 它封装了 IPRateLimiter 的创建和 RateLimitMiddleware 的应用。
// limit: 每秒允许的请求数。
// burst: 允许的突发请求数。
func RateLimit(limit int, burst int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rate.Limit(limit), burst)
	return RateLimitMiddleware(limiter)
}
