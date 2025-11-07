package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ecommerce/internal/groupbuy/model"
	"ecommerce/internal/groupbuy/repository"
	"ecommerce/pkg/idgen"
)

var (
	ErrActivityNotFound      = errors.New("拼团活动不存在")
	ErrActivityNotStarted    = errors.New("拼团活动未开始")
	ErrActivityEnded         = errors.New("拼团活动已结束")
	ErrActivitySoldOut       = errors.New("拼团活动已售罄")
	ErrGroupNotFound         = errors.New("拼团不存在")
	ErrGroupFull             = errors.New("拼团已满")
	ErrGroupExpired          = errors.New("拼团已过期")
	ErrAlreadyJoined         = errors.New("已参与该拼团")
	ErrPurchaseLimitExceeded = errors.New("超过购买限制")
)

// GroupBuyService 拼团服务接口
type GroupBuyService interface {
	// 活动管理
	CreateActivity(ctx context.Context, activity *model.GroupBuyActivity) (*model.GroupBuyActivity, error)
	UpdateActivity(ctx context.Context, activity *model.GroupBuyActivity) (*model.GroupBuyActivity, error)
	GetActivity(ctx context.Context, id uint64) (*model.GroupBuyActivity, error)
	ListActivities(ctx context.Context, status string, pageSize, pageNum int32) ([]*model.GroupBuyActivity, int64, error)
	
	// 拼团操作
	StartGroup(ctx context.Context, userID, activityID uint64) (*model.Group, error)
	JoinGroup(ctx context.Context, userID, groupID uint64) (*model.Group, error)
	GetGroup(ctx context.Context, groupID uint64) (*model.Group, error)
	ListGroupsByActivity(ctx context.Context, activityID uint64, status string, pageSize, pageNum int32) ([]*model.Group, int64, error)
	ListUserGroups(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.Group, int64, error)
	
	// 拼团成员
	GetGroupMembers(ctx context.Context, groupID uint64) ([]*model.GroupMember, error)
	
	// 拼团状态检查
	CheckGroupStatus(ctx context.Context, groupID uint64) (*GroupStatusInfo, error)
	ProcessExpiredGroups(ctx context.Context) (int, error)
}

// GroupStatusInfo 拼团状态信息
type GroupStatusInfo struct {
	Group           *model.Group
	Members         []*model.GroupMember
	RemainingSlots  int32
	RemainingTime   time.Duration
	IsComplete      bool
	IsExpired       bool
}

type groupBuyService struct {
	repo        repository.GroupBuyRepo
	redisClient *redis.Client
	logger      *zap.Logger
}

// NewGroupBuyService 创建拼团服务实例
func NewGroupBuyService(
	repo repository.GroupBuyRepo,
	redisClient *redis.Client,
	logger *zap.Logger,
) GroupBuyService {
	return &groupBuyService{
		repo:        repo,
		redisClient: redisClient,
		logger:      logger,
	}
}

// CreateActivity 创建拼团活动
func (s *groupBuyService) CreateActivity(ctx context.Context, activity *model.GroupBuyActivity) (*model.GroupBuyActivity, error) {
	// 验证时间
	if activity.StartTime.After(activity.EndTime) {
		return nil, fmt.Errorf("开始时间不能晚于结束时间")
	}

	// 验证价格
	if activity.GroupPrice >= activity.OriginalPrice {
		return nil, fmt.Errorf("拼团价必须低于原价")
	}

	activity.Status = model.GroupBuyActivityStatusPending
	activity.SoldCount = 0
	activity.CreatedAt = time.Now()
	activity.UpdatedAt = time.Now()

	if err := s.repo.CreateActivity(ctx, activity); err != nil {
		s.logger.Error("创建拼团活动失败", zap.Error(err))
		return nil, err
	}

	s.logger.Info("创建拼团活动成功", zap.Uint64("activityID", activity.ID))
	return activity, nil
}

// UpdateActivity 更新拼团活动
func (s *groupBuyService) UpdateActivity(ctx context.Context, activity *model.GroupBuyActivity) (*model.GroupBuyActivity, error) {
	existing, err := s.repo.GetActivityByID(ctx, activity.ID)
	if err != nil {
		return nil, ErrActivityNotFound
	}

	// 已开始的活动不允许修改关键信息
	if existing.Status == model.GroupBuyActivityStatusOngoing {
		return nil, fmt.Errorf("进行中的活动不允许修改")
	}

	activity.UpdatedAt = time.Now()
	if err := s.repo.UpdateActivity(ctx, activity); err != nil {
		s.logger.Error("更新拼团活动失败", zap.Error(err))
		return nil, err
	}

	return activity, nil
}

