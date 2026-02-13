package domain

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

var (
	ErrConfirmationSchedulerNotRunning = errors.New("confirmation scheduler not running")
	ErrConfirmationAlreadyScheduled    = errors.New("confirmation already scheduled")
	ErrConfirmationNotFound            = errors.New("confirmation not found")
)

type ConfirmationType int8

const (
	ConfirmationTypeAutoReceive   ConfirmationType = 1
	ConfirmationTypeAutoReview    ConfirmationType = 2
	ConfirmationTypeAutoClose     ConfirmationType = 3
	ConfirmationTypePaymentTimeout ConfirmationType = 4
)

func (t ConfirmationType) String() string {
	switch t {
	case ConfirmationTypeAutoReceive:
		return "AUTO_RECEIVE"
	case ConfirmationTypeAutoReview:
		return "AUTO_REVIEW"
	case ConfirmationTypeAutoClose:
		return "AUTO_CLOSE"
	case ConfirmationTypePaymentTimeout:
		return "PAYMENT_TIMEOUT"
	default:
		return "UNKNOWN"
	}
}

type ConfirmationStatus int8

const (
	ConfirmationStatusPending    ConfirmationStatus = 1
	ConfirmationStatusReminded   ConfirmationStatus = 2
	ConfirmationStatusExecuted   ConfirmationStatus = 3
	ConfirmationStatusCancelled  ConfirmationStatus = 4
	ConfirmationStatusFailed     ConfirmationStatus = 5
)

