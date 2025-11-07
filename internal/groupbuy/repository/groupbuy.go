package repository

import (
	"context"

	"gorm.io/gorm"

	"ecommerce/internal/groupbuy/model"
)

// GroupBuyRepo 拼团仓储接口
type GroupBuyRepo interface {
	// 拼团活动
	CreateActivity(ctx context.Context, activity *model.GroupBuyActivity) error
	UpdateActivity(ctx context.Context, activity *model.GroupBuyActivity) error
	GetActivityByID(ctx context.Context, id uint64) (*model.GroupBuyActivity, error)
	ListActivities(ctx context.Context, status string, pageSize, pageNum int32) ([]*model.GroupBuyActivity, int64, error)
	
	// 拼团
	CreateGroup(ctx context.Context, group *model.GroupBuyGroup) error
	UpdateGroup(ctx context.Context, group *model.GroupBuyGroup) error
	GetGroupByID(ctx context.Context, id uint64) (*model.GroupBuyGroup, error)
	GetGroupByNo(ctx context.Context, groupNo string) (*model.GroupBuyGroup, error)
	ListGroups(ctx context.Context, activityID uint64, status string, pageSize, pageNum int32) ([]*model.GroupBuyGroup, int64, error)
	
	// 拼团成员
	CreateMember(ctx context.Context, member *model.GroupBuyMember) error
	GetMembersByGroupID(ctx context.Context, groupID uint64) ([]*model.GroupBuyMember, error)
	GetMemberCount(ctx context.Context, groupID uint64) (int64, error)
	CheckUserInGroup(ctx context.Context, userID, groupID uint64) (bool, error)
	
	// 事务
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type groupBuyRepo struct {
	db *gorm.DB
}

// NewGroupBuyRepo 创建拼团仓储实例
func NewGroupBuyRepo(db *gorm.DB) GroupBuyRepo {
	return &groupBuyRepo{db: db}
}

// CreateActivity 创建拼团活动
func (r *groupBuyRepo) CreateActivity(ctx context.Context, activity *model.GroupBuyActivity) error {
	return r.db.WithContext(ctx).Create(activity).Error
}

// UpdateActivity 更新拼团活动
func (r *groupBuyRepo) UpdateActivity(ctx context.Context, activity *model.GroupBuyActivity) error {
	return r.db.WithContext(ctx).Save(activity).Error
}

// GetActivityByID 根据ID获取拼团活动
func (r *groupBuyRepo) GetActivityByID(ctx context.Context, id uint64) (*model.GroupBuyActivity, error) {
	var activity model.GroupBuyActivity
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&activity).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

// ListActivities 获取拼团活动列表
func (r *groupBuyRepo) ListActivities(ctx context.Context, status string, pageSize, pageNum int32) ([]*model.GroupBuyActivity, int64, error) {
	var activities []*model.GroupBuyActivity
	var total int64

	query := r.db.WithContext(ctx).Model(&model.GroupBuyActivity{})
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	err := query.Offset(int(offset)).Limit(int(pageSize)).Order("created_at DESC").Find(&activities).Error
	if err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

// CreateGroup 创建拼团
func (r *groupBuyRepo) CreateGroup(ctx context.Context, group *model.GroupBuyGroup) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// UpdateGroup 更新拼团
func (r *groupBuyRepo) UpdateGroup(ctx context.Context, group *model.GroupBuyGroup) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// GetGroupByID 根据ID获取拼团
func (r *groupBuyRepo) GetGroupByID(ctx context.Context, id uint64) (*model.GroupBuyGroup, error) {
	var group model.GroupBuyGroup
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetGroupByNo 根据团号获取拼团
func (r *groupBuyRepo) GetGroupByNo(ctx context.Context, groupNo string) (*model.GroupBuyGroup, error) {
	var group model.GroupBuyGroup
	err := r.db.WithContext(ctx).Where("group_no = ?", groupNo).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// ListGroups 获取拼团列表
func (r *groupBuyRepo) ListGroups(ctx context.Context, activityID uint64, status string, pageSize, pageNum int32) ([]*model.GroupBuyGroup, int64, error) {
	var groups []*model.GroupBuyGroup
	var total int64

	query := r.db.WithContext(ctx).Model(&model.GroupBuyGroup{})
	
	if activityID > 0 {
		query = query.Where("activity_id = ?", activityID)
	}
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	err := query.Offset(int(offset)).Limit(int(pageSize)).Order("created_at DESC").Find(&groups).Error
	if err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

// CreateMember 创建拼团成员
func (r *groupBuyRepo) CreateMember(ctx context.Context, member *model.GroupBuyMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// GetMembersByGroupID 获取拼团成员列表
func (r *groupBuyRepo) GetMembersByGroupID(ctx context.Context, groupID uint64) ([]*model.GroupBuyMember, error) {
	var members []*model.GroupBuyMember
	err := r.db.WithContext(ctx).Where("group_id = ?", groupID).Order("created_at ASC").Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

// GetMemberCount 获取拼团成员数量
func (r *groupBuyRepo) GetMemberCount(ctx context.Context, groupID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.GroupBuyMember{}).Where("group_id = ?", groupID).Count(&count).Error
	return count, err
}

// CheckUserInGroup 检查用户是否在拼团中
func (r *groupBuyRepo) CheckUserInGroup(ctx context.Context, userID, groupID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.GroupBuyMember{}).
		Where("user_id = ? AND group_id = ?", userID, groupID).
		Count(&count).Error
	return count > 0, err
}

// InTx 在事务中执行
func (r *groupBuyRepo) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, "tx", tx))
	})
}
