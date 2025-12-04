package entity

import (
	"database/sql/driver" // 导入数据库驱动接口。
	"encoding/json"       // 导入JSON编码/解码库。
	"errors"              // 导入标准错误处理库。
	"time"                // 导入时间包。

	"gorm.io/gorm" // 导入GORM库。
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
	BlacklistTypeUser   BlacklistType = "user"   // 用户ID。
	BlacklistTypeIP     BlacklistType = "ip"     // IP地址。
	BlacklistTypeDevice BlacklistType = "device" // 设备ID。
	BlacklistTypeEmail  BlacklistType = "email"  // 邮箱。
	BlacklistTypePhone  BlacklistType = "phone"  // 手机号。
)

// StringMap 定义了一个map[string]string类型，实现了 sql.Scanner 和 driver.Valuer 接口，
// 允许GORM将Go的map[string]string类型作为JSON字符串存储到数据库，并从数据库读取。
type StringMap map[string]string

// Value 实现 driver.Valuer 接口，将 StringMap 转换为数据库可以存储的值（JSON字节数组）。
func (m StringMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m) // 将map编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 StringMap。
func (m *StringMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte) // 期望数据库返回字节数组。
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m) // 将JSON字节数组解码为map。
}

// RiskAnalysisResult 实体代表一次风险分析的结果。
// 它包含了被分析用户、风险分数、风险等级以及具体的风险项详情。
type RiskAnalysisResult struct {
	gorm.Model           // 嵌入gorm.Model。
	UserID     uint64    `gorm:"index;not null;comment:用户ID" json:"user_id"`           // 关联的用户ID，索引字段。
	RiskScore  int32     `gorm:"not null;comment:风险分数" json:"risk_score"`              // 风险评估的总分数。
	RiskLevel  RiskLevel `gorm:"type:tinyint;not null;comment:风险等级" json:"risk_level"` // 风险等级。
	RiskItems  string    `gorm:"type:json;comment:风险项详情" json:"risk_items"`            // 具体的风险项列表，存储为JSON字符串（类型为 []*RiskItem）。
}

// RiskItem 值对象定义了具体的风险项。
// 通常作为 RiskAnalysisResult.RiskItems 字段的JSON存储内容。
type RiskItem struct {
	Type      RiskType  `json:"type"`      // 风险类型。
	Level     RiskLevel `json:"level"`     // 该风险项的等级。
	Score     int32     `json:"score"`     // 该风险项的分数。
	Reason    string    `json:"reason"`    // 风险原因。
	Timestamp time.Time `json:"timestamp"` // 风险发生时间。
}

// Blacklist 实体代表一个黑名单条目。
// 用于阻止或限制特定用户、IP、设备等的访问。
type Blacklist struct {
	gorm.Model               // 嵌入gorm.Model。
	Type       BlacklistType `gorm:"type:varchar(32);not null;index;comment:类型" json:"type"`  // 黑名单类型。
	Value      string        `gorm:"type:varchar(255);not null;index;comment:值" json:"value"` // 黑名单值，例如IP地址、用户ID。
	Reason     string        `gorm:"type:varchar(255);comment:原因" json:"reason"`              // 加入黑名单的原因。
	ExpiresAt  time.Time     `gorm:"index;comment:过期时间" json:"expires_at"`                    // 黑名单的过期时间，索引字段。
}

// IsActive 检查黑名单条目是否仍在有效期内。
func (b *Blacklist) IsActive() bool {
	return time.Now().Before(b.ExpiresAt) // 当前时间在过期时间之前则为活跃。
}

// DeviceFingerprint 实体代表设备的指纹信息。
// 用于识别设备，辅助风险评估和用户行为分析。
type DeviceFingerprint struct {
	gorm.Model           // 嵌入gorm.Model。
	UserID     uint64    `gorm:"index;not null;comment:用户ID" json:"user_id"`                           // 关联的用户ID，索引字段。
	DeviceID   string    `gorm:"type:varchar(128);uniqueIndex;not null;comment:设备ID" json:"device_id"` // 设备唯一标识符，唯一索引，不允许为空。
	DeviceInfo StringMap `gorm:"type:json;comment:设备信息" json:"device_info"`                            // 设备的详细信息，存储为JSON。
}

// UserBehavior 实体记录了用户的行为数据。
// 用于分析用户的操作模式，发现异常行为。
type UserBehavior struct {
	gorm.Model                  // 嵌入gorm.Model。
	UserID            uint64    `gorm:"uniqueIndex;not null;comment:用户ID" json:"user_id"`          // 用户ID，唯一索引，不允许为空。
	LastLoginIP       string    `gorm:"type:varchar(64);comment:最后登录IP" json:"last_login_ip"`      // 最后一次登录的IP地址。
	LastLoginTime     time.Time `gorm:"comment:最后登录时间" json:"last_login_time"`                     // 最后一次登录的时间。
	LastLoginDevice   string    `gorm:"type:varchar(128);comment:最后登录设备" json:"last_login_device"` // 最后一次登录的设备ID。
	PurchasedCategory StringMap `gorm:"type:json;comment:已购类目" json:"purchased_category"`          // 用户已购买的商品类目，存储为JSON（例如，map[category_name]string{"true"}）。
}

// RiskRule 实体定义了一条风险评估规则。
// 这些规则用于指导风险引擎如何计算风险分数和确定风险等级。
type RiskRule struct {
	gorm.Model          // 嵌入gorm.Model。
	Name       string   `gorm:"type:varchar(128);uniqueIndex;not null;comment:规则名称" json:"name"`   // 规则名称，唯一索引，不允许为空。
	Type       RiskType `gorm:"type:varchar(32);not null;comment:规则类型" json:"type"`                // 规则关联的风险类型。
	Condition  string   `gorm:"type:text;not null;comment:规则条件(JSON/Expression)" json:"condition"` // 规则条件，可以是JSON格式或表达式字符串，用于动态规则引擎。
	Score      int32    `gorm:"not null;comment:风险分数" json:"score"`                                // 满足此规则时增加或减少的风险分数。
	Enabled    bool     `gorm:"default:true;comment:是否启用" json:"enabled"`                          // 规则是否启用。
}
