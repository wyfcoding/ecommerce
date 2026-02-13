// 变更说明：新增自动扣款记录和续费提醒的MySQL持久化实现
package mysql

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/subscription/domain"
	"gorm.io/gorm"
)

// AutoDebitRecordModel 自动扣款记录模型
type AutoDebitRecordModel struct {
	gorm.Model
	RecordID       string                  `gorm:"column:record_id;type:varchar(64);uniqueIndex;not null;comment:扣款记录ID"`
	SubscriptionID uint                    `gorm:"column:subscription_id;not null;index;comment:订阅ID"`
	Amount         uint64                  `gorm:"column:amount;not null;comment:扣款金额(分)"`
	AttemptCount   int                     `gorm:"column:attempt_count;not null;default:0;comment:尝试次数"`
	LastAttempt    time.Time               `gorm:"column:last_attempt;not null;comment:最后尝试时间"`
	Status         domain.DebitStatus      `gorm:"column:status;type:varchar(20);not null;comment:扣款状态"`
	ErrorMessage   string                  `gorm:"column:error_message;type:varchar(512);comment:错误信息"`
}

// RenewalReminderModel 续费提醒模型
type RenewalReminderModel struct {
	gorm.Model
	SubscriptionID uint                    `gorm:"column:subscription_id;not null;index;comment:订阅ID"`
	UserID         uint64                  `gorm:"column:user_id;not null;index;comment:用户ID"`
	PlanName       string                  `gorm:"column:plan_name;type:varchar(128);comment:计划名称"`
	EndDate        time.Time               `gorm:"column:end_date;not null;comment:结束时间"`
	DaysRemaining  int                     `gorm:"column:days_remaining;not null;comment:剩余天数"`
	ReminderType   domain.ReminderType     `gorm:"column:reminder_type;type:varchar(20);not null;comment:提醒类型"`
	SentAt         time.Time               `gorm:"column:sent_at;not null;comment:发送时间"`
}

// ReminderSentLogModel 提醒发送日志模型
type ReminderSentLogModel struct {
	gorm.Model
	SubscriptionID uint                `gorm:"column:subscription_id;not null;uniqueIndex:idx_sub_type;comment:订阅ID"`
	ReminderType   domain.ReminderType `gorm:"column:reminder_type;type:varchar(20);not null;uniqueIndex:idx_sub_type;comment:提醒类型"`
	SentAt         time.Time           `gorm:"column:sent_at;not null;comment:发送时间"`
}

func (AutoDebitRecordModel) TableName() string    { return "auto_debit_records" }
func (RenewalReminderModel) TableName() string    { return "renewal_reminders" }
func (ReminderSentLogModel) TableName() string    { return "reminder_sent_logs" }

// AutoDebitRepository 自动扣款仓储实现
type AutoDebitRepository struct {
	db *gorm.DB
}

// NewAutoDebitRepository 创建自动扣款仓储
func NewAutoDebitRepository(db *gorm.DB) *AutoDebitRepository {
	return &AutoDebitRepository{db: db}
}

func (r *AutoDebitRepository) Save(ctx context.Context, record *domain.AutoDebitRecord) error {
	model := toAutoDebitRecordModel(record)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *AutoDebitRepository) Get(ctx context.Context, recordID string) (*domain.AutoDebitRecord, error) {
	var model AutoDebitRecordModel
	if err := r.db.WithContext(ctx).Where("record_id = ?", recordID).First(&model).Error; err != nil {
		return nil, err
	}
	return toAutoDebitRecord(&model), nil
}

func (r *AutoDebitRepository) GetBySubscriptionID(ctx context.Context, subscriptionID uint, limit int) ([]*domain.AutoDebitRecord, error) {
	var models []AutoDebitRecordModel
	if err := r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&models).Error; err != nil {
		return nil, err
	}
	
	records := make([]*domain.AutoDebitRecord, len(models))
	for i, m := range models {
		records[i] = toAutoDebitRecord(&m)
	}
	return records, nil
}

