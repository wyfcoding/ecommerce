package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// ErrCircuitOpen 是当熔断器处于“打开”状态时返回的错误。
var ErrCircuitOpen = errors.New("circuit breaker is open")

// State 定义了熔断器的不同状态。
type State int

const (
	StateClosed   State = iota // 关闭状态：服务正常运行，所有请求都通过。
	StateOpen                  // 打开状态：服务被熔断，所有请求快速失败。
	StateHalfOpen              // 半开状态：熔断器在一定时间后尝试允许少量请求通过，探测服务是否恢复。
)

// CircuitBreaker 结构体实现了熔断器模式。
// 它通过监控服务的错误率来动态调整服务的可用性，避免雪崩效应。
type CircuitBreaker struct {
	mu              sync.Mutex // 互斥锁，用于保护熔断器内部状态的并发访问。
	state           State      // 熔断器当前的状态 (Closed, Open, HalfOpen)。
	failureCount    int        // 在Closed状态下，连续失败的次数。
	successCount    int        // 在HalfOpen状态下，连续成功的次数。
	lastFailureTime time.Time  // 上次失败的时间，用于计算Open状态的超时。

	maxFailures     int           // 在Closed状态下，允许的最大连续失败次数，超过则熔断器打开。
	timeout         time.Duration // 在Open状态下，熔断器保持打开的时间，超时后转为HalfOpen。
	halfOpenSuccess int           // 在HalfOpen状态下，需要连续成功的次数，达到则转为Closed。
}

// NewCircuitBreaker 创建并返回一个新的 CircuitBreaker 实例。
// maxFailures: 在Closed状态下，触发熔断所需的连续失败次数。
// timeout: 熔断器保持Open状态的持续时间。
// halfOpenSuccess: 在HalfOpen状态下，成功恢复所需的连续成功调用次数。
func NewCircuitBreaker(maxFailures int, timeout time.Duration, halfOpenSuccess int) *CircuitBreaker {
	return &CircuitBreaker{
		state:           StateClosed, // 初始状态为关闭。
		maxFailures:     maxFailures,
		timeout:         timeout,
		halfOpenSuccess: halfOpenSuccess,
	}
}

// Call 执行一个受熔断器保护的函数。
// fn: 待执行的业务逻辑函数。
// 返回函数执行的结果。如果熔断器处于Open状态，则直接返回 ErrCircuitOpen。
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock() // 加锁以保护状态修改和检查。

	// 如果熔断器处于Open状态，检查是否已达到超时时间，如果是则切换到HalfOpen状态。
	if cb.state == StateOpen && time.Since(cb.lastFailureTime) > cb.timeout {
		cb.state = StateHalfOpen
		cb.successCount = 0 // 重置HalfOpen状态下的成功计数。
	}

	// 如果熔断器处于Open状态，则直接返回错误，不执行业务逻辑。
	if cb.state == StateOpen {
		cb.mu.Unlock() // 解锁。
		return ErrCircuitOpen
	}

	cb.mu.Unlock() // 解锁，允许业务逻辑执行。

	// 执行业务函数。
	err := fn()

	cb.mu.Lock()         // 业务逻辑执行完毕后再次加锁，更新状态。
	defer cb.mu.Unlock() // 确保函数退出时解锁。

	if err != nil {
		cb.onFailure() // 处理失败情况。
		return err
	}

	cb.onSuccess() // 处理成功情况。
	return nil
}

// onFailure 处理函数调用失败后的逻辑。
// 根据熔断器当前状态，更新失败计数并判断是否需要切换到Open状态。
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++               // 增加失败计数。
	cb.lastFailureTime = time.Now() // 更新上次失败时间。
	cb.successCount = 0             // 在Closed或HalfOpen状态下失败，重置成功计数。

	// 如果在Closed状态下连续失败次数达到阈值，则将熔断器切换到Open状态。
	if cb.state == StateClosed && cb.failureCount >= cb.maxFailures {
		cb.state = StateOpen
	} else if cb.state == StateHalfOpen {
		// 在HalfOpen状态下如果再次失败，则立即切换回Open状态。
		cb.state = StateOpen
	}
}

// onSuccess 处理函数调用成功后的逻辑。
// 根据熔断器当前状态，更新成功计数并判断是否需要切换到Closed状态。
func (cb *CircuitBreaker) onSuccess() {
	// 成功调用会重置失败计数。
	cb.failureCount = 0

	// 如果熔断器处于HalfOpen状态，增加成功计数。
	if cb.state == StateHalfOpen {
		cb.successCount++
		// 如果在HalfOpen状态下连续成功次数达到阈值，则将熔断器切换回Closed状态。
		if cb.successCount >= cb.halfOpenSuccess {
			cb.state = StateClosed
		}
	} else if cb.state == StateClosed {
		// 在Closed状态下成功，不做特殊处理。
	}
}

// GetState 返回熔断器当前的运行状态。
func (cb *CircuitBreaker) GetState() State {
	cb.mu.Lock()         // 加锁，保护状态读取。
	defer cb.mu.Unlock() // 确保函数退出时解锁。
	return cb.state
}
