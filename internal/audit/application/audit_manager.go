package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/audit/domain"
	"github.com/wyfcoding/pkg/algorithm"
	"github.com/wyfcoding/pkg/idgen"
)

// AuditManager 处理审计模块的写操作和业务逻辑。
type AuditManager struct {
	repo        domain.AuditRepository
	idGenerator idgen.Generator
	logger      *slog.Logger
}

// NewAuditManager 创建并返回一个新的 AuditManager 实例。
func NewAuditManager(repo domain.AuditRepository, idGenerator idgen.Generator, logger *slog.Logger) *AuditManager {
	return &AuditManager{
		repo:        repo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// SealLogs 为最近的日志生成“数字封条”（Merkle Root）
func (m *AuditManager) SealLogs(ctx context.Context, limit int) (string, error) {
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
		data[i] = []byte(fmt.Sprintf("%s|%d|%s|%s", l.AuditNo, l.UserID, l.Action, l.CreatedAt.Format(time.RFC3339)))
	}

	// 3. 构建 Merkle Tree 并获取根哈希
	tree := algorithm.NewMerkleTree(data)
	rootHash := tree.RootHashHex()

	m.logger.InfoContext(ctx, "audit logs sealed with merkle root", "count", len(logs), "root_hash", rootHash)
	return rootHash, nil
}

// LogEvent 记录一个审计事件。
func (m *AuditManager) LogEvent(ctx context.Context, userID uint64, username string, eventType domain.AuditEventType, module, action string, opts ...LogOption) error {
	auditNo := fmt.Sprintf("AUD%d", m.idGenerator.Generate())
	log := domain.NewAuditLog(auditNo, userID, username, eventType, module, action)

	for _, opt := range opts {
		opt(log)
	}

	if err := m.repo.CreateLog(ctx, log); err != nil {
		m.logger.ErrorContext(ctx, "failed to create audit log", "user_id", userID, "event_type", eventType, "error", err)
		return err
	}
	return nil
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

// CreatePolicy 创建一个新的审计策略。
func (m *AuditManager) CreatePolicy(ctx context.Context, name, description string) (*domain.AuditPolicy, error) {
	policy := domain.NewAuditPolicy(name, description)
	if err := m.repo.CreatePolicy(ctx, policy); err != nil {
		m.logger.ErrorContext(ctx, "failed to create audit policy", "name", name, "error", err)
		return nil, err
	}
	return policy, nil
}

// UpdatePolicy 更新审计策略。
func (m *AuditManager) UpdatePolicy(ctx context.Context, id uint64, eventTypes, modules []string, enabled bool) error {
	policy, err := m.repo.GetPolicy(ctx, id)
	if err != nil {
		return err
	}

	policy.EventTypes = eventTypes
	policy.Modules = modules
	policy.Enabled = enabled
	policy.UpdatedAt = time.Now()

	return m.repo.UpdatePolicy(ctx, policy)
}

// DeletePolicy 删除审计策略。
func (m *AuditManager) DeletePolicy(ctx context.Context, id uint64) error {
	return m.repo.DeletePolicy(ctx, id)
}

// CreateReport 创建一个新的审计报告。
func (m *AuditManager) CreateReport(ctx context.Context, title, description string) (*domain.AuditReport, error) {
	reportNo := fmt.Sprintf("AUDRPT%d", m.idGenerator.Generate())
	report := domain.NewAuditReport(reportNo, title, description)

	if err := m.repo.CreateReport(ctx, report); err != nil {
		m.logger.ErrorContext(ctx, "failed to create audit report", "title", title, "error", err)
		return nil, err
	}
	return report, nil
}

// GenerateReport 生成审计报告。
func (m *AuditManager) GenerateReport(ctx context.Context, id uint64) error {
	report, err := m.repo.GetReport(ctx, id)
	if err != nil {
		return err
	}

	content := fmt.Sprintf("Audit Report for %s generated at %s", report.Title, time.Now().Format(time.RFC3339))
	report.Generate(content)

	return m.repo.UpdateReport(ctx, report)
}

// DeleteReport 删除审计报告。
func (m *AuditManager) DeleteReport(ctx context.Context, id uint64) error {
	return m.repo.DeleteReport(ctx, id)
}

// DeleteLogsBefore 清理历史日志。
func (m *AuditManager) DeleteLogsBefore(ctx context.Context, beforeTime time.Time) error {
	return m.repo.DeleteLogsBefore(ctx, beforeTime)
}
