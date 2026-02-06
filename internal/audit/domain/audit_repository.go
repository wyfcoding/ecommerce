package domain

import (
	"context"
	"time"
)

// AuditRepository 是审计模块的仓储接口。
type AuditRepository interface {
	// 事务支持
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// --- Log methods ---
	SaveLog(ctx context.Context, log *AuditLog) error
	SaveLogInTx(ctx context.Context, tx any, log *AuditLog) error
	GetLog(ctx context.Context, id uint64) (*AuditLog, error)
	ListLogs(ctx context.Context, query *AuditLogQuery) ([]*AuditLog, int64, error)
	DeleteLog(ctx context.Context, id uint64) error
	DeleteLogInTx(ctx context.Context, tx any, id uint64) error
	DeleteLogsBefore(ctx context.Context, beforeTime time.Time) error

	// --- Policy methods ---
	SavePolicy(ctx context.Context, policy *AuditPolicy) error
	SavePolicyInTx(ctx context.Context, tx any, policy *AuditPolicy) error
	GetPolicy(ctx context.Context, id uint64) (*AuditPolicy, error)
	ListPolicies(ctx context.Context, offset, limit int) ([]*AuditPolicy, int64, error)
	DeletePolicy(ctx context.Context, id uint64) error
	DeletePolicyInTx(ctx context.Context, tx any, id uint64) error

	// --- Report methods ---
	SaveReport(ctx context.Context, report *AuditReport) error
	SaveReportInTx(ctx context.Context, tx any, report *AuditReport) error
	GetReport(ctx context.Context, id uint64) (*AuditReport, error)
	ListReports(ctx context.Context, offset, limit int) ([]*AuditReport, int64, error)
	DeleteReport(ctx context.Context, id uint64) error
	DeleteReportInTx(ctx context.Context, tx any, id uint64) error
}

// AuditLogQuery 结构体定义了查询审计日志的条件。
type AuditLogQuery struct {
	UserID       uint64
	EventType    AuditEventType
	Module       string
	ResourceType string
	StartTime    time.Time
	EndTime      time.Time
	Page         int
	PageSize     int
}
