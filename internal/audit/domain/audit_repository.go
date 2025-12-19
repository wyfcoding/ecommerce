package domain

import (
	"context"
	"time"
)

// AuditRepository 是审计模块的仓储接口。
type AuditRepository interface {
	// --- Log methods ---

	// CreateLog 在数据存储中创建一个新的审计日志记录。
	CreateLog(ctx context.Context, log *AuditLog) error
	// GetLog 根据ID获取审计日志实体。
	GetLog(ctx context.Context, id uint64) (*AuditLog, error)
	// ListLogs 列出所有审计日志实体，支持通过查询条件进行过滤和分页。
	ListLogs(ctx context.Context, query *AuditLogQuery) ([]*AuditLog, int64, error)
	// DeleteLog 根据ID删除审计日志实体。
	DeleteLog(ctx context.Context, id uint64) error
	// DeleteLogsBefore 删除指定时间点之前的审计日志，用于数据清理。
	DeleteLogsBefore(ctx context.Context, beforeTime time.Time) error

	// --- Policy methods ---

	// CreatePolicy 在数据存储中创建一个新的审计策略实体。
	CreatePolicy(ctx context.Context, policy *AuditPolicy) error
	// GetPolicy 根据ID获取审计策略实体。
	GetPolicy(ctx context.Context, id uint64) (*AuditPolicy, error)
	// ListPolicies 列出所有审计策略实体，支持分页。
	ListPolicies(ctx context.Context, offset, limit int) ([]*AuditPolicy, int64, error)
	// UpdatePolicy 更新审计策略实体的信息。
	UpdatePolicy(ctx context.Context, policy *AuditPolicy) error
	// DeletePolicy 根据ID删除审计策略实体。
	DeletePolicy(ctx context.Context, id uint64) error

	// --- Report methods ---

	// CreateReport 在数据存储中创建一个新的审计报告实体。
	CreateReport(ctx context.Context, report *AuditReport) error
	// GetReport 根据ID获取审计报告实体。
	GetReport(ctx context.Context, id uint64) (*AuditReport, error)
	// ListReports 列出所有审计报告实体，支持分页。
	ListReports(ctx context.Context, offset, limit int) ([]*AuditReport, int64, error)
	// UpdateReport 更新审计报告实体的信息。
	UpdateReport(ctx context.Context, report *AuditReport) error
	// DeleteReport 根据ID删除审计报告实体。
	DeleteReport(ctx context.Context, id uint64) error
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
