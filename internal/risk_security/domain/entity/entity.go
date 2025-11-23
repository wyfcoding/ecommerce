package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// RiskLevel 风险等级
type RiskLevel int32

const (
	RiskLevelVeryLow  RiskLevel = 0
	RiskLevelLow      RiskLevel = 1
	RiskLevelMedium   RiskLevel = 2
	RiskLevelHigh     RiskLevel = 3
	RiskLevelCritical RiskLevel = 4
)

// RiskType 风险类型
type RiskType string

const (
	RiskTypeBlacklist            RiskType = "blacklist"
	RiskTypeAnomalousTransaction RiskType = "anomalous_transaction"
	RiskTypeDeviceRisk           RiskType = "device_risk"
	RiskTypeIPRisk               RiskType = "ip_risk"
	RiskTypeBehaviorAnomaly      RiskType = "behavior_anomaly"
)

// BlacklistType 黑名单类型
type BlacklistType string

const (
	BlacklistTypeUser   BlacklistType = "user"
	BlacklistTypeIP     BlacklistType = "ip"
	BlacklistTypeDevice BlacklistType = "device"
	BlacklistTypeEmail  BlacklistType = "email"
	BlacklistTypePhone  BlacklistType = "phone"
)

// StringMap defines a map of strings that implements sql.Scanner and driver.Valuer
type StringMap map[string]string

func (m StringMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *StringMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m)
}

// RiskAnalysisResult 风险分析结果
type RiskAnalysisResult struct {
	gorm.Model
	UserID    uint64    `gorm:"index;not null;comment:用户ID" json:"user_id"`
	RiskScore int32     `gorm:"not null;comment:风险分数" json:"risk_score"`
	RiskLevel RiskLevel `gorm:"type:tinyint;not null;comment:风险等级" json:"risk_level"`
	RiskItems string    `gorm:"type:json;comment:风险项详情" json:"risk_items"` // JSON string of []*RiskItem
}

// RiskItem 风险项 (Value Object, stored as JSON in RiskAnalysisResult)
type RiskItem struct {
	Type      RiskType  `json:"type"`
	Level     RiskLevel `json:"level"`
	Score     int32     `json:"score"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// Blacklist 黑名单
type Blacklist struct {
	gorm.Model
	Type      BlacklistType `gorm:"type:varchar(32);not null;index;comment:类型" json:"type"`
	Value     string        `gorm:"type:varchar(255);not null;index;comment:值" json:"value"`
	Reason    string        `gorm:"type:varchar(255);comment:原因" json:"reason"`
	ExpiresAt time.Time     `gorm:"index;comment:过期时间" json:"expires_at"`
}

// IsActive 是否活跃
func (b *Blacklist) IsActive() bool {
	return time.Now().Before(b.ExpiresAt)
}

// DeviceFingerprint 设备指纹
type DeviceFingerprint struct {
	gorm.Model
	UserID     uint64    `gorm:"index;not null;comment:用户ID" json:"user_id"`
	DeviceID   string    `gorm:"type:varchar(128);uniqueIndex;not null;comment:设备ID" json:"device_id"`
	DeviceInfo StringMap `gorm:"type:json;comment:设备信息" json:"device_info"`
}

// UserBehavior 用户行为
type UserBehavior struct {
	gorm.Model
	UserID            uint64    `gorm:"uniqueIndex;not null;comment:用户ID" json:"user_id"`
	LastLoginIP       string    `gorm:"type:varchar(64);comment:最后登录IP" json:"last_login_ip"`
	LastLoginTime     time.Time `gorm:"comment:最后登录时间" json:"last_login_time"`
	LastLoginDevice   string    `gorm:"type:varchar(128);comment:最后登录设备" json:"last_login_device"`
	PurchasedCategory StringMap `gorm:"type:json;comment:已购类目" json:"purchased_category"` // map[category]bool stored as map[string]string "true"
}

// RiskRule 风险规则 (Configuration)
type RiskRule struct {
	gorm.Model
	Name      string   `gorm:"type:varchar(128);uniqueIndex;not null;comment:规则名称" json:"name"`
	Type      RiskType `gorm:"type:varchar(32);not null;comment:规则类型" json:"type"`
	Condition string   `gorm:"type:text;not null;comment:规则条件(JSON/Expression)" json:"condition"`
	Score     int32    `gorm:"not null;comment:风险分数" json:"score"`
	Enabled   bool     `gorm:"default:true;comment:是否启用" json:"enabled"`
}
