package domain

import "time"

const (
	RiskAnalysisCreatedEventType    = "risksecurity.analysis.created"
	BlacklistAddedEventType         = "risksecurity.blacklist.added"
	BlacklistRemovedEventType       = "risksecurity.blacklist.removed"
	UserBehaviorUpdatedEventType    = "risksecurity.behavior.updated"
	DeviceFingerprintSavedEventType = "risksecurity.device.saved"
)

// RiskDetectedEvent 风险检测事件。
type RiskDetectedEvent struct {
	UserID    uint64    `json:"user_id"`
	RiskType  string    `json:"risk_type"`
	Level     string    `json:"level"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}

// SecurityAlertTriggeredEvent 安全警报触发事件。
type SecurityAlertTriggeredEvent struct {
	Type      string    `json:"type"`
	Detail    string    `json:"detail"`
	Timestamp time.Time `json:"timestamp"`
}

// UserBlockedEvent 用户被锁定事件。
type UserBlockedEvent struct {
	UserID    uint64    `json:"user_id"`
	Reason    string    `json:"reason"`
	Duration  int32     `json:"duration"` // 秒，-1 表示永久
	Timestamp time.Time `json:"timestamp"`
}

// RiskAnalysisCreatedEvent 风险分析结果创建事件。
type RiskAnalysisCreatedEvent struct {
	ResultID  uint64    `json:"result_id"`
	UserID    uint64    `json:"user_id"`
	RiskLevel RiskLevel `json:"risk_level"`
	RiskScore int32     `json:"risk_score"`
	Timestamp time.Time `json:"timestamp"`
}

// BlacklistAddedEvent 黑名单添加事件。
type BlacklistAddedEvent struct {
	BlacklistID uint64        `json:"blacklist_id"`
	Type        BlacklistType `json:"type"`
	Value       string        `json:"value"`
	ExpiresAt   time.Time     `json:"expires_at"`
	Timestamp   time.Time     `json:"timestamp"`
}

// BlacklistRemovedEvent 黑名单移除事件。
type BlacklistRemovedEvent struct {
	BlacklistID uint64        `json:"blacklist_id"`
	Type        BlacklistType `json:"type"`
	Value       string        `json:"value"`
	Timestamp   time.Time     `json:"timestamp"`
}

// UserBehaviorUpdatedEvent 用户行为更新事件。
type UserBehaviorUpdatedEvent struct {
	UserID    uint64    `json:"user_id"`
	IP        string    `json:"ip"`
	DeviceID  string    `json:"device_id"`
	Timestamp time.Time `json:"timestamp"`
}

// DeviceFingerprintSavedEvent 设备指纹保存事件。
type DeviceFingerprintSavedEvent struct {
	DeviceID  string    `json:"device_id"`
	UserID    uint64    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}
