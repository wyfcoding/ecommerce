// Package infrastructure 商家服务基础设施层
// 生成摘要：
// 1) 实现 GORM 仓储
// 假设：
// - 使用 PostgreSQL/MySQL 作为主数据库
package infrastructure

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/merchant/domain"
	"gorm.io/gorm"
)

// GormMerchantRepository GORM 商家仓储实现
type GormMerchantRepository struct {
	db *gorm.DB
}

// NewGormMerchantRepository 创建 GORM 商家仓储
func NewGormMerchantRepository(db *gorm.DB) *GormMerchantRepository {
	return &GormMerchantRepository{db: db}
}

// Save 保存商家
func (r *GormMerchantRepository) Save(ctx context.Context, merchant *domain.Merchant) error {
	return r.db.WithContext(ctx).Create(merchant).Error
}

// Update 更新商家
func (r *GormMerchantRepository) Update(ctx context.Context, merchant *domain.Merchant) error {
	return r.db.WithContext(ctx).Save(merchant).Error
}

// FindByID 根据ID查找商家
func (r *GormMerchantRepository) FindByID(ctx context.Context, id uint) (*domain.Merchant, error) {
	var merchant domain.Merchant
	if err := r.db.WithContext(ctx).First(&merchant, id).Error; err != nil {
		return nil, fmt.Errorf("merchant not found: %w", err)
	}
	return &merchant, nil
}

// FindByMerchantNo 根据商家编号查找
func (r *GormMerchantRepository) FindByMerchantNo(ctx context.Context, merchantNo string) (*domain.Merchant, error) {
	var merchant domain.Merchant
	if err := r.db.WithContext(ctx).Where("merchant_no = ?", merchantNo).First(&merchant).Error; err != nil {
		return nil, fmt.Errorf("merchant not found: %w", err)
	}
	return &merchant, nil
}

// FindByUserID 根据用户ID查找
func (r *GormMerchantRepository) FindByUserID(ctx context.Context, userID uint64) (*domain.Merchant, error) {
	var merchant domain.Merchant
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&merchant).Error; err != nil {
		return nil, fmt.Errorf("merchant not found: %w", err)
	}
	return &merchant, nil
}

// List 列表查询
func (r *GormMerchantRepository) List(ctx context.Context, filter *domain.MerchantFilter) ([]*domain.Merchant, int64, error) {
	query := r.db.WithContext(ctx).Model(&domain.Merchant{})

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.Level != nil {
		query = query.Where("level = ?", *filter.Level)
	}
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where("name LIKE ? OR merchant_no LIKE ?", keyword, keyword)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var merchants []*domain.Merchant
	offset := max((filter.Page-1)*filter.PageSize, 0)

	if err := query.Order("created_at DESC").Offset(offset).Limit(filter.PageSize).Find(&merchants).Error; err != nil {
		return nil, 0, err
	}

	return merchants, total, nil
}

// Delete 删除商家
func (r *GormMerchantRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Merchant{}, id).Error
}

// GormBusinessLicenseRepository GORM 营业执照仓储实现
type GormBusinessLicenseRepository struct {
	db *gorm.DB
}

// NewGormBusinessLicenseRepository 创建营业执照仓储
func NewGormBusinessLicenseRepository(db *gorm.DB) *GormBusinessLicenseRepository {
	return &GormBusinessLicenseRepository{db: db}
}

// Save 保存
func (r *GormBusinessLicenseRepository) Save(ctx context.Context, license *domain.BusinessLicense) error {
	return r.db.WithContext(ctx).Create(license).Error
}

// Update 更新
func (r *GormBusinessLicenseRepository) Update(ctx context.Context, license *domain.BusinessLicense) error {
	return r.db.WithContext(ctx).Save(license).Error
}

// FindByMerchantID 根据商家ID查找
func (r *GormBusinessLicenseRepository) FindByMerchantID(ctx context.Context, merchantID uint) (*domain.BusinessLicense, error) {
	var license domain.BusinessLicense
	if err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&license).Error; err != nil {
		return nil, err
	}
	return &license, nil
}

