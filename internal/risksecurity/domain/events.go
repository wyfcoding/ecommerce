package domain

import "time"

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
