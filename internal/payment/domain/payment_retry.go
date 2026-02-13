package domain

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

var (
	ErrMaxRetriesExceeded = errors.New("maximum retries exceeded")
	ErrRetryNotAllowed    = errors.New("retry not allowed in current state")
)

type RetryStatus int8

const (
	RetryStatusPending RetryStatus = iota
	RetryStatusInProgress
	RetryStatusSuccess
	RetryStatusFailed
	RetryStatusAbandoned
)

type RetryStrategy string

const (
	RetryStrategyFixed       RetryStrategy = "fixed"
	RetryStrategyExponential RetryStrategy = "exponential"
	RetryStrategyLinear      RetryStrategy = "linear"
)

type PaymentRetry struct {
	ID              uint64        `json:"id"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	PaymentID       uint64        `json:"payment_id"`
	PaymentNo       string        `json:"payment_no"`
	OrderID         uint64        `json:"order_id"`
	UserID          uint64        `json:"user_id"`
	Amount          int64         `json:"amount"`
	OriginalChannel string        `json:"original_channel"`
	CurrentChannel  string        `json:"current_channel"`
	Status          RetryStatus   `json:"status"`
	AttemptCount    int           `json:"attempt_count"`
	MaxAttempts     int           `json:"max_attempts"`
	NextRetryAt     *time.Time    `json:"next_retry_at"`
	LastAttemptAt   *time.Time    `json:"last_attempt_at"`
	LastError       string        `json:"last_error"`
	Strategy        RetryStrategy `json:"strategy"`
	BaseDelayMs     int64         `json:"base_delay_ms"`
	MaxDelayMs      int64         `json:"max_delay_ms"`
	Attempts        []*RetryAttempt `json:"attempts"`
}

type RetryAttempt struct {
	ID           uint64    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	RetryID      uint64    `json:"retry_id"`
	AttemptNum   int       `json:"attempt_num"`
	Channel      string    `json:"channel"`
	Amount       int64     `json:"amount"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	Success      bool      `json:"success"`
	ErrorCode    string    `json:"error_code"`
	ErrorMessage string    `json:"error_message"`
	LatencyMs    int64     `json:"latency_ms"`
}

type RetryConfig struct {
	MaxAttempts       int
	Strategy          RetryStrategy
	BaseDelayMs       int64
	MaxDelayMs        int64
	JitterPercent     float64
	RetryableErrors   []string
	NonRetryableErrors []string
}

func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:       3,
		Strategy:          RetryStrategyExponential,
		BaseDelayMs:       1000,
		MaxDelayMs:        60000,
		JitterPercent:     0.1,
		RetryableErrors:   []string{"TIMEOUT", "NETWORK_ERROR", "GATEWAY_BUSY", "INSUFFICIENT_FUNDS"},
		NonRetryableErrors: []string{"INVALID_CARD", "CARD_DECLINED", "FRAUD_DETECTED"},
	}
}

func NewPaymentRetry(paymentID uint64, paymentNo string, orderID, userID uint64, amount int64, originalChannel string, config *RetryConfig) *PaymentRetry {
	return &PaymentRetry{
		PaymentID:       paymentID,
		PaymentNo:       paymentNo,
		OrderID:         orderID,
		UserID:          userID,
		Amount:          amount,
		OriginalChannel: originalChannel,
		CurrentChannel:  originalChannel,
		Status:          RetryStatusPending,
		AttemptCount:    0,
		MaxAttempts:     config.MaxAttempts,
		Strategy:        config.Strategy,
		BaseDelayMs:     config.BaseDelayMs,
		MaxDelayMs:      config.MaxDelayMs,
		Attempts:        []*RetryAttempt{},
	}
}

func (r *PaymentRetry) CanRetry() bool {
	return r.Status == RetryStatusPending || r.Status == RetryStatusInProgress
}

func (r *PaymentRetry) HasMoreAttempts() bool {
	return r.AttemptCount < r.MaxAttempts
}

func (r *PaymentRetry) CalculateNextRetryDelay() time.Duration {
	if r.AttemptCount == 0 {
		return 0
	}

	var delayMs int64
	switch r.Strategy {
	case RetryStrategyFixed:
		delayMs = r.BaseDelayMs
	case RetryStrategyExponential:
		delayMs = r.BaseDelayMs * int64(1<<(r.AttemptCount-1))
		if delayMs > r.MaxDelayMs {
			delayMs = r.MaxDelayMs
		}
	case RetryStrategyLinear:
		delayMs = r.BaseDelayMs * int64(r.AttemptCount)
		if delayMs > r.MaxDelayMs {
			delayMs = r.MaxDelayMs
		}
	default:
		delayMs = r.BaseDelayMs
	}

	return time.Duration(delayMs) * time.Millisecond
}

func (r *PaymentRetry) StartAttempt(channel string) *RetryAttempt {
	r.AttemptCount++
	now := time.Now()
	attempt := &RetryAttempt{
		RetryID:    r.ID,
		AttemptNum: r.AttemptCount,
		Channel:    channel,
		Amount:     r.Amount,
		StartedAt:  now,
	}
	r.Attempts = append(r.Attempts, attempt)
	r.LastAttemptAt = &now
	r.Status = RetryStatusInProgress
	r.CurrentChannel = channel
	return attempt
}

