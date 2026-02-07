package domain

import "time"

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

// RiskAnalysisResult 实体代表一次风险分析的综合结果。
type RiskAnalysisResult struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uint64    `json:"user_id"`
	RiskScore int32     `json:"risk_score"`
	RiskLevel RiskLevel `json:"risk_level"`
	RiskItems string    `json:"risk_items"`
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
	ID        uint          `json:"id"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Type      BlacklistType `json:"type"`
	Value     string        `json:"value"`
	Reason    string        `json:"reason"`
	ExpiresAt time.Time     `json:"expires_at"`
}

// IsActive 检查黑名单条目是否仍在有效期内。
func (b *Blacklist) IsActive() bool {
	return time.Now().Before(b.ExpiresAt)
}

// DeviceFingerprint 实体代表设备的指纹信息及关联用户。
type DeviceFingerprint struct {
	ID         uint              `json:"id"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	UserID     uint64            `json:"user_id"`
	DeviceID   string            `json:"device_id"`
	DeviceInfo map[string]string `json:"device_info"`
}

// UserBehavior 实体记录了用户的关键行为数据快照。
type UserBehavior struct {
	ID                uint              `json:"id"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	UserID            uint64            `json:"user_id"`
	LastLoginIP       string            `json:"last_login_ip"`
	LastLoginTime     time.Time         `json:"last_login_time"`
	LastLoginDevice   string            `json:"last_login_device"`
	PurchasedCategory map[string]string `json:"purchased_category"`
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
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Type      RiskType  `json:"type"`
	Condition string    `json:"condition"`
	Score     int32     `json:"score"`
	Enabled   bool      `json:"enabled"`
}
