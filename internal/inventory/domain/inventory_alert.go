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
	ErrAlertNotFound      = errors.New("alert not found")
	ErrAlertAlreadyHandled = errors.New("alert already handled")
	ErrAlertNotActive     = errors.New("alert not active")
)

type AlertType int8

const (
	AlertTypeLowStock      AlertType = 1
	AlertTypeOutOfStock    AlertType = 2
	AlertTypeOverStock     AlertType = 3
	AlertTypeSlowMoving    AlertType = 4
	AlertTypeExpiring      AlertType = 5
	AlertTypeAbnormalSales AlertType = 6
	AlertTypePriceChange   AlertType = 7
	AlertTypeSupplierIssue AlertType = 8
)

func (t AlertType) String() string {
	switch t {
	case AlertTypeLowStock:
		return "LOW_STOCK"
	case AlertTypeOutOfStock:
		return "OUT_OF_STOCK"
	case AlertTypeOverStock:
		return "OVER_STOCK"
	case AlertTypeSlowMoving:
		return "SLOW_MOVING"
	case AlertTypeExpiring:
		return "EXPIRING"
	case AlertTypeAbnormalSales:
		return "ABNORMAL_SALES"
	case AlertTypePriceChange:
		return "PRICE_CHANGE"
	case AlertTypeSupplierIssue:
		return "SUPPLIER_ISSUE"
	default:
		return "UNKNOWN"
	}
}

type AlertSeverity int8

const (
	AlertSeverityInfo     AlertSeverity = 1
	AlertSeverityWarning  AlertSeverity = 2
	AlertSeverityCritical AlertSeverity = 3
	AlertSeverityUrgent   AlertSeverity = 4
)

func (s AlertSeverity) String() string {
	switch s {
	case AlertSeverityInfo:
		return "INFO"
	case AlertSeverityWarning:
		return "WARNING"
	case AlertSeverityCritical:
		return "CRITICAL"
	case AlertSeverityUrgent:
		return "URGENT"
	default:
		return "UNKNOWN"
	}
}

type AlertStatus int8

const (
	AlertStatusActive    AlertStatus = 1
	AlertStatusAcknowledged AlertStatus = 2
	AlertStatusResolved  AlertStatus = 3
	AlertStatusIgnored   AlertStatus = 4
	AlertStatusExpired   AlertStatus = 5
)

func (s AlertStatus) String() string {
	switch s {
	case AlertStatusActive:
		return "ACTIVE"
	case AlertStatusAcknowledged:
		return "ACKNOWLEDGED"
	case AlertStatusResolved:
		return "RESOLVED"
	case AlertStatusIgnored:
		return "IGNORED"
	case AlertStatusExpired:
		return "EXPIRED"
	default:
		return "UNKNOWN"
	}
}