func (r *PaymentRetry) CompleteAttempt(attempt *RetryAttempt, success bool, errorCode, errorMessage string, latencyMs int64) {
	now := time.Now()
	attempt.FinishedAt = &now
	attempt.Success = success
	attempt.ErrorCode = errorCode
	attempt.ErrorMessage = errorMessage
	attempt.LatencyMs = latencyMs

	if success {
		r.Status = RetryStatusSuccess
	} else {
		r.LastError = errorMessage
		if r.AttemptCount >= r.MaxAttempts {
			r.Status = RetryStatusFailed
		} else {
			nextDelay := r.CalculateNextRetryDelay()
			nextRetry := now.Add(nextDelay)
			r.NextRetryAt = &nextRetry
			r.Status = RetryStatusPending
		}
	}
}

func (r *PaymentRetry) Abandon() {
	r.Status = RetryStatusAbandoned
}

func (r *PaymentRetry) IsRetryableError(errorCode string, config *RetryConfig) bool {
	for _, nonRetryable := range config.NonRetryableErrors {
		if errorCode == nonRetryable {
			return false
		}
	}
	for _, retryable := range config.RetryableErrors {
		if errorCode == retryable {
			return true
		}
	}
	return false
}

type RetryManager struct {
	retries    map[uint64]*PaymentRetry
	mu         sync.RWMutex
	config     *RetryConfig
	logger     *slog.Logger
	repository RetryRepository
}

func NewRetryManager(config *RetryConfig, logger *slog.Logger, repo RetryRepository) *RetryManager {
	return &RetryManager{
		retries:    make(map[uint64]*PaymentRetry),
		config:     config,
		logger:     logger,
		repository: repo,
	}
}

func (m *RetryManager) CreateRetry(ctx context.Context, paymentID uint64, paymentNo string, orderID, userID uint64, amount int64, channel string) (*PaymentRetry, error) {
	retry := NewPaymentRetry(paymentID, paymentNo, orderID, userID, amount, channel, m.config)

	m.mu.Lock()
	m.retries[paymentID] = retry
	m.mu.Unlock()

	if m.repository != nil {
		if err := m.repository.Save(ctx, retry); err != nil {
			m.logger.Error("failed to save retry record", "payment_id", paymentID, "error", err)
		}
	}

	return retry, nil
}

func (m *RetryManager) GetRetry(ctx context.Context, paymentID uint64) (*PaymentRetry, error) {
	m.mu.RLock()
	retry, exists := m.retries[paymentID]
	m.mu.RUnlock()

	if exists {
		return retry, nil
	}

	if m.repository != nil {
		return m.repository.FindByPaymentID(ctx, paymentID)
	}

	return nil, errors.New("retry not found")
}

func (m *RetryManager) ExecuteRetry(ctx context.Context, paymentID uint64, executor func(channel string, amount int64) (bool, string, string, int64)) error {
	m.mu.Lock()
	retry, exists := m.retries[paymentID]
	if !exists {
		m.mu.Unlock()
		return errors.New("retry not found")
	}
	m.mu.Unlock()

	if !retry.CanRetry() || !retry.HasMoreAttempts() {
		return ErrMaxRetriesExceeded
	}

	attempt := retry.StartAttempt(retry.CurrentChannel)

	success, errorCode, errorMessage, latencyMs := executor(attempt.Channel, attempt.Amount)

	retry.CompleteAttempt(attempt, success, errorCode, errorMessage, latencyMs)

	if m.repository != nil {
		if err := m.repository.Save(ctx, retry); err != nil {
			m.logger.Error("failed to update retry record", "payment_id", paymentID, "error", err)
		}
	}

	if success {
		m.logger.Info("payment retry succeeded", "payment_id", paymentID, "attempt", attempt.AttemptNum)
		return nil
	}

	if !retry.HasMoreAttempts() {
		m.logger.Warn("payment retry exhausted", "payment_id", paymentID, "attempts", retry.AttemptCount)
		return ErrMaxRetriesExceeded
	}

	m.logger.Info("payment retry failed, will retry", "payment_id", paymentID, "attempt", attempt.AttemptNum, "next_retry", retry.NextRetryAt)
	return nil
}

func (m *RetryManager) SwitchChannel(ctx context.Context, paymentID uint64, newChannel string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	retry, exists := m.retries[paymentID]
	if !exists {
		return errors.New("retry not found")
	}

	retry.CurrentChannel = newChannel
	return nil
}

func (m *RetryManager) GetPendingRetries(ctx context.Context) ([]*PaymentRetry, error) {
	if m.repository != nil {
		return m.repository.FindPending(ctx)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var pending []*PaymentRetry
	for _, retry := range m.retries {
		if retry.Status == RetryStatusPending && retry.NextRetryAt != nil && time.Now().After(*retry.NextRetryAt) {
			pending = append(pending, retry)
		}
	}
	return pending, nil
}

type RetryRepository interface {
	Save(ctx context.Context, retry *PaymentRetry) error
	FindByID(ctx context.Context, id uint64) (*PaymentRetry, error)
	FindByPaymentID(ctx context.Context, paymentID uint64) (*PaymentRetry, error)
	FindPending(ctx context.Context) ([]*PaymentRetry, error)
	Update(ctx context.Context, retry *PaymentRetry) error
}
