package persistence

import (
	"context"
	"errors"
	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain/entity"     // 导入积分商城领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain/repository" // 导入积分商城领域的仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type pointsRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewPointsRepository 创建并返回一个新的 pointsRepository 实例。
func NewPointsRepository(db *gorm.DB) repository.PointsRepository {
	return &pointsRepository{db: db}
}

// --- 商品管理 (Product methods) ---

// SaveProduct 将积分商品实体保存到数据库。
// 如果实体已存在，则更新；如果不存在，则创建。
func (r *pointsRepository) SaveProduct(ctx context.Context, product *entity.PointsProduct) error {
	return r.db.WithContext(ctx).Save(product).Error
}

// GetProduct 根据ID从数据库获取积分商品记录。
// 如果记录未找到，则返回nil。
func (r *pointsRepository) GetProduct(ctx context.Context, id uint64) (*entity.PointsProduct, error) {
	var product entity.PointsProduct
	if err := r.db.WithContext(ctx).First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &product, nil
}

// ListProducts 从数据库列出所有积分商品记录，支持通过状态过滤和分页。
func (r *pointsRepository) ListProducts(ctx context.Context, status *entity.PointsProductStatus, offset, limit int) ([]*entity.PointsProduct, int64, error) {
	var list []*entity.PointsProduct
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PointsProduct{})
	if status != nil { // 如果提供了状态，则按状态过滤。
		db = db.Where("status = ?", *status)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 订单管理 (Order methods) ---

// SaveOrder 将积分订单实体保存到数据库。
func (r *pointsRepository) SaveOrder(ctx context.Context, order *entity.PointsOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// GetOrder 根据ID从数据库获取积分订单记录。
// 如果记录未找到，则返回nil。
func (r *pointsRepository) GetOrder(ctx context.Context, id uint64) (*entity.PointsOrder, error) {
	var order entity.PointsOrder
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &order, nil
}

// ListOrders 从数据库列出指定用户ID的所有积分订单记录，支持通过状态过滤和分页。
func (r *pointsRepository) ListOrders(ctx context.Context, userID uint64, status *entity.PointsOrderStatus, offset, limit int) ([]*entity.PointsOrder, int64, error) {
	var list []*entity.PointsOrder
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PointsOrder{}).Where("user_id = ?", userID)
	if status != nil { // 如果提供了状态，则按状态过滤。
		db = db.Where("status = ?", *status)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 账户与流水管理 (Account & Transaction methods) ---

// GetAccount 根据用户ID从数据库获取积分账户记录。
// 如果记录未找到，则会自动创建一个新的账户并保存。
func (r *pointsRepository) GetAccount(ctx context.Context, userID uint64) (*entity.PointsAccount, error) {
	var account entity.PointsAccount
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果账户不存在，则创建新账户并保存。
			account = entity.PointsAccount{UserID: userID}
			if err := r.db.WithContext(ctx).Save(&account).Error; err != nil {
				return nil, err
			}
			return &account, nil
		}
		return nil, err // 其他错误则返回。
	}
	return &account, nil
}

// SaveAccount 将积分账户实体保存到数据库。
func (r *pointsRepository) SaveAccount(ctx context.Context, account *entity.PointsAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

// SaveTransaction 将积分交易流水实体保存到数据库。
func (r *pointsRepository) SaveTransaction(ctx context.Context, tx *entity.PointsTransaction) error {
	return r.db.WithContext(ctx).Save(tx).Error
}

// ListTransactions 从数据库列出指定用户ID的所有积分交易流水记录，支持分页。
func (r *pointsRepository) ListTransactions(ctx context.Context, userID uint64, offset, limit int) ([]*entity.PointsTransaction, int64, error) {
	var list []*entity.PointsTransaction
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PointsTransaction{}).Where("user_id = ?", userID)

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
