package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
	"gorm.io/gorm"
)

type flashSaleRepository struct {
	db *gorm.DB
}

// NewFlashSaleRepository 创建并返回一个新的 flashSaleRepository 实例。
func NewFlashSaleRepository(db *gorm.DB) domain.FlashSaleRepository {
	return &flashSaleRepository{db: db}
}

// --- tx helpers ---

func (r *flashSaleRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *flashSaleRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *flashSaleRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *flashSaleRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- 活动管理 (Flashsale methods) ---

func (r *flashSaleRepository) SaveFlashsale(ctx context.Context, flashsale *domain.Flashsale) error {
	return r.saveFlashsaleWithTx(ctx, r.db, flashsale)
}

func (r *flashSaleRepository) SaveFlashsaleInTx(ctx context.Context, tx any, flashsale *domain.Flashsale) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveFlashsaleWithTx(ctx, gormTx, flashsale)
}

func (r *flashSaleRepository) GetFlashsale(ctx context.Context, id uint64) (*domain.Flashsale, error) {
	var flashsale FlashsaleModel
	if err := r.db.WithContext(ctx).First(&flashsale, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toFlashsale(&flashsale), nil
}

func (r *flashSaleRepository) ListFlashsales(ctx context.Context, query *domain.FlashsaleQuery) ([]*domain.Flashsale, int64, error) {
	var list []*FlashsaleModel
	var total int64

	db := r.db.WithContext(ctx).Model(&FlashsaleModel{})
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
	if err := db.Offset(offset).Limit(pageSize).Order("start_time asc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.Flashsale, len(list))
	for i, item := range list {
		items[i] = toFlashsale(item)
	}
	return items, total, nil
}

func (r *flashSaleRepository) UpdateStock(ctx context.Context, id uint64, quantity int32) error {
	return r.db.WithContext(ctx).Model(&FlashsaleModel{}).
		Where("id = ? AND sold_count + ? <= total_stock", id, quantity).
		UpdateColumn("sold_count", gorm.Expr("sold_count + ?", quantity)).Error
}

// --- 订单管理 (FlashsaleOrder methods) ---

func (r *flashSaleRepository) SaveOrder(ctx context.Context, order *domain.FlashsaleOrder) error {
	return r.saveOrderWithTx(ctx, r.db, order)
}

func (r *flashSaleRepository) SaveOrderInTx(ctx context.Context, tx any, order *domain.FlashsaleOrder) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveOrderWithTx(ctx, gormTx, order)
}

func (r *flashSaleRepository) GetOrder(ctx context.Context, id uint64) (*domain.FlashsaleOrder, error) {
	var order FlashsaleOrderModel
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toFlashsaleOrder(&order), nil
}

func (r *flashSaleRepository) GetUserOrders(ctx context.Context, userID, flashsaleID uint64) ([]*domain.FlashsaleOrder, error) {
	var list []*FlashsaleOrderModel
	if err := r.db.WithContext(ctx).Where("user_id = ? AND flashsale_id = ?", userID, flashsaleID).Find(&list).Error; err != nil {
		return nil, err
	}
	items := make([]*domain.FlashsaleOrder, len(list))
	for i, item := range list {
		items[i] = toFlashsaleOrder(item)
	}
	return items, nil
}

func (r *flashSaleRepository) ListOrders(ctx context.Context, query *domain.FlashsaleOrderQuery) ([]*domain.FlashsaleOrder, int64, error) {
	var list []*FlashsaleOrderModel
	var total int64

	db := r.db.WithContext(ctx).Model(&FlashsaleOrderModel{})
	if query != nil {
		if query.UserID > 0 {
			db = db.Where("user_id = ?", query.UserID)
		}
		if query.FlashsaleID > 0 {
			db = db.Where("flashsale_id = ?", query.FlashsaleID)
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

	items := make([]*domain.FlashsaleOrder, len(list))
	for i, item := range list {
		items[i] = toFlashsaleOrder(item)
	}
	return items, total, nil
}

func (r *flashSaleRepository) CountUserBought(ctx context.Context, userID, flashsaleID uint64) (int32, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&FlashsaleOrderModel{}).
		Where("user_id = ? AND flashsale_id = ? AND status != ?", userID, flashsaleID, domain.FlashsaleOrderStatusCancelled).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return int32(total), nil
}

func (r *flashSaleRepository) saveFlashsaleWithTx(ctx context.Context, tx *gorm.DB, flashsale *domain.Flashsale) error {
	model := toFlashsaleModel(flashsale)
	if model == nil {
		return nil
	}
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	flashsale.ID = model.ID
	flashsale.CreatedAt = model.CreatedAt
	flashsale.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *flashSaleRepository) saveOrderWithTx(ctx context.Context, tx *gorm.DB, order *domain.FlashsaleOrder) error {
	model := toFlashsaleOrderModel(order)
	if model == nil {
		return nil
	}
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	order.ID = model.ID
	order.CreatedAt = model.CreatedAt
	order.UpdatedAt = model.UpdatedAt
	return nil
}
