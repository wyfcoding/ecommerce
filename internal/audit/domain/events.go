package domain

import "time"

// AuditLogCreatedEvent 审计日志创建事件。
type AuditLogCreatedEvent struct {
	AuditID   uint64    `json:"audit_id"`
	Source    string    `json:"source"`
	Action    string    `json:"action"`
	TargetID  uint64    `json:"target_id"`
	Timestamp time.Time `json:"timestamp"`
}

// AuditAlertTriggeredEvent 审计预警触发事件。
type AuditAlertTriggeredEvent struct {
	AuditID   uint64    `json:"audit_id"`
	Level     string    `json:"level"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}
