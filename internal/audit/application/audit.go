package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/audit/domain"
)

// Audit 结构体定义了审计管理模块的应用服务。
// 它是一个门面（Facade），将复杂的审计逻辑委托给 Manager 和 Query 处理。
type Audit struct {
	manager *AuditManager
	query   *AuditQuery
}

// NewAudit 创建并返回一个新的 Audit 实例。
func NewAudit(manager *AuditManager, query *AuditQuery) *Audit {
	return &Audit{
		manager: manager,
		query:   query,
	}
}

// LogEvent 记录一个审计事件。
func (s *Audit) LogEvent(ctx context.Context, userID uint64, username string, eventType domain.AuditEventType, module, action string, opts ...LogOption) error {
	return s.manager.LogEvent(ctx, userID, username, eventType, module, action, opts...)
}

// QueryLogs 根据条件查询审计日志记录。
func (s *Audit) QueryLogs(ctx context.Context, query *domain.AuditLogQuery) ([]*domain.AuditLog, int64, error) {
	return s.query.ListLogs(ctx, query)
}

// CreatePolicy 创建一个新的审计策略。
func (s *Audit) CreatePolicy(ctx context.Context, name, description string) (*domain.AuditPolicy, error) {
	return s.manager.CreatePolicy(ctx, name, description)
}

// UpdatePolicy 更新现有的审计策略配置。
func (s *Audit) UpdatePolicy(ctx context.Context, id uint64, eventTypes, modules []string, enabled bool) error {
	return s.manager.UpdatePolicy(ctx, id, eventTypes, modules, enabled)
}

// ListPolicies 获取审计策略列表（分页）。
func (s *Audit) ListPolicies(ctx context.Context, page, pageSize int) ([]*domain.AuditPolicy, int64, error) {
	offset := (page - 1) * pageSize
	return s.query.ListPolicies(ctx, offset, pageSize)
}

// CreateReport 创建一个新的审计报告任务。
func (s *Audit) CreateReport(ctx context.Context, title, description string) (*domain.AuditReport, error) {
	return s.manager.CreateReport(ctx, title, description)
}

// GenerateReport 触发审计报告的内容生成过程。
func (s *Audit) GenerateReport(ctx context.Context, id uint64) error {
	return s.manager.GenerateReport(ctx, id)
}

// ListReports 获取所有审计报告列表。
func (s *Audit) ListReports(ctx context.Context, page, pageSize int) ([]*domain.AuditReport, int64, error) {
	offset := (page - 1) * pageSize
	return s.query.ListReports(ctx, offset, pageSize)
}

// DeleteLogsBefore 清理指定时间之前的历史审计日志。
func (s *Audit) DeleteLogsBefore(ctx context.Context, beforeTime time.Time) error {
	return s.manager.DeleteLogsBefore(ctx, beforeTime)
}
