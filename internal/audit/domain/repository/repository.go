package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/audit/domain/entity" // 导入审计领域的实体定义。
	"time"                                                        // 导入时间包，用于查询条件。
)

// AuditRepository 是审计模块的仓储接口。
// 它定义了对审计日志、审计策略和审计报告进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type AuditRepository interface {
	// --- Log methods ---

	// CreateLog 在数据存储中创建一个新的审计日志记录。
	// ctx: 上下文。
	// log: 待创建的审计日志实体。
	CreateLog(ctx context.Context, log *entity.AuditLog) error
	// GetLog 根据ID获取审计日志实体。
	GetLog(ctx context.Context, id uint64) (*entity.AuditLog, error)
	// ListLogs 列出所有审计日志实体，支持通过查询条件进行过滤和分页。
	ListLogs(ctx context.Context, query *AuditLogQuery) ([]*entity.AuditLog, int64, error)
	// DeleteLog 根据ID删除审计日志实体。
	DeleteLog(ctx context.Context, id uint64) error
	// DeleteLogsBefore 删除指定时间点之前的审计日志，用于数据清理。
	DeleteLogsBefore(ctx context.Context, beforeTime time.Time) error

	// --- Policy methods ---

	// CreatePolicy 在数据存储中创建一个新的审计策略实体。
	CreatePolicy(ctx context.Context, policy *entity.AuditPolicy) error
	// GetPolicy 根据ID获取审计策略实体。
	GetPolicy(ctx context.Context, id uint64) (*entity.AuditPolicy, error)
	// ListPolicies 列出所有审计策略实体，支持分页。
	ListPolicies(ctx context.Context, offset, limit int) ([]*entity.AuditPolicy, int64, error)
	// UpdatePolicy 更新审计策略实体的信息。
	UpdatePolicy(ctx context.Context, policy *entity.AuditPolicy) error
	// DeletePolicy 根据ID删除审计策略实体。
	DeletePolicy(ctx context.Context, id uint64) error

	// --- Report methods ---

	// CreateReport 在数据存储中创建一个新的审计报告实体。
	CreateReport(ctx context.Context, report *entity.AuditReport) error
	// GetReport 根据ID获取审计报告实体。
	GetReport(ctx context.Context, id uint64) (*entity.AuditReport, error)
	// ListReports 列出所有审计报告实体，支持分页。
	ListReports(ctx context.Context, offset, limit int) ([]*entity.AuditReport, int64, error)
	// UpdateReport 更新审计报告实体的信息。
	UpdateReport(ctx context.Context, report *entity.AuditReport) error
	// DeleteReport 根据ID删除审计报告实体。
	DeleteReport(ctx context.Context, id uint64) error
}

// AuditLogQuery 结构体定义了查询审计日志的条件。
// 它用于在仓储层进行数据过滤和分页。
type AuditLogQuery struct {
	UserID       uint64                // 根据用户ID过滤。
	EventType    entity.AuditEventType // 根据事件类型过滤。
	Module       string                // 根据模块过滤。
	ResourceType string                // 根据资源类型过滤。
	StartTime    time.Time             // 查询的起始时间。
	EndTime      time.Time             // 查询的结束时间。
	Page         int                   // 页码，用于分页。
	PageSize     int                   // 每页数量，用于分页。
}
