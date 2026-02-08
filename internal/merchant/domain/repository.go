// Package domain 商家服务仓储接口定义
package domain

import "context"

// MerchantRepository 商家仓储接口
type MerchantRepository interface {
	// Save 保存商家
	Save(ctx context.Context, merchant *Merchant) error
	// Update 更新商家
	Update(ctx context.Context, merchant *Merchant) error
	// FindByID 根据ID查找商家
	FindByID(ctx context.Context, id uint) (*Merchant, error)
	// FindByMerchantNo 根据商家编号查找
	FindByMerchantNo(ctx context.Context, merchantNo string) (*Merchant, error)
	// FindByUserID 根据用户ID查找
	FindByUserID(ctx context.Context, userID uint64) (*Merchant, error)
	// List 列表查询
	List(ctx context.Context, filter *MerchantFilter) ([]*Merchant, int64, error)
	// Delete 删除商家
	Delete(ctx context.Context, id uint) error
}

// MerchantFilter 商家过滤条件
type MerchantFilter struct {
	// Status 状态筛选
	Status *MerchantStatus
	// Type 类型筛选
	Type *MerchantType
	// Level 等级筛选
	Level *MerchantLevel
	// Keyword 关键词搜索（商家名称/编号）
	Keyword string
	// Page 页码
	Page int
	// PageSize 每页数量
	PageSize int
}

// BusinessLicenseRepository 营业执照仓储接口
type BusinessLicenseRepository interface {
	// Save 保存营业执照
	Save(ctx context.Context, license *BusinessLicense) error
	// Update 更新营业执照
	Update(ctx context.Context, license *BusinessLicense) error
	// FindByMerchantID 根据商家ID查找
	FindByMerchantID(ctx context.Context, merchantID uint) (*BusinessLicense, error)
	// Delete 删除
	Delete(ctx context.Context, merchantID uint) error
}

// BankAccountRepository 银行账户仓储接口
type BankAccountRepository interface {
	// Save 保存银行账户
	Save(ctx context.Context, account *BankAccount) error
	// Update 更新银行账户
	Update(ctx context.Context, account *BankAccount) error
	// FindByMerchantID 根据商家ID查找
	FindByMerchantID(ctx context.Context, merchantID uint) (*BankAccount, error)
	// Delete 删除
	Delete(ctx context.Context, merchantID uint) error
}

// StoreRepository 店铺仓储接口
type StoreRepository interface {
	// Save 保存店铺
	Save(ctx context.Context, store *Store) error
	// Update 更新店铺
	Update(ctx context.Context, store *Store) error
	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uint) (*Store, error)
	// FindByStoreNo 根据店铺编号查找
	FindByStoreNo(ctx context.Context, storeNo string) (*Store, error)
	// ListByMerchantID 根据商家ID列出店铺
	ListByMerchantID(ctx context.Context, merchantID uint, page, pageSize int) ([]*Store, int64, error)
	// Delete 删除
	Delete(ctx context.Context, id uint) error
}

// MerchantSettingsRepository 商家设置仓储接口
type MerchantSettingsRepository interface {
	// Save 保存设置
	Save(ctx context.Context, settings *MerchantSettings) error
	// Update 更新设置
	Update(ctx context.Context, settings *MerchantSettings) error
	// FindByMerchantID 根据商家ID查找
	FindByMerchantID(ctx context.Context, merchantID uint) (*MerchantSettings, error)
}
