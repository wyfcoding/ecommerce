package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/crossborder/domain"
	"gorm.io/gorm"
)

type CrossBorderRepository struct {
	db *gorm.DB
}

func NewCrossBorderRepository(db *gorm.DB) domain.CrossBorderRepository {
	return &CrossBorderRepository{db: db}
}

func (r *CrossBorderRepository) SaveDeclaration(ctx context.Context, decl *domain.CustomsDeclaration) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(decl).Error; err != nil {
			return err
		}
		
		for i := range decl.Items {
			decl.Items[i].DeclarationID = decl.DeclarationID
			if err := tx.Create(&decl.Items[i]).Error; err != nil {
				return err
			}
		}
		
		return nil
	})
}

func (r *CrossBorderRepository) UpdateDeclaration(ctx context.Context, decl *domain.CustomsDeclaration) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(decl).Error; err != nil {
			return err
		}
		
		for i := range decl.Items {
			if decl.Items[i].ID == 0 {
				decl.Items[i].DeclarationID = decl.DeclarationID
				if err := tx.Create(&decl.Items[i]).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Save(&decl.Items[i]).Error; err != nil {
					return err
				}
			}
		}
		
		for i := range decl.Documents {
			if decl.Documents[i].ID == 0 {
				decl.Documents[i].DeclarationID = decl.DeclarationID
				if err := tx.Create(&decl.Documents[i]).Error; err != nil {
					return err
				}
			}
		}
		
		return nil
	})
}

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

func (r *CrossBorderRepository) GetDeclarationByOrder(ctx context.Context, orderID string) (*domain.CustomsDeclaration, error) {
	var decl domain.CustomsDeclaration
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Documents").
		Where("order_id = ?", orderID).
		First(&decl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &decl, nil
}

func (r *CrossBorderRepository) ListDeclarations(ctx context.Context, page, pageSize int, status domain.DeclarationStatus, userID uint64) ([]*domain.CustomsDeclaration, int64, error) {
	query := r.db.WithContext(ctx).Model(&domain.CustomsDeclaration{})
	
	if status != 0 {
		query = query.Where("status = ?", status)
	}
	if userID > 0 {
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

func (r *CrossBorderRepository) WithTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, "tx", tx)
		return fn(txCtx)
	})
}

type HSCodeRepository struct {
	db *gorm.DB
}

func NewHSCodeRepository(db *gorm.DB) domain.HSCodeRepository {
	return &HSCodeRepository{db: db}
}

func (r *HSCodeRepository) Save(ctx context.Context, hsCode *domain.HSCode) error {
	return r.db.WithContext(ctx).Create(hsCode).Error
}

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

func (r *HSCodeRepository) Search(ctx context.Context, keyword string, page, pageSize int) ([]*domain.HSCode, int64, error) {
	query := r.db.WithContext(ctx).Model(&domain.HSCode{}).Where("active = ?", true)
	
	if keyword != "" {
		query = query.Where("code LIKE ? OR description LIKE ? OR description_en LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	var hsCodes []*domain.HSCode
	offset := (page - 1) * pageSize
	if err := query.Order("code").
		Offset(offset).
		Limit(pageSize).
		Find(&hsCodes).Error; err != nil {
		return nil, 0, err
	}
	
	return hsCodes, total, nil
}

func (r *HSCodeRepository) Update(ctx context.Context, hsCode *domain.HSCode) error {
	return r.db.WithContext(ctx).Save(hsCode).Error
}

type CrossBorderOrderRepository struct {
	db *gorm.DB
}

func NewCrossBorderOrderRepository(db *gorm.DB) domain.CrossBorderOrderRepository {
	return &CrossBorderOrderRepository{db: db}
}

func (r *CrossBorderOrderRepository) Save(ctx context.Context, order *domain.CrossBorderOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

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

func (r *CrossBorderOrderRepository) Update(ctx context.Context, order *domain.CrossBorderOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

type CustomsDocumentRepository struct {
	db *gorm.DB
}

func NewCustomsDocumentRepository(db *gorm.DB) domain.CustomsDocumentRepository {
	return &CustomsDocumentRepository{db: db}
}

func (r *CustomsDocumentRepository) Save(ctx context.Context, doc *domain.CustomsDocument) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *CustomsDocumentRepository) GetByDeclarationID(ctx context.Context, declarationID string) ([]*domain.CustomsDocument, error) {
	var docs []*domain.CustomsDocument
	if err := r.db.WithContext(ctx).
		Where("declaration_id = ?", declarationID).
		Order("created_at DESC").
		Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *CustomsDocumentRepository) Delete(ctx context.Context, documentID string) error {
	return r.db.WithContext(ctx).Where("document_id = ?", documentID).Delete(&domain.CustomsDocument{}).Error
}

type ClearanceEventRepository struct {
	db *gorm.DB
}

func NewClearanceEventRepository(db *gorm.DB) domain.ClearanceEventRepository {
	return &ClearanceEventRepository{db: db}
}

func (r *ClearanceEventRepository) Save(ctx context.Context, event *domain.ClearanceEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

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
