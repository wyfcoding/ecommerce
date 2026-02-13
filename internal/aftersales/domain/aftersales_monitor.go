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
	ErrTimeoutAlertNotFound   = errors.New("timeout alert not found")
	ErrTimeoutAlertHandled    = errors.New("timeout alert already handled")
	ErrInvalidMonitorConfig   = errors.New("invalid monitor config")
)

type TimeoutAlertType int8

const (
	TimeoutAlertTypeMerchantResponse TimeoutAlertType = 1
	TimeoutAlertTypeUserReply        TimeoutAlertType = 2
	TimeoutAlertTypePlatformHandle   TimeoutAlertType = 3
	TimeoutAlertTypeReturnReceive    TimeoutAlertType = 4
	TimeoutAlertTypeRefundProcess    TimeoutAlertType = 5
	TimeoutAlertTypeArbitration      TimeoutAlertType = 6
)

func (t TimeoutAlertType) String() string {
	switch t {
	case TimeoutAlertTypeMerchantResponse:
		return "MERCHANT_RESPONSE"
	case TimeoutAlertTypeUserReply:
		return "USER_REPLY"
	case TimeoutAlertTypePlatformHandle:
		return "PLATFORM_HANDLE"
	case TimeoutAlertTypeReturnReceive:
		return "RETURN_RECEIVE"
	case TimeoutAlertTypeRefundProcess:
		return "REFUND_PROCESS"
	case TimeoutAlertTypeArbitration:
		return "ARBITRATION"
	default:
		return "UNKNOWN"
	}
}

type TimeoutAlertSeverity int8

const (
	TimeoutSeverityWarning  TimeoutAlertSeverity = 1
	TimeoutSeverityCritical TimeoutAlertSeverity = 2
	TimeoutSeverityUrgent   TimeoutAlertSeverity = 3
)

func (s TimeoutAlertSeverity) String() string {
	switch s {
	case TimeoutSeverityWarning:
		return "WARNING"
	case TimeoutSeverityCritical:
		return "CRITICAL"
	case TimeoutSeverityUrgent:
		return "URGENT"
	default:
		return "UNKNOWN"
	}
}

type TimeoutAlertStatus int8

const (
	TimeoutAlertStatusPending    TimeoutAlertStatus = 1
	TimeoutAlertStatusNotified   TimeoutAlertStatus = 2
	TimeoutAlertStatusHandled    TimeoutAlertStatus = 3
	TimeoutAlertStatusEscalated  TimeoutAlertStatus = 4
	TimeoutAlertStatusClosed     TimeoutAlertStatus = 5
)

func (s TimeoutAlertStatus) String() string {
	switch s {
	case TimeoutAlertStatusPending:
		return "PENDING"
	case TimeoutAlertStatusNotified:
		return "NOTIFIED"
	case TimeoutAlertStatusHandled:
		return "HANDLED"
	case TimeoutAlertStatusEscalated:
		return "ESCALATED"
	case TimeoutAlertStatusClosed:
		return "CLOSED"
	default:
		return "UNKNOWN"
	}
}

type MonitorConfig struct {
	MerchantResponseHours int `json:"merchant_response_hours"`
	UserReplyHours        int `json:"user_reply_hours"`
	PlatformHandleHours   int `json:"platform_handle_hours"`
	ReturnReceiveDays     int `json:"return_receive_days"`
	RefundProcessHours    int `json:"refund_process_hours"`
	ArbitrationDays       int `json:"arbitration_days"`
	WarningBeforeHours    int `json:"warning_before_hours"`
	RemindIntervalHours   int `json:"remind_interval_hours"`
	MaxRemindCount        int `json:"max_remind_count"`
}

func DefaultMonitorConfig() *MonitorConfig {
	return &MonitorConfig{
		MerchantResponseHours: 48,
		UserReplyHours:        72,
		PlatformHandleHours:   24,
		ReturnReceiveDays:     7,
		RefundProcessHours:    48,
		ArbitrationDays:       7,
		WarningBeforeHours:    4,
		RemindIntervalHours:   12,
		MaxRemindCount:        3,
	}
}

