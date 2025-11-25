package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain/repository"

	"gorm.io/gorm"
)

type groupbuyRepository struct {
	db *gorm.DB
}

func NewGroupbuyRepository(db *gorm.DB) repository.GroupbuyRepository {
	return &groupbuyRepository{db: db}
}

// Groupbuy methods

func (r *groupbuyRepository) CreateGroupbuy(ctx context.Context, groupbuy *entity.Groupbuy) error {
	return r.db.WithContext(ctx).Create(groupbuy).Error
}

func (r *groupbuyRepository) GetGroupbuyByID(ctx context.Context, id uint64) (*entity.Groupbuy, error) {
	var groupbuy entity.Groupbuy
	if err := r.db.WithContext(ctx).First(&groupbuy, id).Error; err != nil {
		return nil, err
	}
	return &groupbuy, nil
}

func (r *groupbuyRepository) UpdateGroupbuy(ctx context.Context, groupbuy *entity.Groupbuy) error {
	return r.db.WithContext(ctx).Save(groupbuy).Error
}

func (r *groupbuyRepository) ListGroupbuys(ctx context.Context, page, pageSize int) ([]*entity.Groupbuy, int64, error) {
	var groupbuys []*entity.Groupbuy
	var total int64

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Model(&entity.Groupbuy{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Order("created_at desc").Find(&groupbuys).Error; err != nil {
		return nil, 0, err
	}

	return groupbuys, total, nil
}

// GroupbuyTeam methods

func (r *groupbuyRepository) CreateTeam(ctx context.Context, team *entity.GroupbuyTeam) error {
	return r.db.WithContext(ctx).Create(team).Error
}

func (r *groupbuyRepository) GetTeamByID(ctx context.Context, id uint64) (*entity.GroupbuyTeam, error) {
	var team entity.GroupbuyTeam
	if err := r.db.WithContext(ctx).First(&team, id).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *groupbuyRepository) GetTeamByNo(ctx context.Context, teamNo string) (*entity.GroupbuyTeam, error) {
	var team entity.GroupbuyTeam
	if err := r.db.WithContext(ctx).Where("team_no = ?", teamNo).First(&team).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *groupbuyRepository) UpdateTeam(ctx context.Context, team *entity.GroupbuyTeam) error {
	return r.db.WithContext(ctx).Save(team).Error
}

func (r *groupbuyRepository) ListTeamsByGroupbuyID(ctx context.Context, groupbuyID uint64, page, pageSize int) ([]*entity.GroupbuyTeam, int64, error) {
	var teams []*entity.GroupbuyTeam
	var total int64

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Model(&entity.GroupbuyTeam{}).Where("groupbuy_id = ?", groupbuyID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Where("groupbuy_id = ?", groupbuyID).Offset(offset).Limit(pageSize).Order("created_at desc").Find(&teams).Error; err != nil {
		return nil, 0, err
	}

	return teams, total, nil
}

// GroupbuyOrder methods

func (r *groupbuyRepository) CreateOrder(ctx context.Context, order *entity.GroupbuyOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *groupbuyRepository) GetOrderByID(ctx context.Context, id uint64) (*entity.GroupbuyOrder, error) {
	var order entity.GroupbuyOrder
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *groupbuyRepository) UpdateOrder(ctx context.Context, order *entity.GroupbuyOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *groupbuyRepository) ListOrdersByTeamID(ctx context.Context, teamID uint64) ([]*entity.GroupbuyOrder, error) {
	var orders []*entity.GroupbuyOrder
	if err := r.db.WithContext(ctx).Where("team_id = ?", teamID).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *groupbuyRepository) ListOrdersByUserID(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.GroupbuyOrder, int64, error) {
	var orders []*entity.GroupbuyOrder
	var total int64

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Model(&entity.GroupbuyOrder{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Offset(offset).Limit(pageSize).Order("created_at desc").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
