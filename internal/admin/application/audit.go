package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
)

// Audit 审计服务
type Audit struct {
	repo   domain.AuditRepository
	logger *slog.Logger
}

// NewAudit 定义了 NewAudit 相关的服务逻辑。
func NewAudit(repo domain.AuditRepository, logger *slog.Logger) *Audit {
	return &Audit{repo: repo, logger: logger}
}

// LogAction 记录操作日志 (通常建议异步调用)
func (s *Audit) LogAction(ctx context.Context, log *domain.AuditLog) {
	// 在实际生产中，这里可能会投递到 Kafka 或者异步写入，避免阻塞主业务
	// 这里为了简单，直接写入 DB，或者启动 goroutine 写入
	go func() {
		// 创建一个新的 context 避免原有 ctx 取消导致写入失败
		bgCtx := context.Background()
		if err := s.repo.Save(bgCtx, log); err != nil {
			s.logger.Error("failed to save audit log", "error", err, "action", log.Action)
		}
	}()
}

func (s *Audit) QueryLogs(ctx context.Context, filter map[string]any, page, pageSize int) ([]*domain.AuditLog, int64, error) {
	return s.repo.Find(ctx, filter, page, pageSize)
}