// Delete 删除
func (r *GormBusinessLicenseRepository) Delete(ctx context.Context, merchantID uint) error {
	return r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Delete(&domain.BusinessLicense{}).Error
}

// GormBankAccountRepository GORM 银行账户仓储实现
type GormBankAccountRepository struct {
	db *gorm.DB
}

// NewGormBankAccountRepository 创建银行账户仓储
func NewGormBankAccountRepository(db *gorm.DB) *GormBankAccountRepository {
	return &GormBankAccountRepository{db: db}
}

// Save 保存
func (r *GormBankAccountRepository) Save(ctx context.Context, account *domain.BankAccount) error {
	return r.db.WithContext(ctx).Create(account).Error
}

// Update 更新
func (r *GormBankAccountRepository) Update(ctx context.Context, account *domain.BankAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

// FindByMerchantID 根据商家ID查找
func (r *GormBankAccountRepository) FindByMerchantID(ctx context.Context, merchantID uint) (*domain.BankAccount, error) {
	var account domain.BankAccount
	if err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// Delete 删除
func (r *GormBankAccountRepository) Delete(ctx context.Context, merchantID uint) error {
	return r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Delete(&domain.BankAccount{}).Error
}

// GormStoreRepository GORM 店铺仓储实现
type GormStoreRepository struct {
	db *gorm.DB
}

// NewGormStoreRepository 创建店铺仓储
func NewGormStoreRepository(db *gorm.DB) *GormStoreRepository {
	return &GormStoreRepository{db: db}
}

// Save 保存
func (r *GormStoreRepository) Save(ctx context.Context, store *domain.Store) error {
	return r.db.WithContext(ctx).Create(store).Error
}

// Update 更新
func (r *GormStoreRepository) Update(ctx context.Context, store *domain.Store) error {
	return r.db.WithContext(ctx).Save(store).Error
}

// FindByID 根据ID查找
func (r *GormStoreRepository) FindByID(ctx context.Context, id uint) (*domain.Store, error) {
	var store domain.Store
	if err := r.db.WithContext(ctx).First(&store, id).Error; err != nil {
		return nil, err
	}
	return &store, nil
}

// FindByStoreNo 根据店铺编号查找
func (r *GormStoreRepository) FindByStoreNo(ctx context.Context, storeNo string) (*domain.Store, error) {
	var store domain.Store
	if err := r.db.WithContext(ctx).Where("store_no = ?", storeNo).First(&store).Error; err != nil {
		return nil, err
	}
	return &store, nil
}

// ListByMerchantID 根据商家ID列出店铺
func (r *GormStoreRepository) ListByMerchantID(ctx context.Context, merchantID uint, page, pageSize int) ([]*domain.Store, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&domain.Store{}).Where("merchant_id = ?", merchantID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var stores []*domain.Store
	offset := max((page-1)*pageSize, 0)

	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&stores).Error; err != nil {
		return nil, 0, err
	}

	return stores, total, nil
}

// Delete 删除
func (r *GormStoreRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Store{}, id).Error
}

// GormMerchantSettingsRepository GORM 商家设置仓储实现
type GormMerchantSettingsRepository struct {
	db *gorm.DB
}

// NewGormMerchantSettingsRepository 创建商家设置仓储
func NewGormMerchantSettingsRepository(db *gorm.DB) *GormMerchantSettingsRepository {
	return &GormMerchantSettingsRepository{db: db}
}

// Save 保存
func (r *GormMerchantSettingsRepository) Save(ctx context.Context, settings *domain.MerchantSettings) error {
	return r.db.WithContext(ctx).Create(settings).Error
}

// Update 更新
func (r *GormMerchantSettingsRepository) Update(ctx context.Context, settings *domain.MerchantSettings) error {
	return r.db.WithContext(ctx).Save(settings).Error
}

// FindByMerchantID 根据商家ID查找
func (r *GormMerchantSettingsRepository) FindByMerchantID(ctx context.Context, merchantID uint) (*domain.MerchantSettings, error) {
	var settings domain.MerchantSettings
	if err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&settings).Error; err != nil {
		return nil, err
	}
	return &settings, nil
}
