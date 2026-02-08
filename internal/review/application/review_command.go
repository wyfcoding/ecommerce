package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	orderv1 "github.com/wyfcoding/ecommerce/goapi/order/v1"
	"github.com/wyfcoding/ecommerce/internal/review/domain"
	algorithm "github.com/wyfcoding/pkg/algorithm/structures"
	"github.com/wyfcoding/pkg/contextx"
	"github.com/wyfcoding/pkg/messagequeue"
)

// ReviewCommandService 处理评论模块的写操作和核心业务流程。
type ReviewCommandService struct {
	repo        domain.ReviewRepository
	publisher   messagequeue.EventPublisher
	logger      *slog.Logger
	simHash     *algorithm.SimHash
	orderClient orderv1.OrderServiceClient
}

// NewReviewCommandService 创建并返回一个新的 ReviewCommandService 实例。
func NewReviewCommandService(
	repo domain.ReviewRepository,
	publisher messagequeue.EventPublisher,
	logger *slog.Logger,
	orderClient orderv1.OrderServiceClient,
) *ReviewCommandService {
	return &ReviewCommandService{
		repo:        repo,
		publisher:   publisher,
		logger:      logger,
		simHash:     algorithm.NewSimHash(),
		orderClient: orderClient,
	}
}

// CreateReview 提交一条新的评论。
func (m *ReviewCommandService) CreateReview(ctx context.Context, userID, productID, orderID, skuID uint64, rating int, content string, images []string) (*domain.Review, error) {
	// 1. 校验订单状态：仅允许已完成的订单评价
	if m.orderClient != nil && orderID > 0 {
		order, err := m.orderClient.GetOrderByID(ctx, &orderv1.GetOrderByIDRequest{Id: orderID})
		if err != nil {
			return nil, fmt.Errorf("failed to verify order status: %w", err)
		}
		if order == nil {
			return nil, fmt.Errorf("order not found")
		}
		if order.Status != orderv1.OrderStatus_COMPLETED {
			return nil, fmt.Errorf("only completed orders can be reviewed (current status: %s)", order.Status.String())
		}
		if order.UserId != userID {
			return nil, fmt.Errorf("permission denied: order does not belong to user")
		}
	}

	// 简单校验：评分范围。
	if rating < 1 || rating > 5 {
		return nil, fmt.Errorf("rating must be between 1 and 5")
	}

	// --- 查重逻辑集成 ---
	// 获取该商品最近的几条评论进行相似度对比
	recentReviews, _, err := m.repo.List(ctx, productID, nil, 0, 20)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to list recent reviews for spam check", "error", err, "product_id", productID)
		// 如果获取列表失败，由于是辅助查重，可以选择继续或者报错。
		// 这里选择报错，以保证数据的严谨性。
		return nil, err
	}
	newHash := m.simHash.Calculate(content, algorithm.DefaultTokenizer)

	isSpam := false
	for _, r := range recentReviews {
		existingHash := m.simHash.Calculate(r.Content, algorithm.DefaultTokenizer)
		// 海明距离 <= 3 通常认为高度相似
		if m.simHash.HammingDistance(newHash, existingHash) <= 3 {
			isSpam = true
			break
		}
	}

	status := domain.ReviewStatusPending
	if isSpam {
		m.logger.WarnContext(ctx, "suspected spam review detected", "user_id", userID, "product_id", productID)
		// 策略：如果是垃圾内容，可以设为拒绝，或者进入人工审核队列
		status = domain.ReviewStatusRejected
	}

	review := &domain.Review{
		UserID:    userID,
		ProductID: productID,
		OrderID:   orderID,
		SkuID:     skuID,
		Rating:    rating,
		Content:   content,
		Images:    domain.StringArray(images),
		Status:    status,
	}

	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveInTx(ctx, tx, review); err != nil {
			return err
		}
		event := &domain.ReviewCreatedEvent{
			ReviewID:  review.ID,
			UserID:    review.UserID,
			ProductID: review.ProductID,
			Rating:    int32(review.Rating),
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.ReviewCreatedEventType, fmt.Sprintf("%d", review.ID), event)
	}); err != nil {
		m.logger.Error("failed to save review", "error", err)
		return nil, err
	}

	return review, nil
}

// AuditReview 审核评论。
func (m *ReviewCommandService) AuditReview(ctx context.Context, reviewID uint64, approved bool) error {
	review, err := m.repo.Get(ctx, reviewID)
	if err != nil {
		return err
	}
	if review == nil {
		return fmt.Errorf("review not found")
	}

	if approved {
		review.Status = domain.ReviewStatusApproved
	} else {
		review.Status = domain.ReviewStatusRejected
	}

	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveInTx(ctx, tx, review); err != nil {
			return err
		}
		event := &domain.ReviewUpdatedEvent{
			ReviewID:  review.ID,
			Rating:    int32(review.Rating),
			Status:    int32(review.Status),
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.ReviewUpdatedEventType, fmt.Sprintf("%d", review.ID), event)
	})
}

// DeleteReview 删除评论。
func (m *ReviewCommandService) DeleteReview(ctx context.Context, reviewID uint64, userID uint64) error {
	review, err := m.repo.Get(ctx, reviewID)
	if err != nil {
		return err
	}
	if review == nil {
		return fmt.Errorf("review not found")
	}

	// 权限校验：
	// 1. 评论所有者可以删除
	// 2. 管理员 (ADMIN/SUPER_ADMIN) 可以删除
	if review.UserID != userID {
		role := contextx.GetRole(ctx)
		if role != "ADMIN" && role != "SUPER_ADMIN" {
			return fmt.Errorf("permission denied: only owner or admin can delete reviews")
		}
	}

	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.DeleteInTx(ctx, tx, reviewID); err != nil {
			return err
		}
		event := &domain.ReviewDeletedEvent{
			ReviewID:  review.ID,
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.ReviewDeletedEventType, fmt.Sprintf("%d", review.ID), event)
	})
}
