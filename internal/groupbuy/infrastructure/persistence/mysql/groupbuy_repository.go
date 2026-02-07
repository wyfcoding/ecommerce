package mysql

import (
	"context"
	"errors"

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

// --- tx helpers ---

func (r *groupbuyRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *groupbuyRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *groupbuyRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *groupbuyRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- Groupbuy methods ---

func (r *groupbuyRepository) CreateGroupbuy(ctx context.Context, groupbuy *domain.Groupbuy) error {
	return r.createGroupbuyWithTx(ctx, r.db, groupbuy)
}

func (r *groupbuyRepository) CreateGroupbuyInTx(ctx context.Context, tx any, groupbuy *domain.Groupbuy) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.createGroupbuyWithTx(ctx, gormTx, groupbuy)
}

func (r *groupbuyRepository) GetGroupbuyByID(ctx context.Context, id uint64) (*domain.Groupbuy, error) {
	var groupbuy GroupbuyModel
	if err := r.db.WithContext(ctx).First(&groupbuy, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toGroupbuy(&groupbuy), nil
}

func (r *groupbuyRepository) UpdateGroupbuy(ctx context.Context, groupbuy *domain.Groupbuy) error {
	return r.updateGroupbuyWithTx(ctx, r.db, groupbuy)
}

func (r *groupbuyRepository) UpdateGroupbuyInTx(ctx context.Context, tx any, groupbuy *domain.Groupbuy) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.updateGroupbuyWithTx(ctx, gormTx, groupbuy)
}

func (r *groupbuyRepository) ListGroupbuys(ctx context.Context, query *domain.GroupbuyQuery) ([]*domain.Groupbuy, int64, error) {
	var list []*GroupbuyModel
	var total int64

	db := r.db.WithContext(ctx).Model(&GroupbuyModel{})
	if query != nil {
		if query.Status != nil {
			db = db.Where("status = ?", *query.Status)
		}
		if query.ProductID > 0 {
			db = db.Where("product_id = ?", query.ProductID)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := 1
	pageSize := 20
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.Groupbuy, len(list))
	for i, item := range list {
		items[i] = toGroupbuy(item)
	}
	return items, total, nil
}

// --- GroupbuyTeam methods ---

func (r *groupbuyRepository) CreateTeam(ctx context.Context, team *domain.GroupbuyTeam) error {
	return r.createTeamWithTx(ctx, r.db, team)
}

func (r *groupbuyRepository) CreateTeamInTx(ctx context.Context, tx any, team *domain.GroupbuyTeam) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.createTeamWithTx(ctx, gormTx, team)
}

func (r *groupbuyRepository) GetTeamByID(ctx context.Context, id uint64) (*domain.GroupbuyTeam, error) {
	var team GroupbuyTeamModel
	if err := r.db.WithContext(ctx).First(&team, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toGroupbuyTeam(&team), nil
}

func (r *groupbuyRepository) GetTeamByNo(ctx context.Context, teamNo string) (*domain.GroupbuyTeam, error) {
	var team GroupbuyTeamModel
	if err := r.db.WithContext(ctx).Where("team_no = ?", teamNo).First(&team).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toGroupbuyTeam(&team), nil
}

func (r *groupbuyRepository) UpdateTeam(ctx context.Context, team *domain.GroupbuyTeam) error {
	return r.updateTeamWithTx(ctx, r.db, team)
}

func (r *groupbuyRepository) UpdateTeamInTx(ctx context.Context, tx any, team *domain.GroupbuyTeam) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.updateTeamWithTx(ctx, gormTx, team)
}

func (r *groupbuyRepository) ListTeamsByGroupbuyID(ctx context.Context, query *domain.GroupbuyTeamQuery) ([]*domain.GroupbuyTeam, int64, error) {
	var list []*GroupbuyTeamModel
	var total int64

	db := r.db.WithContext(ctx).Model(&GroupbuyTeamModel{})
	if query != nil {
		if query.GroupbuyID > 0 {
			db = db.Where("groupbuy_id = ?", query.GroupbuyID)
		}
		if query.Status != nil {
			db = db.Where("status = ?", *query.Status)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := 1
	pageSize := 20
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.GroupbuyTeam, len(list))
	for i, item := range list {
		items[i] = toGroupbuyTeam(item)
	}
	return items, total, nil
}

// --- GroupbuyOrder methods ---

func (r *groupbuyRepository) CreateOrder(ctx context.Context, order *domain.GroupbuyOrder) error {
	return r.createOrderWithTx(ctx, r.db, order)
}

func (r *groupbuyRepository) CreateOrderInTx(ctx context.Context, tx any, order *domain.GroupbuyOrder) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.createOrderWithTx(ctx, gormTx, order)
}

func (r *groupbuyRepository) GetOrderByID(ctx context.Context, id uint64) (*domain.GroupbuyOrder, error) {
	var order GroupbuyOrderModel
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toGroupbuyOrder(&order), nil
}

func (r *groupbuyRepository) UpdateOrder(ctx context.Context, order *domain.GroupbuyOrder) error {
	return r.updateOrderWithTx(ctx, r.db, order)
}

func (r *groupbuyRepository) UpdateOrderInTx(ctx context.Context, tx any, order *domain.GroupbuyOrder) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.updateOrderWithTx(ctx, gormTx, order)
}

