package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain/repository"
	"errors"

	"gorm.io/gorm"
)

type pointsRepository struct {
	db *gorm.DB
}

func NewPointsRepository(db *gorm.DB) repository.PointsRepository {
	return &pointsRepository{db: db}
}

// 商品
func (r *pointsRepository) SaveProduct(ctx context.Context, product *entity.PointsProduct) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *pointsRepository) GetProduct(ctx context.Context, id uint64) (*entity.PointsProduct, error) {
	var product entity.PointsProduct
	if err := r.db.WithContext(ctx).First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *pointsRepository) ListProducts(ctx context.Context, status *entity.PointsProductStatus, offset, limit int) ([]*entity.PointsProduct, int64, error) {
	var list []*entity.PointsProduct
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PointsProduct{})
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

// 订单
func (r *pointsRepository) SaveOrder(ctx context.Context, order *entity.PointsOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *pointsRepository) GetOrder(ctx context.Context, id uint64) (*entity.PointsOrder, error) {
	var order entity.PointsOrder
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *pointsRepository) ListOrders(ctx context.Context, userID uint64, status *entity.PointsOrderStatus, offset, limit int) ([]*entity.PointsOrder, int64, error) {
	var list []*entity.PointsOrder
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PointsOrder{}).Where("user_id = ?", userID)
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

// 账户 & 流水
func (r *pointsRepository) GetAccount(ctx context.Context, userID uint64) (*entity.PointsAccount, error) {
	var account entity.PointsAccount
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create if not exists
			account = entity.PointsAccount{UserID: userID}
			if err := r.db.WithContext(ctx).Save(&account).Error; err != nil {
				return nil, err
			}
			return &account, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *pointsRepository) SaveAccount(ctx context.Context, account *entity.PointsAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *pointsRepository) SaveTransaction(ctx context.Context, tx *entity.PointsTransaction) error {
	return r.db.WithContext(ctx).Save(tx).Error
}

func (r *pointsRepository) ListTransactions(ctx context.Context, userID uint64, offset, limit int) ([]*entity.PointsTransaction, int64, error) {
	var list []*entity.PointsTransaction
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PointsTransaction{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
