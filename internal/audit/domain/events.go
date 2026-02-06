package domain

import "time"

const (
	AuditLogCreatedEventType      = "audit.log.created"
	AuditLogDeletedEventType      = "audit.log.deleted"
	AuditPolicyCreatedEventType   = "audit.policy.created"
	AuditPolicyUpdatedEventType   = "audit.policy.updated"
	AuditPolicyDeletedEventType   = "audit.policy.deleted"
	AuditReportCreatedEventType   = "audit.report.created"
	AuditReportUpdatedEventType   = "audit.report.updated"
	AuditReportGeneratedEventType = "audit.report.generated"
	AuditReportPublishedEventType = "audit.report.published"
	AuditReportDeletedEventType   = "audit.report.deleted"
	AuditAlertTriggeredEventType  = "audit.alert.triggered"
)

// AuditLogCreatedEvent 审计日志创建事件。
type AuditLogCreatedEvent struct {
	AuditID   uint64    `json:"audit_id"`
	AuditNo   string    `json:"audit_no"`
	EventType string    `json:"event_type"`
	Module    string    `json:"module"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}

// AuditLogDeletedEvent 审计日志删除事件。
type AuditLogDeletedEvent struct {
	AuditID   uint64    `json:"audit_id"`
	Timestamp time.Time `json:"timestamp"`
}

// AuditPolicyCreatedEvent 审计策略创建事件。
type AuditPolicyCreatedEvent struct {
	PolicyID  uint64    `json:"policy_id"`
	Timestamp time.Time `json:"timestamp"`
}

// AuditPolicyUpdatedEvent 审计策略更新事件。
type AuditPolicyUpdatedEvent struct {
	PolicyID  uint64    `json:"policy_id"`
	Timestamp time.Time `json:"timestamp"`
}

// AuditPolicyDeletedEvent 审计策略删除事件。
type AuditPolicyDeletedEvent struct {
	PolicyID  uint64    `json:"policy_id"`
	Timestamp time.Time `json:"timestamp"`
}

// AuditReportCreatedEvent 审计报告创建事件。
type AuditReportCreatedEvent struct {
	ReportID  uint64    `json:"report_id"`
	Timestamp time.Time `json:"timestamp"`
}

// AuditReportUpdatedEvent 审计报告更新事件。
type AuditReportUpdatedEvent struct {
	ReportID  uint64    `json:"report_id"`
	Timestamp time.Time `json:"timestamp"`
}

// AuditReportGeneratedEvent 审计报告生成事件。
type AuditReportGeneratedEvent struct {
	ReportID  uint64    `json:"report_id"`
	Timestamp time.Time `json:"timestamp"`
}

// AuditReportPublishedEvent 审计报告发布事件。
type AuditReportPublishedEvent struct {
	ReportID  uint64    `json:"report_id"`
	Timestamp time.Time `json:"timestamp"`
}

// AuditReportDeletedEvent 审计报告删除事件。
type AuditReportDeletedEvent struct {
	ReportID  uint64    `json:"report_id"`
	Timestamp time.Time `json:"timestamp"`
}

// AuditAlertTriggeredEvent 审计预警触发事件。
type AuditAlertTriggeredEvent struct {
	AuditID   uint64    `json:"audit_id"`
	Level     string    `json:"level"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}
