// 生成摘要：新增结算读模型投影服务，消费事件后刷新 Redis/ES 读侧。
// 假设：读模型以结算ID为主键，写模型为最终一致性来源。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
)

// SettlementProjectionService 负责将事件转换为读模型更新。
type SettlementProjectionService struct {
	repo       domain.SettlementRepository
	readRepo   domain.SettlementReadRepository
	searchRepo domain.SettlementSearchRepository
	logger     *slog.Logger
}

// NewSettlementProjectionService 创建结算投影服务。
func NewSettlementProjectionService(repo domain.SettlementRepository, readRepo domain.SettlementReadRepository, searchRepo domain.SettlementSearchRepository, logger *slog.Logger) *SettlementProjectionService {
	return &SettlementProjectionService{
		repo:       repo,
		readRepo:   readRepo,
		searchRepo: searchRepo,
		logger:     logger,
	}
}

// OnSettlementCreated 处理结算创建事件。
func (s *SettlementProjectionService) OnSettlementCreated(ctx context.Context, event *domain.SettlementCreatedEvent) error {
	return s.refreshReadModel(ctx, event.SettlementNo)
}

// OnSettlementProcessed 处理结算处理中事件。
func (s *SettlementProjectionService) OnSettlementProcessed(ctx context.Context, event *domain.SettlementProcessedEvent) error {
	return s.refreshReadModel(ctx, event.SettlementNo)
}

// OnSettlementCompleted 处理结算完成事件。
func (s *SettlementProjectionService) OnSettlementCompleted(ctx context.Context, event *domain.SettlementCompletedEvent) error {
	return s.refreshReadModel(ctx, event.SettlementNo)
}

// OnSettlementFailed 处理结算失败事件。
func (s *SettlementProjectionService) OnSettlementFailed(ctx context.Context, event *domain.SettlementFailedEvent) error {
	return s.refreshReadModel(ctx, event.SettlementNo)
}

func (s *SettlementProjectionService) refreshReadModel(ctx context.Context, settlementNo string) error {
	if settlementNo == "" {
		return nil
	}

	settlement, err := s.repo.GetSettlementByNo(ctx, settlementNo)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load settlement for projection", "settlement_no", settlementNo, "error", err)
		return err
	}

	if settlement == nil {
		if s.readRepo != nil {
			_ = s.readRepo.Delete(ctx, 0, settlementNo)
		}
		if s.searchRepo != nil {
			// 无法确定 ID 时不执行删除
		}
		return nil
	}

	if s.readRepo != nil {
		if err := s.readRepo.Save(ctx, settlement); err != nil {
			s.logger.ErrorContext(ctx, "failed to save settlement read model", "settlement_no", settlementNo, "error", err)
			return err
		}
	}
	if s.searchRepo != nil {
		if err := s.searchRepo.Index(ctx, settlement); err != nil {
			s.logger.ErrorContext(ctx, "failed to index settlement search model", "settlement_no", settlementNo, "error", err)
			return err
		}
	}

	return nil
}
