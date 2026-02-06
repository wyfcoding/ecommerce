package domain

import "time"

const (
	MetricRecordedEventType   = "analytics.metric.recorded"
	MetricDeletedEventType    = "analytics.metric.deleted"
	DashboardCreatedEventType = "analytics.dashboard.created"
	DashboardUpdatedEventType = "analytics.dashboard.updated"
	DashboardDeletedEventType = "analytics.dashboard.deleted"
	ReportCreatedEventType    = "analytics.report.created"
	ReportUpdatedEventType    = "analytics.report.updated"
	ReportPublishedEventType  = "analytics.report.published"
	ReportDeletedEventType    = "analytics.report.deleted"
)

// MetricRecordedEvent 指标记录事件。
type MetricRecordedEvent struct {
	MetricID  uint64    `json:"metric_id"`
	Timestamp time.Time `json:"timestamp"`
}

// MetricDeletedEvent 指标删除事件。
type MetricDeletedEvent struct {
	MetricID  uint64    `json:"metric_id"`
	Timestamp time.Time `json:"timestamp"`
}

// DashboardCreatedEvent 仪表板创建事件。
type DashboardCreatedEvent struct {
	DashboardID uint64    `json:"dashboard_id"`
	UserID      uint64    `json:"user_id"`
	Timestamp   time.Time `json:"timestamp"`
}

// DashboardUpdatedEvent 仪表板更新事件。
type DashboardUpdatedEvent struct {
	DashboardID uint64    `json:"dashboard_id"`
	UserID      uint64    `json:"user_id"`
	Timestamp   time.Time `json:"timestamp"`
}

// DashboardDeletedEvent 仪表板删除事件。
type DashboardDeletedEvent struct {
	DashboardID uint64    `json:"dashboard_id"`
	UserID      uint64    `json:"user_id"`
	Timestamp   time.Time `json:"timestamp"`
}

// ReportCreatedEvent 报告创建事件。
type ReportCreatedEvent struct {
	ReportID  uint64    `json:"report_id"`
	UserID    uint64    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

// ReportUpdatedEvent 报告更新事件。
type ReportUpdatedEvent struct {
	ReportID  uint64    `json:"report_id"`
	UserID    uint64    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

// ReportPublishedEvent 报告发布事件。
type ReportPublishedEvent struct {
	ReportID  uint64    `json:"report_id"`
	UserID    uint64    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

// ReportDeletedEvent 报告删除事件。
type ReportDeletedEvent struct {
	ReportID  uint64    `json:"report_id"`
	UserID    uint64    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}
