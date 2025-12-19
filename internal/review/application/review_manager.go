package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/review/domain"
)

// ReviewManager 处理评论模块的写操作和核心业务流程。
type ReviewManager struct {
	repo   domain.ReviewRepository
	logger *slog.Logger
}

// NewReviewManager 创建并返回一个新的 ReviewManager 实例。
func NewReviewManager(repo domain.ReviewRepository, logger *slog.Logger) *ReviewManager {
	return &ReviewManager{
		repo:   repo,
		logger: logger,
	}
}

// CreateReview 提交一条新的评论。
func (m *ReviewManager) CreateReview(ctx context.Context, userID, productID, orderID, skuID uint64, rating int, content string, images []string) (*domain.Review, error) {
	// 简单校验：评分范围。
	if rating < 1 || rating > 5 {
		return nil, fmt.Errorf("rating must be between 1 and 5")
	}

	review := &domain.Review{
		UserID:    userID,
		ProductID: productID,
		OrderID:   orderID,
		SkuID:     skuID,
		Rating:    rating,
		Content:   content,
		Images:    domain.StringArray(images),
		Status:    domain.ReviewStatusPending, // 初始状态为待审核。
	}

	if err := m.repo.Save(ctx, review); err != nil {
		m.logger.Error("failed to save review", "error", err)
		return nil, err
	}

	return review, nil
}

// AuditReview 审核评论。
func (m *ReviewManager) AuditReview(ctx context.Context, reviewID uint64, approved bool) error {
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

	return m.repo.Save(ctx, review)
}

// DeleteReview 删除评论。
func (m *ReviewManager) DeleteReview(ctx context.Context, reviewID uint64, userID uint64) error {
	review, err := m.repo.Get(ctx, reviewID)
	if err != nil {
		return err
	}
	if review == nil {
		return fmt.Errorf("review not found")
	}

	// 权限校验：只有评论主或管理员可删除（此处简化为校验UserID）。
	if userID > 0 && review.UserID != userID {
		return fmt.Errorf("permission denied")
	}

	return m.repo.Delete(ctx, reviewID)
}
