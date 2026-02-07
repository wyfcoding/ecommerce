// 生成摘要：新增积分商城读模型投影服务，消费事件后刷新 Redis/ES 读侧。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain"
)

// PointsmallProjectionService 负责将积分商城事件投影到读模型。
type PointsmallProjectionService struct {
	repo             domain.PointsRepository
	productReadRepo  domain.PointsProductReadRepository
	orderReadRepo    domain.PointsOrderReadRepository
	accountReadRepo  domain.PointsAccountReadRepository
	productSearchRepo domain.PointsProductSearchRepository
	orderSearchRepo  domain.PointsOrderSearchRepository
	logger           *slog.Logger
}

// NewPointsmallProjectionService 创建投影服务。
func NewPointsmallProjectionService(
	repo domain.PointsRepository,
	productReadRepo domain.PointsProductReadRepository,
	orderReadRepo domain.PointsOrderReadRepository,
	accountReadRepo domain.PointsAccountReadRepository,
	productSearchRepo domain.PointsProductSearchRepository,
	orderSearchRepo domain.PointsOrderSearchRepository,
	logger *slog.Logger,
) *PointsmallProjectionService {
	return &PointsmallProjectionService{
		repo:              repo,
		productReadRepo:   productReadRepo,
		orderReadRepo:     orderReadRepo,
		accountReadRepo:   accountReadRepo,
		productSearchRepo: productSearchRepo,
		orderSearchRepo:   orderSearchRepo,
		logger:            logger,
	}
}

func (s *PointsmallProjectionService) OnProductCreated(ctx context.Context, event *domain.PointsProductCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshProduct(ctx, event.ProductID)
}

func (s *PointsmallProjectionService) OnStockUpdated(ctx context.Context, event *domain.PointsStockUpdatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshProduct(ctx, event.ItemID)
}

func (s *PointsmallProjectionService) OnOrderCreated(ctx context.Context, event *domain.PointsOrderCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshOrder(ctx, event.OrderID)
}

func (s *PointsmallProjectionService) OnAccountUpdated(ctx context.Context, event *domain.PointsAccountUpdatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshAccount(ctx, event.UserID)
}

func (s *PointsmallProjectionService) refreshProduct(ctx context.Context, productID uint64) error {
	if s.productReadRepo == nil && s.productSearchRepo == nil {
		return nil
	}
	product, err := s.repo.GetProduct(ctx, productID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load product for projection", "product_id", productID, "error", err)
		return err
	}
	if product == nil {
		if s.productReadRepo != nil {
			_ = s.productReadRepo.Delete(ctx, productID)
		}
		if s.productSearchRepo != nil {
			_ = s.productSearchRepo.Delete(ctx, productID)
		}
		return nil
	}
	if s.productReadRepo != nil {
		if err := s.productReadRepo.Save(ctx, product); err != nil {
			s.logger.ErrorContext(ctx, "failed to save product cache", "product_id", productID, "error", err)
			return err
		}
	}
	if s.productSearchRepo != nil {
		if err := s.productSearchRepo.Index(ctx, product); err != nil {
			s.logger.ErrorContext(ctx, "failed to index product", "product_id", productID, "error", err)
			return err
		}
	}
	return nil
}

func (s *PointsmallProjectionService) refreshOrder(ctx context.Context, orderID uint64) error {
	if s.orderReadRepo == nil && s.orderSearchRepo == nil {
		return nil
	}
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load order for projection", "order_id", orderID, "error", err)
		return err
	}
	if order == nil {
		if s.orderReadRepo != nil {
			_ = s.orderReadRepo.Delete(ctx, orderID)
		}
		if s.orderSearchRepo != nil {
			_ = s.orderSearchRepo.Delete(ctx, orderID)
		}
		return nil
	}
	if s.orderReadRepo != nil {
		if err := s.orderReadRepo.Save(ctx, order); err != nil {
			s.logger.ErrorContext(ctx, "failed to save order cache", "order_id", orderID, "error", err)
			return err
		}
	}
	if s.orderSearchRepo != nil {
		if err := s.orderSearchRepo.Index(ctx, order); err != nil {
			s.logger.ErrorContext(ctx, "failed to index order", "order_id", orderID, "error", err)
			return err
		}
	}
	return nil
}

func (s *PointsmallProjectionService) refreshAccount(ctx context.Context, userID uint64) error {
	if s.accountReadRepo == nil {
		return nil
	}
	account, err := s.repo.GetAccount(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load account for projection", "user_id", userID, "error", err)
		return err
	}
	if account == nil {
		_ = s.accountReadRepo.DeleteByUserID(ctx, userID)
		return nil
	}
	if err := s.accountReadRepo.Save(ctx, account); err != nil {
		s.logger.ErrorContext(ctx, "failed to save account cache", "user_id", userID, "error", err)
		return err
	}
	return nil
}