type InventoryAlert struct {
	ID              uint          `json:"id"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	AlertNo         string        `json:"alert_no"`
	AlertType       AlertType     `json:"alert_type"`
	Severity        AlertSeverity `json:"severity"`
	Status          AlertStatus   `json:"status"`
	SkuID           uint64        `json:"sku_id"`
	SkuCode         string        `json:"sku_code"`
	ProductName     string        `json:"product_name"`
	WarehouseID     uint64        `json:"warehouse_id"`
	WarehouseName   string        `json:"warehouse_name"`
	CurrentStock    int32         `json:"current_stock"`
	Threshold       int32         `json:"threshold"`
	RecommendedQty  int32         `json:"recommended_qty"`
	Message         string        `json:"message"`
	Details         string        `json:"details"`
	Source          string        `json:"source"`
	SourceData      map[string]any `json:"source_data"`
	AcknowledgedBy  uint64        `json:"acknowledged_by"`
	AcknowledgedAt  *time.Time    `json:"acknowledged_at"`
	ResolvedBy      uint64        `json:"resolved_by"`
	ResolvedAt      *time.Time    `json:"resolved_at"`
	ResolutionNote  string        `json:"resolution_note"`
	IgnoredBy       uint64        `json:"ignored_by"`
	IgnoredAt       *time.Time    `json:"ignored_at"`
	IgnoreReason    string        `json:"ignore_reason"`
	ExpiresAt       *time.Time    `json:"expires_at"`
	RemindedAt      *time.Time    `json:"reminded_at"`
	ReminderCount   int           `json:"reminder_count"`
	MaxReminders    int           `json:"max_reminders"`
	NotifiedRoles   []string      `json:"notified_roles"`
	NotifiedUsers   []uint64      `json:"notified_users"`
}

type AlertRule struct {
	ID               uint          `json:"id"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
	Name             string        `json:"name"`
	Description      string        `json:"description"`
	AlertType        AlertType     `json:"alert_type"`
	Severity         AlertSeverity `json:"severity"`
	ThresholdValue   int32         `json:"threshold_value"`
	ThresholdPercent float64       `json:"threshold_percent"`
	ComparisonOp     string        `json:"comparison_op"`
	TimeWindow       time.Duration `json:"time_window"`
	CooldownPeriod   time.Duration `json:"cooldown_period"`
	Enabled          bool          `json:"enabled"`
	NotifyRoles      []string      `json:"notify_roles"`
	NotifyUsers      []uint64      `json:"notify_users"`
	AutoResolve      bool          `json:"auto_resolve"`
	MaxReminders     int           `json:"max_reminders"`
	ReminderInterval time.Duration `json:"reminder_interval"`
}

type AlertNotification struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	AlertID      uint      `json:"alert_id"`
	AlertNo      string    `json:"alert_no"`
	RecipientID  uint64    `json:"recipient_id"`
	RecipientName string   `json:"recipient_name"`
	Channel      string    `json:"channel"`
	Status       string    `json:"status"`
	SentAt       *time.Time `json:"sent_at"`
	DeliveredAt  *time.Time `json:"delivered_at"`
	ReadAt       *time.Time `json:"read_at"`
	Error        string    `json:"error"`
}

type AlertConfig struct {
	DefaultCooldown      time.Duration `json:"default_cooldown"`
	DefaultExpiresAfter  time.Duration `json:"default_expires_after"`
	MaxActiveAlerts      int           `json:"max_active_alerts"`
	EnableAutoRemind     bool          `json:"enable_auto_remind"`
	RemindInterval       time.Duration `json:"remind_interval"`
	MaxReminders         int           `json:"max_reminders"`
	EnableEscalation     bool          `json:"enable_escalation"`
	EscalationAfter      time.Duration `json:"escalation_after"`
}

func DefaultAlertConfig() *AlertConfig {
	return &AlertConfig{
		DefaultCooldown:      time.Hour * 4,
		DefaultExpiresAfter:  time.Hour * 24 * 7,
		MaxActiveAlerts:      1000,
		EnableAutoRemind:     true,
		RemindInterval:       time.Hour * 4,
		MaxReminders:         3,
		EnableEscalation:     true,
		EscalationAfter:      time.Hour * 24,
	}
}

func NewInventoryAlert(alertNo string, alertType AlertType, severity AlertSeverity, skuID uint64, skuCode, productName string, warehouseID uint64, warehouseName string, currentStock, threshold int32, message, details, source string) *InventoryAlert {
	return &InventoryAlert{
		AlertNo:       alertNo,
		AlertType:     alertType,
		Severity:      severity,
		Status:        AlertStatusActive,
		SkuID:         skuID,
		SkuCode:       skuCode,
		ProductName:   productName,
		WarehouseID:   warehouseID,
		WarehouseName: warehouseName,
		CurrentStock:  currentStock,
		Threshold:     threshold,
		Message:       message,
		Details:       details,
		Source:        source,
		SourceData:    make(map[string]any),
		ReminderCount: 0,
		MaxReminders:  3,
		NotifiedRoles: make([]string, 0),
		NotifiedUsers: make([]uint64, 0),
	}
}

