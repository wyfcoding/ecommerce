package persistence

import (
	"context"
	"errors" // 导入标准错误处理库。

	"github.com/wyfcoding/ecommerce/internal/usertier/domain/entity"     // 导入用户等级领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/usertier/domain/repository" // 导入用户等级领域的仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type userTierRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewUserTierRepository 创建并返回一个新的 userTierRepository 实例。
func NewUserTierRepository(db *gorm.DB) repository.UserTierRepository {
	return &userTierRepository{db: db}
}

// --- 用户等级 (UserTier methods) ---

// SaveUserTier 将用户等级实体保存到数据库。
// 此方法实现了“upsert”逻辑：如果用户ID对应的等级记录已存在，则更新；否则创建新记录。
func (r *userTierRepository) SaveUserTier(ctx context.Context, tier *entity.UserTier) error {
	var existing entity.UserTier
	// 尝试查找现有记录。
	err := r.db.WithContext(ctx).Where("user_id = ?", tier.UserID).First(&existing).Error
	if err == nil {
		// 如果找到现有记录，则更新其ID和CreatedAt，然后保存。
		tier.ID = existing.ID
		tier.CreatedAt = existing.CreatedAt
		return r.db.WithContext(ctx).Save(tier).Error
	}
	// 如果未找到记录，则创建新记录。
	return r.db.WithContext(ctx).Create(tier).Error
}

// GetUserTier 根据用户ID从数据库获取用户等级记录。
// 如果记录未找到，则返回nil。
func (r *userTierRepository) GetUserTier(ctx context.Context, userID uint64) (*entity.UserTier, error) {
	var tier entity.UserTier
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&tier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &tier, nil
}

// --- 等级配置 (TierConfig methods) ---

// SaveTierConfig 将等级配置实体保存到数据库。
func (r *userTierRepository) SaveTierConfig(ctx context.Context, config *entity.TierConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// GetTierConfig 根据等级级别从数据库获取等级配置记录。
// 如果记录未找到，则返回nil。
func (r *userTierRepository) GetTierConfig(ctx context.Context, level entity.TierLevel) (*entity.TierConfig, error) {
	var config entity.TierConfig
	if err := r.db.WithContext(ctx).Where("level = ?", level).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &config, nil
}

// ListTierConfigs 从数据库列出所有等级配置记录，按等级升序排列。
func (r *userTierRepository) ListTierConfigs(ctx context.Context) ([]*entity.TierConfig, error) {
	var list []*entity.TierConfig
	// 按等级Level升序排列。
	if err := r.db.WithContext(ctx).Order("level asc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- 积分 (PointsAccount & PointsLog methods) ---

// GetPointsAccount 根据用户ID从数据库获取积分账户记录。
// 如果记录未找到，则返回nil。
func (r *userTierRepository) GetPointsAccount(ctx context.Context, userID uint64) (*entity.PointsAccount, error) {
	var account entity.PointsAccount
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &account, nil
}

// SavePointsAccount 将积分账户实体保存到数据库。
// 此方法实现了“upsert”逻辑：如果用户ID对应的积分账户已存在，则更新；否则创建新记录。
func (r *userTierRepository) SavePointsAccount(ctx context.Context, account *entity.PointsAccount) error {
	var existing entity.PointsAccount
	// 尝试查找现有记录。
	err := r.db.WithContext(ctx).Where("user_id = ?", account.UserID).First(&existing).Error
	if err == nil {
		// 如果找到现有记录，则更新其ID和CreatedAt，然后保存。
		account.ID = existing.ID
		account.CreatedAt = existing.CreatedAt
		return r.db.WithContext(ctx).Save(account).Error
	}
	// 如果未找到记录，则创建新记录。
	return r.db.WithContext(ctx).Create(account).Error
}

// SavePointsLog 将积分日志实体保存到数据库。
func (r *userTierRepository) SavePointsLog(ctx context.Context, log *entity.PointsLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// ListPointsLogs 从数据库列出指定用户ID的所有积分日志记录，支持分页。
func (r *userTierRepository) ListPointsLogs(ctx context.Context, userID uint64, offset, limit int) ([]*entity.PointsLog, int64, error) {
	var list []*entity.PointsLog
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PointsLog{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil { // 统计总记录数。
		return nil, 0, err
	}

	// 应用分页和排序（按ID降序，即最新日志在前）。
	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 兑换 (Exchange & ExchangeRecord methods) ---

// GetExchange 根据ID从数据库获取兑换商品记录。
// 如果记录未找到，则返回nil。
func (r *userTierRepository) GetExchange(ctx context.Context, id uint64) (*entity.Exchange, error) {
	var exchange entity.Exchange
	if err := r.db.WithContext(ctx).First(&exchange, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &exchange, nil
}

// ListExchanges 从数据库列出所有兑换商品记录，支持分页。
func (r *userTierRepository) ListExchanges(ctx context.Context, offset, limit int) ([]*entity.Exchange, int64, error) {
	var list []*entity.Exchange
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Exchange{})
	if err := db.Count(&total).Error; err != nil { // 统计总记录数。
		return nil, 0, err
	}

	// 应用分页和排序（按ID降序）。
	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// SaveExchange 将兑换商品实体保存到数据库。
func (r *userTierRepository) SaveExchange(ctx context.Context, exchange *entity.Exchange) error {
	return r.db.WithContext(ctx).Save(exchange).Error
}

// SaveExchangeRecord 将兑换记录实体保存到数据库。
func (r *userTierRepository) SaveExchangeRecord(ctx context.Context, record *entity.ExchangeRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// ListExchangeRecords 从数据库列出指定用户ID的所有兑换记录，支持分页。
func (r *userTierRepository) ListExchangeRecords(ctx context.Context, userID uint64, offset, limit int) ([]*entity.ExchangeRecord, int64, error) {
	var list []*entity.ExchangeRecord
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.ExchangeRecord{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil { // 统计总记录数。
		return nil, 0, err
	}

	// 应用分页和排序（按ID降序）。
	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
