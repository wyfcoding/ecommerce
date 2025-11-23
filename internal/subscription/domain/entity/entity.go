package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// SubscriptionStatus 订阅状态
type SubscriptionStatus int8

const (
	SubscriptionStatusActive   SubscriptionStatus = 1 // 激活
	SubscriptionStatusExpired  SubscriptionStatus = 2 // 过期
	SubscriptionStatusCanceled SubscriptionStatus = 3 // 取消
	SubscriptionStatusPaused   SubscriptionStatus = 4 // 暂停
)

// StringArray defines a slice of strings that implements sql.Scanner and driver.Valuer
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a)
}

// SubscriptionPlan 订阅计划实体
type SubscriptionPlan struct {
	gorm.Model
	Name        string      `gorm:"type:varchar(128);not null;comment:计划名称" json:"name"`
	Description string      `gorm:"type:varchar(255);comment:描述" json:"description"`
	Price       uint64      `gorm:"not null;comment:价格(分)" json:"price"`
	Duration    int32       `gorm:"not null;comment:时长(天)" json:"duration"`
	Features    StringArray `gorm:"type:json;comment:特性列表" json:"features"`
	Enabled     bool        `gorm:"not null;default:true;comment:是否启用" json:"enabled"`
}

// Subscription 订阅实体
type Subscription struct {
	gorm.Model
	UserID     uint64             `gorm:"index;not null;comment:用户ID" json:"user_id"`
	PlanID     uint64             `gorm:"index;not null;comment:计划ID" json:"plan_id"`
	Status     SubscriptionStatus `gorm:"type:tinyint;not null;default:1;comment:状态" json:"status"`
	StartDate  time.Time          `gorm:"not null;comment:开始时间" json:"start_date"`
	EndDate    time.Time          `gorm:"not null;comment:结束时间" json:"end_date"`
	AutoRenew  bool               `gorm:"not null;default:true;comment:自动续订" json:"auto_renew"`
	CanceledAt *time.Time         `gorm:"comment:取消时间" json:"canceled_at"`
}

// IsActive 是否激活
func (s *Subscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive && time.Now().Before(s.EndDate)
}
