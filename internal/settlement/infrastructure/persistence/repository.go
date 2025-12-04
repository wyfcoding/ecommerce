package persistence

import (
	"context"
	"errors" // 导入标准错误处理库。

	"github.com/wyfcoding/ecommerce/internal/settlement/domain/entity"     // 导入结算领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/settlement/domain/repository" // 导入结算领域的仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type settlementRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewSettlementRepository 创建并返回一个新的 settlementRepository 实例。
func NewSettlementRepository(db *gorm.DB) repository.SettlementRepository {
	return &settlementRepository{db: db}
}

// --- 结算单管理 (Settlement methods) ---

// SaveSettlement 将结算单实体保存到数据库。
// 如果实体已存在，则更新；如果不存在，则创建。
func (r *settlementRepository) SaveSettlement(ctx context.Context, settlement *entity.Settlement) error {
	return r.db.WithContext(ctx).Save(settlement).Error
}

// GetSettlement 根据ID从数据库获取结算单记录。
// 如果记录未找到，则返回nil。
func (r *settlementRepository) GetSettlement(ctx context.Context, id uint64) (*entity.Settlement, error) {
	var settlement entity.Settlement
	if err := r.db.WithContext(ctx).First(&settlement, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &settlement, nil
}

// GetSettlementByNo 根据结算单号从数据库获取结算单记录。
// 如果记录未找到，则返回nil。
func (r *settlementRepository) GetSettlementByNo(ctx context.Context, no string) (*entity.Settlement, error) {
	var settlement entity.Settlement
	// 按结算单号过滤。
	if err := r.db.WithContext(ctx).Where("settlement_no = ?", no).First(&settlement).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &settlement, nil
}

// ListSettlements 从数据库列出指定商户ID的所有结算单记录，支持通过状态过滤和分页。
func (r *settlementRepository) ListSettlements(ctx context.Context, merchantID uint64, status *entity.SettlementStatus, offset, limit int) ([]*entity.Settlement, int64, error) {
	var list []*entity.Settlement
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Settlement{})
	if merchantID > 0 { // 如果提供了商户ID，则按商户ID过滤。
		db = db.Where("merchant_id = ?", merchantID)
	}
	if status != nil { // 如果提供了状态，则按状态过滤。
		db = db.Where("status = ?", *status)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 结算明细管理 (SettlementDetail methods) ---

// SaveSettlementDetail 将结算明细实体保存到数据库。
func (r *settlementRepository) SaveSettlementDetail(ctx context.Context, detail *entity.SettlementDetail) error {
	return r.db.WithContext(ctx).Save(detail).Error
}

// ListSettlementDetails 从数据库列出指定结算单ID的所有结算明细记录。
func (r *settlementRepository) ListSettlementDetails(ctx context.Context, settlementID uint64) ([]*entity.SettlementDetail, error) {
	var list []*entity.SettlementDetail
	// 按结算单ID过滤。
	if err := r.db.WithContext(ctx).Where("settlement_id = ?", settlementID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- 商户账户管理 (MerchantAccount methods) ---

// GetMerchantAccount 根据商户ID从数据库获取商户账户记录。
// 如果记录未找到，则返回nil。
func (r *settlementRepository) GetMerchantAccount(ctx context.Context, merchantID uint64) (*entity.MerchantAccount, error) {
	var account entity.MerchantAccount
	// 按商户ID过滤。
	if err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &account, nil
}

// SaveMerchantAccount 将商户账户实体保存到数据库。
func (r *settlementRepository) SaveMerchantAccount(ctx context.Context, account *entity.MerchantAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}
