package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain"

	"gorm.io/gorm"
)

type groupbuyRepository struct {
	db *gorm.DB
}

// NewGroupbuyRepository 创建并返回一个新的 groupbuyRepository 实例。
func NewGroupbuyRepository(db *gorm.DB) domain.GroupbuyRepository {
	return &groupbuyRepository{db: db}
}

// --- Groupbuy methods ---

// CreateGroupbuy 在数据库中创建一个新的拼团活动记录。
func (r *groupbuyRepository) CreateGroupbuy(ctx context.Context, groupbuy *domain.Groupbuy) error {
	return r.db.WithContext(ctx).Create(groupbuy).Error
}

// GetGroupbuyByID 根据ID从数据库获取拼团活动记录。
func (r *groupbuyRepository) GetGroupbuyByID(ctx context.Context, id uint64) (*domain.Groupbuy, error) {
	var groupbuy domain.Groupbuy
	if err := r.db.WithContext(ctx).First(&groupbuy, id).Error; err != nil {
		return nil, err
	}
	return &groupbuy, nil
}

// UpdateGroupbuy 更新数据库中的拼团活动记录。
func (r *groupbuyRepository) UpdateGroupbuy(ctx context.Context, groupbuy *domain.Groupbuy) error {
	return r.db.WithContext(ctx).Save(groupbuy).Error
}

// ListGroupbuys 从数据库列出所有拼团活动记录，支持分页。
func (r *groupbuyRepository) ListGroupbuys(ctx context.Context, page, pageSize int) ([]*domain.Groupbuy, int64, error) {
	var groupbuys []*domain.Groupbuy
	var total int64

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Model(&domain.Groupbuy{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Order("created_at desc").Find(&groupbuys).Error; err != nil {
		return nil, 0, err
	}

	return groupbuys, total, nil
}

// --- GroupbuyTeam methods ---

// CreateTeam 在数据库中创建一个新的拼团团队记录。
func (r *groupbuyRepository) CreateTeam(ctx context.Context, team *domain.GroupbuyTeam) error {
	return r.db.WithContext(ctx).Create(team).Error
}

// GetTeamByID 根据ID从数据库获取拼团团队记录。
func (r *groupbuyRepository) GetTeamByID(ctx context.Context, id uint64) (*domain.GroupbuyTeam, error) {
	var team domain.GroupbuyTeam
	if err := r.db.WithContext(ctx).First(&team, id).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

// GetTeamByNo 根据团队编号从数据库获取拼团团队记录。
func (r *groupbuyRepository) GetTeamByNo(ctx context.Context, teamNo string) (*domain.GroupbuyTeam, error) {
	var team domain.GroupbuyTeam
	if err := r.db.WithContext(ctx).Where("team_no = ?", teamNo).First(&team).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

// UpdateTeam 更新数据库中的拼团团队记录。
func (r *groupbuyRepository) UpdateTeam(ctx context.Context, team *domain.GroupbuyTeam) error {
	return r.db.WithContext(ctx).Save(team).Error
}

// ListTeamsByGroupbuyID 从数据库列出指定拼团活动ID的所有团队记录，支持分页。
func (r *groupbuyRepository) ListTeamsByGroupbuyID(ctx context.Context, groupbuyID uint64, page, pageSize int) ([]*domain.GroupbuyTeam, int64, error) {
	var teams []*domain.GroupbuyTeam
	var total int64

	offset := (page - 1) * pageSize
	db := r.db.WithContext(ctx).Model(&domain.GroupbuyTeam{}).Where("groupbuy_id = ?", groupbuyID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&teams).Error; err != nil {
		return nil, 0, err
	}

	return teams, total, nil
}

// --- GroupbuyOrder methods ---

// CreateOrder 在数据库中创建一个新的拼团订单记录。
func (r *groupbuyRepository) CreateOrder(ctx context.Context, order *domain.GroupbuyOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// GetOrderByID 根据ID从数据库获取拼团订单记录。
func (r *groupbuyRepository) GetOrderByID(ctx context.Context, id uint64) (*domain.GroupbuyOrder, error) {
	var order domain.GroupbuyOrder
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateOrder 更新数据库中的拼团订单记录。
func (r *groupbuyRepository) UpdateOrder(ctx context.Context, order *domain.GroupbuyOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// ListOrdersByTeamID 从数据库列出指定团队ID的所有拼团订单记录。
func (r *groupbuyRepository) ListOrdersByTeamID(ctx context.Context, teamID uint64) ([]*domain.GroupbuyOrder, error) {
	var orders []*domain.GroupbuyOrder
	if err := r.db.WithContext(ctx).Where("team_id = ?", teamID).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// ListOrdersByUserID 从数据库列出指定用户ID的所有拼团订单记录，支持分页。
func (r *groupbuyRepository) ListOrdersByUserID(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.GroupbuyOrder, int64, error) {
	var orders []*domain.GroupbuyOrder
	var total int64

	offset := (page - 1) * pageSize
	db := r.db.WithContext(ctx).Model(&domain.GroupbuyOrder{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
