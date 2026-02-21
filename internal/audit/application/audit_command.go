package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/audit/domain"
	algorithm "github.com/wyfcoding/pkg/algos/infra"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/messagequeue"
)

// AuditCommandService 处理审计模块的写操作和业务逻辑。
type AuditCommandService struct {
	repo        domain.AuditRepository
	publisher   messagequeue.EventPublisher
	idGenerator idgen.Generator
	logger      *slog.Logger
}

// NewAuditCommandService 创建并返回一个新的 AuditCommandService 实例。
func NewAuditCommandService(repo domain.AuditRepository, publisher messagequeue.EventPublisher, idGenerator idgen.Generator, logger *slog.Logger) *AuditCommandService {
	return &AuditCommandService{
		repo:        repo,
		publisher:   publisher,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// SealLogs 为最近的日志生成“数字封条”（Merkle Root）
func (m *AuditCommandService) SealLogs(ctx context.Context, limit int) (string, error) {
	// 1. 获取最近的日志
	query := &domain.AuditLogQuery{
		Page:     1,
		PageSize: limit,
	}
	logs, _, err := m.repo.ListLogs(ctx, query)
	if err != nil {
		return "", err
	}
	if len(logs) == 0 {
		return "", nil
	}

	// 2. 提取数据用于构建 Merkle Tree
	data := make([][]byte, len(logs))
	for i, l := range logs {
		// 将关键字段拼接作为节点数据
		data[i] = fmt.Appendf(nil, "%s|%d|%s|%s", l.AuditNo, l.UserID, l.Action, l.CreatedAt.Format(time.RFC3339))
	}

	// 3. 构建 Merkle Tree 并获取根哈希
	tree := algorithm.NewMerkleTree(data)
	rootHash := tree.RootHashHex()

	m.logger.InfoContext(ctx, "audit logs sealed with merkle root", "count", len(logs), "root_hash", rootHash)
	return rootHash, nil
}

// LogEvent 记录一个审计事件。
func (m *AuditCommandService) LogEvent(ctx context.Context, userID uint64, username string, eventType domain.AuditEventType, module, action string, opts ...LogOption) error {
	auditNo := fmt.Sprintf("AUD%d", m.idGenerator.Generate())
	// 默认系统来源为 ECOMMERCE，可以通过 WithSystem 选项覆盖
	system := "ECOMMERCE"
	log := domain.NewAuditLog(auditNo, system, userID, username, eventType, module, action)

	for _, opt := range opts {
		opt(log)
	}

	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveLogInTx(ctx, tx, log); err != nil {
			m.logger.ErrorContext(ctx, "failed to create audit log", "user_id", userID, "event_type", eventType, "error", err)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.AuditLogCreatedEvent{
			AuditID:   uint64(log.ID),
			AuditNo:   log.AuditNo,
			EventType: string(log.EventType),
			Module:    log.Module,
			Action:    log.Action,
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.AuditLogCreatedEventType, fmt.Sprintf("%d", log.ID), event)
	})
}

// LogOption 定义了用于配置审计日志的函数式选项类型。
type LogOption func(*domain.AuditLog)

// WithError 是一个 LogOption，用于向审计日志添加错误信息。
func WithError(errMsg string) LogOption {
	return func(l *domain.AuditLog) {
		l.SetError(errMsg)
	}
}

// WithResource 是一个 LogOption，用于向审计日志添加资源信息。
func WithResource(resourceType, resourceID string) LogOption {
	return func(l *domain.AuditLog) {
		l.SetResource(resourceType, resourceID)
	}
}

// WithChange 是一个 LogOption，用于向审计日志添加变更前后数据信息。
func WithChange(oldValue, newValue string) LogOption {
	return func(l *domain.AuditLog) {
		l.SetChange(oldValue, newValue)
	}
}

// WithClientInfo 是一个 LogOption，用于向审计日志添加客户端信息。
func WithClientInfo(ip, userAgent string) LogOption {
	return func(l *domain.AuditLog) {
		l.SetClientInfo(ip, userAgent)
	}
}

// WithDuration 是一个 LogOption，用于向审计日志添加操作耗时。
func WithDuration(duration int64) LogOption {
	return func(l *domain.AuditLog) {
		l.SetDuration(duration)
	}
}

// WithSystem 是一个 LogOption，用于设置审计日志的系统来源。
func WithSystem(system string) LogOption {
	return func(l *domain.AuditLog) {
		l.System = system
	}
}

// CreatePolicy 创建一个新的审计策略。
func (m *AuditCommandService) CreatePolicy(ctx context.Context, name, description string) (*domain.AuditPolicy, error) {
	policy := domain.NewAuditPolicy(name, description)
	err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SavePolicyInTx(ctx, tx, policy); err != nil {
			m.logger.ErrorContext(ctx, "failed to create audit policy", "name", name, "error", err)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.AuditPolicyCreatedEvent{
			PolicyID:  uint64(policy.ID),
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.AuditPolicyCreatedEventType, fmt.Sprintf("%d", policy.ID), event)
	})
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// UpdatePolicy 更新审计策略。
func (m *AuditCommandService) UpdatePolicy(ctx context.Context, id uint64, eventTypes, modules []string, enabled bool) error {
	policy, err := m.repo.GetPolicy(ctx, id)
	if err != nil {
		return err
	}

	policy.EventTypes = eventTypes
	policy.Modules = modules
	policy.Enabled = enabled
	policy.UpdatedAt = time.Now()

	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SavePolicyInTx(ctx, tx, policy); err != nil {
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.AuditPolicyUpdatedEvent{
			PolicyID:  uint64(policy.ID),
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.AuditPolicyUpdatedEventType, fmt.Sprintf("%d", policy.ID), event)
	})
}