// GetActivity 获取拼团活动详情
func (s *groupBuyService) GetActivity(ctx context.Context, id uint64) (*model.GroupBuyActivity, error) {
	activity, err := s.repo.GetActivityByID(ctx, id)
	if err != nil {
		return nil, ErrActivityNotFound
	}
	return activity, nil
}

// ListActivities 获取拼团活动列表
func (s *groupBuyService) ListActivities(ctx context.Context, status string, pageSize, pageNum int32) ([]*model.GroupBuyActivity, int64, error) {
	activities, total, err := s.repo.ListActivities(ctx, status, pageSize, pageNum)
	if err != nil {
		s.logger.Error("获取拼团活动列表失败", zap.Error(err))
		return nil, 0, err
	}
	return activities, total, nil
}

// StartGroup 发起拼团
func (s *groupBuyService) StartGroup(ctx context.Context, userID, activityID uint64) (*model.Group, error) {
	// 1. 验证活动
	activity, err := s.repo.GetActivityByID(ctx, activityID)
	if err != nil {
		return nil, ErrActivityNotFound
	}

	now := time.Now()
	if now.Before(activity.StartTime) {
		return nil, ErrActivityNotStarted
	}
	if now.After(activity.EndTime) {
		return nil, ErrActivityEnded
	}
	if activity.Status != model.GroupBuyActivityStatusOngoing {
		return nil, fmt.Errorf("活动状态异常")
	}

	// 2. 检查库存
	if activity.SoldCount >= activity.StockQuantity {
		return nil, ErrActivitySoldOut
	}

	// 3. 检查用户购买限制
	userPurchased, err := s.repo.GetUserPurchaseCount(ctx, userID, activityID)
	if err != nil {
		s.logger.Error("获取用户购买记录失败", zap.Error(err))
	}
	if userPurchased >= activity.LimitPerUser {
		return nil, ErrPurchaseLimitExceeded
	}

	// 4. 创建拼团
	groupNo := fmt.Sprintf("GB%d", idgen.GenID())
	expireTime := now.Add(time.Duration(activity.TimeLimit) * time.Hour)

	group := &model.Group{
		GroupNo:         groupNo,
		ActivityID:      activityID,
		LeaderID:        userID,
		RequiredMembers: activity.RequiredMembers,
		CurrentMembers:  1,
		Status:          model.GroupStatusRecruiting,
		ExpireTime:      expireTime,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.CreateGroup(ctx, group); err != nil {
		s.logger.Error("创建拼团失败", zap.Error(err))
		return nil, err
	}

	// 5. 添加团长为成员
	member := &model.GroupMember{
		GroupID:   group.ID,
		UserID:    userID,
		IsLeader:  true,
		JoinTime:  now,
		CreatedAt: now,
	}

	if err := s.repo.AddGroupMember(ctx, member); err != nil {
		s.logger.Error("添加团长失败", zap.Error(err))
		return nil, err
	}

	// 6. 检查是否立即成团（如果只需要1人）
	if group.IsComplete() {
		s.completeGroup(ctx, group)
	}

	s.logger.Info("发起拼团成功",
		zap.String("groupNo", groupNo),
		zap.Uint64("userID", userID),
		zap.Uint64("activityID", activityID))

	return group, nil
}

// JoinGroup 加入拼团
func (s *groupBuyService) JoinGroup(ctx context.Context, userID, groupID uint64) (*model.Group, error) {
	// 1. 获取拼团信息
	group, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, ErrGroupNotFound
	}

	// 2. 验证拼团状态
	if !group.CanJoin() {
		if group.IsComplete() {
			return nil, ErrGroupFull
		}
		if group.IsExpired() {
			return nil, ErrGroupExpired
		}
		return nil, fmt.Errorf("拼团状态异常")
	}

	// 3. 检查是否已参与
	isMember, err := s.repo.IsGroupMember(ctx, groupID, userID)
	if err != nil {
		s.logger.Error("检查成员失败", zap.Error(err))
	}
	if isMember {
		return nil, ErrAlreadyJoined
	}

	// 4. 验证活动
	activity, err := s.repo.GetActivityByID(ctx, group.ActivityID)
	if err != nil {
		return nil, ErrActivityNotFound
	}

	// 5. 检查用户购买限制
	userPurchased, err := s.repo.GetUserPurchaseCount(ctx, userID, activity.ID)
	if err != nil {
		s.logger.Error("获取用户购买记录失败", zap.Error(err))
	}
	if userPurchased >= activity.LimitPerUser {
		return nil, ErrPurchaseLimitExceeded
	}

	// 6. 使用Redis确保并发安全
	lockKey := fmt.Sprintf("groupbuy:lock:%d", groupID)
	locked, err := s.redisClient.SetNX(ctx, lockKey, 1, 5*time.Second).Result()
	if err != nil || !locked {
		return nil, fmt.Errorf("系统繁忙，请稍后重试")
	}
	defer s.redisClient.Del(ctx, lockKey)

	// 7. 再次检查人数（防止并发）
	group, err = s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.CurrentMembers >= group.RequiredMembers {
		return nil, ErrGroupFull
	}

	// 8. 添加成员
	member := &model.GroupMember{
		GroupID:   groupID,
		UserID:    userID,
		IsLeader:  false,
		JoinTime:  time.Now(),
		CreatedAt: time.Now(),
	}

	if err := s.repo.AddGroupMember(ctx, member); err != nil {
		s.logger.Error("添加成员失败", zap.Error(err))
		return nil, err
	}

	// 9. 更新拼团人数
	group.CurrentMembers++
	group.UpdatedAt = time.Now()

	// 10. 检查是否成团
	if group.IsComplete() {
		s.completeGroup(ctx, group)
	} else {
		if err := s.repo.UpdateGroup(ctx, group); err != nil {
			s.logger.Error("更新拼团失败", zap.Error(err))
			return nil, err
		}
	}

	s.logger.Info("加入拼团成功",
		zap.Uint64("groupID", groupID),
		zap.Uint64("userID", userID))

	return group, nil
}

