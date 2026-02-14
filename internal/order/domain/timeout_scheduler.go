package domain

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type TimeoutCallback func(orderID string) error

type TimeoutTask struct {
	OrderID   string
	Timeout   time.Duration
	Callback  TimeoutCallback
	Timer     *time.Timer
	Cancelled bool
}

type RedisBasedTimeoutScheduler struct {
	logger    *slog.Logger
	tasks     map[string]*TimeoutTask
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	redis     RedisClient
	keyPrefix string
}

type RedisClient interface {
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) (int64, error)
	ZAdd(ctx context.Context, key string, members ...ZMember) (int64, error)
	ZRangeByScore(ctx context.Context, key string, opt *ZRangeBy) ([]string, error)
	ZRem(ctx context.Context, key string, members ...any) (int64, error)
}

type ZMember struct {
	Score  float64
	Member string
}

type ZRangeBy struct {
	Min    string
	Max    string
	Offset int64
	Count  int64
}

func NewRedisBasedTimeoutScheduler(logger *slog.Logger, redis RedisClient, keyPrefix string) *RedisBasedTimeoutScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &RedisBasedTimeoutScheduler{
		logger:    logger,
		tasks:     make(map[string]*TimeoutTask),
		ctx:       ctx,
		cancel:    cancel,
		redis:     redis,
		keyPrefix: keyPrefix,
	}
}

func (s *RedisBasedTimeoutScheduler) ScheduleTimeout(orderID string, timeout time.Duration, callback func(orderID string)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[orderID]; exists {
		return nil
	}

	task := &TimeoutTask{
		OrderID: orderID,
		Timeout: timeout,
		Callback: func(orderID string) error {
			callback(orderID)
			return nil
		},
		Cancelled: false,
	}

	task.Timer = time.AfterFunc(timeout, func() {
		s.executeCallback(orderID)
	})

	s.tasks[orderID] = task

	zsetKey := s.keyPrefix + ":timeout:zset"
	score := float64(time.Now().Add(timeout).Unix())
	s.redis.ZAdd(s.ctx, zsetKey, ZMember{Score: score, Member: orderID})

	s.logger.Debug("scheduled order timeout", "order_id", orderID, "timeout", timeout)
	return nil
}

func (s *RedisBasedTimeoutScheduler) executeCallback(orderID string) {
	s.mu.RLock()
	task, exists := s.tasks[orderID]
	s.mu.RUnlock()

	if !exists || task.Cancelled {
		return
	}

	if err := task.Callback(orderID); err != nil {
		s.logger.Error("timeout callback failed", "order_id", orderID, "error", err)
	}

	s.mu.Lock()
	delete(s.tasks, orderID)
	s.mu.Unlock()

	zsetKey := s.keyPrefix + ":timeout:zset"
	s.redis.ZRem(s.ctx, zsetKey, orderID)
}

func (s *RedisBasedTimeoutScheduler) CancelTimeout(orderID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[orderID]
	if !exists {
		return nil
	}

	task.Cancelled = true
	if task.Timer != nil {
		task.Timer.Stop()
	}

	delete(s.tasks, orderID)

	zsetKey := s.keyPrefix + ":timeout:zset"
	s.redis.ZRem(s.ctx, zsetKey, orderID)

	s.logger.Debug("cancelled order timeout", "order_id", orderID)
	return nil
}

func (s *RedisBasedTimeoutScheduler) Start() {
	go s.pollExpiredTimeouts()
}

func (s *RedisBasedTimeoutScheduler) pollExpiredTimeouts() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkExpiredTimeouts()
		}
	}
}

func (s *RedisBasedTimeoutScheduler) checkExpiredTimeouts() {
	zsetKey := s.keyPrefix + ":timeout:zset"
	now := float64(time.Now().Unix())

	members, err := s.redis.ZRangeByScore(s.ctx, zsetKey, &ZRangeBy{
		Min: "-inf",
		Max: string(rune(int(now))),
	})
	if err != nil {
		s.logger.Error("failed to query expired timeouts", "error", err)
		return
	}

	for _, orderID := range members {
		s.executeCallback(orderID)
	}
}

func (s *RedisBasedTimeoutScheduler) Stop() {
	s.cancel()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, task := range s.tasks {
		if task.Timer != nil {
			task.Timer.Stop()
		}
	}
	s.tasks = make(map[string]*TimeoutTask)
}

type InMemoryTimeoutScheduler struct {
	logger *slog.Logger
	tasks  map[string]*TimeoutTask
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewInMemoryTimeoutScheduler(logger *slog.Logger) *InMemoryTimeoutScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &InMemoryTimeoutScheduler{
		logger: logger,
		tasks:  make(map[string]*TimeoutTask),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *InMemoryTimeoutScheduler) ScheduleTimeout(orderID string, timeout time.Duration, callback func(orderID string)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[orderID]; exists {
		return nil
	}

	task := &TimeoutTask{
		OrderID:   orderID,
		Timeout:   timeout,
		Callback:  func(orderID string) error { callback(orderID); return nil },
		Cancelled: false,
	}

	task.Timer = time.AfterFunc(timeout, func() {
		s.executeCallback(orderID)
	})

	s.tasks[orderID] = task
	s.logger.Debug("scheduled order timeout", "order_id", orderID, "timeout", timeout)
	return nil
}

func (s *InMemoryTimeoutScheduler) executeCallback(orderID string) {
	s.mu.RLock()
	task, exists := s.tasks[orderID]
	s.mu.RUnlock()

	if !exists || task.Cancelled {
		return
	}

	if err := task.Callback(orderID); err != nil {
		s.logger.Error("timeout callback failed", "order_id", orderID, "error", err)
	}

	s.mu.Lock()
	delete(s.tasks, orderID)
	s.mu.Unlock()
}

func (s *InMemoryTimeoutScheduler) CancelTimeout(orderID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[orderID]
	if !exists {
		return nil
	}

	task.Cancelled = true
	if task.Timer != nil {
		task.Timer.Stop()
	}

	delete(s.tasks, orderID)
	s.logger.Debug("cancelled order timeout", "order_id", orderID)
	return nil
}

func (s *InMemoryTimeoutScheduler) Start() {
}

func (s *InMemoryTimeoutScheduler) Stop() {
	s.cancel()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, task := range s.tasks {
		if task.Timer != nil {
			task.Timer.Stop()
		}
	}
	s.tasks = make(map[string]*TimeoutTask)
}
