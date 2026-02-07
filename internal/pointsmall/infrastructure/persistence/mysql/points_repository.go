package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain"
	"gorm.io/gorm"
)

type pointsRepository struct {
	db *gorm.DB
}

// NewPointsRepository 创建并返回一个新的 pointsRepository 实例。
func NewPointsRepository(db *gorm.DB) domain.PointsRepository {
	return &pointsRepository{db: db}
}

// --- tx helpers ---

func (r *pointsRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *pointsRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *pointsRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *pointsRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- 商品管理 (Product methods) ---

func (r *pointsRepository) SaveProduct(ctx context.Context, product *domain.PointsProduct) error {
	return r.saveProductWithTx(ctx, r.db, product)
}

func (r *pointsRepository) SaveProductInTx(ctx context.Context, tx any, product *domain.PointsProduct) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveProductWithTx(ctx, gormTx, product)
}

func (r *pointsRepository) GetProduct(ctx context.Context, id uint64) (*domain.PointsProduct, error) {
	var product PointsProductModel
	if err := r.db.WithContext(ctx).First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toProduct(&product), nil
}

func (r *pointsRepository) ListProducts(ctx context.Context, query *domain.PointsProductQuery) ([]*domain.PointsProduct, int64, error) {
	var list []*PointsProductModel
	var total int64

	db := r.db.WithContext(ctx).Model(&PointsProductModel{})
	if query != nil {
		if query.Status != nil {
			db = db.Where("status = ?", *query.Status)
		}
		if query.Keyword != "" {
			like := "%" + query.Keyword + "%"
			db = db.Where("name LIKE ? OR description LIKE ?", like, like)
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

	items := make([]*domain.PointsProduct, len(list))
	for i, item := range list {
		items[i] = toProduct(item)
	}
	return items, total, nil
}

// --- 订单管理 (Order methods) ---

func (r *pointsRepository) SaveOrder(ctx context.Context, order *domain.PointsOrder) error {
	return r.saveOrderWithTx(ctx, r.db, order)
}

func (r *pointsRepository) SaveOrderInTx(ctx context.Context, tx any, order *domain.PointsOrder) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveOrderWithTx(ctx, gormTx, order)
}

func (r *pointsRepository) GetOrder(ctx context.Context, id uint64) (*domain.PointsOrder, error) {
	var order PointsOrderModel
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toOrder(&order), nil
}

func (r *pointsRepository) ListOrders(ctx context.Context, query *domain.PointsOrderQuery) ([]*domain.PointsOrder, int64, error) {
	var list []*PointsOrderModel
	var total int64

	db := r.db.WithContext(ctx).Model(&PointsOrderModel{})
	if query != nil {
		if query.UserID > 0 {
			db = db.Where("user_id = ?", query.UserID)
		}
		if query.Status != nil {
			db = db.Where("status = ?", *query.Status)
		}
		if query.OrderNo != "" {
			db = db.Where("order_no = ?", query.OrderNo)
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

	items := make([]*domain.PointsOrder, len(list))
	for i, item := range list {
		items[i] = toOrder(item)
	}
	return items, total, nil
}

// --- 账户与流水管理 (Account & Transaction methods) ---

func (r *pointsRepository) GetAccount(ctx context.Context, userID uint64) (*domain.PointsAccount, error) {
	var account PointsAccountModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toAccount(&account), nil
}

func (r *pointsRepository) SaveAccount(ctx context.Context, account *domain.PointsAccount) error {
	return r.saveAccountWithTx(ctx, r.db, account)
}

func (r *pointsRepository) SaveAccountInTx(ctx context.Context, tx any, account *domain.PointsAccount) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveAccountWithTx(ctx, gormTx, account)
}

func (r *pointsRepository) SaveTransaction(ctx context.Context, tx *domain.PointsTransaction) error {
	return r.saveTransactionWithTx(ctx, r.db, tx)
}

func (r *pointsRepository) SaveTransactionInTx(ctx context.Context, tx any, transaction *domain.PointsTransaction) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveTransactionWithTx(ctx, gormTx, transaction)
}

func (r *pointsRepository) ListTransactions(ctx context.Context, query *domain.PointsTransactionQuery) ([]*domain.PointsTransaction, int64, error) {
	var list []*PointsTransactionModel
	var total int64

	db := r.db.WithContext(ctx).Model(&PointsTransactionModel{})
	if query != nil {
		if query.UserID > 0 {
			db = db.Where("user_id = ?", query.UserID)
		}
		if query.Type != "" {
			db = db.Where("type = ?", query.Type)
		}
		if query.RefID != "" {
			db = db.Where("ref_id = ?", query.RefID)
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

	items := make([]*domain.PointsTransaction, len(list))
	for i, item := range list {
		items[i] = toTransaction(item)
	}
	return items, total, nil
}

// --- internal helpers ---

func (r *pointsRepository) saveProductWithTx(ctx context.Context, tx *gorm.DB, product *domain.PointsProduct) error {
	if product == nil {
		return nil
	}
	model := toProductModel(product)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toProduct(model); synced != nil {
		*product = *synced
	}
	return nil
}

func (r *pointsRepository) saveOrderWithTx(ctx context.Context, tx *gorm.DB, order *domain.PointsOrder) error {
	if order == nil {
		return nil
	}
	model := toOrderModel(order)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toOrder(model); synced != nil {
		*order = *synced
	}
	return nil
}

func (r *pointsRepository) saveAccountWithTx(ctx context.Context, tx *gorm.DB, account *domain.PointsAccount) error {
	if account == nil {
		return nil
	}
	model := toAccountModel(account)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toAccount(model); synced != nil {
		*account = *synced
	}
	return nil
}

func (r *pointsRepository) saveTransactionWithTx(ctx context.Context, tx *gorm.DB, transaction *domain.PointsTransaction) error {
	if transaction == nil {
		return nil
	}
	model := toTransactionModel(transaction)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toTransaction(model); synced != nil {
		*transaction = *synced
	}
	return nil
}