type TimeoutAlert struct {
	ID             uint                `json:"id"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	AlertNo        string              `json:"alert_no"`
	AftersalesID   uint64              `json:"aftersales_id"`
	AftersalesNo   string              `json:"aftersales_no"`
	AlertType      TimeoutAlertType    `json:"alert_type"`
	Severity       TimeoutAlertSeverity `json:"severity"`
	Status         TimeoutAlertStatus  `json:"status"`
	Deadline       time.Time           `json:"deadline"`
	WarningAt      *time.Time          `json:"warning_at"`
	NotifiedAt     *time.Time          `json:"notified_at"`
	HandledAt      *time.Time          `json:"handled_at"`
	EscalatedAt    *time.Time          `json:"escalated_at"`
	ClosedAt       *time.Time          `json:"closed_at"`
	HandlerID      uint64              `json:"handler_id"`
	HandlerName    string              `json:"handler_name"`
	HandleNote     string              `json:"handle_note"`
	RemindCount    int                 `json:"remind_count"`
	MaxRemindCount int                 `json:"max_remind_count"`
	LastRemindAt   *time.Time          `json:"last_remind_at"`
	NotifyTargets  []string            `json:"notify_targets"`
}

type AftersalesMonitor struct {
	logger     *slog.Logger
	config     *MonitorConfig
	alerts     map[uint64]*TimeoutAlert
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	running    bool
	repository TimeoutAlertRepository
	notifier   TimeoutNotifier
}

type TimeoutNotifier interface {
	NotifyTimeout(ctx context.Context, alert *TimeoutAlert) error
	NotifyWarning(ctx context.Context, alert *TimeoutAlert) error
	NotifyEscalation(ctx context.Context, alert *TimeoutAlert) error
}

func NewAftersalesMonitor(logger *slog.Logger, config *MonitorConfig, repo TimeoutAlertRepository, notifier TimeoutNotifier) *AftersalesMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &AftersalesMonitor{
		logger:     logger,
		config:     config,
		alerts:     make(map[uint64]*TimeoutAlert),
		ctx:        ctx,
		cancel:     cancel,
		running:    false,
		repository: repo,
		notifier:   notifier,
	}
}

func NewTimeoutAlert(alertNo string, aftersalesID uint64, aftersalesNo string, alertType TimeoutAlertType, deadline time.Time) *TimeoutAlert {
	return &TimeoutAlert{
		AlertNo:        alertNo,
		AftersalesID:   aftersalesID,
		AftersalesNo:   aftersalesNo,
		AlertType:      alertType,
		Severity:       TimeoutSeverityWarning,
		Status:         TimeoutAlertStatusPending,
		Deadline:       deadline,
		RemindCount:    0,
		MaxRemindCount: 3,
		NotifyTargets:  make([]string, 0),
	}
}

func (a *TimeoutAlert) IsTimeout() bool {
	return time.Now().After(a.Deadline)
}

func (a *TimeoutAlert) IsWarning() bool {
	if a.WarningAt == nil {
		return false
	}
	return time.Now().After(*a.WarningAt) && !a.IsTimeout()
}

func (a *TimeoutAlert) CanRemind() bool {
	return a.RemindCount < a.MaxRemindCount && a.Status != TimeoutAlertStatusHandled
}

func (a *TimeoutAlert) MarkNotified() {
	now := time.Now()
	a.NotifiedAt = &now
	a.Status = TimeoutAlertStatusNotified
}

func (a *TimeoutAlert) MarkHandled(handlerID uint64, handlerName, note string) {
	now := time.Now()
	a.Status = TimeoutAlertStatusHandled
	a.HandledAt = &now
	a.HandlerID = handlerID
	a.HandlerName = handlerName
	a.HandleNote = note
}

func (a *TimeoutAlert) MarkEscalated() {
	now := time.Now()
	a.Status = TimeoutAlertStatusEscalated
	a.EscalatedAt = &now
	a.Severity = TimeoutSeverityUrgent
}

func (a *TimeoutAlert) MarkClosed() {
	now := time.Now()
	a.Status = TimeoutAlertStatusClosed
	a.ClosedAt = &now
}

func (a *TimeoutAlert) RecordRemind() {
	now := time.Now()
	a.RemindCount++
	a.LastRemindAt = &now
}

func (m *AftersalesMonitor) StartMonitoring(aftersalesID uint64, aftersalesNo string, alertType TimeoutAlertType, startTime time.Time) *TimeoutAlert {
	var deadline time.Time
	switch alertType {
	case TimeoutAlertTypeMerchantResponse:
		deadline = startTime.Add(time.Duration(m.config.MerchantResponseHours) * time.Hour)
	case TimeoutAlertTypeUserReply:
		deadline = startTime.Add(time.Duration(m.config.UserReplyHours) * time.Hour)
	case TimeoutAlertTypePlatformHandle:
		deadline = startTime.Add(time.Duration(m.config.PlatformHandleHours) * time.Hour)
	case TimeoutAlertTypeReturnReceive:
		deadline = startTime.AddDate(0, 0, m.config.ReturnReceiveDays)
	case TimeoutAlertTypeRefundProcess:
		deadline = startTime.Add(time.Duration(m.config.RefundProcessHours) * time.Hour)
	case TimeoutAlertTypeArbitration:
		deadline = startTime.AddDate(0, 0, m.config.ArbitrationDays)
	default:
		deadline = startTime.Add(time.Duration(m.config.PlatformHandleHours) * time.Hour)
	}

	alertNo := generateAlertNo(alertType, aftersalesID)
	alert := NewTimeoutAlert(alertNo, aftersalesID, aftersalesNo, alertType, deadline)

	warningAt := deadline.Add(-time.Duration(m.config.WarningBeforeHours) * time.Hour)
	alert.WarningAt = &warningAt

	m.mu.Lock()
	m.alerts[aftersalesID] = alert
	m.mu.Unlock()

	if m.repository != nil {
		m.repository.SaveAlert(m.ctx, alert)
	}

	m.logger.Info("started monitoring aftersales timeout", "aftersales_id", aftersalesID, "alert_type", alertType, "deadline", deadline)
	return alert
}

func (m *AftersalesMonitor) StopMonitoring(aftersalesID uint64) {
	m.mu.Lock()
	delete(m.alerts, aftersalesID)
	m.mu.Unlock()

	m.logger.Info("stopped monitoring aftersales timeout", "aftersales_id", aftersalesID)
}

func (m *AftersalesMonitor) HandleAlert(aftersalesID uint64, handlerID uint64, handlerName, note string) error {
	m.mu.Lock()
	alert, exists := m.alerts[aftersalesID]
	m.mu.Unlock()

	if !exists {
		return ErrTimeoutAlertNotFound
	}

	alert.MarkHandled(handlerID, handlerName, note)

	if m.repository != nil {
		m.repository.UpdateAlert(m.ctx, alert)
	}

	m.mu.Lock()
	delete(m.alerts, aftersalesID)
	m.mu.Unlock()

	m.logger.Info("handled timeout alert", "aftersales_id", aftersalesID, "handler", handlerName)
	return nil
}

func (m *AftersalesMonitor) Start() {
	m.mu.Lock()
	m.running = true
	m.mu.Unlock()

	go m.runMonitorLoop()
	go m.runReminderLoop()

	m.logger.Info("aftersales monitor started")
}

func (m *AftersalesMonitor) Stop() {
	m.mu.Lock()
	m.running = false
	m.mu.Unlock()

	m.cancel()

	m.logger.Info("aftersales monitor stopped")
}

func (m *AftersalesMonitor) runMonitorLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkTimeouts()
		}
	}
}

func (m *AftersalesMonitor) runReminderLoop() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.sendReminders()
		}
	}
}

func (m *AftersalesMonitor) checkTimeouts() {
	m.mu.RLock()
	var timeoutAlerts []*TimeoutAlert
	var warningAlerts []*TimeoutAlert

	for _, alert := range m.alerts {
		if alert.IsTimeout() && alert.Status == TimeoutAlertStatusPending {
			timeoutAlerts = append(timeoutAlerts, alert)
		} else if alert.IsWarning() && alert.Status == TimeoutAlertStatusPending {
			warningAlerts = append(warningAlerts, alert)
		}
	}
	m.mu.RUnlock()

	for _, alert := range timeoutAlerts {
		alert.Severity = TimeoutSeverityCritical
		if m.notifier != nil {
			m.notifier.NotifyTimeout(m.ctx, alert)
		}
		alert.MarkNotified()

		if m.repository != nil {
			m.repository.UpdateAlert(m.ctx, alert)
		}

		m.logger.Warn("aftersales timeout detected", "aftersales_id", alert.AftersalesID, "alert_type", alert.AlertType)
	}

	for _, alert := range warningAlerts {
		if m.notifier != nil {
			m.notifier.NotifyWarning(m.ctx, alert)
		}

		m.logger.Info("aftersales timeout warning", "aftersales_id", alert.AftersalesID, "alert_type", alert.AlertType)
	}
}

func (m *AftersalesMonitor) sendReminders() {
	m.mu.RLock()
	var alertsToRemind []*TimeoutAlert
	for _, alert := range m.alerts {
		if alert.CanRemind() && (alert.IsTimeout() || alert.IsWarning()) {
			alertsToRemind = append(alertsToRemind, alert)
		}
	}
	m.mu.RUnlock()

	for _, alert := range alertsToRemind {
		if m.notifier != nil {
			m.notifier.NotifyTimeout(m.ctx, alert)
		}
		alert.RecordRemind()

		if m.repository != nil {
			m.repository.UpdateAlert(m.ctx, alert)
		}

		m.logger.Info("sent timeout reminder", "aftersales_id", alert.AftersalesID, "remind_count", alert.RemindCount)
	}
}

func (m *AftersalesMonitor) LoadActiveAlerts(ctx context.Context) error {
	if m.repository == nil {
		return nil
	}

	alerts, err := m.repository.FindActiveAlerts(ctx, 1000)
	if err != nil {
		return err
	}

	m.mu.Lock()
	for _, alert := range alerts {
		m.alerts[alert.AftersalesID] = alert
	}
	m.mu.Unlock()

	m.logger.Info("loaded active timeout alerts", "count", len(alerts))
	return nil
}

func generateAlertNo(alertType TimeoutAlertType, aftersalesID uint64) string {
	return fmt.Sprintf("TO%s%010d", alertType.String()[:2], aftersalesID)
}

type TimeoutAlertRepository interface {
	SaveAlert(ctx context.Context, alert *TimeoutAlert) error
	UpdateAlert(ctx context.Context, alert *TimeoutAlert) error
	FindAlertByID(ctx context.Context, id uint64) (*TimeoutAlert, error)
	FindAlertByAftersalesID(ctx context.Context, aftersalesID uint64) (*TimeoutAlert, error)
	FindActiveAlerts(ctx context.Context, limit int) ([]*TimeoutAlert, error)
	FindTimeoutAlerts(ctx context.Context) ([]*TimeoutAlert, error)
	DeleteAlert(ctx context.Context, id uint64) error
}