func (a *InventoryAlert) SetExpiresAt(expiresAt time.Time) {
	a.ExpiresAt = &expiresAt
}

func (a *InventoryAlert) SetRecommendedQty(qty int32) {
	a.RecommendedQty = qty
}

func (a *InventoryAlert) SetSourceData(data map[string]any) {
	a.SourceData = data
}

func (a *InventoryAlert) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

func (a *InventoryAlert) IsActive() bool {
	return a.Status == AlertStatusActive
}

func (a *InventoryAlert) CanRemind() bool {
	return a.Status == AlertStatusActive && a.ReminderCount < a.MaxReminders
}

func (a *InventoryAlert) Acknowledge(userID uint64) error {
	if a.Status != AlertStatusActive {
		return ErrAlertNotActive
	}

	now := time.Now()
	a.Status = AlertStatusAcknowledged
	a.AcknowledgedBy = userID
	a.AcknowledgedAt = &now

	return nil
}

func (a *InventoryAlert) Resolve(userID uint64, note string) error {
	if a.Status == AlertStatusResolved || a.Status == AlertStatusIgnored {
		return ErrAlertAlreadyHandled
	}

	now := time.Now()
	a.Status = AlertStatusResolved
	a.ResolvedBy = userID
	a.ResolvedAt = &now
	a.ResolutionNote = note

	return nil
}

func (a *InventoryAlert) Ignore(userID uint64, reason string) error {
	if a.Status == AlertStatusResolved || a.Status == AlertStatusIgnored {
		return ErrAlertAlreadyHandled
	}

	now := time.Now()
	a.Status = AlertStatusIgnored
	a.IgnoredBy = userID
	a.IgnoredAt = &now
	a.IgnoreReason = reason

	return nil
}

func (a *InventoryAlert) MarkReminded() {
	now := time.Now()
	a.RemindedAt = &now
	a.ReminderCount++
}

func (a *InventoryAlert) AddNotifiedRole(role string) {
	a.NotifiedRoles = append(a.NotifiedRoles, role)
}

func (a *InventoryAlert) AddNotifiedUser(userID uint64) {
	a.NotifiedUsers = append(a.NotifiedUsers, userID)
}

type AlertNotifier interface {
	Notify(ctx context.Context, alert *InventoryAlert, recipients []uint64) error
	NotifyByRole(ctx context.Context, alert *InventoryAlert, roles []string) error
	NotifyAll(ctx context.Context, alert *InventoryAlert) error
}

type AlertMonitor struct {
	logger     *slog.Logger
	config     *AlertConfig
	rules      []*AlertRule
	alerts     map[uint64]*InventoryAlert
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	running    bool
	repository AlertRepository
	notifier   AlertNotifier
}

func NewAlertMonitor(logger *slog.Logger, config *AlertConfig, repo AlertRepository, notifier AlertNotifier) *AlertMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &AlertMonitor{
		logger:     logger,
		config:     config,
		rules:      make([]*AlertRule, 0),
		alerts:     make(map[uint64]*InventoryAlert),
		ctx:        ctx,
		cancel:     cancel,
		running:    false,
		repository: repo,
		notifier:   notifier,
	}
}

func (m *AlertMonitor) AddRule(rule *AlertRule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rules = append(m.rules, rule)
}

