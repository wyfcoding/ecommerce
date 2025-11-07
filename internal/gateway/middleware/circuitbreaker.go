package middleware

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CircuitState 熔断器状态
type CircuitState int

const (
	StateClosed CircuitState = iota // 关闭状态 (正常)
	StateOpen                        // 打开状态 (熔断)
	StateHalfOpen                    // 半开状态 (尝试恢复)
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	mu sync.RWMutex

	state         CircuitState
	failureCount  int       // 失败次数
	successCount  int       // 成功次数 (半开状态下)
	lastFailTime  time.Time // 最后失败时间
	lastStateTime time.Time // 最后状态变更时间

	// 配置参数
	maxFailures     int           // 最大失败次数
	timeout         time.Duration // 熔断超时时间
	halfOpenSuccess int           // 半开状态下需要的成功次数
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(maxFailures int, timeout time.Duration, halfOpenSuccess int) *CircuitBreaker {
	return &CircuitBreaker{
		state:           StateClosed,
		maxFailures:     maxFailures,
		timeout:         timeout,
		halfOpenSuccess: halfOpenSuccess,
		lastStateTime:   time.Now(),
	}
}

// Call 执行调用
func (cb *CircuitBreaker) Call(fn func() error) error {
	if !cb.canProceed() {
		return errors.New("circuit breaker is open")
	}

	err := fn()
	cb.recordResult(err)
	return err
}

// canProceed 检查是否可以继续执行
func (cb *CircuitBreaker) canProceed() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// 检查是否超过超时时间，可以尝试恢复
		if time.Since(cb.lastStateTime) > cb.timeout {
			cb.setState(StateHalfOpen)
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// recordResult 记录调用结果
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

// onSuccess 处理成功调用
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		cb.failureCount = 0
	case StateHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.halfOpenSuccess {
			zap.S().Infof("circuit breaker recovered to CLOSED state")
			cb.setState(StateClosed)
			cb.failureCount = 0
			cb.successCount = 0
		}
	}
}

// onFailure 处理失败调用
func (cb *CircuitBreaker) onFailure() {
	cb.lastFailTime = time.Now()

	switch cb.state {
	case StateClosed:
		cb.failureCount++
		if cb.failureCount >= cb.maxFailures {
			zap.S().Warnf("circuit breaker opened due to %d failures", cb.failureCount)
			cb.setState(StateOpen)
		}
	case StateHalfOpen:
		zap.S().Warnf("circuit breaker reopened due to failure in half-open state")
		cb.setState(StateOpen)
		cb.successCount = 0
	}
}

// setState 设置状态
func (cb *CircuitBreaker) setState(state CircuitState) {
	cb.state = state
	cb.lastStateTime = time.Now()
}

// GetState 获取当前状态
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset 重置熔断器
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.lastStateTime = time.Now()
}

// CircuitBreakerMiddleware 熔断器中间件
func CircuitBreakerMiddleware(cb *CircuitBreaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := cb.Call(func() error {
			c.Next()
			// 根据 HTTP 状态码判断是否失败
			if c.Writer.Status() >= 500 {
				return errors.New("server error")
			}
			return nil
		})

		if err != nil && err.Error() == "circuit breaker is open" {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "service temporarily unavailable",
			})
			c.Abort()
		}
	}
}

// ServiceCircuitBreaker 服务级别的熔断器管理
type ServiceCircuitBreaker struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex

	maxFailures     int
	timeout         time.Duration
	halfOpenSuccess int
}

// NewServiceCircuitBreaker 创建服务熔断器管理器
func NewServiceCircuitBreaker(maxFailures int, timeout time.Duration, halfOpenSuccess int) *ServiceCircuitBreaker {
	return &ServiceCircuitBreaker{
		breakers:        make(map[string]*CircuitBreaker),
		maxFailures:     maxFailures,
		timeout:         timeout,
		halfOpenSuccess: halfOpenSuccess,
	}
}

// GetBreaker 获取或创建服务的熔断器
func (scb *ServiceCircuitBreaker) GetBreaker(serviceName string) *CircuitBreaker {
	scb.mu.RLock()
	breaker, exists := scb.breakers[serviceName]
	scb.mu.RUnlock()

	if !exists {
		scb.mu.Lock()
		breaker, exists = scb.breakers[serviceName]
		if !exists {
			breaker = NewCircuitBreaker(scb.maxFailures, scb.timeout, scb.halfOpenSuccess)
			scb.breakers[serviceName] = breaker
		}
		scb.mu.Unlock()
	}

	return breaker
}

// ServiceCircuitBreakerMiddleware 基于服务的熔断器中间件
func ServiceCircuitBreakerMiddleware(scb *ServiceCircuitBreaker, serviceNameFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName := serviceNameFunc(c)
		breaker := scb.GetBreaker(serviceName)

		err := breaker.Call(func() error {
			c.Next()
			if c.Writer.Status() >= 500 {
				return errors.New("server error")
			}
			return nil
		})

		if err != nil && err.Error() == "circuit breaker is open" {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "service temporarily unavailable",
				"service": serviceName,
			})
			c.Abort()
		}
	}
}
