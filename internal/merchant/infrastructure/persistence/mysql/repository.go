// Package mysql 商家服务 MySQL 仓储实现
// 生成摘要：
// 1) 实现 MerchantRepository 接口，使用 GORM 操作 MySQL 数据库
// 2) 包含商家、店铺、营业执照、银行账户等实体的完整 CRUD 操作
// 3) 支持事务操作，确保数据一致性
package mysql

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/merchant/domain"
	"github.com/wyfcoding/pkg/database"
	"gorm.io/gorm"
)

// MerchantRepositoryImpl MySQL 商家仓储实现
type MerchantRepositoryImpl struct {
	db     *database.DB
	logger *slog.Logger
}

// NewMerchantRepository 创建 MySQL 商家仓储实例
func NewMerchantRepository(db *database.DB, logger *slog.Logger) domain.MerchantRepository {
	return &MerchantRepositoryImpl{
		db:     db,
		logger: logger.With("module", "merchant_repository"),
	}
}

// Create 创建商家
func (r *MerchantRepositoryImpl) Create(merchant *domain.Merchant) error {
	start := time.Now()

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 创建商家主记录
		if err := tx.Create(merchant).Error; err != nil {
			return fmt.Errorf("create merchant: %w", err)
		}

		// 创建关联记录
		if merchant.BusinessLicense != nil {
			merchant.BusinessLicense.MerchantID = merchant.ID
			if err := tx.Create(merchant.BusinessLicense).Error; err != nil {
				return fmt.Errorf("create business license: %w", err)
			}
		}

		if merchant.BankAccount != nil {
			merchant.BankAccount.MerchantID = merchant.ID
			if err := tx.Create(merchant.BankAccount).Error; err != nil {
				return fmt.Errorf("create bank account: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		r.logger.Error("failed to create merchant",
			"merchant_no", merchant.MerchantNo, "error", err, "duration", time.Since(start))
		return err
	}

	r.logger.Info("merchant created",
		"merchant_id", merchant.ID, "merchant_no", merchant.MerchantNo, "duration", time.Since(start))
	return nil
}

// Update 更新商家信息
func (r *MerchantRepositoryImpl) Update(merchant *domain.Merchant) error {
	start := time.Now()

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domain.Merchant{}).Where("id = ?", merchant.ID).Updates(merchant).Error; err != nil {
			return fmt.Errorf("update merchant: %w", err)
		}

		// 更新关联记录
		if merchant.BusinessLicense != nil {
			if err := tx.Model(&domain.BusinessLicense{}).Where("merchant_id = ?", merchant.ID).Updates(merchant.BusinessLicense).Error; err != nil {
				return fmt.Errorf("update business license: %w", err)
			}
		}

		if merchant.BankAccount != nil {
			if err := tx.Model(&domain.BankAccount{}).Where("merchant_id = ?", merchant.ID).Updates(merchant.BankAccount).Error; err != nil {
				return fmt.Errorf("update bank account: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		r.logger.Error("failed to update merchant",
			"merchant_id", merchant.ID, "error", err, "duration", time.Since(start))
		return err
	}

	r.logger.Info("merchant updated",
		"merchant_id", merchant.ID, "duration", time.Since(start))
	return nil
}

// FindByID 根据ID查找商家
func (r *MerchantRepositoryImpl) FindByID(id uint) (*domain.Merchant, error) {
	var merchant domain.Merchant
	err := r.db.Preload("BusinessLicense").Preload("BankAccount").Preload("Stores").First(&merchant, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find merchant by id: %w", err)
	}
	return &merchant, nil
}

// FindByMerchantNo 根据商家编号查找
func (r *MerchantRepositoryImpl) FindByMerchantNo(merchantNo string) (*domain.Merchant, error) {
	var merchant domain.Merchant
	err := r.db.Preload("BusinessLicense").Preload("BankAccount").Preload("Stores").
		Where("merchant_no = ?", merchantNo).First(&merchant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find merchant by no: %w", err)
	}
	return &merchant, nil
}

// FindByUserID 根据用户ID查找商家
func (r *MerchantRepositoryImpl) FindByUserID(userID uint64) (*domain.Merchant, error) {
	var merchant domain.Merchant
	err := r.db.Preload("BusinessLicense").Preload("BankAccount").Preload("Stores").
		Where("user_id = ?", userID).First(&merchant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find merchant by user id: %w", err)
	}
	return &merchant, nil
}

// FindByStatus 根据状态查找商家列表
func (r *MerchantRepositoryImpl) FindByStatus(status domain.MerchantStatus, limit, offset int) ([]*domain.Merchant, int64, error) {
	var merchants []*domain.Merchant
	var total int64

	// 获取总数
	if err := r.db.Model(&domain.Merchant{}).Where("status = ?", int(status)).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count merchants by status: %w", err)
	}

	// 获取分页数据
	err := r.db.Preload("BusinessLicense").Preload("BankAccount").
		Where("status = ?", int(status)).
		Limit(limit).Offset(offset).
		Order("created_at DESC").
		Find(&merchants).Error
	if err != nil {
		return nil, 0, fmt.Errorf("find merchants by status: %w", err)
	}

	return merchants, total, nil
}