// DeletePolicy 删除审计策略。
func (m *AuditCommandService) DeletePolicy(ctx context.Context, id uint64) error {
	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.DeletePolicyInTx(ctx, tx, id); err != nil {
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.AuditPolicyDeletedEvent{
			PolicyID:  id,
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.AuditPolicyDeletedEventType, fmt.Sprintf("%d", id), event)
	})
}

// CreateReport 创建一个新的审计报告。
func (m *AuditCommandService) CreateReport(ctx context.Context, title, description string) (*domain.AuditReport, error) {
	reportNo := fmt.Sprintf("AUDRPT%d", m.idGenerator.Generate())
	report := domain.NewAuditReport(reportNo, title, description)

	err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveReportInTx(ctx, tx, report); err != nil {
			m.logger.ErrorContext(ctx, "failed to create audit report", "title", title, "error", err)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.AuditReportCreatedEvent{
			ReportID:  uint64(report.ID),
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.AuditReportCreatedEventType, fmt.Sprintf("%d", report.ID), event)
	})
	if err != nil {
		return nil, err
	}
	return report, nil
}

// GenerateReport 生成审计报告。
func (m *AuditCommandService) GenerateReport(ctx context.Context, id uint64) error {
	report, err := m.repo.GetReport(ctx, id)
	if err != nil {
		return err
	}

	content := fmt.Sprintf("Audit Report for %s generated at %s", report.Title, time.Now().Format(time.RFC3339))
	report.Generate(content)

	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveReportInTx(ctx, tx, report); err != nil {
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.AuditReportGeneratedEvent{
			ReportID:  uint64(report.ID),
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.AuditReportGeneratedEventType, fmt.Sprintf("%d", report.ID), event)
	})
}

// DeleteReport 删除审计报告。
func (m *AuditCommandService) DeleteReport(ctx context.Context, id uint64) error {
	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.DeleteReportInTx(ctx, tx, id); err != nil {
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.AuditReportDeletedEvent{
			ReportID:  id,
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.AuditReportDeletedEventType, fmt.Sprintf("%d", id), event)
	})
}

// DeleteLogsBefore 清理历史日志。
func (m *AuditCommandService) DeleteLogsBefore(ctx context.Context, beforeTime time.Time) error {
	return m.repo.DeleteLogsBefore(ctx, beforeTime)
}
