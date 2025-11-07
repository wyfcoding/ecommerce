package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ecommerce/internal/aftersales/model"
)

// --- 接口定义 ---

// AftersalesRepository 定义了售后数据仓库的接口。
// 提供了对售后申请及其关联商品项的 CRUD 操作和查询功能。
type AftersalesRepository interface {
	// CreateApplication 在数据库中创建一个新的售后申请及其关联的商品项。
	CreateApplication(ctx context.Context, app *model.AftersalesApplication) error
	// GetApplicationByID 根据售后申请的唯一ID获取其详情，并预加载关联的商品项。
	GetApplicationByID(ctx context.Context, id uint) (*model.AftersalesApplication, error)
	// GetApplicationBySN 根据售后申请单号获取其详情，并预加载关联的商品项。
	GetApplicationBySN(ctx context.Context, sn string) (*model.AftersalesApplication, error)
	// ListApplications 列出售后申请列表，支持按用户ID和状态过滤，并支持分页。
	ListApplications(ctx context.Context, userID *uint, statusFilter *model.ApplicationStatus, page, pageSize int) ([]model.AftersalesApplication, int, error)
	// UpdateApplication 更新售后申请的现有信息，例如状态、备注等。
	UpdateApplication(ctx context.Context, app *model.AftersalesApplication) error
}

// --- 数据库模型 ---

// DBAftersalesApplication 对应数据库中的售后申请表。
type DBAftersalesApplication struct {
	gorm.Model
	ApplicationSN   string `gorm:"type:varchar(100);uniqueIndex;not null;comment:售后申请单号"`
	UserID          uint   `gorm:"not null;index;comment:用户ID"`
	OrderID         uint   `gorm:"not null;index;comment:订单ID"`
	OrderSN         string `gorm:"type:varchar(100);index;comment:订单号"`
	Type            string `gorm:"type:varchar(20);not null;comment:售后申请类型 (RETURN, EXCHANGE, REPAIR)"`
	Status          string `gorm:"type:varchar(30);not null;comment:售后申请状态"`
	Reason          string `gorm:"type:varchar(255);not null;comment:申请原因"`
	UserRemarks     string `gorm:"type:text;comment:用户备注"`
	AdminRemarks    string `gorm:"type:text;comment:管理员备注"`
	RefundAmount    float64 `gorm:"type:decimal(10,2);comment:最终退款金额"`

	Items []DBAftersalesItem `gorm:"foreignKey:ApplicationID"`
}

// TableName 自定义 DBAftersalesApplication 对应的表名。
func (DBAftersalesApplication) TableName() string {
	return "aftersales_applications"
}

// DBAftersalesItem 对应数据库中的售后商品项表。
type DBAftersalesItem struct {
	gorm.Model
	ApplicationID uint   `gorm:"not null;index;comment:所属售后申请ID"`
	OrderItemID   uint   `gorm:"not null;comment:原始订单项ID"`
	ProductID     uint   `gorm:"not null;comment:商品ID"`
	ProductSKU    string `gorm:"type:varchar(100);comment:商品SKU"`
	Quantity      int    `gorm:"not null;comment:售后数量"`
}

// TableName 自定义 DBAftersalesItem 对应的表名。
func (DBAftersalesItem) TableName() string {
	return "aftersales_items"
}

// --- 数据层核心 ---

// Data 封装了所有数据库操作的 GORM 客户端。
type Data struct {
	db *gorm.DB
}

// NewData 创建一个新的 Data 实例，并执行数据库迁移。
func NewData(db *gorm.DB) (*Data, func(), error) {
	d := &Data{
		db: db,
	}
	zap.S().Info("Running database migrations for aftersales service...")
	// 自动迁移所有相关的数据库表
	if err := db.AutoMigrate(
		&DBAftersalesApplication{},
		&DBAftersalesItem{},
	); err != nil {
		zap.S().Errorf("Failed to migrate aftersales database: %v", err)
		return nil, nil, fmt.Errorf("failed to migrate aftersales database: %w", err)
	}

	cleanup := func() {
		zap.S().Info("Closing aftersales data layer...")
		// 可以在这里添加数据库连接关闭逻辑，如果 GORM 提供了的话
	}

	return d, cleanup, nil
}

// --- AftersalesRepository 实现 ---

