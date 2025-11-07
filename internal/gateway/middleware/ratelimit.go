package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter 限流器接口
type RateLimiter interface {
	Allow(key string) bool
}

// TokenBucketLimiter 令牌桶限流器
type TokenBucketLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit // 每秒生成的令牌数
	burst    int        // 桶容量
}

// NewTokenBucketLimiter 创建令牌桶限流器
// r: 每秒生成的令牌数
// b: 桶容量
func NewTokenBucketLimiter(r rate.Limit, b int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// getLimiter 获取或创建限流器
func (l *TokenBucketLimiter) getLimiter(key string) *rate.Limiter {
	l.mu.RLock()
	limiter, exists := l.limiters[key]
	l.mu.RUnlock()

	if !exists {
		l.mu.Lock()
		// 双重检查
		limiter, exists = l.limiters[key]
		if !exists {
			limiter = rate.NewLimiter(l.rate, l.burst)
			l.limiters[key] = limiter
		}
		l.mu.Unlock()
	}

	return limiter
}

// Allow 检查是否允许请求
func (l *TokenBucketLimiter) Allow(key string) bool {
	limiter := l.getLimiter(key)
	return limiter.Allow()
}

// Cleanup 定期清理不活跃的限流器
func (l *TokenBucketLimiter) Cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			l.mu.Lock()
			// 简单实现：清空所有限流器
			// 生产环境应该记录最后访问时间，只清理长时间未使用的
			l.limiters = make(map[string]*rate.Limiter)
			l.mu.Unlock()
		}
	}()
}

// RateLimitMiddleware 限流中间件
// limiter: 限流器实例
// keyFunc: 生成限流 key 的函数 (例如: 根据 IP、用户ID 等)
func RateLimitMiddleware(limiter RateLimiter, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "too many requests, please try again later",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// IPRateLimitMiddleware 基于 IP 的限流中间件
func IPRateLimitMiddleware(r rate.Limit, b int) gin.HandlerFunc {
	limiter := NewTokenBucketLimiter(r, b)
	limiter.Cleanup(5 * time.Minute) // 每 5 分钟清理一次

	return RateLimitMiddleware(limiter, func(c *gin.Context) string {
		return c.ClientIP()
	})
}

// UserRateLimitMiddleware 基于用户的限流中间件
func UserRateLimitMiddleware(r rate.Limit, b int) gin.HandlerFunc {
	limiter := NewTokenBucketLimiter(r, b)
	limiter.Cleanup(5 * time.Minute)

	return RateLimitMiddleware(limiter, func(c *gin.Context) string {
		// 优先使用用户 ID
		if userID, exists := c.Get("user_id"); exists {
			return userID.(string)
		}
		// 未登录用户使用 IP
		return c.ClientIP()
	})
}

// APIRateLimitMiddleware 基于 API 路径的限流中间件
func APIRateLimitMiddleware(r rate.Limit, b int) gin.HandlerFunc {
	limiter := NewTokenBucketLimiter(r, b)
	limiter.Cleanup(5 * time.Minute)

	return RateLimitMiddleware(limiter, func(c *gin.Context) string {
		// 使用 IP + 路径作为 key
		return c.ClientIP() + ":" + c.Request.URL.Path
	})
}
