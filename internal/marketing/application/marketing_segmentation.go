// Package application 提供了营销模块的业务逻辑处理。
package application

import (
	"context"
	"log/slog"

	couponv1 "github.com/wyfcoding/ecommerce/goapi/coupon/v1"
	"github.com/wyfcoding/ecommerce/internal/marketing/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// UserSegmentService 提供了基于 Roaring Bitmap（高效位图算法）的海量用户标签筛选与定向营销服务。
// 此服务能够以毫秒级响应完成数亿级用户的交、并、差集运算，适用于精准营销与人群画像分析。
type UserSegmentService struct {
	repo      domain.MarketingRepository          // 营销仓储，提供原始标签数据
	couponCli couponv1.CouponServiceClient        // 远程优惠券服务客户端
	logger    *slog.Logger                        // 结构化日志记录器
	tagCache  map[string]*algorithm.RoaringBitmap // 内存热点标签位图缓存
}

// NewUserSegmentService 初始化并返回一个新的用户分群服务实例。
func NewUserSegmentService(repo domain.MarketingRepository, couponCli couponv1.CouponServiceClient, logger *slog.Logger) *UserSegmentService {
	return &UserSegmentService{
		repo:      repo,
		couponCli: couponCli,
		logger:    logger,
		tagCache:  make(map[string]*algorithm.RoaringBitmap),
	}
}

// LoadTag 将用户 ID 列表加载到内存位图中。
func (s *UserSegmentService) LoadTag(tagName string, userIDs []uint32) {
	bm := algorithm.NewRoaringBitmap()
	for _, id := range userIDs {
		bm.Add(id)
	}
	s.tagCache[tagName] = bm
	s.logger.Info("tag data loaded into bitmap", "tag", tagName, "count", len(userIDs))
}

// LoadTagFromDB 同步数据库中的标签数据至内存位图缓存中。
func (s *UserSegmentService) LoadTagFromDB(ctx context.Context, tagName string) error {
	userIDs, err := s.repo.GetUserIDsByTag(ctx, tagName)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load tag data from database", "tag", tagName, "error", err)
		return err
	}

	s.LoadTag(tagName, userIDs)
	return nil
}

// TargetUsers 执行精准人群圈选。
// 逻辑：对传入的所有标签位图执行逻辑与（AND）运算，筛选出同时具备所有标签的用户 ID 集合。
func (s *UserSegmentService) TargetUsers(tags []string) []uint32 {
	if len(tags) == 0 {
		return nil
	}

	var result *algorithm.RoaringBitmap

	for _, tagName := range tags {
		bm, ok := s.tagCache[tagName]
		if !ok {
			return nil
		}

		if result == nil {
			result = algorithm.NewRoaringBitmap().Or(bm)
		} else {
			result = result.And(bm)
		}
	}

	if result == nil {
		return nil
	}

	return result.ToList()
}

// DistributeCouponsToSegment 针对选定的人群切片批量分发优惠券。
// 流程：位图圈选 -> 批量 RPC 发放 -> 进度审计日志。
func (s *UserSegmentService) DistributeCouponsToSegment(ctx context.Context, couponID uint64, tags []string) error {
	targetIDs := s.TargetUsers(tags)
	s.logger.InfoContext(ctx, "crowd segmentation finished, starting distribution",
		"target_count", len(targetIDs),
		"tags", tags,
		"coupon_id", couponID,
	)

	if len(targetIDs) == 0 {
		return nil
	}

	successCount := 0
	if s.couponCli != nil {
		for _, userID := range targetIDs {
			_, err := s.couponCli.IssueCoupon(ctx, &couponv1.IssueCouponRequest{
				UserId:   uint64(userID),
				CouponId: couponID,
			})
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to issue targeted coupon", "user_id", userID, "coupon_id", couponID, "error", err)
			} else {
				successCount++
			}
		}
	} else {
		s.logger.WarnContext(ctx, "skipping real distribution: coupon client not initialized")
	}

	s.logger.InfoContext(ctx, "targeted coupon distribution task finished",
		"total", len(targetIDs),
		"success", successCount,
		"coupon_id", couponID,
	)

	return nil
}