func (s ConfirmationStatus) String() string {
	switch s {
	case ConfirmationStatusPending:
		return "PENDING"
	case ConfirmationStatusReminded:
		return "REMINDED"
	case ConfirmationStatusExecuted:
		return "EXECUTED"
	case ConfirmationStatusCancelled:
		return "CANCELLED"
	case ConfirmationStatusFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

type OrderConfirmationTask struct {
	ID               uint              `json:"id"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	TaskNo           string            `json:"task_no"`
	OrderID          uint64            `json:"order_id"`
	OrderNo          string            `json:"order_no"`
	UserID           uint64            `json:"user_id"`
	ConfirmationType ConfirmationType  `json:"confirmation_type"`
	Status           ConfirmationStatus `json:"status"`
	ScheduledAt      time.Time         `json:"scheduled_at"`
	RemindBefore     time.Duration     `json:"remind_before"`
	RemindedAt       *time.Time        `json:"reminded_at"`
	ExecutedAt       *time.Time        `json:"executed_at"`
	CancelledAt      *time.Time        `json:"cancelled_at"`
	CancelReason     string            `json:"cancel_reason"`
	Attempts         int               `json:"attempts"`
	MaxAttempts      int               `json:"max_attempts"`
	LastError        string            `json:"last_error"`
	ExtraData        map[string]any    `json:"extra_data"`
}

func NewOrderConfirmationTask(taskNo string, orderID uint64, orderNo string, userID uint64, confirmationType ConfirmationType, scheduledAt time.Time, remindBefore time.Duration) *OrderConfirmationTask {
	return &OrderConfirmationTask{
		TaskNo:           taskNo,
		OrderID:          orderID,
		OrderNo:          orderNo,
		UserID:           userID,
		ConfirmationType: confirmationType,
		Status:           ConfirmationStatusPending,
		ScheduledAt:      scheduledAt,
		RemindBefore:     remindBefore,
		Attempts:         0,
		MaxAttempts:      3,
		ExtraData:        make(map[string]any),
	}
}

func (t *OrderConfirmationTask) ShouldRemind() bool {
	if t.Status != ConfirmationStatusPending {
		return false
	}
	if t.RemindedAt != nil {
		return false
	}
	return time.Now().After(t.ScheduledAt.Add(-t.RemindBefore))
}

func (t *OrderConfirmationTask) ShouldExecute() bool {
	if t.Status != ConfirmationStatusPending && t.Status != ConfirmationStatusReminded {
		return false
	}
	return time.Now().After(t.ScheduledAt) || time.Now().Equal(t.ScheduledAt)
}

func (t *OrderConfirmationTask) MarkReminded() {
	now := time.Now()
	t.RemindedAt = &now
	t.Status = ConfirmationStatusReminded
}

func (t *OrderConfirmationTask) MarkExecuted() {
	now := time.Now()
	t.ExecutedAt = &now
	t.Status = ConfirmationStatusExecuted
}

func (t *OrderConfirmationTask) MarkCancelled(reason string) {
	now := time.Now()
	t.CancelledAt = &now
	t.CancelReason = reason
	t.Status = ConfirmationStatusCancelled
}

func (t *OrderConfirmationTask) MarkFailed(err string) {
	t.LastError = err
	t.Attempts++
	if t.Attempts >= t.MaxAttempts {
		t.Status = ConfirmationStatusFailed
	}
}

func (t *OrderConfirmationTask) CanRetry() bool {
	return t.Attempts < t.MaxAttempts && t.Status != ConfirmationStatusCancelled
}

type ConfirmationConfig struct {
	AutoReceiveDays      int           `json:"auto_receive_days"`
	AutoReviewDays       int           `json:"auto_review_days"`
	AutoCloseDays        int           `json:"auto_close_days"`
	PaymentTimeoutMinutes int          `json:"payment_timeout_minutes"`
	RemindBeforeHours    int           `json:"remind_before_hours"`
	MaxRetryAttempts     int           `json:"max_retry_attempts"`
	RetryInterval        time.Duration `json:"retry_interval"`
}

func DefaultConfirmationConfig() *ConfirmationConfig {
	return &ConfirmationConfig{
		AutoReceiveDays:      7,
		AutoReviewDays:       15,
		AutoCloseDays:        30,
		PaymentTimeoutMinutes: 30,
		RemindBeforeHours:    24,
		MaxRetryAttempts:     3,
		RetryInterval:        time.Minute * 5,
	}
}

type ConfirmationScheduler struct {
	logger     *slog.Logger
	config     *ConfirmationConfig
	tasks      map[uint64]*OrderConfirmationTask
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	running    bool
	repository ConfirmationTaskRepository
	executor   ConfirmationExecutor
	reminder   ConfirmationReminder
}

type ConfirmationTaskRepository interface {
	Save(ctx context.Context, task *OrderConfirmationTask) error
	FindByID(ctx context.Context, id uint64) (*OrderConfirmationTask, error)
	FindByOrderID(ctx context.Context, orderID uint64, confirmationType ConfirmationType) (*OrderConfirmationTask, error)
	FindPending(ctx context.Context, limit int) ([]*OrderConfirmationTask, error)
	FindDueReminders(ctx context.Context) ([]*OrderConfirmationTask, error)
	FindDueExecutions(ctx context.Context) ([]*OrderConfirmationTask, error)
	Delete(ctx context.Context, id uint64) error
}

type ConfirmationExecutor interface {
	ExecuteAutoReceive(ctx context.Context, orderID uint64) error
	ExecuteAutoReview(ctx context.Context, orderID uint64) error
	ExecuteAutoClose(ctx context.Context, orderID uint64) error
	ExecutePaymentTimeout(ctx context.Context, orderID uint64) error
}

type ConfirmationReminder interface {
	SendReceiveReminder(ctx context.Context, orderID uint64, userID uint64) error
	SendReviewReminder(ctx context.Context, orderID uint64, userID uint64) error
	SendPaymentReminder(ctx context.Context, orderID uint64, userID uint64) error
}

func NewConfirmationScheduler(logger *slog.Logger, config *ConfirmationConfig, repo ConfirmationTaskRepository, executor ConfirmationExecutor, reminder ConfirmationReminder) *ConfirmationScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &ConfirmationScheduler{
		logger:     logger,
		config:     config,
		tasks:      make(map[uint64]*OrderConfirmationTask),
		ctx:        ctx,
		cancel:     cancel,
		running:    false,
		repository: repo,
		executor:   executor,
		reminder:   reminder,
	}
}

func (s *ConfirmationScheduler) ScheduleAutoReceive(orderID uint64, orderNo string, userID uint64, shippedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[orderID]; exists {
		return ErrConfirmationAlreadyScheduled
	}

	scheduledAt := shippedAt.AddDate(0, 0, s.config.AutoReceiveDays)
	remindBefore := time.Duration(s.config.RemindBeforeHours) * time.Hour

	taskNo := generateTaskNo(ConfirmationTypeAutoReceive, orderID)
	task := NewOrderConfirmationTask(taskNo, orderID, orderNo, userID, ConfirmationTypeAutoReceive, scheduledAt, remindBefore)

	s.tasks[orderID] = task

	if s.repository != nil {
		if err := s.repository.Save(s.ctx, task); err != nil {
			s.logger.Error("failed to save confirmation task", "order_id", orderID, "error", err)
		}
	}

	s.logger.Info("scheduled auto receive confirmation", "order_id", orderID, "scheduled_at", scheduledAt)
	return nil
}

func (s *ConfirmationScheduler) ScheduleAutoReview(orderID uint64, orderNo string, userID uint64, receivedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	scheduledAt := receivedAt.AddDate(0, 0, s.config.AutoReviewDays)
	remindBefore := time.Duration(s.config.RemindBeforeHours) * time.Hour

	taskNo := generateTaskNo(ConfirmationTypeAutoReview, orderID)
	task := NewOrderConfirmationTask(taskNo, orderID, orderNo, userID, ConfirmationTypeAutoReview, scheduledAt, remindBefore)

	s.tasks[orderID] = task

	if s.repository != nil {
		if err := s.repository.Save(s.ctx, task); err != nil {
			s.logger.Error("failed to save auto review task", "order_id", orderID, "error", err)
		}
	}

	s.logger.Info("scheduled auto review confirmation", "order_id", orderID, "scheduled_at", scheduledAt)
	return nil
}

func (s *ConfirmationScheduler) SchedulePaymentTimeout(orderID uint64, orderNo string, userID uint64, createdAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	scheduledAt := createdAt.Add(time.Duration(s.config.PaymentTimeoutMinutes) * time.Minute)
	remindBefore := time.Duration(s.config.RemindBeforeHours/24) * time.Hour

	taskNo := generateTaskNo(ConfirmationTypePaymentTimeout, orderID)
	task := NewOrderConfirmationTask(taskNo, orderID, orderNo, userID, ConfirmationTypePaymentTimeout, scheduledAt, remindBefore)

	s.tasks[orderID] = task

	if s.repository != nil {
		if err := s.repository.Save(s.ctx, task); err != nil {
			s.logger.Error("failed to save payment timeout task", "order_id", orderID, "error", err)
		}
	}

	s.logger.Info("scheduled payment timeout confirmation", "order_id", orderID, "scheduled_at", scheduledAt)
	return nil
}

func (s *ConfirmationScheduler) Cancel(orderID uint64, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[orderID]
	if !exists {
		return ErrConfirmationNotFound
	}

	task.MarkCancelled(reason)

	if s.repository != nil {
		if err := s.repository.Save(s.ctx, task); err != nil {
			s.logger.Error("failed to cancel confirmation task", "order_id", orderID, "error", err)
		}
	}

	delete(s.tasks, orderID)

	s.logger.Info("cancelled confirmation task", "order_id", orderID, "reason", reason)
	return nil
}

func (s *ConfirmationScheduler) Start() {
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	go s.runReminderLoop()
	go s.runExecutionLoop()

	s.logger.Info("confirmation scheduler started")
}

func (s *ConfirmationScheduler) Stop() {
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	s.cancel()

	s.logger.Info("confirmation scheduler stopped")
}

func (s *ConfirmationScheduler) runReminderLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.processReminders()
		}
	}
}

func (s *ConfirmationScheduler) runExecutionLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.processExecutions()
		}
	}
}

func (s *ConfirmationScheduler) processReminders() {
	s.mu.RLock()
	var tasksToRemind []*OrderConfirmationTask
	for _, task := range s.tasks {
		if task.ShouldRemind() {
			tasksToRemind = append(tasksToRemind, task)
		}
	}
	s.mu.RUnlock()

	for _, task := range tasksToRemind {
		if err := s.sendReminder(task); err != nil {
			s.logger.Error("failed to send reminder", "order_id", task.OrderID, "error", err)
			continue
		}

		task.MarkReminded()

		if s.repository != nil {
			if err := s.repository.Save(s.ctx, task); err != nil {
				s.logger.Error("failed to update task after reminder", "order_id", task.OrderID, "error", err)
			}
		}
	}
}

func (s *ConfirmationScheduler) processExecutions() {
	s.mu.RLock()
	var tasksToExecute []*OrderConfirmationTask
	for _, task := range s.tasks {
		if task.ShouldExecute() {
			tasksToExecute = append(tasksToExecute, task)
		}
	}
	s.mu.RUnlock()

	for _, task := range tasksToExecute {
		if err := s.executeTask(task); err != nil {
			s.logger.Error("failed to execute confirmation task", "order_id", task.OrderID, "error", err)
			task.MarkFailed(err.Error())

			if s.repository != nil {
				s.repository.Save(s.ctx, task)
			}
			continue
		}

		task.MarkExecuted()

		if s.repository != nil {
			s.repository.Save(s.ctx, task)
		}

		s.mu.Lock()
		delete(s.tasks, task.OrderID)
		s.mu.Unlock()

		s.logger.Info("executed confirmation task", "order_id", task.OrderID, "type", task.ConfirmationType)
	}
}

func (s *ConfirmationScheduler) sendReminder(task *OrderConfirmationTask) error {
	if s.reminder == nil {
		return nil
	}

	switch task.ConfirmationType {
	case ConfirmationTypeAutoReceive:
		return s.reminder.SendReceiveReminder(s.ctx, task.OrderID, task.UserID)
	case ConfirmationTypeAutoReview:
		return s.reminder.SendReviewReminder(s.ctx, task.OrderID, task.UserID)
	case ConfirmationTypePaymentTimeout:
		return s.reminder.SendPaymentReminder(s.ctx, task.OrderID, task.UserID)
	default:
		return nil
	}
}

func (s *ConfirmationScheduler) executeTask(task *OrderConfirmationTask) error {
	if s.executor == nil {
		return errors.New("executor not configured")
	}

	switch task.ConfirmationType {
	case ConfirmationTypeAutoReceive:
		return s.executor.ExecuteAutoReceive(s.ctx, task.OrderID)
	case ConfirmationTypeAutoReview:
		return s.executor.ExecuteAutoReview(s.ctx, task.OrderID)
	case ConfirmationTypeAutoClose:
		return s.executor.ExecuteAutoClose(s.ctx, task.OrderID)
	case ConfirmationTypePaymentTimeout:
		return s.executor.ExecutePaymentTimeout(s.ctx, task.OrderID)
	default:
		return errors.New("unknown confirmation type")
	}
}

func (s *ConfirmationScheduler) LoadPendingTasks(ctx context.Context) error {
	if s.repository == nil {
		return nil
	}

	tasks, err := s.repository.FindPending(ctx, 1000)
	if err != nil {
		return err
	}

	s.mu.Lock()
	for _, task := range tasks {
		s.tasks[task.OrderID] = task
	}
	s.mu.Unlock()

	s.logger.Info("loaded pending confirmation tasks", "count", len(tasks))
	return nil
}

func generateTaskNo(confirmationType ConfirmationType, orderID uint64) string {
	return fmt.Sprintf("CF%s%010d", confirmationType.String()[:2], orderID)
}
