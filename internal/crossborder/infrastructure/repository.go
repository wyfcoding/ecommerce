// 变更说明：完善跨境电商仓储 MySQL 实现
package infrastructure

import (
	"context"
	"errors"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/crossborder/domain"
	"gorm.io/gorm"
)

// CrossBorderRepository 跨境电商仓储实现
type CrossBorderRepository struct {
	db *gorm.DB
}

// NewCrossBorderRepository 创建跨境电商仓储
func NewCrossBorderRepository(db *gorm.DB) domain.CrossBorderRepository {
	return &CrossBorderRepository{db: db}
}

// SaveDeclaration 保存报关单
func (r *CrossBorderRepository) SaveDeclaration(ctx context.Context, decl *domain.CustomsDeclaration) error {
	return r.db.WithContext(ctx).Create(decl).Error
}

// UpdateDeclaration 更新报关单
func (r *CrossBorderRepository) UpdateDeclaration(ctx context.Context, decl *domain.CustomsDeclaration) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(decl).Error; err != nil {
			return err
		}
		for _, item := range decl.Items {
			item.DeclarationID = decl.DeclarationID
			if err := tx.Save(&item).Error; err != nil {
				return err
			}
		}
		for _, doc := range decl.Documents {
			doc.DeclarationID = decl.DeclarationID
			if err := tx.Save(&doc).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetDeclaration 获取报关单
func (r *CrossBorderRepository) GetDeclaration(ctx context.Context, declarationID string) (*domain.CustomsDeclaration, error) {
	var decl domain.CustomsDeclaration
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Documents").
		Preload("ClearanceEvents").
		Where("declaration_id = ?", declarationID).
		First(&decl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &decl, nil
}

// GetDeclarationByOrder 根据订单ID获取报关单
func (r *CrossBorderRepository) GetDeclarationByOrder(ctx context.Context, orderID string) (*domain.CustomsDeclaration, error) {
	var decl domain.CustomsDeclaration
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Documents").
		Preload("ClearanceEvents").
		Where("order_id = ?", orderID).
		Order("created_at DESC").
		First(&decl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &decl, nil
}

// ListDeclarations 获取报关单列表
func (r *CrossBorderRepository) ListDeclarations(ctx context.Context, page, pageSize int, status domain.DeclarationStatus, userID uint64) ([]*domain.CustomsDeclaration, int64, error) {
	query := r.db.WithContext(ctx).Model(&domain.CustomsDeclaration{})

	if status != 0 {
		query = query.Where("status = ?", status)
	}
	if userID != 0 {
		query = query.Where("user_id = ?", userID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var decls []*domain.CustomsDeclaration
	offset := (page - 1) * pageSize
	if err := query.Preload("Items").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&decls).Error; err != nil {
		return nil, 0, err
	}

	return decls, total, nil
}

// WithTx 在事务中执行
func (r *CrossBorderRepository) WithTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, "tx", tx)
		return fn(txCtx)
	})
}

// HSCodeRepository HS编码仓储实现
type HSCodeRepository struct {
	db *gorm.DB
}

// NewHSCodeRepository 创建HS编码仓储
func NewHSCodeRepository(db *gorm.DB) domain.HSCodeRepository {
	return &HSCodeRepository{db: db}
}

// Save 保存HS编码
func (r *HSCodeRepository) Save(ctx context.Context, hsCode *domain.HSCode) error {
	return r.db.WithContext(ctx).Create(hsCode).Error
}

