// Package application 营销应用层
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/marketing/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// UserSegmentService 基于位图的海量用户圈选服务
type UserSegmentService struct {
	repo   domain.MarketingRepository
	logger *slog.Logger
	// 内存缓存常用的标签位图（真实生产中可存储在 Redis 中并定期同步）
	tagCache map[string]*algorithm.RoaringBitmap
}

func NewUserSegmentService(repo domain.MarketingRepository, logger *slog.Logger) *UserSegmentService {
	return &UserSegmentService{
		repo:     repo,
		logger:   logger,
		tagCache: make(map[string]*algorithm.RoaringBitmap),
	}
}

// LoadTag 手动加载标签数据
func (s *UserSegmentService) LoadTag(tagName string, userIDs []uint32) {
	bm := algorithm.NewRoaringBitmap()
	for _, id := range userIDs {
		bm.Add(id)
	}
	s.tagCache[tagName] = bm
	s.logger.Info("tag loaded into bitmap", "tag", tagName, "count", len(userIDs))
}

// LoadTagFromDB 从数据库加载并刷新标签位图
func (s *UserSegmentService) LoadTagFromDB(ctx context.Context, tagName string) error {
	userIDs, err := s.repo.GetUserIDsByTag(ctx, tagName)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load tag from db", "tag", tagName, "error", err)
		return err
	}

	s.LoadTag(tagName, userIDs)
	return nil
}

// TargetUsers 精准圈选：筛选出同时满足多个标签的用户
func (s *UserSegmentService) TargetUsers(tags []string) []uint32 {
	if len(tags) == 0 {
		return nil
	}

	var result *algorithm.RoaringBitmap

	for _, tagName := range tags {
		bm, ok := s.tagCache[tagName]
		if !ok {
			continue
		}

		if result == nil {
			// 第一个标签，克隆一份
			result = algorithm.NewRoaringBitmap().Or(bm)
		} else {
			// 执行交集运算 (AND)
			result = result.And(bm)
		}
	}

	if result == nil {
		return nil
	}

	return result.ToList()
}

// DistributeCouponsToSegment 针对选定人群批量分发优惠券
func (s *UserSegmentService) DistributeCouponsToSegment(ctx context.Context, couponID string, tags []string) error {
	// 1. 毫秒级位图运算，圈定目标人群
	targetIDs := s.TargetUsers(tags)

	s.logger.Info("segmentation finished", "target_count", len(targetIDs), "tags", tags)

	// 2. 异步批量发放（顶级架构中应发送到 Kafka）
	for _, userID := range targetIDs {
		// 模拟发券逻辑
		_ = userID
	}

	return nil
}
