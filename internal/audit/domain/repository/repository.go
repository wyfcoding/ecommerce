package repository

import (
	"context"
	"ecommerce/internal/audit/domain/entity"
	"time"
)

// AuditRepository 审计服务仓储接口
type AuditRepository interface {
	// Log methods
	CreateLog(ctx context.Context, log *entity.AuditLog) error
	GetLog(ctx context.Context, id uint64) (*entity.AuditLog, error)
	ListLogs(ctx context.Context, query *AuditLogQuery) ([]*entity.AuditLog, int64, error)
	DeleteLog(ctx context.Context, id uint64) error
	DeleteLogsBefore(ctx context.Context, beforeTime time.Time) error

	// Policy methods
	CreatePolicy(ctx context.Context, policy *entity.AuditPolicy) error
	GetPolicy(ctx context.Context, id uint64) (*entity.AuditPolicy, error)
	ListPolicies(ctx context.Context, offset, limit int) ([]*entity.AuditPolicy, int64, error)
	UpdatePolicy(ctx context.Context, policy *entity.AuditPolicy) error
	DeletePolicy(ctx context.Context, id uint64) error

	// Report methods
	CreateReport(ctx context.Context, report *entity.AuditReport) error
	GetReport(ctx context.Context, id uint64) (*entity.AuditReport, error)
	ListReports(ctx context.Context, offset, limit int) ([]*entity.AuditReport, int64, error)
	UpdateReport(ctx context.Context, report *entity.AuditReport) error
	DeleteReport(ctx context.Context, id uint64) error
}

// AuditLogQuery 审计日志查询条件
type AuditLogQuery struct {
	UserID       uint64
	EventType    entity.AuditEventType
	Module       string
	ResourceType string
	StartTime    time.Time
	EndTime      time.Time
	Page         int
	PageSize     int
}