// Get 获取HS编码
func (r *HSCodeRepository) Get(ctx context.Context, code string) (*domain.HSCode, error) {
	var hsCode domain.HSCode
	if err := r.db.WithContext(ctx).Where("code = ? AND active = ?", code, true).First(&hsCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &hsCode, nil
}

// GetByCodes 批量获取HS编码
func (r *HSCodeRepository) GetByCodes(ctx context.Context, codes []string) (map[string]*domain.HSCode, error) {
	var hsCodes []*domain.HSCode
	if err := r.db.WithContext(ctx).
		Where("code IN ? AND active = ?", codes, true).
		Find(&hsCodes).Error; err != nil {
		return nil, err
	}

	result := make(map[string]*domain.HSCode)
	for _, hs := range hsCodes {
		result[hs.Code] = hs
	}
	return result, nil
}

// Search 搜索HS编码
func (r *HSCodeRepository) Search(ctx context.Context, keyword string, page, pageSize int) ([]*domain.HSCode, int64, error) {
	query := r.db.WithContext(ctx).Model(&domain.HSCode{}).Where("active = ?", true)

	if keyword != "" {
		keyword = "%" + keyword + "%"
		query = query.Where("code LIKE ? OR description LIKE ? OR description_en LIKE ?", keyword, keyword, keyword)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var hsCodes []*domain.HSCode
	offset := (page - 1) * pageSize
	if err := query.Order("code ASC").Offset(offset).Limit(pageSize).Find(&hsCodes).Error; err != nil {
		return nil, 0, err
	}

	return hsCodes, total, nil
}

// Update 更新HS编码
func (r *HSCodeRepository) Update(ctx context.Context, hsCode *domain.HSCode) error {
	return r.db.WithContext(ctx).Save(hsCode).Error
}

// CrossBorderOrderRepository 跨境订单仓储实现
type CrossBorderOrderRepository struct {
	db *gorm.DB
}

// NewCrossBorderOrderRepository 创建跨境订单仓储
func NewCrossBorderOrderRepository(db *gorm.DB) domain.CrossBorderOrderRepository {
	return &CrossBorderOrderRepository{db: db}
}

// Save 保存跨境订单
func (r *CrossBorderOrderRepository) Save(ctx context.Context, order *domain.CrossBorderOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// Get 获取跨境订单
func (r *CrossBorderOrderRepository) Get(ctx context.Context, crossBorderOrderID string) (*domain.CrossBorderOrder, error) {
	var order domain.CrossBorderOrder
	if err := r.db.WithContext(ctx).Where("cross_border_order_id = ?", crossBorderOrderID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetByOrderID 根据订单ID获取
func (r *CrossBorderOrderRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.CrossBorderOrder, error) {
	var order domain.CrossBorderOrder
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// Update 更新跨境订单
func (r *CrossBorderOrderRepository) Update(ctx context.Context, order *domain.CrossBorderOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// CustomsDocumentRepository 报关证件仓储实现
type CustomsDocumentRepository struct {
	db *gorm.DB
}

// NewCustomsDocumentRepository 创建报关证件仓储
func NewCustomsDocumentRepository(db *gorm.DB) domain.CustomsDocumentRepository {
	return &CustomsDocumentRepository{db: db}
}

// Save 保存证件
func (r *CustomsDocumentRepository) Save(ctx context.Context, doc *domain.CustomsDocument) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

// GetByDeclarationID 根据报关单ID获取证件列表
func (r *CustomsDocumentRepository) GetByDeclarationID(ctx context.Context, declarationID string) ([]*domain.CustomsDocument, error) {
	var docs []*domain.CustomsDocument
	if err := r.db.WithContext(ctx).
		Where("declaration_id = ?", declarationID).
		Order("created_at ASC").
		Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

// Delete 删除证件
func (r *CustomsDocumentRepository) Delete(ctx context.Context, documentID string) error {
	return r.db.WithContext(ctx).Where("document_id = ?", documentID).Delete(&domain.CustomsDocument{}).Error
}

// ClearanceEventRepository 清关事件仓储实现
type ClearanceEventRepository struct {
	db *gorm.DB
}

// NewClearanceEventRepository 创建清关事件仓储
func NewClearanceEventRepository(db *gorm.DB) domain.ClearanceEventRepository {
	return &ClearanceEventRepository{db: db}
}

// Save 保存清关事件
func (r *ClearanceEventRepository) Save(ctx context.Context, event *domain.ClearanceEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// GetByDeclarationID 根据报关单ID获取事件列表
func (r *ClearanceEventRepository) GetByDeclarationID(ctx context.Context, declarationID string) ([]*domain.ClearanceEvent, error) {
	var events []*domain.ClearanceEvent
	if err := r.db.WithContext(ctx).
		Where("declaration_id = ?", declarationID).
		Order("occurred_at ASC").
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// GetDB 获取数据库连接
func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok {
		return tx
	}
	return defaultDB
}

// SplitRestrictions 分割限制条件
func SplitRestrictions(restrictions string) []string {
	if restrictions == "" {
		return nil
	}
	return strings.Split(restrictions, ",")
}

// JoinRestrictions 合并限制条件
func JoinRestrictions(restrictions []string) string {
	if len(restrictions) == 0 {
		return ""
	}
	return strings.Join(restrictions, ",")
}
