package persistence

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

// --- 商品管理 (Product methods) ---

func (r *pointsRepository) SaveProduct(ctx context.Context, product *domain.PointsProduct) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *pointsRepository) GetProduct(ctx context.Context, id uint64) (*domain.PointsProduct, error) {
	var product domain.PointsProduct
	if err := r.db.WithContext(ctx).First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *pointsRepository) ListProducts(ctx context.Context, status *domain.PointsProductStatus, offset, limit int) ([]*domain.PointsProduct, int64, error) {
	var list []*domain.PointsProduct
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.PointsProduct{})
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 订单管理 (Order methods) ---

func (r *pointsRepository) SaveOrder(ctx context.Context, order *domain.PointsOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *pointsRepository) GetOrder(ctx context.Context, id uint64) (*domain.PointsOrder, error) {
	var order domain.PointsOrder
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *pointsRepository) ListOrders(ctx context.Context, userID uint64, status *domain.PointsOrderStatus, offset, limit int) ([]*domain.PointsOrder, int64, error) {
	var list []*domain.PointsOrder
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.PointsOrder{}).Where("user_id = ?", userID)
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 账户与流水管理 (Account & Transaction methods) ---

func (r *pointsRepository) GetAccount(ctx context.Context, userID uint64) (*domain.PointsAccount, error) {
	var account domain.PointsAccount
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果账户不存在，则创建新账户并保存。
			account = domain.PointsAccount{UserID: userID}
			if err := r.db.WithContext(ctx).Save(&account).Error; err != nil {
				return nil, err
			}
			return &account, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *pointsRepository) SaveAccount(ctx context.Context, account *domain.PointsAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *pointsRepository) SaveTransaction(ctx context.Context, tx *domain.PointsTransaction) error {
	return r.db.WithContext(ctx).Save(tx).Error
}

func (r *pointsRepository) ListTransactions(ctx context.Context, userID uint64, offset, limit int) ([]*domain.PointsTransaction, int64, error) {
	var list []*domain.PointsTransaction
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.PointsTransaction{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