// aftersalesRepository 是 AftersalesRepository 接口的 GORM 实现。
type aftersalesRepository struct {
	*Data
}

// NewAftersalesRepository 创建一个新的 AftersalesRepository 实例。
func NewAftersalesRepository(data *Data) AftersalesRepository {
	return &aftersalesRepository{data}
}

// CreateApplication 在数据库中创建一个新的售后申请及其关联的商品项。
// 这是一个事务性操作，确保申请和商品项的一致性。
func (r *aftersalesRepository) CreateApplication(ctx context.Context, app *model.AftersalesApplication) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dbApp := fromBizAftersalesApplication(app)
		if err := tx.Create(dbApp).Error; err != nil {
			zap.S().Errorf("Failed to create aftersales application %s in db: %v", app.ApplicationSN, err)
			return fmt.Errorf("failed to create aftersales application: %w", err)
		}
		// GORM 的关联创建会自动处理 dbApp.Items，前提是它们已正确设置 ApplicationID
		// 确保从数据库返回的 ID 更新到业务模型中
		app.ID = dbApp.ID
		for i, item := range dbApp.Items {
			app.Items[i].ID = item.ID
		}
		zap.S().Infof("Aftersales application %s created in db with ID %d", app.ApplicationSN, app.ID)
		return nil
	})
}

// GetApplicationByID 根据售后申请的唯一ID获取其详情，并预加载关联的商品项。
// 如果未找到记录，则返回 nil 和 nil 错误。
func (r *aftersalesRepository) GetApplicationByID(ctx context.Context, id uint) (*model.AftersalesApplication, error) {
	var dbApp DBAftersalesApplication
	if err := r.db.WithContext(ctx).Preload("Items").First(&dbApp, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Record not found
		}
		zap.S().Errorf("Failed to get aftersales application by id %d from db: %v", id, err)
		return nil, fmt.Errorf("failed to get aftersales application: %w", err)
	}
	return toBizAftersalesApplication(&dbApp), nil
}

// GetApplicationBySN 根据售后申请单号获取其详情，并预加载关联的商品项。
// 如果未找到记录，则返回 nil 和 nil 错误。
func (r *aftersalesRepository) GetApplicationBySN(ctx context.Context, sn string) (*model.AftersalesApplication, error) {
	var dbApp DBAftersalesApplication
	if err := r.db.WithContext(ctx).Preload("Items").Where("application_sn = ?", sn).First(&dbApp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Record not found
		}
		zap.S().Errorf("Failed to get aftersales application by SN %s from db: %v", sn, err)
		return nil, fmt.Errorf("failed to get aftersales application: %w", err)
	}
	return toBizAftersalesApplication(&dbApp), nil
}

// ListApplications 列出售后申请列表，支持按用户ID和状态过滤，并支持分页。
func (r *aftersalesRepository) ListApplications(ctx context.Context, userID *uint, statusFilter *model.ApplicationStatus, page, pageSize int) ([]model.AftersalesApplication, int, error) {
	var dbApps []DBAftersalesApplication
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&DBAftersalesApplication{}).Preload("Items")

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if statusFilter != nil {
		query = query.Where("status = ?", string(*statusFilter))
	}

	// 获取总数
	if err := query.Count(&totalCount).Error; err != nil {
		zap.S().Errorf("Failed to count aftersales applications: %v", err)
		return nil, 0, fmt.Errorf("failed to count aftersales applications: %w", err)
	}

	// 分页查询
	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&dbApps).Error; err != nil {
		zap.S().Errorf("Failed to list aftersales applications from db: %v", err)
		return nil, 0, fmt.Errorf("failed to list aftersales applications: %w", err)
	}

	bizApps := make([]model.AftersalesApplication, len(dbApps))
	for i, dbApp := range dbApps {
		bizApps[i] = *toBizAftersalesApplication(&dbApp)
	}

	return bizApps, int(totalCount), nil
}

