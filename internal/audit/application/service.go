package application

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/audit/domain/entity"     // 导入审计领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/audit/domain/repository" // 导入审计领域的仓储接口。
	"github.com/wyfcoding/ecommerce/pkg/idgen"                        // 导入ID生成器接口。

	"log/slog" // 导入结构化日志库。
)

// AuditService 结构体定义了审计管理相关的应用服务。
// 它协调领域层和基础设施层，处理审计日志的记录、审计策略的管理和审计报告的生成等业务逻辑。
type AuditService struct {
	repo        repository.AuditRepository // 依赖AuditRepository接口，用于数据持久化操作。
	idGenerator idgen.Generator            // 依赖ID生成器接口，用于生成唯一的审计日志编号和报告编号。
	logger      *slog.Logger               // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewAuditService 创建并返回一个新的 AuditService 实例。
func NewAuditService(repo repository.AuditRepository, idGenerator idgen.Generator, logger *slog.Logger) *AuditService {
	return &AuditService{
		repo:        repo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// LogEvent 记录一个审计事件。
// ctx: 上下文。
// userID: 执行操作的用户ID。
// username: 执行操作的用户名。
// eventType: 审计事件类型。
// module: 发生事件的模块。
// action: 执行的具体操作。
// opts: 可选的日志选项，用于添加额外信息，如错误信息、资源信息等。
// 返回可能发生的错误。
func (s *AuditService) LogEvent(ctx context.Context, userID uint64, username string, eventType entity.AuditEventType, module, action string, opts ...LogOption) error {
	// 生成唯一的审计日志编号。
	auditNo := fmt.Sprintf("AUD%d", s.idGenerator.Generate())
	// 创建审计日志实体。
	log := entity.NewAuditLog(auditNo, userID, username, eventType, module, action)

	// 应用所有传入的日志选项，为审计日志添加额外信息。
	for _, opt := range opts {
		opt(log)
	}

	// 通过仓储接口保存审计日志。
	if err := s.repo.CreateLog(ctx, log); err != nil {
		s.logger.ErrorContext(ctx, "failed to create audit log", "user_id", userID, "event_type", eventType, "error", err)
		return err
	}
	return nil
}

// LogOption 定义了用于配置审计日志的函数式选项类型。
type LogOption func(*entity.AuditLog)

// WithError 是一个 LogOption，用于向审计日志添加错误信息。
func WithError(errMsg string) LogOption {
	return func(l *entity.AuditLog) {
		l.SetError(errMsg)
	}
}

// WithResource 是一个 LogOption，用于向审计日志添加资源信息。
func WithResource(resourceType, resourceID string) LogOption {
	return func(l *entity.AuditLog) {
		l.SetResource(resourceType, resourceID)
	}
}

// WithChange 是一个 LogOption，用于向审计日志添加变更前后数据信息。
func WithChange(oldValue, newValue string) LogOption {
	return func(l *entity.AuditLog) {
		l.SetChange(oldValue, newValue)
	}
}

// WithClientInfo 是一个 LogOption，用于向审计日志添加客户端信息（IP和User-Agent）。
func WithClientInfo(ip, userAgent string) LogOption {
	return func(l *entity.AuditLog) {
		l.SetClientInfo(ip, userAgent)
	}
}

// WithDuration 是一个 LogOption，用于向审计日志添加操作耗时。
func WithDuration(duration int64) LogOption {
	return func(l *entity.AuditLog) {
		l.SetDuration(duration)
	}
}

// QueryLogs 查询审计日志。
// ctx: 上下文。
// query: 包含过滤条件和分页参数的查询对象。
// 返回审计日志列表、总数和可能发生的错误。
func (s *AuditService) QueryLogs(ctx context.Context, query *repository.AuditLogQuery) ([]*entity.AuditLog, int64, error) {
	return s.repo.ListLogs(ctx, query)
}

// CreatePolicy 创建一个新的审计策略。
// ctx: 上下文。
// name: 策略名称。
// description: 策略描述。
// 返回创建成功的AuditPolicy实体和可能发生的错误。
func (s *AuditService) CreatePolicy(ctx context.Context, name, description string) (*entity.AuditPolicy, error) {
	policy := entity.NewAuditPolicy(name, description) // 创建审计策略实体。
	if err := s.repo.CreatePolicy(ctx, policy); err != nil {
		s.logger.ErrorContext(ctx, "failed to create audit policy", "name", name, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "audit policy created successfully", "policy_id", policy.ID, "name", name)
	return policy, nil
}

// UpdatePolicy 更新审计策略。
// ctx: 上下文。
// id: 审计策略ID。
// eventTypes: 策略关注的事件类型列表。
// modules: 策略关注的模块列表。
// enabled: 策略是否启用。
// 返回可能发生的错误。
func (s *AuditService) UpdatePolicy(ctx context.Context, id uint64, eventTypes, modules []string, enabled bool) error {
	// 获取审计策略实体。
	policy, err := s.repo.GetPolicy(ctx, id)
	if err != nil {
		return err
	}

	// 更新策略属性。
	policy.EventTypes = eventTypes
	policy.Modules = modules
	policy.Enabled = enabled
	policy.UpdatedAt = time.Now() // 更新UpdatedAt字段。

	// 通过仓储接口更新数据库中的策略。
	return s.repo.UpdatePolicy(ctx, policy)
}

// ListPolicies 获取审计策略列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回审计策略列表、总数和可能发生的错误。
func (s *AuditService) ListPolicies(ctx context.Context, page, pageSize int) ([]*entity.AuditPolicy, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListPolicies(ctx, offset, pageSize)
}

// CreateReport 创建一个新的审计报告。
// ctx: 上下文。
// title: 报告标题。
// description: 报告描述。
// 返回创建成功的AuditReport实体和可能发生的错误。
func (s *AuditService) CreateReport(ctx context.Context, title, description string) (*entity.AuditReport, error) {
	// 生成唯一的报告编号。
	reportNo := fmt.Sprintf("AUDRPT%d", s.idGenerator.Generate())
	// 创建审计报告实体。
	report := entity.NewAuditReport(reportNo, title, description)

	// 通过仓储接口保存审计报告。
	if err := s.repo.CreateReport(ctx, report); err != nil {
		s.logger.ErrorContext(ctx, "failed to create audit report", "title", title, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "audit report created successfully", "report_id", report.ID, "title", title)
	return report, nil
}

// GenerateReport 生成指定ID的审计报告内容。
// ctx: 上下文。
// id: 审计报告ID。
// 返回可能发生的错误。
func (s *AuditService) GenerateReport(ctx context.Context, id uint64) error {
	// 获取审计报告实体。
	report, err := s.repo.GetReport(ctx, id)
	if err != nil {
		return err
	}

	// Mock report generation: 模拟报告生成逻辑，实际应根据审计日志数据生成报告内容。
	content := fmt.Sprintf("Audit Report for %s generated at %s", report.Title, time.Now().Format(time.RFC3339))
	report.Generate(content) // 调用实体方法生成报告内容。

	// 更新数据库中的报告。
	return s.repo.UpdateReport(ctx, report)
}

// ListReports 获取审计报告列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回审计报告列表、总数和可能发生的错误。
func (s *AuditService) ListReports(ctx context.Context, page, pageSize int) ([]*entity.AuditReport, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListReports(ctx, offset, pageSize)
}