func (r *groupbuyRepository) ListOrdersByTeamID(ctx context.Context, teamID uint64) ([]*domain.GroupbuyOrder, error) {
	var orders []*GroupbuyOrderModel
	if err := r.db.WithContext(ctx).Where("team_id = ?", teamID).Find(&orders).Error; err != nil {
		return nil, err
	}
	result := make([]*domain.GroupbuyOrder, len(orders))
	for i, item := range orders {
		result[i] = toGroupbuyOrder(item)
	}
	return result, nil
}

func (r *groupbuyRepository) ListOrdersByUserID(ctx context.Context, query *domain.GroupbuyOrderQuery) ([]*domain.GroupbuyOrder, int64, error) {
	var orders []*GroupbuyOrderModel
	var total int64

	db := r.db.WithContext(ctx).Model(&GroupbuyOrderModel{})
	if query != nil {
		if query.UserID > 0 {
			db = db.Where("user_id = ?", query.UserID)
		}
		if query.TeamID > 0 {
			db = db.Where("team_id = ?", query.TeamID)
		}
		if query.Status != nil {
			db = db.Where("status = ?", *query.Status)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := 1
	pageSize := 20
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*domain.GroupbuyOrder, len(orders))
	for i, item := range orders {
		result[i] = toGroupbuyOrder(item)
	}
	return result, total, nil
}

// --- internal helpers ---

func (r *groupbuyRepository) createGroupbuyWithTx(ctx context.Context, tx *gorm.DB, groupbuy *domain.Groupbuy) error {
	if groupbuy == nil {
		return nil
	}
	model := toGroupbuyModel(groupbuy)
	if err := tx.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	if synced := toGroupbuy(model); synced != nil {
		*groupbuy = *synced
	}
	return nil
}

func (r *groupbuyRepository) updateGroupbuyWithTx(ctx context.Context, tx *gorm.DB, groupbuy *domain.Groupbuy) error {
	if groupbuy == nil {
		return nil
	}
	model := toGroupbuyModel(groupbuy)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toGroupbuy(model); synced != nil {
		*groupbuy = *synced
	}
	return nil
}

func (r *groupbuyRepository) createTeamWithTx(ctx context.Context, tx *gorm.DB, team *domain.GroupbuyTeam) error {
	if team == nil {
		return nil
	}
	model := toGroupbuyTeamModel(team)
	if err := tx.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	if synced := toGroupbuyTeam(model); synced != nil {
		*team = *synced
	}
	return nil
}

func (r *groupbuyRepository) updateTeamWithTx(ctx context.Context, tx *gorm.DB, team *domain.GroupbuyTeam) error {
	if team == nil {
		return nil
	}
	model := toGroupbuyTeamModel(team)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toGroupbuyTeam(model); synced != nil {
		*team = *synced
	}
	return nil
}

func (r *groupbuyRepository) createOrderWithTx(ctx context.Context, tx *gorm.DB, order *domain.GroupbuyOrder) error {
	if order == nil {
		return nil
	}
	model := toGroupbuyOrderModel(order)
	if err := tx.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	if synced := toGroupbuyOrder(model); synced != nil {
		*order = *synced
	}
	return nil
}

func (r *groupbuyRepository) updateOrderWithTx(ctx context.Context, tx *gorm.DB, order *domain.GroupbuyOrder) error {
	if order == nil {
		return nil
	}
	model := toGroupbuyOrderModel(order)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toGroupbuyOrder(model); synced != nil {
		*order = *synced
	}
	return nil
}
