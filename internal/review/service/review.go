package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"ecommerce/internal/review/model"
	"ecommerce/internal/review/repository"
	// 伪代码: 模拟 gRPC 客户端
	// orderpb "ecommerce/gen/order/v1"
)

// ReviewService 定义了评论服务的业务逻辑接口
type ReviewService interface {
	CreateReview(ctx context.Context, userID, productID, orderID uint, rating int, title, content string) (*model.Review, error)
	ListProductReviews(ctx context.Context, productID uint, page, pageSize int) ([]model.Review, int64, error)
	ApproveReview(ctx context.Context, reviewID uint) error
	AddComment(ctx context.Context, reviewID, userID uint, content string, isSeller bool) (*model.Comment, error)
	GetProductReviewStats(ctx context.Context, productID uint) (*model.ProductReviewStats, error)
}

// reviewService 是接口的具体实现
type reviewService struct {
	repo   repository.ReviewRepository
	logger *zap.Logger
	// orderClient orderpb.OrderServiceClient
}

// NewReviewService 创建一个新的 reviewService 实例
func NewReviewService(repo repository.ReviewRepository, logger *zap.Logger) ReviewService {
	return &reviewService{repo: repo, logger: logger}
}

// CreateReview 用户提交一条新评论
func (s *reviewService) CreateReview(ctx context.Context, userID, productID, orderID uint, rating int, title, content string) (*model.Review, error) {
	s.logger.Info("Creating review", zap.Uint("userID", userID), zap.Uint("productID", productID))

	// 1. 验证用户是否有权评论 (是否购买过此商品)
	// hasPurchased, err := s.verifyPurchase(ctx, userID, productID, orderID)
	// if err != nil || !hasPurchased {
	// 	 return nil, fmt.Errorf("无权评论该商品或验证失败")
	// }

	// 2. 创建评论，初始状态为 PENDING
	review := &model.Review{
		UserID:    userID,
		ProductID: productID,
		OrderID:   orderID,
		Rating:    rating,
		Title:     title,
		Content:   content,
		Status:    model.StatusPending,
	}

	if err := s.repo.CreateReview(ctx, review); err != nil {
		s.logger.Error("Failed to create review in DB", zap.Error(err))
		return nil, err
	}

	// 3. (可选) 发送消息通知运营人员审核

	return review, nil
}

// ApproveReview 审核通过一条评论
func (s *reviewService) ApproveReview(ctx context.Context, reviewID uint) error {
	s.logger.Info("Approving review", zap.Uint("reviewID", reviewID))

	// 1. 获取评论
	review, err := s.repo.GetReview(ctx, reviewID)
	if err != nil || review == nil {
		return fmt.Errorf("评论不存在")
	}

	// 2. 检查状态，防止重复操作
	if review.Status == model.StatusApproved {
		return nil // 幂等性
	}

	// 3. 在事务中更新评论状态并更新统计数据
	return s.repo.UpdateProductReviewStats(ctx, review.ProductID, review.Rating, 1)
	// 注意：这里简化了流程，实际应该先更新 review 状态，再更新统计
	// 更好的做法是在一个事务中完成这两件事
}

// ListProductReviews 获取某个商品的评论列表
func (s *reviewService) ListProductReviews(ctx context.Context, productID uint, page, pageSize int) ([]model.Review, int64, error) {
	return s.repo.ListReviewsByProduct(ctx, productID, page, pageSize)
}

// AddComment 添加回复
func (s *reviewService) AddComment(ctx context.Context, reviewID, userID uint, content string, isSeller bool) (*model.Comment, error) {
	comment := &model.Comment{
		ReviewID: reviewID,
		UserID:   userID,
		Content:  content,
		IsSeller: isSeller,
	}
	if err := s.repo.CreateComment(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

// GetProductReviewStats 获取商品的评论统计信息
func (s *reviewService) GetProductReviewStats(ctx context.Context, productID uint) (*model.ProductReviewStats, error) {
	// 此处应从 repository 获取，这里暂时省略
	panic("implement me")
}

// verifyPurchase 是一个辅助函数，用于通过 gRPC 调用订单服务验证购买记录
// func (s *reviewService) verifyPurchase(ctx context.Context, userID, productID, orderID uint) (bool, error) {
// 	 resp, err := s.orderClient.GetOrderDetails(ctx, &orderpb.GetOrderDetailsRequest{UserId: userID, OrderId: orderID})
// 	 if err != nil { return false, err }
// 	 for _, item := range resp.Order.Items {
// 		 if item.ProductId == productID {
// 			 return true, nil
// 		 }
// 	 }
// 	 return false, nil
// }