// Delete 删除商家（软删除）
func (r *MerchantRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&domain.Merchant{}, id).Error
}

// StoreRepositoryImpl MySQL 店铺仓储实现
type StoreRepositoryImpl struct {
	db     *database.DB
	logger *slog.Logger
}

// NewStoreRepository 创建 MySQL 店铺仓储实例
func NewStoreRepository(db *database.DB, logger *slog.Logger) domain.StoreRepository {
	return &StoreRepositoryImpl{
		db:     db,
		logger: logger.With("module", "store_repository"),
	}
}

// Create 创建店铺
func (r *StoreRepositoryImpl) Create(store *domain.Store) error {
	start := time.Now()

	err := r.db.Create(store).Error
	if err != nil {
		r.logger.Error("failed to create store",
			"store_no", store.StoreNo, "error", err, "duration", time.Since(start))
		return fmt.Errorf("create store: %w", err)
	}

	r.logger.Info("store created",
		"store_id", store.ID, "store_no", store.StoreNo, "duration", time.Since(start))
	return nil
}

// Update 更新店铺信息
func (r *StoreRepositoryImpl) Update(store *domain.Store) error {
	start := time.Now()

	err := r.db.Model(&domain.Store{}).Where("id = ?", store.ID).Updates(store).Error
	if err != nil {
		r.logger.Error("failed to update store",
			"store_id", store.ID, "error", err, "duration", time.Since(start))
		return fmt.Errorf("update store: %w", err)
	}

	r.logger.Info("store updated",
		"store_id", store.ID, "duration", time.Since(start))
	return nil
}

// FindByID 根据ID查找店铺
func (r *StoreRepositoryImpl) FindByID(id uint) (*domain.Store, error) {
	var store domain.Store
	err := r.db.First(&store, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find store by id: %w", err)
	}
	return &store, nil
}

// FindByStoreNo 根据店铺编号查找
func (r *StoreRepositoryImpl) FindByStoreNo(storeNo string) (*domain.Store, error) {
	var store domain.Store
	err := r.db.Where("store_no = ?", storeNo).First(&store).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find store by no: %w", err)
	}
	return &store, nil
}

// FindByMerchantID 根据商家ID查找所有店铺
func (r *StoreRepositoryImpl) FindByMerchantID(merchantID uint) ([]*domain.Store, error) {
	var stores []*domain.Store
	err := r.db.Where("merchant_id = ?", merchantID).Order("created_at DESC").Find(&stores).Error
	if err != nil {
		return nil, fmt.Errorf("find stores by merchant id: %w", err)
	}
	return stores, nil
}

// FindActiveStores 查找活跃店铺
func (r *StoreRepositoryImpl) FindActiveStores(merchantID uint) ([]*domain.Store, error) {
	var stores []*domain.Store
	err := r.db.Where("merchant_id = ? AND is_open = ?", merchantID, true).
		Order("created_at DESC").Find(&stores).Error
	if err != nil {
		return nil, fmt.Errorf("find active stores: %w", err)
	}
	return stores, nil
}

// Delete 删除店铺（软删除）
func (r *StoreRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&domain.Store{}, id).Error
}

// SettingsRepositoryImpl MySQL 商家设置仓储实现
type SettingsRepositoryImpl struct {
	db     *database.DB
	logger *slog.Logger
}

// NewSettingsRepository 创建 MySQL 商家设置仓储实例
func NewSettingsRepository(db *database.DB, logger *slog.Logger) domain.SettingsRepository {
	return &SettingsRepositoryImpl{
		db:     db,
		logger: logger.With("module", "settings_repository"),
	}
}

// FindByMerchantID 根据商家ID查找设置
func (r *SettingsRepositoryImpl) FindByMerchantID(merchantID uint) (*domain.MerchantSettings, error) {
	var settings domain.MerchantSettings
	err := r.db.Where("merchant_id = ?", merchantID).First(&settings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find settings by merchant id: %w", err)
	}
	return &settings, nil
}

// Save 保存或更新设置
func (r *SettingsRepositoryImpl) Save(settings *domain.MerchantSettings) error {
	start := time.Now()

	err := r.db.Save(settings).Error
	if err != nil {
		r.logger.Error("failed to save settings",
			"merchant_id", settings.MerchantID, "error", err, "duration", time.Since(start))
		return fmt.Errorf("save settings: %w", err)
	}

	r.logger.Info("settings saved",
		"merchant_id", settings.MerchantID, "duration", time.Since(start))
	return nil
}

// Update 更新设置
func (r *SettingsRepositoryImpl) Update(settings *domain.MerchantSettings) error {
	start := time.Now()

	err := r.db.Model(&domain.MerchantSettings{}).Where("merchant_id = ?", settings.MerchantID).Updates(settings).Error
	if err != nil {
		r.logger.Error("failed to update settings",
			"merchant_id", settings.MerchantID, "error", err, "duration", time.Since(start))
		return fmt.Errorf("update settings: %w", err)
	}

	r.logger.Info("settings updated",
		"merchant_id", settings.MerchantID, "duration", time.Since(start))
	return nil
}