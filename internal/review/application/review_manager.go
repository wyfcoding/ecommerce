package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/review/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// ReviewManager 处理评论模块的写操作和核心业务流程。
type ReviewManager struct {
	repo    domain.ReviewRepository
	logger  *slog.Logger
	simHash *algorithm.SimHash
}

// NewReviewManager 创建并返回一个新的 ReviewManager 实例。
func NewReviewManager(repo domain.ReviewRepository, logger *slog.Logger) *ReviewManager {
	return &ReviewManager{
		repo:    repo,
		logger:  logger,
		simHash: algorithm.NewSimHash(),
	}
}

// CreateReview 提交一条新的评论。
func (m *ReviewManager) CreateReview(ctx context.Context, userID, productID, orderID, skuID uint64, rating int, content string, images []string) (*domain.Review, error) {
	// 简单校验：评分范围。
	if rating < 1 || rating > 5 {
		return nil, fmt.Errorf("rating must be between 1 and 5")
	}

	// --- 查重逻辑集成 ---
	// 获取该商品最近的几条评论进行相似度对比
	recentReviews, _, _ := m.repo.List(ctx, productID, nil, 0, 20)
	newHash := m.simHash.Calculate(content)

	isSpam := false
	for _, r := range recentReviews {
		existingHash := m.simHash.Calculate(r.Content)
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
