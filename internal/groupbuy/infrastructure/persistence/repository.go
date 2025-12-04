package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain/entity"     // 导入拼团模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain/repository" // 导入拼团模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type groupbuyRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewGroupbuyRepository 创建并返回一个新的 groupbuyRepository 实例。
// db: GORM数据库连接实例。
func NewGroupbuyRepository(db *gorm.DB) repository.GroupbuyRepository {
	return &groupbuyRepository{db: db}
}

// --- Groupbuy methods ---

// CreateGroupbuy 在数据库中创建一个新的拼团活动记录。
func (r *groupbuyRepository) CreateGroupbuy(ctx context.Context, groupbuy *entity.Groupbuy) error {
	return r.db.WithContext(ctx).Create(groupbuy).Error
}

// GetGroupbuyByID 根据ID从数据库获取拼团活动记录。
func (r *groupbuyRepository) GetGroupbuyByID(ctx context.Context, id uint64) (*entity.Groupbuy, error) {
	var groupbuy entity.Groupbuy
	if err := r.db.WithContext(ctx).First(&groupbuy, id).Error; err != nil {
		return nil, err
	}
	return &groupbuy, nil
}

// UpdateGroupbuy 更新数据库中的拼团活动记录。
func (r *groupbuyRepository) UpdateGroupbuy(ctx context.Context, groupbuy *entity.Groupbuy) error {
	return r.db.WithContext(ctx).Save(groupbuy).Error
}

// ListGroupbuys 从数据库列出所有拼团活动记录，支持分页。
func (r *groupbuyRepository) ListGroupbuys(ctx context.Context, page, pageSize int) ([]*entity.Groupbuy, int64, error) {
	var groupbuys []*entity.Groupbuy
	var total int64

	offset := (page - 1) * pageSize
	// 统计总记录数。
	if err := r.db.WithContext(ctx).Model(&entity.Groupbuy{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Order("created_at desc").Find(&groupbuys).Error; err != nil {
		return nil, 0, err
	}

	return groupbuys, total, nil
}

// --- GroupbuyTeam methods ---

// CreateTeam 在数据库中创建一个新的拼团团队记录。
func (r *groupbuyRepository) CreateTeam(ctx context.Context, team *entity.GroupbuyTeam) error {
	return r.db.WithContext(ctx).Create(team).Error
}

// GetTeamByID 根据ID从数据库获取拼团团队记录。
func (r *groupbuyRepository) GetTeamByID(ctx context.Context, id uint64) (*entity.GroupbuyTeam, error) {
	var team entity.GroupbuyTeam
	if err := r.db.WithContext(ctx).First(&team, id).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

// GetTeamByNo 根据团队编号从数据库获取拼团团队记录。
func (r *groupbuyRepository) GetTeamByNo(ctx context.Context, teamNo string) (*entity.GroupbuyTeam, error) {
	var team entity.GroupbuyTeam
	if err := r.db.WithContext(ctx).Where("team_no = ?", teamNo).First(&team).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

// UpdateTeam 更新数据库中的拼团团队记录。
func (r *groupbuyRepository) UpdateTeam(ctx context.Context, team *entity.GroupbuyTeam) error {
	return r.db.WithContext(ctx).Save(team).Error
}

// ListTeamsByGroupbuyID 从数据库列出指定拼团活动ID的所有团队记录，支持分页。
func (r *groupbuyRepository) ListTeamsByGroupbuyID(ctx context.Context, groupbuyID uint64, page, pageSize int) ([]*entity.GroupbuyTeam, int64, error) {
	var teams []*entity.GroupbuyTeam
	var total int64

	offset := (page - 1) * pageSize
	db := r.db.WithContext(ctx).Model(&entity.GroupbuyTeam{}).Where("groupbuy_id = ?", groupbuyID)

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&teams).Error; err != nil {
		return nil, 0, err
	}

	return teams, total, nil
}

// --- GroupbuyOrder methods ---

// CreateOrder 在数据库中创建一个新的拼团订单记录。
func (r *groupbuyRepository) CreateOrder(ctx context.Context, order *entity.GroupbuyOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// GetOrderByID 根据ID从数据库获取拼团订单记录。
func (r *groupbuyRepository) GetOrderByID(ctx context.Context, id uint64) (*entity.GroupbuyOrder, error) {
	var order entity.GroupbuyOrder
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateOrder 更新数据库中的拼团订单记录。
func (r *groupbuyRepository) UpdateOrder(ctx context.Context, order *entity.GroupbuyOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// ListOrdersByTeamID 从数据库列出指定团队ID的所有拼团订单记录。
func (r *groupbuyRepository) ListOrdersByTeamID(ctx context.Context, teamID uint64) ([]*entity.GroupbuyOrder, error) {
	var orders []*entity.GroupbuyOrder
	if err := r.db.WithContext(ctx).Where("team_id = ?", teamID).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// ListOrdersByUserID 从数据库列出指定用户ID的所有拼团订单记录，支持分页。
func (r *groupbuyRepository) ListOrdersByUserID(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.GroupbuyOrder, int64, error) {
	var orders []*entity.GroupbuyOrder
	var total int64

	offset := (page - 1) * pageSize
	db := r.db.WithContext(ctx).Model(&entity.GroupbuyOrder{}).Where("user_id = ?", userID)

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
