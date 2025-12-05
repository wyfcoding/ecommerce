package application

import (
	"context"
	"errors" // 导入标准错误处理库。

	"github.com/wyfcoding/ecommerce/internal/review/domain/entity"     // 导入评论领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/review/domain/repository" // 导入评论领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// ReviewService 结构体定义了评论管理相关的应用服务。
// 它协调领域层和基础设施层，处理评论的创建、审核、查询以及商品评分统计等业务逻辑。
type ReviewService struct {
	repo   repository.ReviewRepository // 依赖ReviewRepository接口，用于数据持久化操作。
	logger *slog.Logger                // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewReviewService 创建并返回一个新的 ReviewService 实例。
func NewReviewService(repo repository.ReviewRepository, logger *slog.Logger) *ReviewService {
	return &ReviewService{
		repo:   repo,
		logger: logger,
	}
}

// CreateReview 创建一条新的评论。
// ctx: 上下文。
// userID: 评论用户ID。
// productID: 评论商品ID。
// orderID: 关联订单ID。
// skuID: 关联SKU ID。
// rating: 评分（1-5）。
// content: 评论内容。
// images: 评论图片URL列表。
// 返回创建成功的Review实体和可能发生的错误。
func (s *ReviewService) CreateReview(ctx context.Context, userID, productID, orderID, skuID uint64, rating int, content string, images []string) (*entity.Review, error) {
	// 基础输入验证。
	if rating < 1 || rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}
	if content == "" {
		return nil, errors.New("content cannot be empty")
	}

	review := &entity.Review{
		UserID:    userID,
		ProductID: productID,
		OrderID:   orderID,
		SkuID:     skuID,
		Rating:    rating,
		Content:   content,
		Images:    images,
		Status:    entity.ReviewStatusPending, // 新评论默认为待审核状态。
	}

	// 通过仓储接口保存评论。
	if err := s.repo.Save(ctx, review); err != nil {
		s.logger.Error("failed to save review", "error", err)
		return nil, err
	}
	s.logger.Info("review created successfully", "review_id", review.ID)
	return review, nil
}

// ApproveReview 审核通过指定ID的评论。
// ctx: 上下文。
// id: 评论ID。
// 返回可能发生的错误。
func (s *ReviewService) ApproveReview(ctx context.Context, id uint64) error {
	review, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if review == nil {
		return errors.New("review not found")
	}

	review.Status = entity.ReviewStatusApproved // 状态变更为已通过。
	// 通过仓储接口保存更新后的评论。
	return s.repo.Save(ctx, review)
}

// RejectReview 审核拒绝指定ID的评论。
// ctx: 上下文。
// id: 评论ID。
// 返回可能发生的错误。
func (s *ReviewService) RejectReview(ctx context.Context, id uint64) error {
	review, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if review == nil {
		return errors.New("review not found")
	}

	review.Status = entity.ReviewStatusRejected // 状态变更为已拒绝。
	// 通过仓储接口保存更新后的评论。
	return s.repo.Save(ctx, review)
}

// ListReviews 获取评论列表。
// ctx: 上下文。
// productID: 筛选评论的商品ID。
// status: 筛选评论的状态。
// page, pageSize: 分页参数。
// 返回评论列表、总数和可能发生的错误。
func (s *ReviewService) ListReviews(ctx context.Context, productID uint64, status *int, page, pageSize int) ([]*entity.Review, int64, error) {
	offset := (page - 1) * pageSize
	var st *entity.ReviewStatus
	if status != nil { // 如果提供了状态，则按状态过滤。
		s := entity.ReviewStatus(*status)
		st = &s
	}
	return s.repo.List(ctx, productID, st, offset, pageSize)
}

// ListUserReviews 获取用户的评论列表。
// ctx: 上下文。
// userID: 筛选评论的用户ID。
// page, pageSize: 分页参数。
// 返回评论列表、总数和可能发生的错误。
func (s *ReviewService) ListUserReviews(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.Review, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListByUser(ctx, userID, offset, pageSize)
}

// GetProductStats 获取指定商品的评分统计数据。
// ctx: 上下文。
// productID: 商品ID。
// 返回ProductRatingStats实体和可能发生的错误。
func (s *ReviewService) GetProductStats(ctx context.Context, productID uint64) (*entity.ProductRatingStats, error) {
	return s.repo.GetProductStats(ctx, productID)
}
