package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// RiskLevel 定义了风险的等级。
type RiskLevel int32

const (
	RiskLevelVeryLow  RiskLevel = 0 // 极低风险。
	RiskLevelLow      RiskLevel = 1 // 低风险。
	RiskLevelMedium   RiskLevel = 2 // 中风险。
	RiskLevelHigh     RiskLevel = 3 // 高风险。
	RiskLevelCritical RiskLevel = 4 // 严重风险。
)

// RiskType 定义了风险的类型。
type RiskType string

const (
	RiskTypeBlacklist            RiskType = "blacklist"             // 黑名单风险。
	RiskTypeAnomalousTransaction RiskType = "anomalous_transaction" // 异常交易风险。
	RiskTypeDeviceRisk           RiskType = "device_risk"           // 设备风险。
	RiskTypeIPRisk               RiskType = "ip_risk"               // IP风险。
	RiskTypeBehaviorAnomaly      RiskType = "behavior_anomaly"      // 行为异常风险。
)

// BlacklistType 定义了黑名单的类型。
type BlacklistType string

const (
	BlacklistTypeUser   BlacklistType = "user"   // 用户ID黑名单。
	BlacklistTypeIP     BlacklistType = "ip"     // IP地址黑名单。
	BlacklistTypeDevice BlacklistType = "device" // 设备ID黑名单。
	BlacklistTypeEmail  BlacklistType = "email"  // 邮箱黑名单。
	BlacklistTypePhone  BlacklistType = "phone"  // 手机号黑名单。
)

// StringMap 定义了一个map[string]string类型，实现了 sql.Scanner 和 driver.Valuer 接口。
type StringMap map[string]string

// Value 实现 driver.Valuer 接口，用于数据库存储。
func (m StringMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan 实现 sql.Scanner 接口，用于从数据库读取。
func (m *StringMap) Scan(value any) error {
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

// RiskAnalysisResult 实体代表一次风险分析的综合结果。
type RiskAnalysisResult struct {
	gorm.Model
	UserID    uint64    `gorm:"index;not null;comment:用户ID" json:"user_id"`
	RiskScore int32     `gorm:"not null;comment:风险分数" json:"risk_score"`
	RiskLevel RiskLevel `gorm:"type:tinyint;not null;comment:风险等级" json:"risk_level"`
	RiskItems string    `gorm:"type:json;comment:风险项详情" json:"risk_items"`
}

// RiskItem 值对象定义了具体的风险项。
type RiskItem struct {
	Type      RiskType  `json:"type"`      // 风险类型。
	Level     RiskLevel `json:"level"`     // 风险等级。
	Score     int32     `json:"score"`     // 风险评分。
	Reason    string    `json:"reason"`    // 风险触发原因。
	Timestamp time.Time `json:"timestamp"` // 风险识别时间。
}

// Blacklist 实体代表一个黑名单条目。
type Blacklist struct {
	gorm.Model
	Type      BlacklistType `gorm:"type:varchar(32);not null;index;comment:类型" json:"type"`
	Value     string        `gorm:"type:varchar(255);not null;index;comment:值" json:"value"`
	Reason    string        `gorm:"type:varchar(255);comment:原因" json:"reason"`
	ExpiresAt time.Time     `gorm:"index;comment:过期时间" json:"expires_at"`
}

// IsActive 检查黑名单条目是否仍在有效期内。
func (b *Blacklist) IsActive() bool {
	return time.Now().Before(b.ExpiresAt)
}

// DeviceFingerprint 实体代表设备的指纹信息及关联用户。
type DeviceFingerprint struct {
	gorm.Model
	UserID     uint64    `gorm:"index;not null;comment:用户ID" json:"user_id"`
	DeviceID   string    `gorm:"type:varchar(128);uniqueIndex;not null;comment:设备ID" json:"device_id"`
	DeviceInfo StringMap `gorm:"type:json;comment:设备信息" json:"device_info"`
}

// UserBehavior 实体记录了用户的关键行为数据快照。
type UserBehavior struct {
	gorm.Model
	UserID            uint64    `gorm:"uniqueIndex;not null;comment:用户ID" json:"user_id"`
	LastLoginIP       string    `gorm:"type:varchar(64);comment:最后登录IP" json:"last_login_ip"`
	LastLoginTime     time.Time `gorm:"comment:最后登录时间" json:"last_login_time"`
	LastLoginDevice   string    `gorm:"type:varchar(128);comment:最后登录设备" json:"last_login_device"`
	PurchasedCategory StringMap `gorm:"type:json;comment:已购类目" json:"purchased_category"`
}

// RiskContext 定义了风险评估的上下文信息。
type RiskContext struct {
	UserID        uint64 `json:"user_id"`
	IP            string `json:"ip"`
	DeviceID      string `json:"device_id"`
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	OrderID       uint64 `json:"order_id"`
}

// RiskRule 实体定义了一条风险评估规则配置。
type RiskRule struct {
	gorm.Model
	Name      string   `gorm:"type:varchar(128);uniqueIndex;not null;comment:规则名称" json:"name"`
	Type      RiskType `gorm:"type:varchar(32);not null;comment:规则类型" json:"type"`
	Condition string   `gorm:"type:text;not null;comment:规则条件(JSON/Expression)" json:"condition"`
	Score     int32    `gorm:"not null;comment:风险分数" json:"score"`
	Enabled   bool     `gorm:"default:true;comment:是否启用" json:"enabled"`
}
