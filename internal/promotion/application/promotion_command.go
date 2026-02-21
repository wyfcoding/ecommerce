// 变更说明：
// 新增促销引擎应用层命令服务。
// 负责促销活动的创建、激活、暂停、计算等写操作。
package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/promotion/domain"
	"github.com/wyfcoding/pkg/messagequeue"
)

// PromotionCommandService 促销写操作服务。
type PromotionCommandService struct {
	repo      domain.PromotionRepository
	publisher messagequeue.EventPublisher
	engine    *domain.PromotionEngine
	logger    *slog.Logger
}

// NewPromotionCommandService 创建促销命令服务。
func NewPromotionCommandService(
	repo domain.PromotionRepository,
	publisher messagequeue.EventPublisher,
	logger *slog.Logger,
) *PromotionCommandService {
	return &PromotionCommandService{
		repo:      repo,
		publisher: publisher,
		engine:    domain.NewPromotionEngine(),
		logger:    logger,
	}
}

// CreatePromotionCmd 创建促销活动命令。
type CreatePromotionCmd struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Type        domain.PromotionType `json:"type"`
	Scope       domain.PromotionScope `json:"scope"`
	StackMode   domain.StackMode    `json:"stack_mode"`
	Priority    int32               `json:"priority"`
	StartTime   time.Time           `json:"start_time"`
	EndTime     time.Time           `json:"end_time"`
	ScopeIDs    []uint64            `json:"scope_ids"`
	ExcludeIDs  []uint64            `json:"exclude_ids"`
	MerchantID  uint64              `json:"merchant_id"`
	UsageLimit  int64               `json:"usage_limit"`
	UserLimit   int32               `json:"user_limit"`
	Label       string              `json:"label"`
	Rules       []CreateRuleCmd     `json:"rules"`
}

// CreateRuleCmd 创建促销规则命令。
type CreateRuleCmd struct {
	Threshold  decimal.Decimal `json:"threshold"`
	Discount   decimal.Decimal `json:"discount"`
	GiftSKUID  uint64          `json:"gift_sku_id"`
	GiftQty    int32           `json:"gift_qty"`
	FixedPrice decimal.Decimal `json:"fixed_price"`
	SortOrder  int32           `json:"sort_order"`
}

// CreatePromotion 创建促销活动。
func (s *PromotionCommandService) CreatePromotion(ctx context.Context, cmd *CreatePromotionCmd) (*domain.Promotion, error) {
	promo := &domain.Promotion{
		Name:        cmd.Name,
		Description: cmd.Description,
		Type:        cmd.Type,
		Status:      domain.PromotionStatusDraft,
		Scope:       cmd.Scope,
		StackMode:   cmd.StackMode,
		Priority:    cmd.Priority,
		StartTime:   cmd.StartTime,
		EndTime:     cmd.EndTime,
		ScopeIDs:    cmd.ScopeIDs,
		ExcludeIDs:  cmd.ExcludeIDs,
		MerchantID:  cmd.MerchantID,
		UsageLimit:  cmd.UsageLimit,
		UserLimit:   cmd.UserLimit,
		Label:       cmd.Label,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 转换规则。
	rules := make([]*domain.PromotionRule, 0, len(cmd.Rules))
	for _, r := range cmd.Rules {
		rules = append(rules, &domain.PromotionRule{
			Threshold:  r.Threshold,
			Discount:   r.Discount,
			GiftSKUID:  r.GiftSKUID,
			GiftQty:    r.GiftQty,
			FixedPrice: r.FixedPrice,
			SortOrder:  r.SortOrder,
		})
	}
	promo.Rules = rules

	if err := s.repo.Save(ctx, promo); err != nil {
		s.logger.ErrorContext(ctx, "failed to save promotion", "name", cmd.Name, "error", err)
		return nil, fmt.Errorf("save promotion: %w", err)
	}

	// 发布创建事件。
	if s.publisher != nil {
		event := &domain.PromotionCreatedEvent{
			PromotionID: promo.ID,
			Name:        promo.Name,
			Type:        promo.Type,
			StartTime:   promo.StartTime,
			EndTime:     promo.EndTime,
			Timestamp:   time.Now(),
		}
		if err := s.publisher.Publish(ctx, domain.PromotionCreatedEventType, fmt.Sprintf("%d", promo.ID), event); err != nil {
			s.logger.WarnContext(ctx, "failed to publish promotion created event", "promotion_id", promo.ID, "error", err)
		}
	}

	s.logger.InfoContext(ctx, "promotion created", "promotion_id", promo.ID, "name", promo.Name, "type", promo.Type)
	return promo, nil
}

// CalculateCartPromotions 计算购物车促销优惠。
// 这是促销引擎的核心入口，由订单服务在下单前调用。
func (s *PromotionCommandService) CalculateCartPromotions(ctx context.Context, items []*domain.CartItem) (*domain.PromotionCalculation, error) {
	now := time.Now()

	// 1. 获取所有生效中的促销活动。
	promotions, err := s.repo.ListActive(ctx, now)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list active promotions", "error", err)
		return nil, fmt.Errorf("list active promotions: %w", err)
	}

	// 2. 调用促销引擎计算。
	result := s.engine.Calculate(items, promotions)

	s.logger.InfoContext(ctx, "cart promotions calculated",
		"item_count", len(items),
		"promotion_count", len(result.Promotions),
		"original_amount", result.OriginalAmount.String(),
		"total_discount", result.TotalDiscount.String(),
		"final_amount", result.FinalAmount.String(),
	)

	return result, nil
}

// ActivatePromotion 激活促销活动。
func (s *PromotionCommandService) ActivatePromotion(ctx context.Context, id uint64) error {
	promo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get promotion: %w", err)
	}
	if promo == nil {
		return domain.ErrPromotionNotFound
	}

	promo.Status = domain.PromotionStatusActive
	promo.UpdatedAt = time.Now()

	if err := s.repo.Save(ctx, promo); err != nil {
		return fmt.Errorf("save promotion: %w", err)
	}

	s.logger.InfoContext(ctx, "promotion activated", "promotion_id", id)
	return nil
}

// PausePromotion 暂停促销活动。
func (s *PromotionCommandService) PausePromotion(ctx context.Context, id uint64) error {
	promo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get promotion: %w", err)
	}
	if promo == nil {
		return domain.ErrPromotionNotFound
	}

	promo.Status = domain.PromotionStatusPaused
	promo.UpdatedAt = time.Now()

	if err := s.repo.Save(ctx, promo); err != nil {
		return fmt.Errorf("save promotion: %w", err)
	}

	s.logger.InfoContext(ctx, "promotion paused", "promotion_id", id)
	return nil
}

// RecordUsage 记录促销使用（下单成功后调用）。
func (s *PromotionCommandService) RecordUsage(ctx context.Context, promotionID, orderID, userID uint64, discountAmt int64) error {
	if err := s.repo.IncrementUsage(ctx, promotionID); err != nil {
		s.logger.ErrorContext(ctx, "failed to increment promotion usage", "promotion_id", promotionID, "error", err)
		return err
	}

	if s.publisher != nil {
		event := &domain.PromotionUsedEvent{
			PromotionID: promotionID,
			OrderID:     orderID,
			UserID:      userID,
			DiscountAmt: discountAmt,
			Timestamp:   time.Now(),
		}
		if err := s.publisher.Publish(ctx, domain.PromotionUsedEventType, fmt.Sprintf("%d", promotionID), event); err != nil {
			s.logger.WarnContext(ctx, "failed to publish promotion used event", "promotion_id", promotionID, "error", err)
		}
	}

	return nil
}
