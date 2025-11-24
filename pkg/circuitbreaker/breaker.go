package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// State 熔断器状态
type State int

const (
	StateClosed   State = iota // 关闭状态（正常）
	StateOpen                  // 打开状态（熔断）
	StateHalfOpen              // 半开状态（尝试恢复）
)

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	mu              sync.Mutex
	state           State
	failureCount    int
	successCount    int
	lastFailureTime time.Time

	maxFailures     int           // 最大失败次数
	timeout         time.Duration // 超时时间
	halfOpenSuccess int           // 半开状态需要的成功次数
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(maxFailures int, timeout time.Duration, halfOpenSuccess int) *CircuitBreaker {
	return &CircuitBreaker{
		state:           StateClosed,
		maxFailures:     maxFailures,
		timeout:         timeout,
		halfOpenSuccess: halfOpenSuccess,
	}
}

// Call 执行函数调用
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()

	// 检查是否需要从Open转到HalfOpen
	if cb.state == StateOpen && time.Since(cb.lastFailureTime) > cb.timeout {
		cb.state = StateHalfOpen
		cb.successCount = 0
	}

	// 如果是Open状态，直接返回错误
	if cb.state == StateOpen {
		cb.mu.Unlock()
		return ErrCircuitOpen
	}

	cb.mu.Unlock()

	// 执行函数
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

// onFailure 处理失败
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	cb.successCount = 0

	if cb.failureCount >= cb.maxFailures {
		cb.state = StateOpen
	}
}

// onSuccess 处理成功
func (cb *CircuitBreaker) onSuccess() {
	cb.failureCount = 0

	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.halfOpenSuccess {
			cb.state = StateClosed
		}
	}
}

// GetState 获取当前状态
func (cb *CircuitBreaker) GetState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