func (r *AutoDebitRepository) GetPendingRecords(ctx context.Context, limit int) ([]*domain.AutoDebitRecord, error) {
	var models []AutoDebitRecordModel
	if err := r.db.WithContext(ctx).
		Where("status = ?", domain.DebitStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&models).Error; err != nil {
		return nil, err
	}
	
	records := make([]*domain.AutoDebitRecord, len(models))
	for i, m := range models {
		records[i] = toAutoDebitRecord(&m)
	}
	return records, nil
}

func (r *AutoDebitRepository) UpdateStatus(ctx context.Context, recordID string, status domain.DebitStatus, errorMsg string) error {
	return r.db.WithContext(ctx).
		Model(&AutoDebitRecordModel{}).
		Where("record_id = ?", recordID).
		Updates(map[string]interface{}{
			"status":       status,
			"error_message": errorMsg,
		}).Error
}

// RenewalReminderRepository 续费提醒仓储实现
type RenewalReminderRepository struct {
	db *gorm.DB
}

// NewRenewalReminderRepository 创建续费提醒仓储
func NewRenewalReminderRepository(db *gorm.DB) *RenewalReminderRepository {
	return &RenewalReminderRepository{db: db}
}

func (r *RenewalReminderRepository) Save(ctx context.Context, reminder *domain.RenewalReminder) error {
	model := toRenewalReminderModel(reminder)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *RenewalReminderRepository) GetSentReminders(ctx context.Context, subscriptionID uint, reminderType domain.ReminderType) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ReminderSentLogModel{}).
		Where("subscription_id = ? AND reminder_type = ?", subscriptionID, reminderType).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *RenewalReminderRepository) MarkAsSent(ctx context.Context, subscriptionID uint, reminderType domain.ReminderType) error {
	log := &ReminderSentLogModel{
		SubscriptionID: subscriptionID,
		ReminderType:   reminderType,
		SentAt:         time.Now(),
	}
	return r.db.WithContext(ctx).Create(log).Error
}

// 转换函数
func toAutoDebitRecordModel(record *domain.AutoDebitRecord) *AutoDebitRecordModel {
	if record == nil {
		return nil
	}
	return &AutoDebitRecordModel{
		RecordID:       record.RecordID,
		SubscriptionID: record.SubscriptionID,
		Amount:         record.Amount,
		AttemptCount:   record.AttemptCount,
		LastAttempt:    record.LastAttempt,
		Status:         record.Status,
		ErrorMessage:   record.ErrorMessage,
	}
}

func toAutoDebitRecord(model *AutoDebitRecordModel) *domain.AutoDebitRecord {
	if model == nil {
		return nil
	}
	return &domain.AutoDebitRecord{
		RecordID:       model.RecordID,
		SubscriptionID: model.SubscriptionID,
		Amount:         model.Amount,
		AttemptCount:   model.AttemptCount,
		LastAttempt:    model.LastAttempt,
		Status:         model.Status,
		ErrorMessage:   model.ErrorMessage,
	}
}

func toRenewalReminderModel(reminder *domain.RenewalReminder) *RenewalReminderModel {
	if reminder == nil {
		return nil
	}
	return &RenewalReminderModel{
		SubscriptionID: reminder.SubscriptionID,
		UserID:         reminder.UserID,
		PlanName:       reminder.PlanName,
		EndDate:        reminder.EndDate,
		DaysRemaining:  reminder.DaysRemaining,
		ReminderType:   reminder.ReminderType,
		SentAt:         time.Now(),
	}
}

func toRenewalReminder(model *RenewalReminderModel) *domain.RenewalReminder {
	if model == nil {
		return nil
	}
	return &domain.RenewalReminder{
		SubscriptionID: model.SubscriptionID,
		UserID:         model.UserID,
		PlanName:       model.PlanName,
		EndDate:        model.EndDate,
		DaysRemaining:  model.DaysRemaining,
		ReminderType:   model.ReminderType,
	}
}