func (m *AlertMonitor) CheckInventory(ctx context.Context, inv *Inventory, skuCode, productName, warehouseName string) ([]*InventoryAlert, error) {
	var alerts []*InventoryAlert

	m.mu.RLock()
	rules := m.rules
	m.mu.RUnlock()

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		if alert := m.evaluateRule(rule, inv, skuCode, productName, warehouseName); alert != nil {
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

func (m *AlertMonitor) evaluateRule(rule *AlertRule, inv *Inventory, skuCode, productName, warehouseName string) *InventoryAlert {
	var shouldAlert bool
	var message string

	switch rule.AlertType {
	case AlertTypeLowStock:
		shouldAlert, message = m.checkLowStock(rule, inv, skuCode)
	case AlertTypeOutOfStock:
		shouldAlert, message = m.checkOutOfStock(rule, inv, skuCode)
	case AlertTypeOverStock:
		shouldAlert, message = m.checkOverStock(rule, inv, skuCode)
	default:
		return nil
	}

	if !shouldAlert {
		return nil
	}

	alertNo := fmt.Sprintf("ALT%s%04d", time.Now().Format("20060102150405"), inv.SkuID%10000)
	alert := NewInventoryAlert(
		alertNo,
		rule.AlertType,
		rule.Severity,
		inv.SkuID,
		skuCode,
		productName,
		inv.WarehouseID,
		warehouseName,
		inv.AvailableStock,
		rule.ThresholdValue,
		message,
		"",
		"INVENTORY_MONITOR",
	)

	if m.config.DefaultExpiresAfter > 0 {
		alert.SetExpiresAt(time.Now().Add(m.config.DefaultExpiresAfter))
	}

	return alert
}

func (m *AlertMonitor) checkLowStock(rule *AlertRule, inv *Inventory, skuCode string) (bool, string) {
	threshold := rule.ThresholdValue
	if threshold == 0 {
		threshold = inv.WarningThreshold
	}

	if inv.AvailableStock <= threshold {
		message := fmt.Sprintf("库存预警: SKU %s 当前库存 %d，低于阈值 %d", skuCode, inv.AvailableStock, threshold)
		return true, message
	}

	return false, ""
}

func (m *AlertMonitor) checkOutOfStock(rule *AlertRule, inv *Inventory, skuCode string) (bool, string) {
	if inv.AvailableStock == 0 {
		message := fmt.Sprintf("缺货预警: SKU %s 已无库存", skuCode)
		return true, message
	}

	return false, ""
}

func (m *AlertMonitor) checkOverStock(rule *AlertRule, inv *Inventory, skuCode string) (bool, string) {
	if rule.ThresholdValue > 0 && inv.AvailableStock > rule.ThresholdValue {
		message := fmt.Sprintf("库存积压预警: SKU %s 当前库存 %d，超过阈值 %d", skuCode, inv.AvailableStock, rule.ThresholdValue)
		return true, message
	}

	return false, ""
}

func (m *AlertMonitor) CreateAlert(ctx context.Context, alert *InventoryAlert) error {
	m.mu.Lock()
	m.alerts[alert.SkuID] = alert
	m.mu.Unlock()

	if m.repository != nil {
		if err := m.repository.Save(ctx, alert); err != nil {
			m.logger.Error("failed to save alert", "sku_id", alert.SkuID, "error", err)
			return err
		}
	}

	if m.notifier != nil {
		if err := m.notifier.NotifyAll(ctx, alert); err != nil {
			m.logger.Error("failed to send alert notification", "alert_no", alert.AlertNo, "error", err)
		}
	}

	m.logger.Info("created inventory alert", "alert_no", alert.AlertNo, "type", alert.AlertType, "severity", alert.Severity)
	return nil
}

func (m *AlertMonitor) AcknowledgeAlert(ctx context.Context, alertID uint64, userID uint64) error {
	m.mu.Lock()
	alert, exists := m.alerts[alertID]
	m.mu.Unlock()

	if !exists {
		if m.repository != nil {
			var err error
			alert, err = m.repository.FindByID(ctx, alertID)
			if err != nil {
				return ErrAlertNotFound
			}
		} else {
			return ErrAlertNotFound
		}
	}

	if err := alert.Acknowledge(userID); err != nil {
		return err
	}

	if m.repository != nil {
		return m.repository.Update(ctx, alert)
	}

	return nil
}

func (m *AlertMonitor) ResolveAlert(ctx context.Context, alertID uint64, userID uint64, note string) error {
	m.mu.Lock()
	alert, exists := m.alerts[alertID]
	if exists {
		delete(m.alerts, alertID)
	}
	m.mu.Unlock()

	if !exists {
		if m.repository != nil {
			var err error
			alert, err = m.repository.FindByID(ctx, alertID)
			if err != nil {
				return ErrAlertNotFound
			}
		} else {
			return ErrAlertNotFound
		}
	}

	if err := alert.Resolve(userID, note); err != nil {
		return err
	}

	if m.repository != nil {
		return m.repository.Update(ctx, alert)
	}

	return nil
}

func (m *AlertMonitor) Start() {
	m.mu.Lock()
	m.running = true
	m.mu.Unlock()

	go m.runReminderLoop()
	go m.runExpirationCheck()

	m.logger.Info("alert monitor started")
}

func (m *AlertMonitor) Stop() {
	m.mu.Lock()
	m.running = false
	m.mu.Unlock()

	m.cancel()

	m.logger.Info("alert monitor stopped")
}

func (m *AlertMonitor) runReminderLoop() {
	if !m.config.EnableAutoRemind {
		return
	}

	ticker := time.NewTicker(m.config.RemindInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.processReminders()
		}
	}
}

func (m *AlertMonitor) runExpirationCheck() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.processExpirations()
		}
	}
}

