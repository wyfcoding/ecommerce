package entity

import (
	"database/sql/driver" // 导入数据库驱动接口。
	"encoding/json"       // 导入JSON编码/解码库。
	"errors"              // 导入标准错误处理库。
	"time"                // 导入时间包。

	"gorm.io/gorm" // 导入GORM库。
)

// SubscriptionStatus 定义了订阅的生命周期状态。
type SubscriptionStatus int8

const (
	SubscriptionStatusActive   SubscriptionStatus = 1 // 活跃：订阅正在生效中。
	SubscriptionStatusExpired  SubscriptionStatus = 2 // 过期：订阅已到期。
	SubscriptionStatusCanceled SubscriptionStatus = 3 // 取消：订阅已被用户或系统取消。
	SubscriptionStatusPaused   SubscriptionStatus = 4 // 暂停：订阅暂时暂停。
)

// StringArray 定义了一个字符串切片类型，实现了 sql.Scanner 和 driver.Valuer 接口，
// 允许GORM将Go的 []string 类型作为JSON字符串存储到数据库，并从数据库读取。
type StringArray []string

// Value 实现 driver.Valuer 接口，将 StringArray 转换为数据库可以存储的值（JSON字节数组）。
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a) // 将切片编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 StringArray。
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte) // 期望数据库返回字节数组。
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a) // 将JSON字节数组解码为切片。
}

// SubscriptionPlan 实体代表一个订阅计划。
// 它包含了计划的名称、描述、价格、持续时间、包含的功能和启用状态。
type SubscriptionPlan struct {
	gorm.Model              // 嵌入gorm.Model。
	Name        string      `gorm:"type:varchar(128);not null;comment:计划名称" json:"name"` // 计划名称。
	Description string      `gorm:"type:varchar(255);comment:描述" json:"description"`     // 计划描述。
	Price       uint64      `gorm:"not null;comment:价格(分)" json:"price"`                 // 计划价格（单位：分）。
	Duration    int32       `gorm:"not null;comment:时长(天)" json:"duration"`              // 计划持续时长（天）。
	Features    StringArray `gorm:"type:json;comment:特性列表" json:"features"`              // 计划包含的功能列表，存储为JSON。
	Enabled     bool        `gorm:"not null;default:true;comment:是否启用" json:"enabled"`   // 计划是否启用。
}

// Subscription 实体代表用户的订阅记录。
// 它包含了订阅用户、订阅计划、订阅状态、生效时间范围和自动续订设置等。
type Subscription struct {
	gorm.Model                    // 嵌入gorm.Model。
	UserID     uint64             `gorm:"index;not null;comment:用户ID" json:"user_id"`               // 订阅用户ID，索引字段。
	PlanID     uint64             `gorm:"index;not null;comment:计划ID" json:"plan_id"`               // 订阅计划ID，索引字段。
	Status     SubscriptionStatus `gorm:"type:tinyint;not null;default:1;comment:状态" json:"status"` // 订阅状态，默认为活跃。
	StartDate  time.Time          `gorm:"not null;comment:开始时间" json:"start_date"`                  // 订阅开始时间。
	EndDate    time.Time          `gorm:"not null;comment:结束时间" json:"end_date"`                    // 订阅结束时间。
	AutoRenew  bool               `gorm:"not null;default:true;comment:自动续订" json:"auto_renew"`     // 是否自动续订。
	CanceledAt *time.Time         `gorm:"comment:取消时间" json:"canceled_at"`                          // 订阅取消时间。
}

// IsActive 检查订阅是否当前处于活跃状态。
func (s *Subscription) IsActive() bool {
	// 判断订阅状态是否为活跃，并且当前时间是否在订阅有效期内。
	return s.Status == SubscriptionStatusActive && time.Now().Before(s.EndDate)
}