// UpdateApplication 更新售后申请的现有信息，例如状态、备注等。
// 它会更新主申请记录及其关联的商品项。
func (r *aftersalesRepository) UpdateApplication(ctx context.Context, app *model.AftersalesApplication) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dbApp := fromBizAftersalesApplication(app)
		if err := tx.Save(dbApp).Error; err != nil {
			zap.S().Errorf("Failed to update aftersales application %d in db: %v", app.ID, err)
			return fmt.Errorf("failed to update aftersales application: %w", err)
		}
		// 更新关联的 Items
		for _, item := range dbApp.Items {
			if err := tx.Save(&item).Error; err != nil {
				zap.S().Errorf("Failed to update aftersales item %d for application %d in db: %v", item.ID, app.ID, err)
				return fmt.Errorf("failed to update aftersales item: %w", err)
			}
		}
		zap.S().Infof("Aftersales application %d updated in db", app.ID)
		return nil
	})
}

// --- 模型转换辅助函数 ---

// toBizAftersalesApplication 将 DBAftersalesApplication 数据库模型转换为 model.AftersalesApplication 业务领域模型。
func toBizAftersalesApplication(dbApp *DBAftersalesApplication) *model.AftersalesApplication {
	if dbApp == nil {
		return nil
	}
	bizItems := make([]model.AftersalesItem, len(dbApp.Items))
	for i, dbItem := range dbApp.Items {
		bizItems[i] = *toBizAftersalesItem(&dbItem)
	}

	return &model.AftersalesApplication{
		ID:            dbApp.ID,
		ApplicationSN: dbApp.ApplicationSN,
		UserID:        dbApp.UserID,
		OrderID:       dbApp.OrderID,
		OrderSN:       dbApp.OrderSN,
		Type:          model.ApplicationType(dbApp.Type),
		Status:        model.ApplicationStatus(dbApp.Status),
		Reason:        dbApp.Reason,
		UserRemarks:   dbApp.UserRemarks,
		AdminRemarks:  dbApp.AdminRemarks,
		RefundAmount:  dbApp.RefundAmount,
		CreatedAt:     dbApp.CreatedAt,
		UpdatedAt:     dbApp.UpdatedAt,
		Items:         bizItems,
	}
}

// fromBizAftersalesApplication 将 model.AftersalesApplication 业务领域模型转换为 DBAftersalesApplication 数据库模型。
func fromBizAftersalesApplication(bizApp *model.AftersalesApplication) *DBAftersalesApplication {
	if bizApp == nil {
		return nil
	}
	dbItems := make([]DBAftersalesItem, len(bizApp.Items))
	for i, bizItem := range bizApp.Items {
		dbItems[i] = *fromBizAftersalesItem(&bizItem)
	}

	return &DBAftersalesApplication{
		Model:           gorm.Model{ID: bizApp.ID, CreatedAt: bizApp.CreatedAt, UpdatedAt: bizApp.UpdatedAt},
		ApplicationSN:   bizApp.ApplicationSN,
		UserID:          bizApp.UserID,
		OrderID:         bizApp.OrderID,
		OrderSN:         bizApp.OrderSN,
		Type:            string(bizApp.Type),
		Status:          string(bizApp.Status),
		Reason:          bizApp.Reason,
		UserRemarks:     bizApp.UserRemarks,
		AdminRemarks:    bizApp.AdminRemarks,
		RefundAmount:    bizApp.RefundAmount,
		Items:           dbItems,
	}
}

// toBizAftersalesItem 将 DBAftersalesItem 数据库模型转换为 model.AftersalesItem 业务领域模型。
func toBizAftersalesItem(dbItem *DBAftersalesItem) *model.AftersalesItem {
	if dbItem == nil {
		return nil
	}
	return &model.AftersalesItem{
		ID:            dbItem.ID,
		ApplicationID: dbItem.ApplicationID,
		OrderItemID:   dbItem.OrderItemID,
		ProductID:     dbItem.ProductID,
		ProductSKU:    dbItem.ProductSKU,
		Quantity:      dbItem.Quantity,
	}
}

// fromBizAftersalesItem 将 model.AftersalesItem 业务领域模型转换为 DBAftersalesItem 数据库模型。
func fromBizAftersalesItem(bizItem *model.AftersalesItem) *DBAftersalesItem {
	if bizItem == nil {
		return nil
	}
	return &DBAftersalesItem{
		Model:           gorm.Model{ID: bizItem.ID},
		ApplicationID: bizItem.ApplicationID,
		OrderItemID:   bizItem.OrderItemID,
		ProductID:     bizItem.ProductID,
		ProductSKU:    bizItem.ProductSKU,
		Quantity:      bizItem.Quantity,
	}
}