func (m *AlertMonitor) processReminders() {
	m.mu.RLock()
	var alertsToRemind []*InventoryAlert
	for _, alert := range m.alerts {
		if alert.CanRemind() {
			alertsToRemind = append(alertsToRemind, alert)
		}
	}
	m.mu.RUnlock()

	for _, alert := range alertsToRemind {
		if m.notifier != nil {
			if err := m.notifier.NotifyAll(m.ctx, alert); err != nil {
				m.logger.Error("failed to send alert reminder", "alert_no", alert.AlertNo, "error", err)
				continue
			}
		}

		alert.MarkReminded()

		if m.repository != nil {
			m.repository.Update(m.ctx, alert)
		}
	}
}

func (m *AlertMonitor) processExpirations() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for skuID, alert := range m.alerts {
		if alert.IsExpired() {
			alert.Status = AlertStatusExpired
			delete(m.alerts, skuID)

			if m.repository != nil {
				m.repository.Update(m.ctx, alert)
			}
		}
	}
}

func (m *AlertMonitor) LoadActiveAlerts(ctx context.Context) error {
	if m.repository == nil {
		return nil
	}

	alerts, err := m.repository.FindActive(ctx, m.config.MaxActiveAlerts)
	if err != nil {
		return err
	}

	m.mu.Lock()
	for _, alert := range alerts {
		m.alerts[alert.SkuID] = alert
	}
	m.mu.Unlock()

	m.logger.Info("loaded active alerts", "count", len(alerts))
	return nil
}

type AlertRepository interface {
	Save(ctx context.Context, alert *InventoryAlert) error
	FindByID(ctx context.Context, id uint64) (*InventoryAlert, error)
	FindByAlertNo(ctx context.Context, alertNo string) (*InventoryAlert, error)
	FindBySkuID(ctx context.Context, skuID uint64) ([]*InventoryAlert, error)
	FindActive(ctx context.Context, limit int) ([]*InventoryAlert, error)
	FindByType(ctx context.Context, alertType AlertType, limit, offset int) ([]*InventoryAlert, error)
	FindByStatus(ctx context.Context, status AlertStatus, limit, offset int) ([]*InventoryAlert, error)
	Update(ctx context.Context, alert *InventoryAlert) error
}

type AlertRuleRepository interface {
	FindByID(ctx context.Context, id uint64) (*AlertRule, error)
	FindByType(ctx context.Context, alertType AlertType) ([]*AlertRule, error)
	FindEnabled(ctx context.Context) ([]*AlertRule, error)
	Save(ctx context.Context, rule *AlertRule) error
}