// GetGroup 获取拼团详情
func (s *groupBuyService) GetGroup(ctx context.Context, groupID uint64) (*model.Group, error) {
	group, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, ErrGroupNotFound
	}
	return group, nil
}

// ListGroupsByActivity 获取活动的拼团列表
func (s *groupBuyService) ListGroupsByActivity(ctx context.Context, activityID uint64, status string, pageSize, pageNum int32) ([]*model.Group, int64, error) {
	groups, total, err := s.repo.ListGroupsByActivity(ctx, activityID, status, pageSize, pageNum)
	if err != nil {
		s.logger.Error("获取拼团列表失败", zap.Error(err))
		return nil, 0, err
	}
	return groups, total, nil
}

// ListUserGroups 获取用户的拼团列表
func (s *groupBuyService) ListUserGroups(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.Group, int64, error) {
	groups, total, err := s.repo.ListUserGroups(ctx, userID, pageSize, pageNum)
	if err != nil {
		s.logger.Error("获取用户拼团列表失败", zap.Error(err))
		return nil, 0, err
	}
	return groups, total, nil
}

// GetGroupMembers 获取拼团成员
func (s *groupBuyService) GetGroupMembers(ctx context.Context, groupID uint64) ([]*model.GroupMember, error) {
	members, err := s.repo.GetGroupMembers(ctx, groupID)
	if err != nil {
		s.logger.Error("获取拼团成员失败", zap.Error(err))
		return nil, err
	}
	return members, nil
}

// CheckGroupStatus 检查拼团状态
func (s *groupBuyService) CheckGroupStatus(ctx context.Context, groupID uint64) (*GroupStatusInfo, error) {
	group, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, ErrGroupNotFound
	}

	members, err := s.repo.GetGroupMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	remainingTime := group.ExpireTime.Sub(now)
	if remainingTime < 0 {
		remainingTime = 0
	}

	return &GroupStatusInfo{
		Group:          group,
		Members:        members,
		RemainingSlots: group.RequiredMembers - group.CurrentMembers,
		RemainingTime:  remainingTime,
		IsComplete:     group.IsComplete(),
		IsExpired:      group.IsExpired(),
	}, nil
}

// ProcessExpiredGroups 处理过期的拼团
func (s *groupBuyService) ProcessExpiredGroups(ctx context.Context) (int, error) {
	// 获取所有招募中的拼团
	groups, err := s.repo.GetExpiredGroups(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, group := range groups {
		// 标记为失败
		group.Status = model.GroupStatusFailed
		group.UpdatedAt = time.Now()

		if err := s.repo.UpdateGroup(ctx, group); err != nil {
			s.logger.Error("更新拼团状态失败", zap.Error(err))
			continue
		}

		// TODO: 通知用户拼团失败
		// TODO: 退款处理

		count++
	}

	if count > 0 {
		s.logger.Info("处理过期拼团", zap.Int("count", count))
	}

	return count, nil
}

// completeGroup 完成拼团
func (s *groupBuyService) completeGroup(ctx context.Context, group *model.Group) {
	now := time.Now()
	group.Status = model.GroupStatusSuccess
	group.SuccessTime = &now
	group.UpdatedAt = now

	if err := s.repo.UpdateGroup(ctx, group); err != nil {
		s.logger.Error("更新拼团状态失败", zap.Error(err))
		return
	}

	// TODO: 通知所有成员拼团成功
	// TODO: 创建订单

	s.logger.Info("拼团成功", zap.Uint64("groupID", group.ID))
}
