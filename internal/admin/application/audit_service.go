package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
)

// AuditService 审计服务
type AuditService struct {
	repo   domain.AuditRepository
	logger *slog.Logger
}

// NewAuditService 定义了 NewAudit 相关的服务逻辑。
func NewAuditService(repo domain.AuditRepository, logger *slog.Logger) *AuditService {
	return &AuditService{repo: repo, logger: logger}
}

// LogAction 记录操作日志 (通常建议异步调用)
func (s *AuditService) LogAction(ctx context.Context, log *domain.AuditLog) {
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

func (s *AuditService) QueryLogs(ctx context.Context, filter map[string]any, page, pageSize int) ([]*domain.AuditLog, int64, error) {
	return s.repo.Find(ctx, filter, page, pageSize)
}
