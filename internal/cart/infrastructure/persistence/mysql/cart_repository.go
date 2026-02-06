package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/cart/domain"
	"gorm.io/gorm"
)

type cartRepository struct {
	db *gorm.DB
}

// NewCartRepository 创建购物车仓储（MySQL）。
func NewCartRepository(db *gorm.DB) domain.CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *cartRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *cartRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *cartRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (r *cartRepository) Save(ctx context.Context, cart *domain.Cart) error {
	return r.saveWithTx(ctx, r.db, cart)
}

func (r *cartRepository) SaveInTx(ctx context.Context, tx any, cart *domain.Cart) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveWithTx(ctx, gormTx, cart)
}

func (r *cartRepository) GetByUserID(ctx context.Context, userID uint64) (*domain.Cart, error) {
	var model CartModel
	if err := r.db.WithContext(ctx).Preload("Items").Where("user_id = ?", userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toCart(&model), nil
}

func (r *cartRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Select("Items").Delete(&CartModel{}, id).Error
}

func (r *cartRepository) DeleteInTx(ctx context.Context, tx any, id uint64) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.WithContext(ctx).Select("Items").Delete(&CartModel{}, id).Error
}

func (r *cartRepository) Clear(ctx context.Context, cartID uint64) error {
	return r.db.WithContext(ctx).Where("cart_id = ?", cartID).Delete(&CartItemModel{}).Error
}

func (r *cartRepository) ClearInTx(ctx context.Context, tx any, cartID uint64) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.WithContext(ctx).Where("cart_id = ?", cartID).Delete(&CartItemModel{}).Error
}

func (r *cartRepository) saveWithTx(ctx context.Context, tx *gorm.DB, cart *domain.Cart) error {
	if cart == nil {
		return nil
	}

	if cart.ID == 0 {
		model := &CartModel{
			UserID:            cart.UserID,
			AppliedCouponCode: cart.AppliedCouponCode,
		}
		if err := tx.WithContext(ctx).Create(model).Error; err != nil {
			return err
		}
		cart.ID = uint64(model.ID)
		cart.CreatedAt = model.CreatedAt
		cart.UpdatedAt = model.UpdatedAt
	} else {
		updates := map[string]any{
			"user_id":             cart.UserID,
			"applied_coupon_code": cart.AppliedCouponCode,
		}
		if err := tx.WithContext(ctx).Model(&CartModel{}).Where("id = ?", cart.ID).Updates(updates).Error; err != nil {
			return err
		}
	}

	if err := tx.WithContext(ctx).Where("cart_id = ?", cart.ID).Delete(&CartItemModel{}).Error; err != nil {
		return err
	}
	if len(cart.Items) == 0 {
		return nil
	}

	items := make([]CartItemModel, 0, len(cart.Items))
	for _, item := range cart.Items {
		if item == nil {
			continue
		}
		items = append(items, CartItemModel{
			CartID:          cart.ID,
			ProductID:       item.ProductID,
			SkuID:           item.SkuID,
			ProductName:     item.ProductName,
			SkuName:         item.SkuName,
			Price:           item.Price,
			Quantity:        item.Quantity,
			ProductImageURL: item.ProductImageURL,
		})
	}

	if len(items) == 0 {
		return nil
	}
	return tx.WithContext(ctx).Create(&items).Error
}
