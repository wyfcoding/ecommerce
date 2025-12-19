package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// SubscriptionStatus 定义了订阅的生命周期状态。
type SubscriptionStatus int8

const (
	SubscriptionStatusActive   SubscriptionStatus = 1 // 活跃：订阅正在生效中。
	SubscriptionStatusExpired  SubscriptionStatus = 2 // 过期：订阅已到期。
	SubscriptionStatusCanceled SubscriptionStatus = 3 // 取消：订阅已被用户或系统取消。
	SubscriptionStatusPaused   SubscriptionStatus = 4 // 暂停：订阅暂时暂停。
)

// StringArray 定义了一个字符串切片类型，实现了 sql.Scanner 和 driver.Valuer 接口。
type StringArray []string

// Value 实现 driver.Valuer 接口。
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

// Scan 实现 sql.Scanner 接口。
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

// SubscriptionPlan 实体代表一个订阅计划。
type SubscriptionPlan struct {
	gorm.Model
	Name        string      `gorm:"type:varchar(128);not null;comment:计划名称" json:"name"`
	Description string      `gorm:"type:varchar(255);comment:描述" json:"description"`
	Price       uint64      `gorm:"not null;comment:价格(分)" json:"price"`
	Duration    int32       `gorm:"not null;comment:时长(天)" json:"duration"`
	Features    StringArray `gorm:"type:json;comment:特性列表" json:"features"`
	Enabled     bool        `gorm:"not null;default:true;comment:是否启用" json:"enabled"`
}

// Subscription 实体代表用户的订阅记录。
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

// IsActive 检查订阅是否当前处于活跃状态。
func (s *Subscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive && time.Now().Before(s.EndDate)
}
